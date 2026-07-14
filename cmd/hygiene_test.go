package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
)

func TestBuildHygieneFindings(t *testing.T) {
	now := time.Now()
	project := &github.Project{Items: map[string][]github.ProjectItem{
		"In Progress": {
			{Number: 1, Title: "Stale", Status: "In Progress", State: "OPEN", UpdatedAt: now.Add(-10 * 24 * time.Hour), Fields: map[string]string{"Workstream": "Core"}},
			{Number: 2, Title: "Missing fields", Status: "In Progress", State: "OPEN", UpdatedAt: now, Assignees: []string{"alice"}},
			{Number: 3, Title: "Closed active", Status: "In Progress", State: "CLOSED", UpdatedAt: now, Assignees: []string{"alice"}, Fields: map[string]string{"Target Date": "2026-01-01", "Workstream": "Core"}},
			{Number: 4, Title: "Blocked", Status: "Blocked", State: "OPEN", UpdatedAt: now, Assignees: []string{"alice"}, Fields: map[string]string{"Target Date": "2026-01-01", "Workstream": "Core"}},
		},
		"Done": {
			{Number: 5, Title: "Open done", Status: "Done", State: "OPEN", UpdatedAt: now, Assignees: []string{"alice"}},
		},
	}}

	findings := buildHygieneFindings(project, 7*24*time.Hour, []string{"Target Date", "Workstream"}, []string{"Owner", "DRI"})
	kinds := map[string]int{}
	for _, finding := range findings {
		kinds[finding.Kind]++
	}
	for _, kind := range []string{"stale-active", "unowned-active", "missing-field", "closed-not-done", "blocked-without-link", "done-open"} {
		if kinds[kind] == 0 {
			t.Errorf("expected finding kind %q", kind)
		}
	}
}

func TestHygieneMarkdown(t *testing.T) {
	report := hygieneReport{
		Title:   "Roadmap",
		Project: 25,
		Findings: []hygieneFinding{{
			Kind:        "stale-active",
			Description: "Active item has not been updated recently",
			Item:        github.ProjectItem{Number: 1, Title: "Task", URL: "https://github.com/o/r/issues/1"},
			Detail:      "8d ago",
		}},
	}
	var b strings.Builder
	printHygieneMarkdown(&b, report)
	got := b.String()
	if !strings.Contains(got, "# Hygiene: Roadmap (#25)") || !strings.Contains(got, "[#1](https://github.com/o/r/issues/1)") {
		t.Fatalf("unexpected markdown:\n%s", got)
	}
}
