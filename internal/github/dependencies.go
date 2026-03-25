package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// dependencyQuery fetches tracked-by and tracking relationships for an issue.
const dependencyQuery = `query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      trackedByIssues(first: 10) {
        nodes {
          number
          title
          url
          state
        }
      }
      trackedInIssues(first: 10) {
        nodes {
          number
          title
          url
          state
        }
      }
    }
  }
}`

type dependencyResponse struct {
	Data struct {
		Repository struct {
			Issue *struct {
				TrackedByIssues struct {
					Nodes []DependencyRef `json:"nodes"`
				} `json:"trackedByIssues"`
				TrackedInIssues struct {
					Nodes []DependencyRef `json:"nodes"`
				} `json:"trackedInIssues"`
			} `json:"issue"`
		} `json:"repository"`
	} `json:"data"`
}

// GetDependencies fetches tracked-by and tracking relationships for an issue.
// Returns (blockedBy, blocks, error). Best-effort: returns nil slices on failure.
func GetDependencies(ctx context.Context, owner, repo string, number int) ([]DependencyRef, []DependencyRef, error) {
	payload, err := GraphQL(ctx, dependencyQuery, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	})
	if err != nil {
		return nil, nil, err
	}
	var resp dependencyResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, nil, err
	}
	if resp.Data.Repository.Issue == nil {
		return nil, nil, nil
	}
	return resp.Data.Repository.Issue.TrackedByIssues.Nodes,
		resp.Data.Repository.Issue.TrackedInIssues.Nodes,
		nil
}

// EnrichWithDependencies populates BlockedBy and Blocks on every Issue item in
// the project. It fetches dependencies concurrently and is best-effort: errors
// on individual items are silently ignored so the rest of the project still
// renders correctly.
func EnrichWithDependencies(ctx context.Context, project *Project) error {
	if project == nil {
		return nil
	}

	type work struct {
		status string
		index  int
	}

	var items []work
	for status, list := range project.Items {
		for i, item := range list {
			if !strings.EqualFold(item.ContentType, "Issue") {
				continue
			}
			if item.Repository == "" || item.Number == 0 {
				continue
			}
			items = append(items, work{status: status, index: i})
		}
	}

	if len(items) == 0 {
		return nil
	}

	// Limit concurrency to avoid overwhelming the gh CLI.
	const maxConcurrency = 5
	sem := make(chan struct{}, maxConcurrency)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, w := range items {
		wg.Add(1)
		go func(w work) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			item := project.Items[w.status][w.index]
			owner, repo := splitOwnerRepo(item.Repository)
			if owner == "" || repo == "" {
				return
			}

			blockedBy, blocks, err := GetDependencies(ctx, owner, repo, item.Number)
			if err != nil {
				// Best-effort: skip items that fail.
				return
			}

			mu.Lock()
			project.Items[w.status][w.index].BlockedBy = blockedBy
			project.Items[w.status][w.index].Blocks = blocks
			mu.Unlock()
		}(w)
	}

	wg.Wait()
	return nil
}

// splitOwnerRepo splits "owner/repo" into its two components.
func splitOwnerRepo(nwo string) (string, string) {
	parts := strings.SplitN(nwo, "/", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// FormatBlockedBy returns a human-readable summary of blocking issues.
func FormatBlockedBy(deps []DependencyRef) string {
	if len(deps) == 0 {
		return ""
	}
	var parts []string
	for _, d := range deps {
		state := strings.ToLower(d.State)
		parts = append(parts, fmt.Sprintf("#%d %s (%s)", d.Number, d.Title, state))
	}
	return strings.Join(parts, ", ")
}
