package cmd

import "testing"

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
