package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Issue struct {
	ID         string `json:"id"`
	Number     int    `json:"number"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

func CurrentUser(ctx context.Context) (string, error) {
	payload, err := runGH(ctx, "api", "user", "--jq", ".login")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(payload)), nil
}

func CreateIssueComment(ctx context.Context, owner string, repo string, number int, body string) error {
	path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, number)
	_, err := API(ctx, "POST", path, map[string]string{"body": body})
	return err
}

func ParseIssueCreateOutput(payload []byte) (*Issue, error) {
	var issue Issue
	if err := json.Unmarshal(payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

const contentIDQuery = `query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issueOrPullRequest(number: $number) {
      ... on Issue { id }
      ... on PullRequest { id }
    }
  }
}`

type contentIDResponse struct {
	Data struct {
		Repository struct {
			IssueOrPullRequest struct {
				ID string `json:"id"`
			} `json:"issueOrPullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

// GetContentID returns the GraphQL node ID for an issue or pull request.
func GetContentID(ctx context.Context, owner string, repo string, number int) (string, error) {
	payload, err := GraphQL(ctx, contentIDQuery, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	})
	if err != nil {
		return "", err
	}
	var resp contentIDResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", err
	}
	id := resp.Data.Repository.IssueOrPullRequest.ID
	if id == "" {
		return "", fmt.Errorf("issue or PR #%d not found in %s/%s", number, owner, repo)
	}
	return id, nil
}
