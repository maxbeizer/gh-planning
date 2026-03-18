package github

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Activity struct {
	User           string        `json:"user"`
	Since          time.Time     `json:"since"`
	PRsMerged      []SearchIssue `json:"prsMerged"`
	PRsOpened      []SearchIssue `json:"prsOpened"`
	PRsOpen        []SearchIssue `json:"prsOpen"`
	PRsInReview    []SearchIssue `json:"prsInReview"`
	PRsDraft       []SearchIssue `json:"prsDraft"`
	IssuesOpened   []SearchIssue `json:"issuesOpened"`
	IssuesClosed   []SearchIssue `json:"issuesClosed"`
	Reviews        []SearchIssue `json:"reviews"`
	ReviewRequests []SearchIssue `json:"reviewRequests"`
	AssignedIssues []SearchIssue `json:"assignedIssues"`
	Blocked        []SearchIssue `json:"blocked"`
	LastActivity   time.Time     `json:"lastActivity"`
}

func FetchActivity(ctx context.Context, user string, since time.Time) (Activity, error) {
	queryDate := since.Format(time.RFC3339)
	activity := Activity{User: user, Since: since}

	// Each entry maps a search query to where its result is stored.
	type search struct {
		query string
		dest  *[]SearchIssue
	}
	searches := []search{
		{fmt.Sprintf("author:%s type:pr is:merged merged:>%s", user, queryDate), &activity.PRsMerged},
		{fmt.Sprintf("author:%s type:pr created:>%s", user, queryDate), &activity.PRsOpened},
		{fmt.Sprintf("author:%s type:issue created:>%s", user, queryDate), &activity.IssuesOpened},
		{fmt.Sprintf("author:%s type:issue is:closed closed:>%s", user, queryDate), &activity.IssuesClosed},
		{fmt.Sprintf("reviewed-by:%s type:pr updated:>%s", user, queryDate), &activity.Reviews},
		{fmt.Sprintf("author:%s type:pr is:open", user), &activity.PRsOpen},
		{fmt.Sprintf("author:%s type:pr is:open review:required", user), &activity.PRsInReview},
		{fmt.Sprintf("author:%s type:pr is:open draft:true", user), &activity.PRsDraft},
		{fmt.Sprintf("review-requested:%s type:pr is:open", user), &activity.ReviewRequests},
		{fmt.Sprintf("assignee:%s type:issue is:open sort:updated", user), &activity.AssignedIssues},
		{fmt.Sprintf("assignee:%s is:open label:blocked", user), &activity.Blocked},
	}

	var wg sync.WaitGroup
	errs := make([]error, len(searches))
	for i, s := range searches {
		wg.Add(1)
		go func(idx int, query string, dest *[]SearchIssue) {
			defer wg.Done()
			items, err := SearchIssues(ctx, query)
			if err != nil {
				errs[idx] = err
				return
			}
			*dest = items
		}(i, s.query, s.dest)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return activity, err
		}
	}

	activity.LastActivity = latestActivityTime(activity)
	return activity, nil
}

func latestActivityTime(activity Activity) time.Time {
	latest := time.Time{}
	all := [][]SearchIssue{
		activity.PRsMerged,
		activity.PRsOpened,
		activity.PRsOpen,
		activity.PRsInReview,
		activity.PRsDraft,
		activity.IssuesOpened,
		activity.IssuesClosed,
		activity.Reviews,
		activity.ReviewRequests,
		activity.AssignedIssues,
		activity.Blocked,
	}
	for _, items := range all {
		for _, item := range items {
			latest = maxTime(latest, item.UpdatedAt)
			latest = maxTime(latest, item.ClosedAt)
			latest = maxTime(latest, item.CreatedAt)
		}
	}
	return latest
}

func maxTime(a time.Time, b time.Time) time.Time {
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}
	if b.After(a) {
		return b
	}
	return a
}
