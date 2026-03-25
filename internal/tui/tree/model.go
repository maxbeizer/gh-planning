package tree

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
)

// statusOrder defines preferred ordering (matches listview/board).
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

// Model is the Bubble Tea model for the tree view tab.
type Model struct {
	ctx       *context.ProgramContext
	keys      keys.NavigationKeyMap
	roots     []*Node // top-level items (no parent)
	flatNodes []*Node // flattened visible list for rendering
	cursor    int
	offset    int
	ready     bool
}

// New creates a tree view Model wired to the given ProgramContext.
func New(ctx *context.ProgramContext) Model {
	return Model{
		ctx:  ctx,
		keys: keys.NewNavigationKeyMap(),
	}
}

// SetItems builds the tree from parent/child relationships in project data.
func (m *Model) SetItems(items map[string][]github.ProjectItem) {
	// 1. Collect all items across all statuses into a flat list.
	var all []github.ProjectItem
	for _, list := range items {
		all = append(all, list...)
	}

	// 2. Build a map of issue number → Node.
	nodeByNumber := make(map[int]*Node)
	for i := range all {
		item := &all[i]
		if item.Number > 0 {
			nodeByNumber[item.Number] = &Node{Item: item}
		}
	}

	// 3. For items with ParentIssue, attach as child of parent node.
	// 4. Items with no parent (or parent not in project) become roots.
	var roots []*Node
	for _, node := range nodeByNumber {
		if node.Item.ParentIssue != nil {
			if parent, ok := nodeByNumber[node.Item.ParentIssue.Number]; ok {
				parent.Children = append(parent.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	// Sort roots by status order, then by number.
	sort.Slice(roots, func(i, j int) bool {
		ri, rj := statusRank(roots[i].Item.Status), statusRank(roots[j].Item.Status)
		if ri != rj {
			return ri < rj
		}
		return roots[i].Item.Number < roots[j].Item.Number
	})

	// Sort children by number.
	for _, node := range nodeByNumber {
		if len(node.Children) > 1 {
			sort.Slice(node.Children, func(i, j int) bool {
				return node.Children[i].Item.Number < node.Children[j].Item.Number
			})
		}
	}

	// Set depths.
	var setDepth func(nodes []*Node, depth int)
	setDepth = func(nodes []*Node, depth int) {
		for _, n := range nodes {
			n.Depth = depth
			setDepth(n.Children, depth+1)
		}
	}
	setDepth(roots, 0)

	m.roots = roots
	m.cursor = 0
	m.offset = 0
	m.ready = true
	m.rebuildFlat()
}

// rebuildFlat walks the tree and collects visible (expanded) nodes.
func (m *Model) rebuildFlat() {
	m.flatNodes = m.flatNodes[:0]
	var walk func(nodes []*Node)
	walk = func(nodes []*Node) {
		for _, n := range nodes {
			m.flatNodes = append(m.flatNodes, n)
			if n.Expanded && n.hasChildren() {
				walk(n.Children)
			}
		}
	}
	walk(m.roots)
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages for tree navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.ready || len(m.flatNodes) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.flatNodes)-1 {
				m.cursor++
			}
			m.ensureVisible()

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureVisible()

		case key.Matches(msg, m.keys.Select):
			m.toggleExpand()
		}
	}

	return m, nil
}

// toggleExpand expands or collapses the node at the cursor.
func (m *Model) toggleExpand() {
	if m.cursor >= len(m.flatNodes) {
		return
	}
	node := m.flatNodes[m.cursor]
	if !node.hasChildren() {
		return
	}
	node.Expanded = !node.Expanded
	m.rebuildFlat()
	// Clamp cursor if it now points past the last node.
	if m.cursor >= len(m.flatNodes) {
		m.cursor = len(m.flatNodes) - 1
	}
	m.ensureVisible()
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
