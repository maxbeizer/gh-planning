package theme

import "charm.land/lipgloss/v2"

// GitHub color palette — matches internal/tui/styles.go but using Lipgloss v2 API.
var (
	// Primary colors
	Blue   = lipgloss.Color("#58A6FF") // GitHub blue — links, interactive, primary
	Gray   = lipgloss.Color("#8B949E") // muted gray — secondary info
	Green  = lipgloss.Color("#3FB950") // green — done, success, closed
	Amber  = lipgloss.Color("#D29922") // amber — in progress, warning
	Red    = lipgloss.Color("#F85149") // red — blocked, error, danger
	Purple = lipgloss.Color("#BC8CFF") // purple — accent, headings
	Dim    = lipgloss.Color("#484F58") // dim gray — stale, borders
	Bright = lipgloss.Color("#F0F6FC") // bright white — titles, emphasis

	// Backgrounds
	BgDark    = lipgloss.Color("#0D1117") // GitHub dark background
	BgSurface = lipgloss.Color("#161B22") // slightly raised surface
	BgHover   = lipgloss.Color("#1F2937") // hover/selected background

	// Status semantics — use these for consistent meaning across views.
	StatusDone       = Green
	StatusInProgress = Amber
	StatusBlocked    = Red
	StatusBacklog    = Gray
	StatusStale      = Dim
)
