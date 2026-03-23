package cmd

import "testing"

func TestKindPrefix(t *testing.T) {
	tests := []struct {
		kind string
		want string
	}{
		{"decision", "🎯"},
		{"blocker", "🚫"},
		{"hypothesis", "💡"},
		{"tried", "🔄"},
		{"result", "✅"},
		{"progress", "📝"},
		{"unknown", "📝"},
		{"", "📝"},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := kindPrefix(tt.kind)
			if got != tt.want {
				t.Errorf("kindPrefix(%q) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}
