package footer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
)

// Model is a render-only footer/status bar component.
type Model struct {
	ctx           *tuictx.ProgramContext
	lastRefresh   time.Time
	itemCount     int
	filteredCount int // -1 if no filter active
}

// New creates a footer Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{
		ctx:           ctx,
		filteredCount: -1,
	}
}

// SetRefreshTime records when data was last fetched.
func (m *Model) SetRefreshTime(t time.Time) {
	m.lastRefresh = t
}

// SetItemCount updates the displayed item counts.
// Pass filtered = -1 when no filter is active.
func (m *Model) SetItemCount(total, filtered int) {
	m.itemCount = total
	m.filteredCount = filtered
}

// ViewWithStatus renders the status bar with a custom right-side status message.
func (m Model) ViewWithStatus(status string) string {
	th := m.ctx.Theme
	w := m.ctx.Width

	left := m.renderLeft()
	center := m.renderCenter()
	right := status + " "

	leftW := lipgloss.Width(left)
	centerW := lipgloss.Width(center)
	rightW := lipgloss.Width(right)

	totalContent := leftW + centerW + rightW
	totalGap := w - totalContent
	if totalGap < 2 {
		bar := left + " " + center + " " + right
		return th.StatusBar.Width(w).Render(bar)
	}

	leftGap := (w - centerW) / 2 - leftW
	if leftGap < 1 {
		leftGap = 1
	}
	rightGap := w - leftW - leftGap - centerW - rightW
	if rightGap < 1 {
		rightGap = 1
	}

	bar := left + strings.Repeat(" ", leftGap) + center + strings.Repeat(" ", rightGap) + right
	return th.StatusBar.Width(w).Render(bar)
}

// View renders the full-width status bar.
func (m Model) View() string {
	th := m.ctx.Theme
	w := m.ctx.Width

	left := m.renderLeft()
	center := m.renderCenter()
	right := m.renderRight()

	leftW := lipgloss.Width(left)
	centerW := lipgloss.Width(center)
	rightW := lipgloss.Width(right)

	// Distribute gaps so center is visually centered.
	totalContent := leftW + centerW + rightW
	totalGap := w - totalContent
	if totalGap < 2 {
		// Not enough room — just join with single spaces.
		bar := left + " " + center + " " + right
		return th.StatusBar.Width(w).Render(bar)
	}

	leftGap := (w - centerW) / 2 - leftW
	if leftGap < 1 {
		leftGap = 1
	}
	rightGap := w - leftW - leftGap - centerW - rightW
	if rightGap < 1 {
		rightGap = 1
	}

	bar := left + strings.Repeat(" ", leftGap) + center + strings.Repeat(" ", rightGap) + right
	return th.StatusBar.Width(w).Render(bar)
}

func (m Model) renderLeft() string {
	profile := m.ctx.ProfileName
	owner := m.ctx.Owner
	num := m.ctx.ProjectNumber

	if profile != "" {
		return fmt.Sprintf(" %s · %s/#%d", profile, owner, num)
	}
	return fmt.Sprintf(" %s/#%d", owner, num)
}

func (m Model) renderCenter() string {
	if m.itemCount == 0 {
		return ""
	}
	if m.filteredCount >= 0 {
		return fmt.Sprintf("%d/%d filtered", m.filteredCount, m.itemCount)
	}
	return fmt.Sprintf("%d items", m.itemCount)
}

func (m Model) renderRight() string {
	if m.lastRefresh.IsZero() {
		return "not yet loaded "
	}
	return fmt.Sprintf("refreshed %s ", formatDuration(time.Since(m.lastRefresh)))
}

// formatDuration produces a human-friendly relative time string.
func formatDuration(d time.Duration) string {
	secs := int(math.Round(d.Seconds()))
	switch {
	case secs < 60:
		return "just now"
	case secs < 3600:
		m := secs / 60
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	default:
		h := secs / 3600
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	}
}
