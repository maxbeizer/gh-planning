package tree

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/maxbeizer/gh-planning/internal/tui/board"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
)

// View renders the tree view within the available content area.
func (m Model) View() string {
	if !m.ready || len(m.flatNodes) == 0 {
		return m.ctx.Theme.Muted.Render("No items to display")
	}

	visHeight := m.ctx.ContentHeight()
	var b strings.Builder

	// Scroll-up indicator.
	if m.offset > 0 {
		b.WriteString(m.ctx.Theme.Dimmed.Render("  ↑ more"))
		b.WriteString("\n")
		visHeight--
	}

	// Reserve a line for the down indicator if needed.
	hasMore := m.offset+visHeight < len(m.flatNodes)
	if hasMore {
		visHeight--
	}

	end := m.offset + visHeight
	if end > len(m.flatNodes) {
		end = len(m.flatNodes)
	}

	for i := m.offset; i < end; i++ {
		node := m.flatNodes[i]
		active := i == m.cursor
		b.WriteString(renderNode(m.ctx, node, active))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	if hasMore {
		b.WriteString("\n")
		b.WriteString(m.ctx.Theme.Dimmed.Render("  ↓ more"))
	}

	return b.String()
}

// renderNode renders a single tree node as one line.
func renderNode(ctx *context.ProgramContext, node *Node, active bool) string {
	th := ctx.Theme
	width := ctx.ContentWidth()

	// Build the line content.
	var parts []string

	// Indent: 2 spaces per depth level.
	indent := strings.Repeat("  ", node.Depth)
	parts = append(parts, indent)

	// Expand indicator.
	if node.hasChildren() {
		if node.Expanded {
			parts = append(parts, "▼ ")
		} else {
			parts = append(parts, "▶ ")
		}
	} else {
		parts = append(parts, "· ")
	}

	// Issue number.
	number := fmt.Sprintf("#%-5d", node.Item.Number)
	parts = append(parts, number+" ")

	// Issue type badge.
	badge := board.IssueTypeBadge(node.Item.Fields["Issue Type"])
	if badge != "" {
		parts = append(parts, badge+" ")
	}

	// State indicator.
	state := stateIndicator(node.Item.State)

	// Progress bar for parents with sub-issues.
	progress := ""
	summary := node.Item.SubIssuesSummary
	if summary.Total > 0 {
		progress = fmt.Sprintf(" [%d/%d] %s", summary.Completed, summary.Total, progressBar(summary.Completed, summary.Total, 5))
	}

	// Calculate remaining width for title.
	suffix := " " + state + progress
	usedWidth := 0
	for _, p := range parts {
		usedWidth += runewidth.StringWidth(p)
	}
	suffixWidth := runewidth.StringWidth(suffix)
	titleWidth := width - usedWidth - suffixWidth
	if titleWidth < 10 {
		titleWidth = 10
	}

	// Truncate title.
	title := truncate(node.Item.Title, titleWidth)
	parts = append(parts, title)
	parts = append(parts, suffix)

	line := strings.Join(parts, "")

	// Pad line to full width for consistent highlighting.
	lineWidth := runewidth.StringWidth(line)
	if lineWidth < width {
		line += strings.Repeat(" ", width-lineWidth)
	}

	if active {
		return th.RowActive.Render(line)
	}
	return th.Row.Render(line)
}

// stateIndicator returns an emoji for the issue state.
func stateIndicator(state string) string {
	switch strings.ToLower(state) {
	case "closed":
		return "✅"
	default:
		return "🔵"
	}
}

// progressBar renders a filled/empty block bar.
func progressBar(completed, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := width * completed / total
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// truncate shortens s to fit within maxWidth columns.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	w := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxWidth-1 {
			return s[:i] + "…"
		}
		w += rw
	}
	return s
}
