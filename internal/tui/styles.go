package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#58A6FF") // GitHub blue
	ColorSecondary = lipgloss.Color("#8B949E") // muted gray
	ColorSuccess   = lipgloss.Color("#3FB950") // green
	ColorWarning   = lipgloss.Color("#D29922") // amber
	ColorDanger    = lipgloss.Color("#F85149") // red
	ColorAccent    = lipgloss.Color("#BC8CFF") // purple
	ColorDim       = lipgloss.Color("#484F58") // dim gray
	ColorBright    = lipgloss.Color("#F0F6FC") // bright white
)

// Text styles
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		MarginBottom(1)

	Subtitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBright)

	Heading = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	Muted = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	Success = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	Warning = lipgloss.NewStyle().
		Foreground(ColorWarning)

	Danger = lipgloss.NewStyle().
		Foreground(ColorDanger)

	Command = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	Code = lipgloss.NewStyle().
		Foreground(ColorBright).
		Background(lipgloss.Color("#161B22")).
		Padding(0, 1)

	Dimmed = lipgloss.NewStyle().
		Foreground(ColorDim)
)

// Layout styles
var (
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDim).
		Padding(1, 2)

	ActiveBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	HelpBar = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		MarginTop(1)
)

// Badge renders a small labeled badge.
func Badge(label string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(label)
}

// ProgressDots renders filled/empty dots for step progress.
func ProgressDots(current, total int) string {
	dots := ""
	for i := 0; i < total; i++ {
		if i < current {
			dots += Success.Render("●") + " "
		} else if i == current {
			dots += lipgloss.NewStyle().Foreground(ColorPrimary).Render("◉") + " "
		} else {
			dots += Dimmed.Render("○") + " "
		}
	}
	return dots
}
