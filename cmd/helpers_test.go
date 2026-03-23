package cmd

import (
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/maxbeizer/gh-planning/internal/github"
)

func TestResolveIssueInput_URL(t *testing.T) {
	repo, num, err := resolveIssueInput("https://github.com/maxbeizer/app/issues/42", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != "maxbeizer/app" {
		t.Errorf("repo = %q, want %q", repo, "maxbeizer/app")
	}
	if num != 42 {
		t.Errorf("number = %d, want 42", num)
	}
}

func TestResolveIssueInput_FullRef(t *testing.T) {
	repo, num, err := resolveIssueInput("github/github#123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != "github/github" {
		t.Errorf("repo = %q, want %q", repo, "github/github")
	}
	if num != 123 {
		t.Errorf("number = %d, want 123", num)
	}
}

func TestResolveIssueInput_BareNumberWithOverride(t *testing.T) {
	repo, num, err := resolveIssueInput("99", "myorg/myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != "myorg/myrepo" {
		t.Errorf("repo = %q, want %q", repo, "myorg/myrepo")
	}
	if num != 99 {
		t.Errorf("number = %d, want 99", num)
	}
}

func TestResolveIssueInput_InvalidInput(t *testing.T) {
	_, _, err := resolveIssueInput("not-a-number", "")
	if err == nil {
		t.Fatal("expected error for invalid input, got nil")
	}
}

func TestResolveIssueInput_URLWithPR(t *testing.T) {
	repo, num, err := resolveIssueInput("https://github.com/owner/repo/pull/55", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != "owner/repo" {
		t.Errorf("repo = %q, want %q", repo, "owner/repo")
	}
	if num != 55 {
		t.Errorf("number = %d, want 55", num)
	}
}

func TestParseIssueRef_FullRef(t *testing.T) {
	repo, num, err := parseIssueRef("owner/repo#42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo != "owner/repo" {
		t.Errorf("repo = %q, want %q", repo, "owner/repo")
	}
	if num != 42 {
		t.Errorf("number = %d, want 42", num)
	}
}

func TestParseIssueRef_EmptyRepo(t *testing.T) {
	_, _, err := parseIssueRef("#42")
	if err == nil {
		t.Fatal("expected error for empty repo, got nil")
	}
}

func TestParseIssueRef_InvalidNumber(t *testing.T) {
	_, _, err := parseIssueRef("owner/repo#abc")
	if err == nil {
		t.Fatal("expected error for invalid number, got nil")
	}
}

func TestParseIssueURL_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantRepo string
		wantNum  int
	}{
		{"https://github.com/owner/repo/issues/1", "owner/repo", 1},
		{"https://github.com/org/project/pull/99", "org/project", 99},
		{"http://github.com/a/b/issues/5", "a/b", 5},
	}
	for _, tt := range tests {
		repo, num, err := parseIssueURL(tt.input)
		if err != nil {
			t.Errorf("parseIssueURL(%q) error: %v", tt.input, err)
			continue
		}
		if repo != tt.wantRepo {
			t.Errorf("parseIssueURL(%q) repo = %q, want %q", tt.input, repo, tt.wantRepo)
		}
		if num != tt.wantNum {
			t.Errorf("parseIssueURL(%q) number = %d, want %d", tt.input, num, tt.wantNum)
		}
	}
}

func TestParseIssueURL_Invalid(t *testing.T) {
	_, _, err := parseIssueURL("https://github.com/too-short")
	if err == nil {
		t.Fatal("expected error for short URL, got nil")
	}
}

func TestParseTrackOutput_WithURL(t *testing.T) {
	output := `Created issue #42
https://github.com/maxbeizer/app/issues/42
Added to project`
	ref, num := parseTrackOutput(output)
	if ref != "maxbeizer/app#42" {
		t.Errorf("ref = %q, want %q", ref, "maxbeizer/app#42")
	}
	if num != 42 {
		t.Errorf("number = %d, want 42", num)
	}
}

func TestParseTrackOutput_NoURL(t *testing.T) {
	ref, num := parseTrackOutput("some random output with no URL")
	if ref != "" {
		t.Errorf("ref = %q, want empty", ref)
	}
	if num != 0 {
		t.Errorf("number = %d, want 0", num)
	}
}

func TestParseTrackOutput_URLOnMultipleLines(t *testing.T) {
	output := `✅ Created issue
✅ Added to project
https://github.com/org/repo/issues/100`
	ref, num := parseTrackOutput(output)
	if ref != "org/repo#100" {
		t.Errorf("ref = %q, want %q", ref, "org/repo#100")
	}
	if num != 100 {
		t.Errorf("number = %d, want 100", num)
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{",,,", []string{}},
	}
	for _, tt := range tests {
		got := splitAndTrim(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitAndTrim(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		err   bool
	}{
		{"", 0, false},
		{"1h", time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"2d", 48 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"1s", time.Second, false},
		{"abc", 0, true},
		{"xd", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.err && err == nil {
				t.Fatalf("parseDuration(%q) expected error, got nil", tt.input)
			}
			if !tt.err && err != nil {
				t.Fatalf("parseDuration(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestHumanizeDuration(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{5 * time.Minute, "5m ago"},
		{30 * time.Minute, "30m ago"},
		{2 * time.Hour, "2h ago"},
		{23 * time.Hour, "23h ago"},
		{48 * time.Hour, "2d ago"},
		{6 * 24 * time.Hour, "6d ago"},
		{14 * 24 * time.Hour, "2w ago"},
		{-5 * time.Minute, "5m ago"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := humanizeDuration(tt.input)
			if got != tt.want {
				t.Errorf("humanizeDuration(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProjectURL(t *testing.T) {
	got := projectURL("maxbeizer", 42)
	want := "https://github.com/users/maxbeizer/projects/42"
	if got != want {
		t.Errorf("projectURL() = %q, want %q", got, want)
	}
}

func TestFilterProjectItems(t *testing.T) {
	now := time.Now()
	project := &github.Project{
		Items: map[string][]github.ProjectItem{
			"In Progress": {
				{Number: 1, Title: "Task 1", Assignees: []string{"alice"}, UpdatedAt: now},
				{Number: 2, Title: "Task 2", Assignees: []string{"bob"}, UpdatedAt: now.Add(-48 * time.Hour)},
			},
			"Done": {
				{Number: 3, Title: "Task 3", Assignees: []string{"alice"}, UpdatedAt: now},
			},
		},
	}

	t.Run("no filters", func(t *testing.T) {
		result := filterProjectItems(project, "", 0, nil)
		total := 0
		for _, items := range result {
			total += len(items)
		}
		if total != 3 {
			t.Errorf("expected 3 items, got %d", total)
		}
	})

	t.Run("filter by assignee", func(t *testing.T) {
		result := filterProjectItems(project, "alice", 0, nil)
		total := 0
		for _, items := range result {
			total += len(items)
		}
		if total != 2 {
			t.Errorf("expected 2 items for alice, got %d", total)
		}
	})

	t.Run("exclude status", func(t *testing.T) {
		result := filterProjectItems(project, "", 0, []string{"Done"})
		if _, ok := result["Done"]; ok {
			t.Error("expected Done to be excluded")
		}
		if len(result["In Progress"]) != 2 {
			t.Errorf("expected 2 In Progress items, got %d", len(result["In Progress"]))
		}
	})

	t.Run("stale filter", func(t *testing.T) {
		result := filterProjectItems(project, "", 24*time.Hour, nil)
		total := 0
		for _, items := range result {
			total += len(items)
		}
		if total != 1 {
			t.Errorf("expected 1 stale item, got %d", total)
		}
	})
}

func TestDecorateStatus(t *testing.T) {
	tests := []struct {
		status string
		prefix string
	}{
		{"In Progress", "🔵"},
		{"Backlog", "📋"},
		{"Done", "✅"},
		{"Closed", "✅"},
		{"In Review", "🔍"},
		{"Needs Review", "🔍"},
		{"Something Else", "•"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := decorateStatus(tt.status, tt.status)
			if got[:len(tt.prefix)] != tt.prefix {
				t.Errorf("decorateStatus(%q) = %q, want prefix %q", tt.status, got, tt.prefix)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		fits  bool
	}{
		{"short string", "hello", 10, true},
		{"exact fit", "hello", 5, true},
		{"needs truncation", "hello world this is long", 10, false},
		{"emoji string", "🔵 In Progress", 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.max)
			width := runewidth.StringWidth(result)
			if width > tt.max {
				t.Errorf("truncate(%q, %d) visual width = %d, exceeds max", tt.input, tt.max, width)
			}
			if tt.fits && result != tt.input {
				t.Errorf("truncate(%q, %d) = %q, expected unchanged", tt.input, tt.max, result)
			}
		})
	}
}

func TestFindStatusOption(t *testing.T) {
	options := map[string]string{
		"In Progress": "opt-1",
		"Done":        "opt-2",
		"Backlog":     "opt-3",
	}

	t.Run("exact match", func(t *testing.T) {
		id, ok := findStatusOption(options, "Done")
		if !ok || id != "opt-2" {
			t.Errorf("expected opt-2, got %q (ok=%v)", id, ok)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		id, ok := findStatusOption(options, "done")
		if !ok || id != "opt-2" {
			t.Errorf("expected opt-2 (case insensitive), got %q (ok=%v)", id, ok)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := findStatusOption(options, "Nonexistent")
		if ok {
			t.Error("expected not found")
		}
	})

	t.Run("multiple names fallback", func(t *testing.T) {
		id, ok := findStatusOption(options, "Nonexistent", "Backlog")
		if !ok || id != "opt-3" {
			t.Errorf("expected opt-3 (fallback), got %q (ok=%v)", id, ok)
		}
	})
}
