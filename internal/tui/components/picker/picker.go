package picker

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
)

// Option represents a single selectable option in the picker.
type Option struct {
	Name  string
	ID    string // GraphQL field value option ID
	Emoji string
}

// Model is a floating overlay picker for selecting from a list of options.
type Model struct {
	ctx      *tuictx.ProgramContext
	options  []Option
	cursor   int
	visible  bool
	title    string
	onSelect func(Option) tea.Cmd
}

// New creates a picker model. onSelect is called when the user confirms a choice.
func New(ctx *tuictx.ProgramContext, options []Option, onSelect func(Option) tea.Cmd) Model {
	return Model{
		ctx:      ctx,
		options:  options,
		title:    "Change Status",
		onSelect: onSelect,
	}
}

// SetVisible shows or hides the picker, resetting cursor on show.
func (m *Model) SetVisible(v bool) {
	m.visible = v
	if v {
		m.cursor = 0
	}
}

// IsVisible returns whether the picker overlay is shown.
func (m Model) IsVisible() bool {
	return m.visible
}

// SetTitle changes the picker heading text.
func (m *Model) SetTitle(title string) {
	m.title = title
}

// SetOnSelect replaces the selection callback.
func (m *Model) SetOnSelect(fn func(Option) tea.Cmd) {
	m.onSelect = fn
}

// SetOptions replaces the option list and resets the cursor.
func (m *Model) SetOptions(options []Option) {
	m.options = options
	m.cursor = 0
}

// Update handles keyboard input for the picker.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.options) - 1
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if len(m.options) > 0 && m.onSelect != nil {
				m.visible = false
				return m, m.onSelect(m.options[m.cursor])
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.visible = false
		}
	}

	return m, nil
}

// View renders the picker as a styled overlay box.
func (m Model) View() string {
	if !m.visible || len(m.options) == 0 {
		return ""
	}

	th := m.ctx.Theme

	var b strings.Builder
	b.WriteString(th.HelpTitle.Render(m.title))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}

		label := opt.Name
		if opt.Emoji != "" {
			label = opt.Emoji + " " + label
		}

		if i == m.cursor {
			b.WriteString(th.RowActive.Render(fmt.Sprintf("%s%s", prefix, label)))
		} else {
			b.WriteString(th.Row.Render(fmt.Sprintf("%s%s", prefix, label)))
		}
		if i < len(m.options)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(th.Muted.Render("j/k navigate • enter select • esc cancel"))

	content := b.String()
	return th.Overlay.Render(content)
}

// OverlayView returns the picker centered within the given dimensions,
// suitable for compositing on top of the base view.
func (m Model) OverlayView(width, height int) string {
	if !m.visible {
		return ""
	}
	overlay := m.View()
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
}
