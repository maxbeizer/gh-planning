package github

import (
	"testing"
)

func TestIsBlocked(t *testing.T) {
	tests := []struct {
		name    string
		item    ProjectItem
		blocked bool
	}{
		{
			name:    "no blockers",
			item:    ProjectItem{},
			blocked: false,
		},
		{
			name: "all blockers closed",
			item: ProjectItem{
				BlockedBy: []DependencyRef{
					{Number: 1, State: "CLOSED"},
					{Number: 2, State: "closed"},
				},
			},
			blocked: false,
		},
		{
			name: "one open blocker",
			item: ProjectItem{
				BlockedBy: []DependencyRef{
					{Number: 1, State: "CLOSED"},
					{Number: 2, State: "OPEN"},
				},
			},
			blocked: true,
		},
		{
			name: "case insensitive open",
			item: ProjectItem{
				BlockedBy: []DependencyRef{
					{Number: 5, State: "open"},
				},
			},
			blocked: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.IsBlocked(); got != tt.blocked {
				t.Errorf("IsBlocked() = %v, want %v", got, tt.blocked)
			}
		})
	}
}

func TestSplitOwnerRepo(t *testing.T) {
	tests := []struct {
		input     string
		wantOwner string
		wantRepo  string
	}{
		{"octocat/hello-world", "octocat", "hello-world"},
		{"org/repo", "org", "repo"},
		{"noslash", "", ""},
		{"", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			owner, repo := splitOwnerRepo(tt.input)
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("splitOwnerRepo(%q) = (%q, %q), want (%q, %q)", tt.input, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestFormatBlockedBy(t *testing.T) {
	tests := []struct {
		name string
		deps []DependencyRef
		want string
	}{
		{"nil", nil, ""},
		{"empty", []DependencyRef{}, ""},
		{
			"single",
			[]DependencyRef{{Number: 42, Title: "Fix bug", State: "OPEN"}},
			"#42 Fix bug (open)",
		},
		{
			"multiple",
			[]DependencyRef{
				{Number: 1, Title: "First", State: "OPEN"},
				{Number: 2, Title: "Second", State: "CLOSED"},
			},
			"#1 First (open), #2 Second (closed)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBlockedBy(tt.deps)
			if got != tt.want {
				t.Errorf("FormatBlockedBy() = %q, want %q", got, tt.want)
			}
		})
	}
}
