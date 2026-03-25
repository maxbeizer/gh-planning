package board

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/theme"
)

// Column represents a single status column on the kanban board.
type Column struct {
	Status string
	Emoji  string
	Items  []github.ProjectItem
	Offset int // scroll offset for overflow
}

// Render draws the column header and visible cards within the given dimensions.
func (c *Column) Render(width, height int, isActive bool, activeCard int, th *theme.Theme) string {
	var b strings.Builder

	// Header: emoji + status + (count)
	header := fmt.Sprintf("%s %s (%d)", c.Emoji, c.Status, len(c.Items))
	if isActive {
		b.WriteString(th.ColumnHeaderActive.Width(width).Render(header))
	} else {
		b.WriteString(th.ColumnHeader.Width(width).Render(header))
	}
	b.WriteString("\n")

	// Calculate how many cards can fit in remaining height.
	// Header takes ~3 lines (text + underline border for active + margin-bottom).
	availableHeight := height - 3
	if availableHeight < 1 {
		availableHeight = 1
	}

	if len(c.Items) == 0 {
		empty := th.Muted.Render("  (empty)")
		b.WriteString("\n") // vertical spacing
		b.WriteString(empty)
		return b.String()
	}

	// Determine visible range based on Offset
	visibleCount := c.visibleCardCount(availableHeight, width)
	start := c.Offset
	end := start + visibleCount
	if end > len(c.Items) {
		end = len(c.Items)
	}

	// Show scroll-up indicator if scrolled down
	if start > 0 {
		above := fmt.Sprintf("  ↑ %d above", start)
		b.WriteString(th.Muted.Render(above))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		cardActive := isActive && i == activeCard
		b.WriteString(RenderCard(c.Items[i], width, cardActive, th))
		b.WriteString("\n")
	}

	// Show overflow indicator if items are clipped below
	if end < len(c.Items) {
		more := fmt.Sprintf("  ↓ +%d more", len(c.Items)-end)
		b.WriteString(th.Muted.Render(more))
	}

	return b.String()
}

// visibleCardCount estimates how many cards fit in the available height.
func (c *Column) visibleCardCount(availableHeight, width int) int {
	if len(c.Items) == 0 {
		return 0
	}
	// Each card: 2 content lines + 2 border lines + 1 margin = ~5 lines
	// But lipgloss border rendering may vary, estimate conservatively.
	cardHeight := 5
	count := availableHeight / cardHeight
	if count < 1 {
		count = 1
	}
	if count > len(c.Items) {
		count = len(c.Items)
	}
	return count
}

// EnsureActiveVisible adjusts the scroll offset so that the active card is visible.
func (c *Column) EnsureActiveVisible(activeCard int, availableHeight, width int) {
	visibleCount := c.visibleCardCount(availableHeight-2, width)
	if visibleCount <= 0 {
		visibleCount = 1
	}

	if activeCard < c.Offset {
		c.Offset = activeCard
	} else if activeCard >= c.Offset+visibleCount {
		c.Offset = activeCard - visibleCount + 1
	}
}

// statusEmoji maps a status string to its emoji. Mirrors cmd/board.go statusEmoji.
func statusEmoji(status string) string {
	lower := strings.ToLower(status)
	switch lower {
	case "in progress":
		return "🔵"
	case "backlog", "ready", "todo":
		return "📋"
	case "done", "closed", "complete", "completed":
		return "✅"
	case "in review", "needs review", "needs my attention":
		return "🔍"
	case "blocked":
		return "🚫"
	default:
		return "•"
	}
}

// renderColumns renders all columns side-by-side using lipgloss.
func renderColumns(columns []Column, colWidth, height, activeCol, activeCard int, th *theme.Theme) string {
	if len(columns) == 0 {
		return th.Muted.Render("  No columns to display")
	}

	rendered := make([]string, len(columns))
	for i := range columns {
		isActive := i == activeCol
		card := -1
		if isActive {
			card = activeCard
		}
		col := th.ColumnBorder.Height(height).Width(colWidth).Render(
			columns[i].Render(colWidth-2, height, isActive, card, th),
		)
		rendered[i] = col
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}
