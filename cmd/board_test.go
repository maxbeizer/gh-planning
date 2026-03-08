package cmd

import (
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/maxbeizer/gh-planning/internal/github"
)

func TestPadOrTruncate_PlainASCII(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  int // expected visual width
	}{
		{"short padded", "hello", 10, 10},
		{"exact fit", "hello", 5, 5},
		{"empty string", "", 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padOrTruncate(tt.input, tt.width)
			got := visualWidth(result)
			if got != tt.want {
				t.Errorf("padOrTruncate(%q, %d) visual width = %d, want %d (result=%q)", tt.input, tt.width, got, tt.want, result)
			}
		})
	}
}

func TestPadOrTruncate_Emoji(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
	}{
		{"crab emoji padded", "🦀 hello", 15},
		{"clipboard emoji padded", "📋 Backlog", 15},
		{"check emoji padded", "✅ Done", 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padOrTruncate(tt.input, tt.width)
			got := visualWidth(result)
			if got != tt.width {
				t.Errorf("padOrTruncate(%q, %d) visual width = %d, want %d (result=%q)", tt.input, tt.width, got, tt.width, result)
			}
		})
	}
}

func TestPadOrTruncate_Truncation(t *testing.T) {
	long := "This is a very long string that should be truncated"
	result := padOrTruncate(long, 10)
	got := visualWidth(result)
	if got != 10 {
		t.Errorf("truncated visual width = %d, want 10 (result=%q)", got, result)
	}
}

func TestPadOrTruncate_TruncationWithEmoji(t *testing.T) {
	s := "🔵 In Progress with lots of text"
	result := padOrTruncate(s, 12)
	got := visualWidth(result)
	if got != 12 {
		t.Errorf("truncated emoji visual width = %d, want 12 (result=%q)", got, result)
	}
}

func TestSortedStatuses_Order(t *testing.T) {
	groups := map[string][]github.ProjectItem{
		"Done":        {},
		"In Progress": {},
		"Backlog":     {},
		"In Review":   {},
	}
	result := sortedStatuses(groups)
	expected := []string{"Backlog", "In Progress", "In Review", "Done"}
	if len(result) != len(expected) {
		t.Fatalf("sortedStatuses length = %d, want %d", len(result), len(expected))
	}
	for i, got := range result {
		if got != expected[i] {
			t.Errorf("sortedStatuses[%d] = %q, want %q", i, got, expected[i])
		}
	}
}

func TestSortedStatuses_UnknownStatusLast(t *testing.T) {
	groups := map[string][]github.ProjectItem{
		"Backlog": {},
		"Custom":  {},
		"Done":    {},
	}
	result := sortedStatuses(groups)
	// Custom should come between Backlog and Done since it has max rank
	// but alphabetically. Actually: Backlog(0), Done(7), Custom(len=11).
	// So: Backlog, Done, Custom
	if result[0] != "Backlog" {
		t.Errorf("first status = %q, want Backlog", result[0])
	}
	if result[len(result)-1] != "Custom" {
		t.Errorf("last status = %q, want Custom", result[len(result)-1])
	}
}

func TestStatusEmoji(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"In Progress", "🔵"},
		{"in progress", "🔵"},
		{"Backlog", "📋"},
		{"Ready", "📋"},
		{"Todo", "📋"},
		{"Done", "✅"},
		{"Closed", "✅"},
		{"Complete", "✅"},
		{"Completed", "✅"},
		{"In Review", "🔍"},
		{"Needs Review", "🔍"},
		{"Needs My Attention", "🔍"},
		{"Blocked", "🚫"},
		{"Something Else", "•"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := statusEmoji(tt.status)
			if got != tt.want {
				t.Errorf("statusEmoji(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func visualWidth(s string) int {
	return runewidth.StringWidth(s)
}
