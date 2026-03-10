package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DashboardTab identifies a tab in the dashboard.
type DashboardTab int

const (
	TabSprint DashboardTab = iota
	TabBoard
	TabBlockers
	TabFocus
)

var tabNames = []string{"Sprint", "Board", "Blockers", "Focus"}

// DashboardData holds all data needed to render the dashboard panels.
type DashboardData struct {
	// Project info
	ProjectTitle  string
	ProjectNumber int
	Owner         string

	// Sprint panel
	SprintItems     map[string][]DashboardItem
	IterationTitle  string
	IterationEnd    time.Time
	HasIteration    bool

	// Board panel (all items minus done)
	BoardItems map[string][]DashboardItem

	// Blockers panel
	CriticalPath  []string
	BlockedItems  []BlockerEntry
	ImpactRanking []ImpactEntry

	// Focus panel
	FocusIssue   string
	FocusElapsed time.Duration
	FocusActive  bool
	RecentLogs   []LogLine
}

// DashboardItem is a simplified item for dashboard rendering.
type DashboardItem struct {
	Number   int
	Title    string
	Status   string
	Assignee string
	URL      string
	Updated  time.Time
	Labels   []string
}

// BlockerEntry represents a blocked relationship.
type BlockerEntry struct {
	Blocked   string
	BlockedBy string
}

// ImpactEntry shows unblocking impact.
type ImpactEntry struct {
	Issue   string
	Unblocks int
}

// LogLine represents a progress log entry.
type LogLine struct {
	Time    time.Time
	Message string
	Kind    string
}

// DashboardModel is the main bubbletea model for the planning dashboard.
type DashboardModel struct {
	Data      DashboardData
	ActiveTab DashboardTab
	Width     int
	Height    int
	ScrollY   int // scroll offset within active panel
	Quit      bool
}

// NewDashboardModel creates a new dashboard model with data.
func NewDashboardModel(data DashboardData) DashboardModel {
	return DashboardModel{
		Data:   data,
		Width:  120,
		Height: 40,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.Quit = true
			return m, tea.Quit
		case "1":
			m.ActiveTab = TabSprint
			m.ScrollY = 0
		case "2":
			m.ActiveTab = TabBoard
			m.ScrollY = 0
		case "3":
			m.ActiveTab = TabBlockers
			m.ScrollY = 0
		case "4":
			m.ActiveTab = TabFocus
			m.ScrollY = 0
		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % 4
			m.ScrollY = 0
		case "shift+tab":
			m.ActiveTab = (m.ActiveTab + 3) % 4
			m.ScrollY = 0
		case "j", "down":
			m.ScrollY++
		case "k", "up":
			if m.ScrollY > 0 {
				m.ScrollY--
			}
		case "g":
			m.ScrollY = 0
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m DashboardModel) View() string {
	var b strings.Builder

	// Title bar
	title := fmt.Sprintf("📊 %s (#%d)", m.Data.ProjectTitle, m.Data.ProjectNumber)
	b.WriteString(Title.Render(title))
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n\n")

	// Active panel content
	content := m.renderActivePanel()

	// Apply scrolling
	lines := strings.Split(content, "\n")
	maxScroll := len(lines) - (m.Height - 8)
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.ScrollY > maxScroll {
		m.ScrollY = maxScroll
	}
	if m.ScrollY < len(lines) {
		lines = lines[m.ScrollY:]
	}

	// Limit visible lines to terminal height minus chrome
	visibleLines := m.Height - 8
	if visibleLines < 5 {
		visibleLines = 5
	}
	if len(lines) > visibleLines {
		lines = lines[:visibleLines]
	}

	b.WriteString(strings.Join(lines, "\n"))

	// Help bar
	b.WriteString("\n\n")
	scrollHint := ""
	if maxScroll > 0 {
		scrollHint = " • j/k scroll • g top"
	}
	b.WriteString(HelpBar.Render("1-4 switch tabs • tab/shift+tab cycle" + scrollHint + " • q quit"))

	return b.String()
}

func (m DashboardModel) renderTabBar() string {
	tabs := make([]string, len(tabNames))
	for i, name := range tabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)

		// Add badge counts
		switch DashboardTab(i) {
		case TabSprint:
			count := countDashItems(m.Data.SprintItems)
			if count > 0 {
				label = fmt.Sprintf(" %d:%s(%d) ", i+1, name, count)
			}
		case TabBoard:
			count := countDashItems(m.Data.BoardItems)
			if count > 0 {
				label = fmt.Sprintf(" %d:%s(%d) ", i+1, name, count)
			}
		case TabBlockers:
			count := len(m.Data.BlockedItems)
			if count > 0 {
				label = fmt.Sprintf(" %d:%s(%d) ", i+1, name, count)
			}
		case TabFocus:
			if m.Data.FocusActive {
				label = fmt.Sprintf(" %d:%s● ", i+1, name)
			}
		}

		if DashboardTab(i) == m.ActiveTab {
			tabs[i] = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBright).
				Background(ColorPrimary).
				Render(label)
		} else {
			tabs[i] = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Render(label)
		}
	}
	return strings.Join(tabs, " ")
}

func (m DashboardModel) renderActivePanel() string {
	switch m.ActiveTab {
	case TabSprint:
		return m.renderSprintPanel()
	case TabBoard:
		return m.renderBoardPanel()
	case TabBlockers:
		return m.renderBlockersPanel()
	case TabFocus:
		return m.renderFocusPanel()
	}
	return ""
}

// --- Sprint Panel ---

func (m DashboardModel) renderSprintPanel() string {
	var b strings.Builder

	if m.Data.HasIteration && m.Data.IterationTitle != "" {
		b.WriteString(Subtitle.Render("🏃 Iteration: " + m.Data.IterationTitle))
		if !m.Data.IterationEnd.IsZero() {
			remaining := time.Until(m.Data.IterationEnd)
			if remaining > 0 {
				b.WriteString(Muted.Render(fmt.Sprintf("  (%d days remaining)", int(remaining.Hours()/24))))
			}
		}
		b.WriteString("\n\n")
	} else {
		b.WriteString(Muted.Render("ℹ️  No iteration field — showing active items as sprint proxy"))
		b.WriteString("\n\n")
	}

	total := countDashItems(m.Data.SprintItems)
	if total == 0 {
		b.WriteString(Muted.Render("  No items in the current sprint."))
		return b.String()
	}

	statuses := sortedKeys(m.Data.SprintItems)
	for _, status := range statuses {
		items := m.Data.SprintItems[status]
		if len(items) == 0 {
			continue
		}
		header := fmt.Sprintf("%s %s (%d)", statusIcon(status), status, len(items))
		b.WriteString(Subtitle.Render(header))
		b.WriteString("\n")
		for _, item := range items {
			b.WriteString(m.renderItemLine(item))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(Muted.Render(fmt.Sprintf("📊 Total: %d items", total)))
	return b.String()
}

// --- Board Panel ---

func (m DashboardModel) renderBoardPanel() string {
	var b strings.Builder

	total := countDashItems(m.Data.BoardItems)
	if total == 0 {
		b.WriteString(Muted.Render("  (no items)"))
		return b.String()
	}

	statuses := sortedStatusKeys(m.Data.BoardItems)
	numCols := len(statuses)
	if numCols == 0 {
		b.WriteString(Muted.Render("  (no items)"))
		return b.String()
	}

	// Calculate column width
	maxColWidth := 40
	colWidth := (m.Width - 2) / numCols
	if colWidth > maxColWidth {
		colWidth = maxColWidth
	}
	if colWidth < 20 {
		colWidth = 20
	}
	innerWidth := colWidth - 2

	// Build column content
	columns := make([][]string, numCols)
	for i, status := range statuses {
		items := m.Data.BoardItems[status]
		header := fmt.Sprintf("%s %s (%d)", statusIcon(status), status, len(items))
		columns[i] = append(columns[i], truncateStr(header, innerWidth))

		maxCards := 12
		for j, item := range items {
			if j >= maxCards {
				columns[i] = append(columns[i], truncateStr(fmt.Sprintf("  +%d more", len(items)-maxCards), innerWidth))
				break
			}
			card := fmt.Sprintf("#%-4d %s", item.Number, item.Title)
			columns[i] = append(columns[i], truncateStr(card, innerWidth))
		}
	}

	// Find max height
	maxRows := 0
	for _, col := range columns {
		if len(col) > maxRows {
			maxRows = len(col)
		}
	}

	// Render columns side by side
	colStyle := lipgloss.NewStyle().Width(colWidth)
	headerStyle := lipgloss.NewStyle().Width(colWidth).Bold(true).Foreground(ColorPrimary)

	for row := 0; row < maxRows; row++ {
		parts := make([]string, numCols)
		for col := range columns {
			if row < len(columns[col]) {
				if row == 0 {
					parts[col] = headerStyle.Render(columns[col][row])
				} else {
					parts[col] = colStyle.Render(columns[col][row])
				}
			} else {
				parts[col] = colStyle.Render("")
			}
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, parts...))
		b.WriteString("\n")
		// Separator after header
		if row == 0 {
			for range numCols {
				b.WriteString(Dimmed.Render(strings.Repeat("─", colWidth)))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// --- Blockers Panel ---

func (m DashboardModel) renderBlockersPanel() string {
	var b strings.Builder

	if len(m.Data.CriticalPath) == 0 && len(m.Data.BlockedItems) == 0 {
		b.WriteString(Success.Render("✅ No blockers tracked"))
		b.WriteString("\n")
		b.WriteString(Muted.Render("Use `gh planning blocked <issue> --by <blocker>` to track dependencies."))
		return b.String()
	}

	// Critical path
	if len(m.Data.CriticalPath) > 0 {
		b.WriteString(Danger.Render(fmt.Sprintf("🔴 Critical Path (%d deep)", len(m.Data.CriticalPath))))
		b.WriteString("\n")
		for i, node := range m.Data.CriticalPath {
			indent := strings.Repeat("    ", i)
			if i == 0 {
				b.WriteString(fmt.Sprintf("  %s%s\n", indent, node))
			} else {
				b.WriteString(fmt.Sprintf("  %s└── blocks %s\n", indent, node))
			}
		}
		b.WriteString("\n")
	}

	// Blocked items
	if len(m.Data.BlockedItems) > 0 {
		b.WriteString(Warning.Render("🟡 Blocked Items"))
		b.WriteString("\n")
		for _, entry := range m.Data.BlockedItems {
			b.WriteString(fmt.Sprintf("  %s ← blocked by %s\n", entry.Blocked, entry.BlockedBy))
		}
		b.WriteString("\n")
	}

	// Impact ranking
	if len(m.Data.ImpactRanking) > 0 {
		b.WriteString(Heading.Render("⚡ Unblocking Impact"))
		b.WriteString("\n")
		for _, entry := range m.Data.ImpactRanking {
			b.WriteString(fmt.Sprintf("  %s → would unblock %d issue(s)\n", entry.Issue, entry.Unblocks))
		}
	}

	return b.String()
}

// --- Focus Panel ---

func (m DashboardModel) renderFocusPanel() string {
	var b strings.Builder

	if !m.Data.FocusActive {
		b.WriteString(Muted.Render("🎯 No active focus session"))
		b.WriteString("\n\n")
		b.WriteString(Muted.Render("Use `gh planning focus <issue>` to start focusing."))
		return b.String()
	}

	// Focus session info
	elapsed := m.Data.FocusElapsed
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	timeStr := fmt.Sprintf("%dh %dm", hours, minutes)
	if hours == 0 {
		timeStr = fmt.Sprintf("%dm", minutes)
	}

	focusBox := ActiveBox.Copy().Width(min(m.Width-4, 60)).Render(
		fmt.Sprintf("🎯 %s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(ColorBright).Render(m.Data.FocusIssue),
			Muted.Render("Elapsed: "+timeStr),
		),
	)
	b.WriteString(focusBox)
	b.WriteString("\n\n")

	// Recent logs
	if len(m.Data.RecentLogs) > 0 {
		b.WriteString(Subtitle.Render("📝 Recent Activity"))
		b.WriteString("\n")
		for _, entry := range m.Data.RecentLogs {
			icon := logKindIcon(entry.Kind)
			age := shortDuration(time.Since(entry.Time))
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				icon,
				entry.Message,
				Dimmed.Render("("+age+")"),
			))
		}
	} else {
		b.WriteString(Muted.Render("No log entries yet. Use `gh planning log \"message\"` to track progress."))
	}

	return b.String()
}

// --- Helpers ---

func (m DashboardModel) renderItemLine(item DashboardItem) string {
	assignee := "—"
	if item.Assignee != "" {
		assignee = "@" + item.Assignee
	}
	age := shortDuration(time.Since(item.Updated))
	return fmt.Sprintf("  #%-4d %-30s %-14s %s",
		item.Number,
		truncateStr(item.Title, 30),
		assignee,
		Dimmed.Render(age),
	)
}

func statusIcon(status string) string {
	switch strings.ToLower(status) {
	case "in progress":
		return "🔵"
	case "backlog", "ready", "todo", "to do":
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

func logKindIcon(kind string) string {
	switch kind {
	case "decision":
		return "🔷"
	case "blocker":
		return "🚫"
	case "hypothesis":
		return "💡"
	case "tried":
		return "🔬"
	case "result":
		return "📊"
	default:
		return "📝"
	}
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	m := int(d.Minutes())
	if m < 60 {
		return fmt.Sprintf("%dm", m)
	}
	h := int(d.Hours())
	if h < 24 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dd", h/24)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

func countDashItems(groups map[string][]DashboardItem) int {
	count := 0
	for _, items := range groups {
		count += len(items)
	}
	return count
}

func sortedKeys(m map[string][]DashboardItem) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// statusOrderDash matches the board column ordering.
var statusOrderDash = []string{
	"backlog", "ready", "todo",
	"in progress", "in review", "needs review", "needs my attention",
	"done", "closed", "complete", "completed",
}

func statusRankDash(status string) int {
	lower := strings.ToLower(status)
	for i, s := range statusOrderDash {
		if s == lower {
			return i
		}
	}
	return len(statusOrderDash)
}

func sortedStatusKeys(m map[string][]DashboardItem) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ri, rj := statusRankDash(keys[i]), statusRankDash(keys[j])
		if ri != rj {
			return ri < rj
		}
		return keys[i] < keys[j]
	})
	return keys
}

// RenderDashboardPlain renders all panels as plain text (for --plain or non-TTY).
func RenderDashboardPlain(data DashboardData) string {
	m := NewDashboardModel(data)
	var b strings.Builder

	b.WriteString(fmt.Sprintf("📊 %s (#%d)\n", data.ProjectTitle, data.ProjectNumber))
	b.WriteString(strings.Repeat("═", 60))
	b.WriteString("\n\n")

	b.WriteString("━━━ Sprint ━━━\n")
	m.ActiveTab = TabSprint
	b.WriteString(m.renderActivePanel())
	b.WriteString("\n\n")

	b.WriteString("━━━ Board ━━━\n")
	m.ActiveTab = TabBoard
	b.WriteString(m.renderActivePanel())
	b.WriteString("\n\n")

	b.WriteString("━━━ Blockers ━━━\n")
	m.ActiveTab = TabBlockers
	b.WriteString(m.renderActivePanel())
	b.WriteString("\n\n")

	b.WriteString("━━━ Focus ━━━\n")
	m.ActiveTab = TabFocus
	b.WriteString(m.renderActivePanel())
	b.WriteString("\n")

	return b.String()
}
