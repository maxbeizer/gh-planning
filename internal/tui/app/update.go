package app

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/copilot"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		// Size the detail overlay to ~80% of the terminal.
		dw := msg.Width * 4 / 5
		dh := msg.Height * 4 / 5
		m.detail.SetSize(dw, dh)
		m.ready = true

	case projectDataMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else if msg.project != nil {
			m.err = nil
			m.board.SetItems(msg.project.Items)
			m.listview.SetItems(msg.project.Items)
			m.tree.SetItems(msg.project.Items)
			m.depgraph.SetItems(msg.project.Items)
			m.footer.SetRefreshTime(time.Now())
		}

	case copilot.DoneMsg:
		// Copilot session ended — refresh data since the user may have made changes.
		m.loading = true
		return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)

	case tea.KeyMsg:
		// When detail pane is visible, delegate all keys to it.
		if m.detail.IsVisible() {
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			return m, cmd
		}

		// When help overlay is visible, only ? and esc dismiss it.
		if m.help.IsVisible() {
			switch {
			case key.Matches(msg, m.keys.Help):
				m.help.Toggle()
			case key.Matches(msg, m.keys.ClearFilter):
				m.help.Toggle()
			}
			return m, nil
		}

		// When search bar is active, delegate keys to it (except esc).
		if m.search.IsActive() {
			if key.Matches(msg, m.keys.ClearFilter) {
				m.search.Deactivate()
				m.board.SetFilter("")
				m.listview.SetFilter("")
				m.updateFooterCounts()
				return m, nil
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			query := m.search.Query()
			m.board.SetFilter(query)
			m.listview.SetFilter(query)
			m.updateFooterCounts()
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.Toggle()
		case key.Matches(msg, m.keys.LaunchCopilot):
			var item *github.ProjectItem
			switch m.activeTab {
			case 0:
				item = m.board.SelectedItem()
			case 1:
				item = m.listview.SelectedItem()
			case 3:
				item = m.depgraph.SelectedItem()
			}
			if item != nil {
				return m, copilot.Launch(item)
			}
		case key.Matches(msg, m.keys.Refresh):
			if !m.loading {
				m.loading = true
				return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)
			}
		case key.Matches(msg, m.keys.NextTab):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case key.Matches(msg, m.keys.PrevTab):
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		case key.Matches(msg, m.keys.Filter):
			cmd := m.search.Activate()
			return m, cmd
		default:
			// Delegate to the active tab's sub-model.
			switch m.activeTab {
			case 0:
				var cmd tea.Cmd
				m.board, cmd = m.board.Update(msg)
				return m, cmd
			case 1:
				m.listview, _ = m.listview.Update(msg)
			case 2:
				m.tree, _ = m.tree.Update(msg)
			case 3:
				var cmd tea.Cmd
				m.depgraph, cmd = m.depgraph.Update(msg)
				return m, cmd
			}
		}
	}

	return m, nil
}

// updateFooterCounts syncs the footer item counts with the current filter state.
func (m *Model) updateFooterCounts() {
	total := m.board.TotalItemCount()
	query := m.search.Query()
	if m.search.IsActive() && query != "" {
		m.footer.SetItemCount(total, m.board.FilteredItemCount())
		m.footer.SetFilter(query)
	} else {
		m.footer.SetItemCount(total, -1)
		m.footer.SetFilter("")
	}
}
