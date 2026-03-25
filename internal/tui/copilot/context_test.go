package copilot

import (
	"strings"
	"testing"

	"github.com/maxbeizer/gh-planning/internal/github"
)

func TestBuildContext_BasicItem(t *testing.T) {
	item := &github.ProjectItem{
		Title:      "Fix login bug",
		Number:     42,
		State:      "OPEN",
		Status:     "In Progress",
		Repository: "acme/app",
		URL:        "https://github.com/acme/app/issues/42",
		Assignees:  []string{"alice", "bob"},
		Labels:     []string{"bug", "auth"},
	}

	got := BuildContext(item)

	for _, want := range []string{
		"## Issue #42: Fix login bug",
		"**State:** OPEN",
		"**Status:** In Progress",
		"**Repository:** acme/app",
		"@alice, @bob",
		"bug, auth",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildContext missing %q\n\nGot:\n%s", want, got)
		}
	}
}

func TestBuildContext_Dependencies(t *testing.T) {
	item := &github.ProjectItem{
		Title:  "Main task",
		Number: 10,
		State:  "OPEN",
		BlockedBy: []github.DependencyRef{
			{Number: 5, Title: "Migration", State: "OPEN"},
		},
		Blocks: []github.DependencyRef{
			{Number: 20, Title: "Deploy"},
		},
	}

	got := BuildContext(item)

	if !strings.Contains(got, "### Dependencies") {
		t.Errorf("missing Dependencies heading\n%s", got)
	}
	if !strings.Contains(got, "#5 (OPEN) — Migration") {
		t.Errorf("missing blocked-by detail\n%s", got)
	}
	if !strings.Contains(got, "#20") {
		t.Errorf("missing blocks detail\n%s", got)
	}
}

func TestBuildContext_SubIssues(t *testing.T) {
	item := &github.ProjectItem{
		Title:  "Epic",
		Number: 100,
		SubIssuesSummary: github.SubIssueSummary{
			Total:     5,
			Completed: 3,
		},
	}

	got := BuildContext(item)

	if !strings.Contains(got, "3/5 complete") {
		t.Errorf("missing sub-issues summary\n%s", got)
	}
}

func TestBuildContext_CustomFields(t *testing.T) {
	item := &github.ProjectItem{
		Title:  "Task",
		Number: 1,
		Fields: map[string]string{
			"Status":   "Done",
			"Priority": "High",
			"Sprint":   "Sprint 23",
		},
	}

	got := BuildContext(item)

	// Status should be skipped (already shown in metadata)
	if strings.Contains(got, "Status: Done") {
		t.Errorf("Status field should be excluded from custom fields\n%s", got)
	}
	if !strings.Contains(got, "Priority: High") {
		t.Errorf("missing Priority field\n%s", got)
	}
	if !strings.Contains(got, "Sprint: Sprint 23") {
		t.Errorf("missing Sprint field\n%s", got)
	}
}
