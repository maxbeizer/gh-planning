package board

import "strings"

// IssueTypeBadge returns an emoji badge for the given issue type.
func IssueTypeBadge(issueType string) string {
	switch strings.ToLower(issueType) {
	case "bug":
		return "🐛"
	case "feature":
		return "✨"
	case "task":
		return "📋"
	case "epic":
		return "🏔️"
	default:
		if issueType != "" {
			return "•"
		}
		return ""
	}
}
