package github

import (
	"context"
	"fmt"
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

	var err error
	activity.PRsMerged, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:merged merged:>%s", user, queryDate))
	if err != nil {
		return activity, err
	}
	activity.PRsOpened, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:pr created:>%s", user, queryDate))
	if err != nil {
		return activity, err
	}
	activity.IssuesOpened, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:issue created:>%s", user, queryDate))
	if err != nil {
		return activity, err
	}
	activity.IssuesClosed, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:issue is:closed closed:>%s", user, queryDate))
	if err != nil {
		return activity, err
	}
	activity.Reviews, err = SearchIssues(ctx, fmt.Sprintf("reviewed-by:%s type:pr updated:>%s", user, queryDate))
	if err != nil {
		return activity, err
	}
	activity.PRsOpen, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:open", user))
	if err != nil {
		return activity, err
	}
	activity.PRsInReview, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:open review:required", user))
	if err != nil {
		return activity, err
	}
	activity.PRsDraft, err = SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:open draft:true", user))
	if err != nil {
		return activity, err
	}
	activity.ReviewRequests, err = SearchIssues(ctx, fmt.Sprintf("review-requested:%s type:pr is:open", user))
	if err != nil {
		return activity, err
	}
	activity.AssignedIssues, err = SearchIssues(ctx, fmt.Sprintf("assignee:%s type:issue is:open sort:updated", user))
	if err != nil {
		return activity, err
	}
	activity.Blocked, err = SearchIssues(ctx, fmt.Sprintf("assignee:%s is:open label:blocked", user))
	if err != nil {
		return activity, err
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
