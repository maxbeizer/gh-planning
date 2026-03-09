package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListItem represents a single item in a filterable list.
type ListItem struct {
	Title       string
	Description string
	Category    string
	Command     string
	Example     string
	DocURL      string
	Detail      string // extra info shown when selected
}

// ListModel is a bubbletea model for browsing and filtering a categorized list.
type ListModel struct {
	Items    []ListItem
	Cursor   int
	Filter   string
	Filtered []int // indices into Items matching the current filter
	Width    int
	Height   int
	Quit     bool
	Expanded bool // whether to show detail of selected item
}

// NewListModel creates a ListModel from a slice of items.
func NewListModel(items []ListItem) ListModel {
	m := ListModel{
		Items:  items,
		Width:  80,
		Height: 24,
	}
	m.applyFilter()
	return m
}

func (m *ListModel) applyFilter() {
	m.Filtered = nil
	filter := strings.ToLower(m.Filter)
	for i, item := range m.Items {
		if filter == "" ||
			strings.Contains(strings.ToLower(item.Title), filter) ||
			strings.Contains(strings.ToLower(item.Description), filter) ||
			strings.Contains(strings.ToLower(item.Category), filter) ||
			strings.Contains(strings.ToLower(item.Command), filter) {
			m.Filtered = append(m.Filtered, i)
		}
	}
	if m.Cursor >= len(m.Filtered) {
		m.Cursor = max(0, len(m.Filtered)-1)
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.Quit = true
			return m, tea.Quit
		case "q":
			if m.Filter == "" {
				m.Quit = true
				return m, tea.Quit
			}
			m.Filter += "q"
			m.applyFilter()
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				m.Expanded = false
			}
		case "down", "j":
			if m.Cursor < len(m.Filtered)-1 {
				m.Cursor++
				m.Expanded = false
			}
		case "enter":
			m.Expanded = !m.Expanded
		case "backspace":
			if len(m.Filter) > 0 {
				m.Filter = m.Filter[:len(m.Filter)-1]
				m.applyFilter()
			}
		default:
			if len(msg.String()) == 1 && msg.String() != " " {
				m.Filter += msg.String()
				m.applyFilter()
			} else if msg.String() == " " {
				m.Filter += " "
				m.applyFilter()
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m ListModel) View() string {
	var b strings.Builder

	// Filter bar
	filterStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDim).
		Padding(0, 1).
		Width(min(m.Width-4, 60))

	filterText := m.Filter
	if filterText == "" {
		filterText = Dimmed.Render("type to filter...")
	}
	b.WriteString(filterStyle.Render("🔍 " + filterText))
	b.WriteString("\n\n")

	if len(m.Filtered) == 0 {
		b.WriteString(Muted.Render("  No matches found."))
		b.WriteString("\n")
	} else {
		// Determine visible window
		maxVisible := max(1, m.Height-10)
		start := 0
		if m.Cursor >= maxVisible {
			start = m.Cursor - maxVisible + 1
		}
		end := min(start+maxVisible, len(m.Filtered))

		lastCategory := ""
		for vi := start; vi < end; vi++ {
			idx := m.Filtered[vi]
			item := m.Items[idx]

			// Category header
			if item.Category != lastCategory {
				lastCategory = item.Category
				b.WriteString("\n")
				b.WriteString(Heading.Render(item.Category))
				b.WriteString("\n")
			}

			// Item line
			cursor := "  "
			style := lipgloss.NewStyle()
			if vi == m.Cursor {
				cursor = Command.Render("▸ ")
				style = style.Bold(true).Foreground(ColorBright)
			}

			cmdStr := ""
			if item.Command != "" {
				cmdStr = Muted.Render(" — ") + Command.Render(item.Command)
			}
			b.WriteString(cursor + style.Render(item.Title) + cmdStr + "\n")

			// Expanded detail for selected item
			if vi == m.Cursor && m.Expanded {
				detail := m.renderDetail(item)
				b.WriteString(detail)
			}
		}

		if len(m.Filtered) > maxVisible {
			b.WriteString(fmt.Sprintf("\n%s", Dimmed.Render(fmt.Sprintf("  showing %d–%d of %d", start+1, end, len(m.Filtered)))))
		}
	}

	// Help bar
	b.WriteString("\n\n")
	b.WriteString(HelpBar.Render("↑↓ navigate • enter expand • type to filter • esc quit"))

	return b.String()
}

func (m ListModel) renderDetail(item ListItem) string {
	var b strings.Builder
	pad := "    "

	if item.Description != "" {
		b.WriteString(pad + Muted.Render(item.Description) + "\n")
	}
	if item.Example != "" {
		exBox := lipgloss.NewStyle().
			MarginLeft(4).
			Foreground(ColorBright).
			Background(lipgloss.Color("#161B22")).
			Padding(0, 1).
			Render("$ " + item.Example)
		b.WriteString(exBox + "\n")
	}
	if item.Detail != "" {
		b.WriteString(pad + item.Detail + "\n")
	}
	if item.DocURL != "" {
		b.WriteString(pad + Muted.Render("📖 ") + lipgloss.NewStyle().Foreground(ColorPrimary).Underline(true).Render(item.DocURL) + "\n")
	}
	b.WriteString("\n")
	return b.String()
}

// RenderListPlain renders the full list as plain text.
func RenderListPlain(items []ListItem) string {
	var b strings.Builder
	lastCat := ""
	for _, item := range items {
		if item.Category != lastCat {
			lastCat = item.Category
			b.WriteString(fmt.Sprintf("\n%s\n%s\n", item.Category, strings.Repeat("─", len(item.Category))))
		}
		b.WriteString(fmt.Sprintf("  %-30s %s\n", item.Command, item.Title))
		if item.Description != "" {
			b.WriteString(fmt.Sprintf("  %-30s %s\n", "", item.Description))
		}
		if item.Example != "" {
			b.WriteString(fmt.Sprintf("  %-30s $ %s\n", "", item.Example))
		}
		if item.DocURL != "" {
			b.WriteString(fmt.Sprintf("  %-30s Docs: %s\n", "", item.DocURL))
		}
		b.WriteString("\n")
	}
	return b.String()
}
