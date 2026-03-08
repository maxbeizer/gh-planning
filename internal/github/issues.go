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
