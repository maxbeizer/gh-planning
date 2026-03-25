package search

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
)

// Model is a filter bar component backed by a textinput bubble.
type Model struct {
	ctx    *tuictx.ProgramContext
	input  textinput.Model
	active bool
	query  string
}

// New creates a search bar Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 128
	ti.Prompt = "🔍 "
	return Model{
		ctx:   ctx,
		input: ti,
	}
}

// Activate focuses the search bar and begins accepting input.
func (m *Model) Activate() tea.Cmd {
	m.active = true
	m.input.Focus()
	return m.input.Focus()
}

// Deactivate unfocuses the search bar and clears the query.
func (m *Model) Deactivate() {
	m.active = false
	m.query = ""
	m.input.SetValue("")
	m.input.Blur()
}

// IsActive reports whether the search bar is currently focused.
func (m *Model) IsActive() bool {
	return m.active
}

// Query returns the current filter text.
func (m *Model) Query() string {
	return m.query
}

// Update processes key messages when the search bar is active.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.query = m.input.Value()
	return m, cmd
}

// View renders the filter bar. Returns empty string when inactive.
func (m Model) View() string {
	if !m.active {
		return ""
	}
	th := m.ctx.Theme
	return th.FilterBar.Width(m.ctx.Width).Render(m.input.View())
}
