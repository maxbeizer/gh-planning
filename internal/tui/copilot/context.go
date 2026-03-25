package copilot

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/github"
)

// BuildContext serializes a ProjectItem into a structured markdown context
// string suitable for passing to a Copilot CLI session.
func BuildContext(item *github.ProjectItem) string {
	var b strings.Builder

	// Title line
	if item.Number > 0 {
		fmt.Fprintf(&b, "## Issue #%d: %s\n", item.Number, item.Title)
	} else {
		fmt.Fprintf(&b, "## %s\n", item.Title)
	}

	// Metadata line
	var meta []string
	if item.State != "" {
		meta = append(meta, fmt.Sprintf("**State:** %s", item.State))
	}
	if item.Status != "" {
		meta = append(meta, fmt.Sprintf("**Status:** %s", item.Status))
	}
	if item.ContentType != "" {
		meta = append(meta, fmt.Sprintf("**Type:** %s", item.ContentType))
	}
	if len(meta) > 0 {
		b.WriteString(strings.Join(meta, "  "))
		b.WriteString("\n")
	}

	if item.Repository != "" {
		fmt.Fprintf(&b, "**Repository:** %s\n", item.Repository)
	}
	if item.URL != "" {
		fmt.Fprintf(&b, "**URL:** %s\n", item.URL)
	}
	if len(item.Assignees) > 0 {
		tagged := make([]string, len(item.Assignees))
		for i, a := range item.Assignees {
			tagged[i] = "@" + a
		}
		fmt.Fprintf(&b, "**Assignees:** %s\n", strings.Join(tagged, ", "))
	}
	if len(item.Labels) > 0 {
		fmt.Fprintf(&b, "**Labels:** %s\n", strings.Join(item.Labels, ", "))
	}

	// Parent issue
	if item.ParentIssue != nil {
		fmt.Fprintf(&b, "**Parent:** #%d — %s\n", item.ParentIssue.Number, item.ParentIssue.Title)
	}

	// Sub-issues
	if item.SubIssuesSummary.Total > 0 {
		fmt.Fprintf(&b, "\n### Sub-Issues (%d/%d complete)\n",
			item.SubIssuesSummary.Completed, item.SubIssuesSummary.Total)
	}

	// Dependencies
	if len(item.BlockedBy) > 0 || len(item.Blocks) > 0 {
		b.WriteString("\n### Dependencies\n")
		if len(item.BlockedBy) > 0 {
			parts := make([]string, len(item.BlockedBy))
			for i, dep := range item.BlockedBy {
				parts[i] = fmt.Sprintf("#%d (%s) — %s", dep.Number, dep.State, dep.Title)
			}
			fmt.Fprintf(&b, "**Blocked by:** %s\n", strings.Join(parts, ", "))
		}
		if len(item.Blocks) > 0 {
			parts := make([]string, len(item.Blocks))
			for i, dep := range item.Blocks {
				parts[i] = fmt.Sprintf("#%d", dep.Number)
			}
			fmt.Fprintf(&b, "**Blocks:** %s\n", strings.Join(parts, ", "))
		}
	}

	// Custom fields
	if len(item.Fields) > 0 {
		b.WriteString("\n### Custom Fields\n")
		for k, v := range item.Fields {
			if strings.EqualFold(k, "Status") {
				continue // already shown above
			}
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}

	return b.String()
}
