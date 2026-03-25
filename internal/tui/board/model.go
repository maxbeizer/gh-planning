package board

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
)

// statusOrder defines the preferred column ordering, mirroring cmd/board.go.
var statusOrder = []string{
	"backlog", "ready", "todo",
	"in progress", "in review", "needs review", "needs my attention",
	"done", "closed", "complete", "completed",
}

// Model is the Bubble Tea model for the kanban board view.
type Model struct {
	ctx       *tuictx.ProgramContext
	keys      keys.NavigationKeyMap
	columns    []Column
	activeCol  int
	activeCard int
	allItems   map[string][]github.ProjectItem // unfiltered items
	filter     string
	ready      bool
}

// New creates a board model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{
		ctx:  ctx,
		keys: keys.NewNavigationKeyMap(),
	}
}

// Init satisfies tea.Model. No initial command needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages for board navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(m.columns) == 0 {
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keys.Left):
			if m.activeCol > 0 {
				m.activeCol--
			} else {
				m.activeCol = len(m.columns) - 1
			}
			// Clamp active card to new column
			m.clampActiveCard()
			m.ensureVisible()

		case key.Matches(msg, m.keys.Right):
			if m.activeCol < len(m.columns)-1 {
				m.activeCol++
			} else {
				m.activeCol = 0
			}
			m.clampActiveCard()
			m.ensureVisible()

		case key.Matches(msg, m.keys.Up):
			col := &m.columns[m.activeCol]
			if len(col.Items) > 0 {
				if m.activeCard > 0 {
					m.activeCard--
				} else {
					m.activeCard = len(col.Items) - 1
				}
				m.ensureVisible()
			}

		case key.Matches(msg, m.keys.Down):
			col := &m.columns[m.activeCol]
			if len(col.Items) > 0 {
				if m.activeCard < len(col.Items)-1 {
					m.activeCard++
				} else {
					m.activeCard = 0
				}
				m.ensureVisible()
			}

		case key.Matches(msg, m.keys.Select):
			// Enter — detail pane will hook in later
		}
	}

	return m, nil
}

// View renders the kanban board.
func (m Model) View() string {
	if !m.ready || len(m.columns) == 0 {
		return m.ctx.Theme.Muted.Render("  Loading board…")
	}

	contentWidth := m.ctx.ContentWidth()
	contentHeight := m.ctx.ContentHeight()
	colWidth := contentWidth / len(m.columns)
	if colWidth < 24 {
		colWidth = 24
	}
	if colWidth > 50 {
		colWidth = 50
	}

	return renderColumns(m.columns, colWidth, contentHeight, m.activeCol, m.activeCard, m.ctx.Theme)
}

// SetItems populates the board columns from project data grouped by status.
func (m *Model) SetItems(items map[string][]github.ProjectItem) {
	m.allItems = items
	m.rebuildColumns()
	m.ready = true
	m.clampActiveCard()
}

// SetFilter applies a filter query. Empty string clears the filter.
func (m *Model) SetFilter(query string) {
	m.filter = query
	m.rebuildColumns()
}

// FilteredItemCount returns the number of items currently visible.
func (m *Model) FilteredItemCount() int {
	n := 0
	for _, col := range m.columns {
		n += len(col.Items)
	}
	return n
}

// TotalItemCount returns the total number of unfiltered items.
func (m *Model) TotalItemCount() int {
	n := 0
	for _, items := range m.allItems {
		n += len(items)
	}
	return n
}

// rebuildColumns reconstructs columns from allItems, applying the current filter.
func (m *Model) rebuildColumns() {
	filtered := filterItems(m.allItems, m.filter)
	statuses := sortedStatuses(filtered)
	m.columns = make([]Column, len(statuses))
	for i, status := range statuses {
		m.columns[i] = Column{
			Status: status,
			Emoji:  statusEmoji(status),
			Items:  filtered[status],
		}
	}
	m.clampActiveCard()
}

// SelectedItem returns a pointer to the currently focused ProjectItem, or nil.
func (m *Model) SelectedItem() *github.ProjectItem {
	if len(m.columns) == 0 {
		return nil
	}
	if m.activeCol < 0 || m.activeCol >= len(m.columns) {
		return nil
	}
	col := &m.columns[m.activeCol]
	if m.activeCard < 0 || m.activeCard >= len(col.Items) {
		return nil
	}
	return &col.Items[m.activeCard]
}

// clampActiveCard ensures activeCard is within bounds for the current column.
func (m *Model) clampActiveCard() {
	if len(m.columns) == 0 {
		m.activeCard = 0
		return
	}
	col := m.columns[m.activeCol]
	if m.activeCard >= len(col.Items) {
		m.activeCard = len(col.Items) - 1
	}
	if m.activeCard < 0 {
		m.activeCard = 0
	}
}

// ensureVisible adjusts the active column's scroll offset so the active card is visible.
func (m *Model) ensureVisible() {
	if len(m.columns) == 0 {
		return
	}
	col := &m.columns[m.activeCol]
	contentHeight := m.ctx.ContentHeight()
	contentWidth := m.ctx.ContentWidth()
	colWidth := contentWidth / len(m.columns)
	if colWidth < 24 {
		colWidth = 24
	}
	if colWidth > 50 {
		colWidth = 50
	}
	col.EnsureActiveVisible(m.activeCard, contentHeight, colWidth)
}

// filterItems returns a copy of items containing only entries matching query.
func filterItems(items map[string][]github.ProjectItem, query string) map[string][]github.ProjectItem {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	out := make(map[string][]github.ProjectItem, len(items))
	for status, list := range items {
		var matched []github.ProjectItem
		for _, item := range list {
			if matchesQuery(item, q) {
				matched = append(matched, item)
			}
		}
		if len(matched) > 0 {
			out[status] = matched
		}
	}
	return out
}

// matchesQuery checks whether a ProjectItem matches a lowercase query string.
func matchesQuery(item github.ProjectItem, q string) bool {
	if strings.Contains(strings.ToLower(item.Title), q) {
		return true
	}
	if strings.Contains(fmt.Sprintf("%d", item.Number), q) {
		return true
	}
	if strings.Contains(strings.ToLower(item.Repository), q) {
		return true
	}
	for _, a := range item.Assignees {
		if strings.Contains(strings.ToLower(a), q) {
			return true
		}
	}
	for _, l := range item.Labels {
		if strings.Contains(strings.ToLower(l), q) {
			return true
		}
	}
	return false
}

// statusRank returns the sort order for a given status string.
func statusRank(status string) int {
	lower := strings.ToLower(status)
	for i, s := range statusOrder {
		if s == lower {
			return i
		}
	}
	return len(statusOrder)
}

// sortedStatuses returns status keys sorted by the canonical order.
func sortedStatuses(groups map[string][]github.ProjectItem) []string {
	statuses := make([]string, 0, len(groups))
	for status := range groups {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool {
		ri, rj := statusRank(statuses[i]), statusRank(statuses[j])
		if ri != rj {
			return ri < rj
		}
		return statuses[i] < statuses[j]
	})
	return statuses
}
