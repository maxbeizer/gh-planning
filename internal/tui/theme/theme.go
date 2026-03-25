package theme

import "charm.land/lipgloss/v2"

// Theme holds all Lipgloss v2 styles used by TUI components.
// Instantiate via New() and pass through ProgramContext.
type Theme struct {
	// Text
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Heading  lipgloss.Style
	Muted    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Danger   lipgloss.Style
	Dimmed   lipgloss.Style
	Code     lipgloss.Style

	// Cards (kanban board)
	Card         lipgloss.Style
	CardActive   lipgloss.Style
	CardStale    lipgloss.Style
	CardBlocked  lipgloss.Style
	CardNumber   lipgloss.Style
	CardTitle    lipgloss.Style
	CardMeta     lipgloss.Style
	CardStaleMeta lipgloss.Style

	// Columns (kanban board)
	ColumnHeader       lipgloss.Style
	ColumnHeaderActive lipgloss.Style
	ColumnBorder       lipgloss.Style

	// Rows (list view)
	Row            lipgloss.Style
	RowActive      lipgloss.Style
	RowAlt         lipgloss.Style
	SectionHeader  lipgloss.Style

	// Detail pane
	DetailLabel lipgloss.Style

	// Help overlay
	HelpKey     lipgloss.Style
	HelpDesc    lipgloss.Style
	HelpHeading lipgloss.Style
	HelpTitle   lipgloss.Style

	// Layout
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	StatusBar   lipgloss.Style
	HelpBar     lipgloss.Style
	FilterBar   lipgloss.Style
	Overlay     lipgloss.Style

	// Badges
	BadgeDone       lipgloss.Style
	BadgeInProgress lipgloss.Style
	BadgeBlocked    lipgloss.Style
	BadgeStale      lipgloss.Style
}

// New creates a Theme with the default GitHub-inspired color palette.
func New() *Theme {
	return &Theme{
		// Text styles
		Title:    lipgloss.NewStyle().Bold(true).Foreground(Blue),
		Subtitle: lipgloss.NewStyle().Bold(true).Foreground(Bright),
		Heading:  lipgloss.NewStyle().Bold(true).Foreground(Purple),
		Muted:    lipgloss.NewStyle().Foreground(Gray),
		Success:  lipgloss.NewStyle().Foreground(Green),
		Warning:  lipgloss.NewStyle().Foreground(Amber),
		Danger:   lipgloss.NewStyle().Foreground(Red),
		Dimmed:   lipgloss.NewStyle().Foreground(Dim),
		Code:     lipgloss.NewStyle().Foreground(Bright).Background(BgSurface).Padding(0, 1),

		// Card styles
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1).
			MarginBottom(1),
		CardActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Blue).
			Background(BgHover).
			Padding(0, 1).
			MarginBottom(1),
		CardStale: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Dim).
			Padding(0, 1).
			MarginBottom(1).
			Foreground(Dim),
		CardBlocked: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Red).
			Padding(0, 1).
			MarginBottom(1),
		CardNumber: lipgloss.NewStyle().Foreground(Blue).Bold(true),
		CardTitle:  lipgloss.NewStyle().Foreground(Bright).Bold(true),
		CardMeta:   lipgloss.NewStyle().Foreground(Gray),
		CardStaleMeta: lipgloss.NewStyle().Foreground(Dim),

		// Column styles
		ColumnHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(Gray).
			Padding(0, 1).
			MarginBottom(1),
		ColumnHeaderActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(Blue).
			Padding(0, 1).
			MarginBottom(1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Blue),
		ColumnBorder: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(Dim),

		// Row styles
		Row:       lipgloss.NewStyle().Padding(0, 1),
		RowActive: lipgloss.NewStyle().Padding(0, 1).Background(BgHover).Foreground(Bright).Bold(true),
		RowAlt:    lipgloss.NewStyle().Padding(0, 1).Background(BgSurface),
		SectionHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(Bright).
			Padding(0, 1).
			MarginTop(1),

		// Detail pane styles
		DetailLabel: lipgloss.NewStyle().Foreground(Gray).Bold(true).Width(12),

		// Help overlay styles
		HelpKey:     lipgloss.NewStyle().Bold(true).Foreground(Blue).Width(12),
		HelpDesc:    lipgloss.NewStyle().Foreground(Gray),
		HelpHeading: lipgloss.NewStyle().Bold(true).Foreground(Purple).MarginTop(1),
		HelpTitle:   lipgloss.NewStyle().Bold(true).Foreground(Bright).MarginBottom(1),

		// Layout styles
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(Blue).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Blue).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(Gray).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(Dim).
			Padding(0, 2),
		StatusBar: lipgloss.NewStyle().
			Foreground(Gray).
			Background(BgSurface).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(Dim),
		HelpBar: lipgloss.NewStyle().
			Foreground(Dim),
		FilterBar: lipgloss.NewStyle().
			Foreground(Bright).
			Background(BgSurface).
			Padding(0, 1),
		Overlay: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Blue).
			Padding(1, 3).
			Background(BgDark),

		// Badge styles
		BadgeDone:       lipgloss.NewStyle().Foreground(Green).Bold(true),
		BadgeInProgress: lipgloss.NewStyle().Foreground(Amber).Bold(true),
		BadgeBlocked:    lipgloss.NewStyle().Foreground(Red).Bold(true),
		BadgeStale:      lipgloss.NewStyle().Foreground(Dim).Bold(true),
	}
}
