package github

import (
	"testing"
	"time"
)

func TestRepositoryNameFromURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://api.github.com/repos/octocat/hello-world", "octocat/hello-world"},
		{"https://api.github.com/repos/org/repo/", "org/repo"},
		{"", ""},
		{"https://api.github.com/repos/single", "repos/single"},
		{"no-slashes", "no-slashes"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RepositoryNameFromURL(tt.input)
			if got != tt.want {
				t.Errorf("RepositoryNameFromURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIssueURL(t *testing.T) {
	tests := []struct {
		name  string
		issue SearchIssue
		want  string
	}{
		{
			name:  "prefers HTMLURL",
			issue: SearchIssue{HTMLURL: "https://github.com/o/r/issues/1", URL: "https://api.github.com/repos/o/r/issues/1"},
			want:  "https://github.com/o/r/issues/1",
		},
		{
			name:  "falls back to URL",
			issue: SearchIssue{URL: "https://api.github.com/repos/o/r/issues/2"},
			want:  "https://api.github.com/repos/o/r/issues/2",
		},
		{
			name:  "both empty",
			issue: SearchIssue{},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IssueURL(tt.issue)
			if got != tt.want {
				t.Errorf("IssueURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMaxTime(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	zero := time.Time{}

	tests := []struct {
		name string
		a, b time.Time
		want time.Time
	}{
		{"b is later", t1, t2, t2},
		{"a is later", t2, t1, t2},
		{"equal", t1, t1, t1},
		{"a is zero", zero, t2, t2},
		{"b is zero", t1, zero, t1},
		{"both zero", zero, zero, zero},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maxTime(tt.a, tt.b)
			if !got.Equal(tt.want) {
				t.Errorf("maxTime(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
