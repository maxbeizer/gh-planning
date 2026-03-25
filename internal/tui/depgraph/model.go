package depgraph

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/mattn/go-runewidth"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
)

// GraphNode represents a single item positioned within the dependency graph.
type GraphNode struct {
	Item      *github.ProjectItem
	Level     int   // depth in dependency chain (0 = no blockers)
	BlockedBy []int // indices into Model.nodes slice
	Blocks    []int // indices into Model.nodes slice
	Circular  bool  // true when the node participates in a cycle
	Critical  bool  // true when the node is on the critical path
}

// Model is the Bubble Tea model for the dependency graph tab.
type Model struct {
	ctx          *context.ProgramContext
	keys         keys.NavigationKeyMap
	nodes        []GraphNode
	cursor       int
	offset       int
	ready        bool
	criticalLen  int // length of the critical path (number of items)
	hasDeps      bool // true when any item has dependency data
}

// New creates a depgraph Model wired to the given ProgramContext.
func New(ctx *context.ProgramContext) Model {
	return Model{
		ctx:  ctx,
		keys: keys.NewNavigationKeyMap(),
	}
}

// SetItems builds the dependency graph from project data.
func (m *Model) SetItems(items map[string][]github.ProjectItem) {
	// Flatten all items across statuses.
	var all []github.ProjectItem
	for _, list := range items {
		all = append(all, list...)
	}

	// Check if any item has dependency data at all.
	m.hasDeps = false
	for i := range all {
		if len(all[i].BlockedBy) > 0 || len(all[i].Blocks) > 0 {
			m.hasDeps = true
			break
		}
	}

	// Build index: issue number → index in all slice.
	numToIdx := make(map[int]int, len(all))
	for i := range all {
		if all[i].Number > 0 {
			numToIdx[all[i].Number] = i
		}
	}

	// Build adjacency: for each item, record which other items block it / it blocks.
	adjacency := make([]adjEntry, len(all))
	for i := range all {
		for _, dep := range all[i].BlockedBy {
			if j, ok := numToIdx[dep.Number]; ok {
				adjacency[i].blockedBy = append(adjacency[i].blockedBy, j)
				adjacency[j].blocks = append(adjacency[j].blocks, i)
			}
		}
	}

	// Compute levels via BFS (longest path from any root).
	levels := make([]int, len(all))
	inCycle := make([]bool, len(all))
	computeLevels(all, adjacency, levels, inCycle)

	// Build graph nodes.
	m.nodes = make([]GraphNode, len(all))
	for i := range all {
		m.nodes[i] = GraphNode{
			Item:      &all[i],
			Level:     levels[i],
			BlockedBy: adjacency[i].blockedBy,
			Blocks:    adjacency[i].blocks,
			Circular:  inCycle[i],
		}
	}

	// Sort nodes: by level ascending, then by issue number.
	sort.SliceStable(m.nodes, func(i, j int) bool {
		if m.nodes[i].Level != m.nodes[j].Level {
			return m.nodes[i].Level < m.nodes[j].Level
		}
		return m.nodes[i].Item.Number < m.nodes[j].Item.Number
	})

	// After sorting, rebuild index for critical path calculation.
	nodeNumToIdx := make(map[int]int, len(m.nodes))
	for i := range m.nodes {
		if m.nodes[i].Item.Number > 0 {
			nodeNumToIdx[m.nodes[i].Item.Number] = i
		}
	}

	// Compute critical path: longest chain of open blocking dependencies.
	m.criticalLen = m.markCriticalPath(nodeNumToIdx)

	m.cursor = 0
	m.offset = 0
	m.ready = true
}

type adjEntry struct {
	blockedBy []int
	blocks    []int
}

// computeLevels assigns dependency depth via iterative BFS and detects cycles.
func computeLevels(all []github.ProjectItem, adjacency []adjEntry, levels []int, inCycle []bool) {
	n := len(all)
	if n == 0 {
		return
	}

	// Kahn's algorithm variant to assign levels and detect cycles.
	inDegree := make([]int, n)
	for i := range adjacency {
		inDegree[i] = len(adjacency[i].blockedBy)
	}

	// Queue starts with nodes that have no blockers.
	var queue []int
	for i := range inDegree {
		if inDegree[i] == 0 {
			levels[i] = 0
			queue = append(queue, i)
		}
	}

	processed := 0
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		processed++

		for _, next := range adjacency[cur].blocks {
			// Level = max(level, blocker's level + 1) for longest path.
			if levels[cur]+1 > levels[next] {
				levels[next] = levels[cur] + 1
			}
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	// Any unprocessed nodes are in a cycle.
	if processed < n {
		for i := range inDegree {
			if inDegree[i] > 0 {
				inCycle[i] = true
				levels[i] = -1 // sentinel: will be grouped specially
			}
		}
	}
}

// markCriticalPath identifies the longest chain of open blockers and marks those
// nodes. Returns the chain length.
func (m *Model) markCriticalPath(nodeNumToIdx map[int]int) int {
	// For each node, compute the longest chain of open dependencies leading to it.
	// "Open" means the blocking item's state is not closed/merged.
	n := len(m.nodes)
	if n == 0 {
		return 0
	}

	chainLen := make([]int, n)
	chainParent := make([]int, n)
	for i := range chainParent {
		chainParent[i] = -1
	}

	maxLen := 0
	maxIdx := -1

	// Process in level order (already sorted). Skip circular nodes.
	for i := range m.nodes {
		node := &m.nodes[i]
		if node.Circular {
			continue
		}
		for _, dep := range node.Item.BlockedBy {
			if isOpen(dep.State) {
				if j, ok := nodeNumToIdx[dep.Number]; ok {
					candidate := chainLen[j] + 1
					if candidate > chainLen[i] {
						chainLen[i] = candidate
						chainParent[i] = j
					}
				}
			}
		}
		if chainLen[i] > maxLen {
			maxLen = chainLen[i]
			maxIdx = i
		}
	}

	// Walk back and mark the critical path.
	if maxIdx >= 0 && maxLen > 0 {
		visited := make(map[int]bool, maxLen+1)
		idx := maxIdx
		for idx >= 0 && !visited[idx] {
			visited[idx] = true
			m.nodes[idx].Critical = true
			idx = chainParent[idx]
		}
		return maxLen + 1 // chain length includes the starting node
	}

	return 0
}

func isOpen(state string) bool {
	s := strings.ToLower(state)
	return s == "open" || s == ""
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key messages for graph navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.ready {
		return m, nil
	}

	rows := m.visibleRows()
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
			// Enter could open detail — return a cmd if wired up.
		}
	}

	return m, nil
}

// SelectedItem returns the currently highlighted item, if any.
func (m *Model) SelectedItem() *github.ProjectItem {
	rows := m.visibleRows()
	if m.cursor >= 0 && m.cursor < len(rows) {
		r := rows[m.cursor]
		if !r.isHeader {
			return r.node.Item
		}
	}
	return nil
}

// ensureVisible adjusts offset so cursor stays in the viewport.
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

// row represents a renderable row: either a level header or a graph node.
type row struct {
	isHeader bool
	level    int
	label    string // header label
	node     *GraphNode
}

// visibleRows builds the flat list of renderable rows grouped by level.
func (m *Model) visibleRows() []row {
	if len(m.nodes) == 0 {
		return nil
	}

	var rows []row
	currentLevel := -999

	for i := range m.nodes {
		n := &m.nodes[i]
		if n.Level != currentLevel {
			currentLevel = n.Level
			rows = append(rows, row{
				isHeader: true,
				level:    currentLevel,
				label:    levelLabel(currentLevel),
			})
		}
		rows = append(rows, row{
			isHeader: false,
			level:    currentLevel,
			node:     n,
		})
	}

	return rows
}

func levelLabel(level int) string {
	switch {
	case level < 0:
		return "Circular Dependencies"
	case level == 0:
		return "Unblocked"
	default:
		return fmt.Sprintf("Blocked (level %d)", level)
	}
}

// View renders the dependency graph as grouped ASCII.
func (m Model) View() string {
	th := m.ctx.Theme

	if !m.ready || len(m.nodes) == 0 {
		if !m.hasDeps {
			return th.Muted.Render("  No dependency data available.\n  Dependencies are fetched from GitHub's tracking relationships.")
		}
		return th.Muted.Render("  No items to display")
	}

	rows := m.visibleRows()
	visHeight := m.ctx.ContentHeight()
	contentWidth := m.ctx.ContentWidth()

	var b strings.Builder

	// Critical path status line.
	if m.criticalLen > 0 {
		b.WriteString(th.Warning.Render(fmt.Sprintf("  ⚡ Critical path: %d items deep", m.criticalLen)))
		b.WriteString("\n")
		visHeight--
	}

	// Scroll-up indicator.
	if m.offset > 0 {
		b.WriteString(th.Dimmed.Render("  ↑ more"))
		b.WriteString("\n")
		visHeight--
	}

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
			b.WriteString(m.renderLevelHeader(r, contentWidth))
		} else {
			b.WriteString(m.renderNodeRow(r, active))
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	if hasMore {
		b.WriteString("\n")
		b.WriteString(th.Dimmed.Render("  ↓ more"))
	}

	return b.String()
}

// renderLevelHeader renders a section header for a dependency level.
func (m *Model) renderLevelHeader(r row, width int) string {
	th := m.ctx.Theme

	emoji := "──"
	switch {
	case r.level < 0:
		emoji = "🔄"
	case r.level == 0:
		emoji = "✅"
	default:
		emoji = "🚫"
	}

	label := fmt.Sprintf(" %s %s ", emoji, r.label)
	labelWidth := runewidth.StringWidth(label)
	remaining := width - labelWidth - 2
	if remaining < 0 {
		remaining = 0
	}
	line := strings.Repeat("─", remaining)

	return th.SectionHeader.Render(label + line)
}

// renderNodeRow renders a single dependency graph item.
func (m *Model) renderNodeRow(r row, active bool) string {
	th := m.ctx.Theme
	node := r.node
	item := node.Item

	// Build the line.
	var line strings.Builder

	// Critical path indicator.
	if node.Critical {
		line.WriteString("⚡")
	} else {
		line.WriteString("  ")
	}

	// Issue number.
	numStr := fmt.Sprintf("#%-5d", item.Number)

	// Content type emoji.
	typeEmoji := "📋"
	switch strings.ToLower(item.ContentType) {
	case "pullrequest":
		typeEmoji = "🔀"
	}

	// Title (truncated to fit).
	maxTitleWidth := m.ctx.ContentWidth() - 40 // leave room for metadata
	if maxTitleWidth < 10 {
		maxTitleWidth = 10
	}
	title := item.Title
	if runewidth.StringWidth(title) > maxTitleWidth {
		title = runewidth.Truncate(title, maxTitleWidth-1, "…")
	}

	// Dependency info.
	depInfo := m.depInfoString(node)

	line.WriteString(fmt.Sprintf("%s %s %s", numStr, typeEmoji, title))
	if depInfo != "" {
		line.WriteString("  ")
		line.WriteString(depInfo)
	}

	text := line.String()

	// Apply styling.
	if active {
		return th.RowActive.Render(text)
	}
	if node.Circular {
		return th.Danger.Render("  " + text)
	}
	if node.Level > 0 && item.IsBlocked() {
		return th.Row.Render(text)
	}
	return th.Row.Render(text)
}

// depInfoString builds the inline dependency annotation for a node.
func (m *Model) depInfoString(node *GraphNode) string {
	th := m.ctx.Theme
	item := node.Item

	if node.Circular {
		return th.Danger.Render("🔄 circular dependency")
	}

	if len(item.BlockedBy) == 0 {
		// Show state badge for unblocked items.
		return m.stateBadge(item)
	}

	// Show "blocked by" references with colored state.
	var parts []string
	for _, dep := range item.BlockedBy {
		ref := fmt.Sprintf("#%d", dep.Number)
		if isOpen(dep.State) {
			ref = th.Danger.Render("🚫 " + ref)
		} else {
			ref = th.Success.Render("✅ " + ref)
		}
		parts = append(parts, ref)
	}

	return "blocked by " + strings.Join(parts, ", ")
}

// stateBadge returns a colored state indicator for an item.
func (m *Model) stateBadge(item *github.ProjectItem) string {
	th := m.ctx.Theme
	status := strings.ToLower(item.Status)

	switch {
	case status == "done" || status == "closed" || status == "complete" || status == "completed":
		return th.Success.Render("✅ " + item.Status)
	case status == "in progress":
		return th.Warning.Render("🔵 " + item.Status)
	case status == "in review" || status == "needs review":
		return th.Warning.Render("🔍 " + item.Status)
	default:
		return th.Muted.Render(item.Status)
	}
}
