package app

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/actions"
	"github.com/maxbeizer/gh-planning/internal/tui/components/picker"
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

	case autoRefreshTickMsg:
		var cmds []tea.Cmd
		if !m.loading && m.autoRefreshInterval > 0 {
			m.loading = true
			cmds = append(cmds, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber))
		}
		cmds = append(cmds, m.scheduleAutoRefresh())
		return m, tea.Batch(cmds...)

	case statusInfoMsg:
		if msg.err != nil {
			m.footer.SetStatus(fmt.Sprintf("Failed to load statuses: %v", msg.err))
			break
		}
		m.statusProjectID = msg.projectID
		m.statusFieldID = msg.statusFieldID
		m.statusOptions = msg.statusOptions
		m.pendingItem = msg.item

		var opts []picker.Option
		for name := range msg.statusOptions {
			opts = append(opts, picker.Option{Name: name, ID: msg.statusOptions[name]})
		}
		m.picker.SetTitle("Change Status")
		m.picker.SetOptions(opts)
		m.picker.SetOnSelect(func(opt picker.Option) tea.Cmd {
			m.mutating = true
			return updateItemStatus(
				m.ctx.Owner, m.ctx.ProjectNumber,
				m.statusProjectID, m.pendingItem.ID,
				m.statusFieldID, opt.ID, opt.Name,
			)
		})
		m.picker.SetVisible(true)

	case statusUpdatedMsg:
		m.mutating = false
		if msg.err != nil {
			m.footer.SetStatus(fmt.Sprintf("Status change failed: %v", msg.err))
		} else {
			m.footer.SetStatus(fmt.Sprintf("✓ Status → %s", msg.newStatus))
			m.loading = true
			return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)
		}

	case copilot.DoneMsg:
		// Copilot session ended — refresh data since the user may have made changes.
		m.loading = true
		return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)

	case actions.AssignResultMsg:
		if msg.Err != nil {
			m.footer.SetStatus(fmt.Sprintf("Assign failed: %v", msg.Err))
		} else {
			m.footer.SetStatus(fmt.Sprintf("Assigned %s", msg.User))
			m.loading = true
			return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)
		}

	case actions.LabelResultMsg:
		if msg.Err != nil {
			m.footer.SetStatus(fmt.Sprintf("Label failed: %v", msg.Err))
		} else {
			m.footer.SetStatus(fmt.Sprintf("Added label %s", msg.Label))
			m.loading = true
			return m, fetchProjectDataFresh(m.ctx.Owner, m.ctx.ProjectNumber)
		}

	case actions.LogResultMsg:
		if msg.Err != nil {
			m.footer.SetStatus(fmt.Sprintf("Log failed: %v", msg.Err))
		} else {
			m.footer.SetStatus("Progress logged")
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

		// When picker overlay is visible, delegate to picker.
		if m.picker.IsVisible() {
			var cmd tea.Cmd
			m.picker, cmd = m.picker.Update(msg)
			return m, cmd
		}

		// When prompt overlay is visible, delegate to prompt.
		if m.prompt.IsVisible() {
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
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
		case key.Matches(msg, m.actionKeys.Assign):
			item := m.activeItem()
			if item == nil {
				break
			}
			team := m.ctx.Config.Team
			if len(team) == 0 {
				m.footer.SetStatus("No team configured — add team members in config")
				break
			}
			opts := make([]picker.Option, len(team))
			for i, u := range team {
				opts[i] = picker.Option{Name: u, ID: u}
			}
			m.picker.SetTitle("Assign User")
			m.picker.SetOptions(opts)
			m.picker.SetOnSelect(func(opt picker.Option) tea.Cmd {
				return actions.AssignUser(item.Repository, item.Number, opt.ID)
			})
			m.picker.SetVisible(true)
		case key.Matches(msg, m.actionKeys.Label):
			item := m.activeItem()
			if item == nil {
				break
			}
			cmd := m.prompt.Show("Add label:", func(label string) tea.Cmd {
				return actions.AddLabel(item.Repository, item.Number, label)
			})
			return m, cmd
		case key.Matches(msg, m.actionKeys.Log):
			item := m.activeItem()
			if item == nil {
				break
			}
			cmd := m.prompt.Show("Log progress:", func(message string) tea.Cmd {
				return actions.LogProgress(message, item.Number)
			})
			return m, cmd
		case key.Matches(msg, m.actionKeys.ChangeStatus):
			item := m.activeItem()
			if item == nil {
				break
			}
			m.footer.SetStatus("Loading statuses...")
			return m, fetchStatusInfo(m.ctx.Owner, m.ctx.ProjectNumber, *item)
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

// activeItem returns the currently selected item from whichever view tab is active.
func (m *Model) activeItem() *github.ProjectItem {
	switch m.activeTab {
	case 0:
		return m.board.SelectedItem()
	case 1:
		return m.listview.SelectedItem()
	case 3:
		return m.depgraph.SelectedItem()
	}
	return nil
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
