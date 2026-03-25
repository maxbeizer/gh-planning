package prompt

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
)

// Model is a small text input overlay with a title and submit callback.
type Model struct {
	ctx      *tuictx.ProgramContext
	input    textinput.Model
	title    string
	visible  bool
	onSubmit func(string) tea.Cmd
}

// New creates a prompt Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Prompt = "▸ "
	return Model{
		ctx:   ctx,
		input: ti,
	}
}

// Show makes the prompt visible with the given title and callback.
func (m *Model) Show(title string, onSubmit func(string) tea.Cmd) tea.Cmd {
	m.title = title
	m.onSubmit = onSubmit
	m.visible = true
	m.input.SetValue("")
	return m.input.Focus()
}

// Hide dismisses the prompt without submitting.
func (m *Model) Hide() {
	m.visible = false
	m.onSubmit = nil
	m.input.Blur()
}

// IsVisible reports whether the prompt overlay is shown.
func (m Model) IsVisible() bool {
	return m.visible
}

// Update handles key input when the prompt is visible.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			val := m.input.Value()
			if val != "" && m.onSubmit != nil {
				cmd := m.onSubmit(val)
				m.Hide()
				return m, cmd
			}
			m.Hide()
			return m, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			m.Hide()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the prompt as a styled overlay box. Returns "" when hidden.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	th := m.ctx.Theme

	content := th.HelpTitle.Render(m.title) + "\n\n" +
		m.input.View() + "\n\n" +
		th.Muted.Render("enter submit • esc cancel")

	return th.Overlay.Render(content)
}

// OverlayView returns the prompt centered within the given dimensions.
func (m Model) OverlayView(width, height int) string {
	if !m.visible {
		return ""
	}
	overlay := m.View()
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
}
