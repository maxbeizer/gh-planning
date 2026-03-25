package app

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
			m.footer.SetRefreshTime(time.Now())
		}

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

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.Toggle()
		case key.Matches(msg, m.keys.Refresh):
			if !m.loading {
				m.loading = true
				return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)
			}
		case key.Matches(msg, m.keys.NextTab):
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
		case key.Matches(msg, m.keys.PrevTab):
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
		default:
			// Delegate to the active tab's sub-model.
			switch m.activeTab {
			case 0:
				var cmd tea.Cmd
				m.board, cmd = m.board.Update(msg)
				return m, cmd
			case 1:
				m.listview, _ = m.listview.Update(msg)
			}
		}
	}

	return m, nil
}
