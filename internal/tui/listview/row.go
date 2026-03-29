package listview

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"

	"github.com/maxbeizer/gh-planning/internal/tui/board"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/theme"
)

// renderSectionHeader renders a collapsible status section header.
func renderSectionHeader(ctx *context.ProgramContext, sec *Section, active bool) string {
	th := ctx.Theme
	width := ctx.ContentWidth()

	chevron := "▼"
	if sec.Collapsed {
		chevron = "▶"
	}

	label := fmt.Sprintf("%s %s %s (%d)", chevron, sec.Emoji, sec.Status, len(sec.Items))

	// Use status-colored accent for the section header.
	headerStyle := sectionHeaderStyle(th, sec.Status)

	if active {
		return th.RowActive.Width(width).Render(label)
	}

	rendered := headerStyle.Render(label)
	// Pad to full width for consistent row appearance.
	renderedWidth := lipgloss.Width(rendered)
	if renderedWidth < width {
		rendered += strings.Repeat(" ", width-renderedWidth)
	}
	return rendered
}

// sectionHeaderStyle returns a bold style colored by status semantics.
func sectionHeaderStyle(th *theme.Theme, status string) lipgloss.Style {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "done") || strings.Contains(lower, "closed") || strings.Contains(lower, "complete"):
		return lipgloss.NewStyle().Bold(true).Foreground(theme.Green).Padding(0, 1)
	case strings.Contains(lower, "progress"):
		return lipgloss.NewStyle().Bold(true).Foreground(theme.Amber).Padding(0, 1)
	case strings.Contains(lower, "review") || strings.Contains(lower, "attention"):
		return lipgloss.NewStyle().Bold(true).Foreground(theme.Purple).Padding(0, 1)
	case strings.Contains(lower, "block"):
		return lipgloss.NewStyle().Bold(true).Foreground(theme.Red).Padding(0, 1)
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(theme.Bright).Padding(0, 1)
	}
}

// renderItemRow renders a single project item row.
func renderItemRow(ctx *context.ProgramContext, r rowEntry, active bool) string {
	th := ctx.Theme
	width := ctx.ContentWidth()

	item := r.item

	// Blocked indicator prefix for items with open dependency blockers.
	blockedPrefix := ""
	if item.IsBlocked() {
		blockedPrefix = "🚫 "
	}

	// Build columns.
	number := fmt.Sprintf("#%d", item.Number)
	repo := shortRepo(item.Repository)
	assignee := "—"
	if len(item.Assignees) > 0 {
		assignee = "@" + item.Assignees[0]
	}
	age := humanizeTime(item.UpdatedAt)

	// Fixed-width columns: "  #123  " (8) + repo (16) + assignee (14) + age (6) + padding (6)
	const (
		numW      = 8
		repoW     = 16
		assigneeW = 14
		ageW      = 6
		padding   = 6 // spaces between columns
	)
	fixedW := numW + repoW + assigneeW + ageW + padding
	titleW := width - fixedW
	if titleW < 10 {
		titleW = 10
	}

	typePrefix := ""
	badge := board.IssueTypeBadge(item.Fields["Issue Type"])
	if badge == "" {
		badge = board.IssueTypeBadge(item.IssueType)
	}
	if badge != "" {
		typePrefix = badge + " "
	}
	title := truncate(blockedPrefix+typePrefix+item.Title, titleW)

	// Sub-issue progress indicator appended to title if present.
	if item.SubIssuesSummary.Total > 0 {
		prog := fmt.Sprintf(" [%d/%d]", item.SubIssuesSummary.Completed, item.SubIssuesSummary.Total)
		progW := runewidth.StringWidth(prog)
		if runewidth.StringWidth(title)+progW <= titleW {
			title = title + prog
		} else {
			title = truncate(item.Title, titleW-progW) + prog
		}
	}

	// Compose the row with styled segments.
	var parts []string
	parts = append(parts, "  ")
	parts = append(parts, th.CardNumber.Render(padRight(number, numW)))
	parts = append(parts, padRight(title, titleW))
	parts = append(parts, " ")
	parts = append(parts, th.Muted.Render(padRight(repo, repoW)))
	parts = append(parts, " ")
	parts = append(parts, th.Muted.Render(padRight(assignee, assigneeW)))
	parts = append(parts, " ")
	parts = append(parts, th.Dimmed.Render(padRight(age, ageW)))

	line := strings.Join(parts, "")

	// Apply row style.
	isBlocked := item.IsBlocked()
	switch {
	case active:
		return th.RowActive.Width(width).Render(stripAnsi(line))
	case isBlocked:
		return th.Danger.Width(width).Render(stripAnsi(line))
	case r.visibleIndex%2 == 0:
		return th.RowAlt.Width(width).Render(stripAnsi(line))
	default:
		return line
	}
}

// truncate shortens s to at most maxW display columns using runewidth,
// appending "…" if truncated.
func truncate(s string, maxW int) string {
	if runewidth.StringWidth(s) <= maxW {
		return s
	}
	w := 0
	for i, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > maxW-1 { // leave room for ellipsis
			return s[:i] + "…"
		}
		w += rw
	}
	return s
}

// padRight pads s with spaces to exactly w display columns.
func padRight(s string, w int) string {
	sw := runewidth.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

// shortRepo returns just the repo name from "owner/repo".
func shortRepo(repo string) string {
	if i := strings.LastIndex(repo, "/"); i >= 0 {
		return repo[i+1:]
	}
	return repo
}

// humanizeTime returns a short human-readable duration since t.
func humanizeTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
}

// stripAnsi is a simple ANSI escape stripper for re-styling composed strings.
func stripAnsi(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
