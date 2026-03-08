package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
)

func shortSessionID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func formatTimestamp(t time.Time) string {
	stamp := t.UTC()
	loc, err := time.LoadLocation("America/Chicago")
	if err == nil {
		stamp = t.In(loc)
	}
	return fmt.Sprintf("%s CT", stamp.Format("Mon Jan 2, 2006 3:04 PM"))
}

func findStatusOption(options map[string]string, names ...string) (string, bool) {
	for _, name := range names {
		for optionName, optionID := range options {
			if strings.EqualFold(optionName, name) {
				return optionID, true
			}
		}
	}
	return "", false
}

func findProjectItemID(ctx context.Context, owner string, project int, repo string, number int) (string, error) {
	projectData, err := github.GetProject(ctx, owner, project)
	if err != nil {
		return "", err
	}
	for _, items := range projectData.Items {
		for _, item := range items {
			if item.Number == number && strings.EqualFold(item.Repository, repo) {
				return item.ID, nil
			}
		}
	}
	return "", fmt.Errorf("issue not found in project")
}
