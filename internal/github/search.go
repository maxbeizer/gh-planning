package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type SearchIssue struct {
	Title         string    `json:"title"`
	Number        int       `json:"number"`
	URL           string    `json:"url"`
	HTMLURL       string    `json:"html_url"`
	RepositoryURL string    `json:"repository_url"`
	State         string    `json:"state"`
	User          struct {
		Login string `json:"login"`
	} `json:"user"`
	Comments  int       `json:"comments"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ClosedAt  time.Time `json:"closed_at"`
	PullRequest *struct{} `json:"pull_request,omitempty"`
}

type searchResponse struct {
	Items []SearchIssue `json:"items"`
}

func SearchIssues(ctx context.Context, query string) ([]SearchIssue, error) {
	encoded := url.QueryEscape(query)
	endpoint := fmt.Sprintf("search/issues?q=%s&per_page=100", encoded)
	payload, err := runGH(ctx, "api", endpoint)
	if err != nil {
		return nil, err
	}
	var resp searchResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func RepositoryNameFromURL(repoURL string) string {
	if repoURL == "" {
		return ""
	}
	parts := strings.Split(strings.TrimSuffix(repoURL, "/"), "/")
	if len(parts) < 2 {
		return repoURL
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	return fmt.Sprintf("%s/%s", owner, repo)
}

func IssueURL(issue SearchIssue) string {
	if issue.HTMLURL != "" {
		return issue.HTMLURL
	}
	return issue.URL
}
