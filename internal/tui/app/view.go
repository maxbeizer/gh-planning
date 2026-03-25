package app

import (
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

	// Search bar (between tab bar and content when active)
	searchBarHeight := 0
	if m.search.IsActive() {
		b.WriteString(m.search.View())
		b.WriteString("\n")
		searchBarHeight = 1
	}

	// Content area
	content := m.renderContent()
	contentHeight := m.ctx.ContentHeight() - searchBarHeight
	// Pad content to fill the available height.
	lines := strings.Count(content, "\n") + 1
	if lines < contentHeight {
		content += strings.Repeat("\n", contentHeight-lines)
	}
	b.WriteString(content)

	// Status bar (show refresh indicator when loading)
	if m.loading {
		b.WriteString(m.footer.ViewWithStatus("⟳ refreshing..."))
	} else {
		b.WriteString(m.footer.View())
	}

	base := b.String()

	// Help overlay — rendered on top when visible.
	if m.help.IsVisible() {
		overlay := m.help.View()
		base = lipgloss.Place(
			m.ctx.Width,
			m.ctx.Height,
			lipgloss.Left,
			lipgloss.Top,
			base,
		)
		// Layer the centered overlay on the base.
		base = overlayString(base, overlay, m.ctx.Width, m.ctx.Height)
	}

	// Detail pane overlay — rendered on top when visible.
	if m.detail.IsVisible() {
		overlay := m.detail.View()
		base = lipgloss.Place(
			m.ctx.Width,
			m.ctx.Height,
			lipgloss.Left,
			lipgloss.Top,
			base,
		)
		centeredOverlay := lipgloss.Place(
			m.ctx.Width,
			m.ctx.Height,
			lipgloss.Center,
			lipgloss.Center,
			overlay,
		)
		base = overlayString(base, centeredOverlay, m.ctx.Width, m.ctx.Height)
	}

	v := tea.NewView(base)
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

	if m.err != nil {
		return th.Danger.Render("  Error: " + m.err.Error())
	}

	switch m.activeTab {
	case 0:
		return m.board.View()
	case 1:
		return m.listview.View()
	case 2:
		return m.tree.View()
	case 3:
		return m.depgraph.View()
	default:
		return th.Muted.Render("Unknown view")
	}
}

// overlayString composites the overlay on top of base, line by line.
// Non-empty overlay lines replace corresponding base lines.
func overlayString(base, overlay string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure base has enough lines.
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	for i, ol := range overlayLines {
		if i < len(baseLines) && strings.TrimSpace(ol) != "" {
			baseLines[i] = ol
		}
	}

	return strings.Join(baseLines, "\n")
}
