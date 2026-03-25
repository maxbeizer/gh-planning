package config

import "testing"

func TestParseSectionFilter(t *testing.T) {
	tests := []struct {
		input     string
		wantField string
		wantValue string
	}{
		{"status:In Progress", "status", "In Progress"},
		{"status:Backlog", "status", "Backlog"},
		{"assignee:octocat", "assignee", "octocat"},
		{"plain text", "", "plain text"},
		{"", "", ""},
		{"status: Done ", "status", "Done"},
		{" field : value ", "field", "value"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			f, v := ParseSectionFilter(tt.input)
			if f != tt.wantField || v != tt.wantValue {
				t.Errorf("ParseSectionFilter(%q) = (%q, %q), want (%q, %q)",
					tt.input, f, v, tt.wantField, tt.wantValue)
			}
		})
	}
}
