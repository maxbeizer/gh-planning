package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// SubIssueItem represents a single sub-issue returned by the REST API.
type SubIssueItem struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	URL    string `json:"url"`
}

// subIssueResponse matches the JSON shape returned by the sub_issues REST endpoint.
type subIssueResponse struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
	URL     string `json:"url"`
}

// GetSubIssues fetches sub-issues for the given repo issue via the REST API.
// repo should be "owner/repo" format.
func GetSubIssues(ctx context.Context, repo string, number int) ([]SubIssueItem, error) {
	path := fmt.Sprintf("repos/%s/issues/%d/sub_issues", repo, number)
	data, err := API(ctx, "", path, nil)
	if err != nil {
		return nil, err
	}
	var raw []subIssueResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing sub-issues response: %w", err)
	}
	items := make([]SubIssueItem, 0, len(raw))
	for _, r := range raw {
		u := r.HTMLURL
		if u == "" {
			u = r.URL
		}
		items = append(items, SubIssueItem{
			Number: r.Number,
			Title:  r.Title,
			State:  r.State,
			URL:    u,
		})
	}
	return items, nil
}

// EnrichWithSubIssues populates SubIssuesSummary and ParentIssue on every
// ProjectItem in the project. It calls the REST sub_issues endpoint for each
// Issue item (PRs are skipped). Errors on individual items are silently
// ignored so that a single API failure does not break the whole board.
func EnrichWithSubIssues(ctx context.Context, project *Project) error {
	if project == nil {
		return nil
	}

	// Collect pointers to all issue items so we can mutate them in place.
	type itemRef struct {
		status string
		index  int
	}
	var refs []itemRef
	for status, items := range project.Items {
		for i := range items {
			if strings.EqualFold(items[i].ContentType, "Issue") {
				refs = append(refs, itemRef{status: status, index: i})
			}
		}
	}
	if len(refs) == 0 {
		return nil
	}

	// Fetch sub-issues concurrently with bounded parallelism.
	const maxWorkers = 5
	type result struct {
		ref   itemRef
		subs  []SubIssueItem
	}

	sem := make(chan struct{}, maxWorkers)
	var mu sync.Mutex
	var results []result

	var wg sync.WaitGroup
	for _, ref := range refs {
		wg.Add(1)
		go func(r itemRef) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			item := &project.Items[r.status][r.index]
			subs, err := GetSubIssues(ctx, item.Repository, item.Number)
			if err != nil || len(subs) == 0 {
				return
			}
			mu.Lock()
			results = append(results, result{ref: r, subs: subs})
			mu.Unlock()
		}(ref)
	}
	wg.Wait()

	// Build a lookup of issue number→(status, index) for parent resolution.
	type issueKey struct {
		repo   string
		number int
	}
	issueIndex := make(map[issueKey]itemRef, len(refs))
	for _, r := range refs {
		item := &project.Items[r.status][r.index]
		issueIndex[issueKey{repo: item.Repository, number: item.Number}] = r
	}

	// Apply sub-issue summaries and parent back-links.
	for _, res := range results {
		parent := &project.Items[res.ref.status][res.ref.index]
		completed := 0
		for _, sub := range res.subs {
			if strings.EqualFold(sub.State, "closed") {
				completed++
			}
			// If the sub-issue is also in the project, mark its parent.
			if childRef, ok := issueIndex[issueKey{repo: parent.Repository, number: sub.Number}]; ok {
				child := &project.Items[childRef.status][childRef.index]
				child.ParentIssue = &ParentRef{
					Number: parent.Number,
					Title:  parent.Title,
					URL:    parent.URL,
				}
			}
		}
		parent.SubIssuesSummary = SubIssueSummary{
			Total:     len(res.subs),
			Completed: completed,
		}
	}

	return nil
}
