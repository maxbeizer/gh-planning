package help

import (
	"strings"

	"charm.land/lipgloss/v2"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
)

// Model is a render-only help overlay component.
type Model struct {
	ctx     *tuictx.ProgramContext
	visible bool
}

// New creates a help Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{ctx: ctx}
}

// Toggle flips the overlay visibility.
func (m *Model) Toggle() { m.visible = !m.visible }

// IsVisible reports whether the overlay is currently shown.
func (m Model) IsVisible() bool { return m.visible }

// section groups keybindings under a heading.
type section struct {
	title    string
	bindings []binding
}

type binding struct {
	key  string
	desc string
}

var sections = []section{
	{
		title: "Global",
		bindings: []binding{
			{"q", "quit"},
			{"?", "help"},
			{"R", "refresh"},
			{"tab", "switch view"},
		},
	},
	{
		title: "Navigation",
		bindings: []binding{
			{"j/k", "up/down"},
			{"h/l", "left/right (board)"},
			{"enter", "select/expand"},
		},
	},
	{
		title: "Actions",
		bindings: []binding{
			{"o", "open in browser"},
			{"esc", "back/close"},
		},
	},
}

// View renders the help overlay. Returns "" when not visible.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	th := m.ctx.Theme

	var b strings.Builder
	b.WriteString(th.HelpTitle.Render("Keyboard Shortcuts"))
	b.WriteString("\n")

	for _, sec := range sections {
		b.WriteString(th.HelpHeading.Render(sec.title))
		b.WriteString("\n")
		for _, bind := range sec.bindings {
			b.WriteString("  ")
			b.WriteString(th.HelpKey.Render(bind.key))
			b.WriteString(th.HelpDesc.Render(bind.desc))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(th.Dimmed.Render("Press ? or esc to close"))

	box := th.Overlay.Render(b.String())

	return lipgloss.Place(
		m.ctx.Width,
		m.ctx.Height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}
