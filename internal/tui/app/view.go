package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	if !m.ready {
		return tea.NewView("Loading...")
	}

	var b strings.Builder

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	// Content area
	content := m.renderContent()
	contentHeight := m.ctx.ContentHeight()
	// Pad content to fill the available height.
	lines := strings.Count(content, "\n") + 1
	if lines < contentHeight {
		content += strings.Repeat("\n", contentHeight-lines)
	}
	b.WriteString(content)

	// Status bar
	b.WriteString(m.renderStatusBar())

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m Model) renderTabBar() string {
	var tabs []string
	for i, name := range m.tabs {
		if i == m.activeTab {
			tabs = append(tabs, m.ctx.Theme.TabActive.Render(name))
		} else {
			tabs = append(tabs, m.ctx.Theme.TabInactive.Render(name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

func (m Model) renderContent() string {
	th := m.ctx.Theme
	switch m.activeTab {
	case 0:
		return th.Muted.Render("Board view — coming soon")
	case 1:
		return th.Muted.Render("List view — coming soon")
	default:
		return th.Muted.Render("Unknown view")
	}
}

func (m Model) renderStatusBar() string {
	th := m.ctx.Theme

	left := fmt.Sprintf(" %s · %s/#%d",
		m.ctx.ProfileName, m.ctx.Owner, m.ctx.ProjectNumber)
	if m.ctx.ProfileName == "" {
		left = fmt.Sprintf(" %s/#%d", m.ctx.Owner, m.ctx.ProjectNumber)
	}

	right := "? help · q quit · tab switch view "
	gap := m.ctx.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return th.StatusBar.Width(m.ctx.Width).Render(bar)
}
