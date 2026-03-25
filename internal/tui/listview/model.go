package listview

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
)

// statusOrder defines preferred section ordering (matches board).
var statusOrder = []string{
	"backlog", "ready", "todo",
	"in progress", "in review", "needs review", "needs my attention",
	"done", "closed", "complete", "completed",
}

func statusRank(status string) int {
	lower := strings.ToLower(status)
	for i, s := range statusOrder {
		if s == lower {
			return i
		}
	}
	return len(statusOrder)
}

func statusEmoji(status string) string {
	switch strings.ToLower(status) {
	case "in progress":
		return "🔵"
	case "backlog", "ready", "todo":
		return "📋"
	case "done", "closed", "complete", "completed":
		return "✅"
	case "in review", "needs review", "needs my attention":
		return "🔍"
	case "blocked":
		return "🚫"
	default:
		return "•"
	}
}

// Section is a group of items under a status heading.
type Section struct {
	Status    string
	Emoji     string
	Items     []github.ProjectItem
	Collapsed bool
}

// Model is the Bubble Tea model for the list view tab.
type Model struct {
	ctx      *context.ProgramContext
	keys     keys.NavigationKeyMap
	sections []Section
	cursor   int // flat index across all visible rows (headers + items)
	offset   int // scroll offset for viewport
	ready    bool
}

// New creates a list view Model wired to the given ProgramContext.
func New(ctx *context.ProgramContext) Model {
	return Model{
		ctx:  ctx,
		keys: keys.NewNavigationKeyMap(),
	}
}

// SetItems populates sections from project data, sorted by status order.
func (m *Model) SetItems(items map[string][]github.ProjectItem) {
	// Collect unique statuses.
	type statusGroup struct {
		name  string
		items []github.ProjectItem
	}
	var groups []statusGroup
	for status, list := range items {
		if len(list) == 0 {
			continue
		}
		groups = append(groups, statusGroup{name: status, items: list})
	}
	sort.Slice(groups, func(i, j int) bool {
		return statusRank(groups[i].name) < statusRank(groups[j].name)
	})

	m.sections = make([]Section, len(groups))
	for i, g := range groups {
		m.sections[i] = Section{
			Status: g.name,
			Emoji:  statusEmoji(g.name),
			Items:  g.items,
		}
	}
	m.cursor = 0
	m.offset = 0
	m.ready = true
}

// rowEntry represents a single renderable row in the flat list.
type rowEntry struct {
	isHeader     bool
	sectionIdx   int
	itemIdx      int // only valid when !isHeader
	item         *github.ProjectItem
	section      *Section
	flatIndex    int
	visibleIndex int // position among visible rows for alt-row striping
}

// flatRows builds the list of visible rows (headers + non-collapsed items).
func (m *Model) flatRows() []rowEntry {
	var rows []rowEntry
	idx := 0
	visIdx := 0
	for si := range m.sections {
		sec := &m.sections[si]
		rows = append(rows, rowEntry{
			isHeader:     true,
			sectionIdx:   si,
			section:      sec,
			flatIndex:    idx,
			visibleIndex: visIdx,
		})
		idx++
		visIdx++
		if !sec.Collapsed {
			for ii := range sec.Items {
				rows = append(rows, rowEntry{
					isHeader:     false,
					sectionIdx:   si,
					itemIdx:      ii,
					item:         &sec.Items[ii],
					section:      sec,
					flatIndex:    idx,
					visibleIndex: visIdx,
				})
				idx++
				visIdx++
			}
		}
	}
	return rows
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages for list navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	rows := m.flatRows()
	if len(rows) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(rows)-1 {
				m.cursor++
			}
			m.ensureVisible()

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureVisible()

		case key.Matches(msg, m.keys.Select):
			if m.cursor < len(rows) && rows[m.cursor].isHeader {
				m.sections[rows[m.cursor].sectionIdx].Collapsed =
					!m.sections[rows[m.cursor].sectionIdx].Collapsed
				// Clamp cursor if it now points past the last row.
				newRows := m.flatRows()
				if m.cursor >= len(newRows) {
					m.cursor = len(newRows) - 1
				}
				m.ensureVisible()
			}
		}
	}

	return m, nil
}

// ensureVisible adjusts offset so the cursor stays in the viewport.
func (m *Model) ensureVisible() {
	visHeight := m.ctx.ContentHeight()
	if visHeight < 1 {
		visHeight = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visHeight {
		m.offset = m.cursor - visHeight + 1
	}
}

// View renders the list view within the available content area.
func (m Model) View() string {
	if !m.ready || len(m.sections) == 0 {
		return m.ctx.Theme.Muted.Render("No items to display")
	}

	rows := m.flatRows()
	visHeight := m.ctx.ContentHeight()

	var b strings.Builder

	// Scroll-up indicator.
	if m.offset > 0 {
		b.WriteString(m.ctx.Theme.Dimmed.Render("  ↑ more"))
		b.WriteString("\n")
		visHeight-- // consumed a line
	}

	// Reserve a line for the down indicator if needed.
	hasMore := m.offset+visHeight < len(rows)
	if hasMore {
		visHeight--
	}

	end := m.offset + visHeight
	if end > len(rows) {
		end = len(rows)
	}

	for i := m.offset; i < end; i++ {
		r := rows[i]
		active := i == m.cursor
		if r.isHeader {
			b.WriteString(renderSectionHeader(m.ctx, r.section, active))
		} else {
			b.WriteString(renderItemRow(m.ctx, r, active))
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	if hasMore {
		b.WriteString("\n")
		b.WriteString(m.ctx.Theme.Dimmed.Render("  ↓ more"))
	}

	return b.String()
}
