package board

import (
	"fmt"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/theme"
)

// staleDuration is the threshold after which a card is considered stale.
const staleDuration = 14 * 24 * time.Hour

// RenderCard renders a single project item as a two-line card.
// Line 1: #N  Title (truncated)
// Line 2: @assignee  repo  2d ago
func RenderCard(item github.ProjectItem, width int, isActive bool, th *theme.Theme) string {
	// Card border + padding consumes 4 chars of width (border left/right + padding left/right)
	innerWidth := width - 4
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Line 1: issue number + title
	number := fmt.Sprintf("#%d", item.Number)
	// Reserve space for number + 2 spaces
	titleWidth := innerWidth - runewidth.StringWidth(number) - 2
	if titleWidth < 4 {
		titleWidth = 4
	}
	title := truncate(item.Title, titleWidth)

	isStale := !item.UpdatedAt.IsZero() && time.Since(item.UpdatedAt) > staleDuration

	// Use dimmed styles for stale cards so all text is noticeably faded.
	numStyle := th.CardNumber
	titleStyle := th.CardTitle
	metaStyle := th.CardMeta
	if isStale {
		numStyle = th.Dimmed
		titleStyle = th.Dimmed
		metaStyle = th.CardStaleMeta
	}

	badge := IssueTypeBadge(item.Fields["Issue Type"])
	if badge == "" {
		badge = IssueTypeBadge(item.IssueType)
	}
	var line1 string
	if badge != "" {
		line1 = numStyle.Render(number) + " " + badge + " " + titleStyle.Render(title)
	} else {
		line1 = numStyle.Render(number) + "  " + titleStyle.Render(title)
	}

	// Line 2: meta — @assignee  repo  age
	var metaParts []string
	if len(item.Assignees) > 0 {
		metaParts = append(metaParts, "@"+item.Assignees[0])
	}
	if item.Repository != "" {
		// Show just the repo name, not the full owner/repo
		parts := strings.SplitN(item.Repository, "/", 2)
		if len(parts) == 2 {
			metaParts = append(metaParts, parts[1])
		} else {
			metaParts = append(metaParts, item.Repository)
		}
	}
	if !item.UpdatedAt.IsZero() {
		metaParts = append(metaParts, humanizeDuration(time.Since(item.UpdatedAt)))
	}
	meta := strings.Join(metaParts, "  ")
	line2 := metaStyle.Render(truncate(meta, innerWidth))

	// Add blocked indicator when the item has open dependency blockers.
	if item.IsBlocked() {
		line2 = th.Danger.Render("🚫 blocked") + "  " + line2
	}

	content := line1 + "\n" + line2

	// Determine if blocked: either the status column contains "block" or
	// the item has open dependency blockers from the enrichment pass.
	isBlocked := strings.Contains(strings.ToLower(item.Status), "block") || item.IsBlocked()

	var style func(strs ...string) string
	switch {
	case isActive:
		style = th.CardActive.Width(width).Render
	case isBlocked:
		style = th.CardBlocked.Width(width).Render
	case isStale:
		style = th.CardStale.Width(width).Render
	default:
		style = th.Card.Width(width).Render
	}

	return style(content)
}

// truncate shortens s to fit within width columns using runewidth.
func truncate(s string, width int) string {
	if runewidth.StringWidth(s) <= width {
		return s
	}
	return runewidth.Truncate(s, width, "…")
}

// humanizeDuration formats a duration into a short human-readable string.
func humanizeDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
