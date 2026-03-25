package detail

import (
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
	"github.com/maxbeizer/gh-planning/internal/tui/theme"
)

// Model is the detail pane for viewing a single project item.
type Model struct {
	ctx        *context.ProgramContext
	keys       keys.NavigationKeyMap
	globalKeys keys.GlobalKeyMap
	item       *github.ProjectItem
	viewport   viewport.Model
	visible    bool
	ready      bool
}

// New creates a detail pane model wired to the given ProgramContext.
func New(ctx *context.ProgramContext) Model {
	vp := viewport.New()
	return Model{
		ctx:        ctx,
		keys:       keys.NewNavigationKeyMap(),
		globalKeys: keys.NewGlobalKeyMap(),
		viewport:   vp,
	}
}

// SetItem sets the item to display and rebuilds rendered content.
func (m *Model) SetItem(item *github.ProjectItem) {
	m.item = item
	m.rebuildContent()
}

// SetVisible shows or hides the detail pane.
func (m *Model) SetVisible(v bool) {
	m.visible = v
}

// IsVisible returns whether the detail pane is currently shown.
func (m Model) IsVisible() bool {
	return m.visible
}

// SetSize updates viewport dimensions for the overlay.
func (m *Model) SetSize(width, height int) {
	// Leave room for the overlay border and padding (2 border + 2*3 padding = 8 horizontal, 2 border + 2*1 padding = 4 vertical)
	innerW := width - 8
	innerH := height - 4
	if innerW < 20 {
		innerW = 20
	}
	if innerH < 5 {
		innerH = 5
	}
	m.viewport.SetWidth(innerW)
	m.viewport.SetHeight(innerH)
	m.ready = true
	if m.item != nil {
		m.rebuildContent()
	}
}

// Update handles key events when the detail pane is visible.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.globalKeys.ClearFilter):
			m.visible = false
			return m, nil
		case key.Matches(msg, m.globalKeys.OpenBrowser):
			if m.item != nil && m.item.URL != "" {
				cmd := exec.Command("open", m.item.URL)
				_ = cmd.Start()
			}
			return m, nil
		}
	}

	// Delegate remaining keys (j/k scroll) to viewport.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the detail pane as a bordered overlay.
func (m Model) View() string {
	if !m.visible || m.item == nil {
		return ""
	}

	th := m.ctx.Theme
	overlayW := m.viewport.Width() + 8 // add back border + padding

	// Title border
	title := fmt.Sprintf(" #%d  %s ", m.item.Number, m.item.Title)
	style := th.Overlay.
		Width(overlayW).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true)

	content := m.viewport.View()

	return style.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			th.Title.Render(title),
			"",
			content,
		),
	)
}

func (m *Model) rebuildContent() {
	if m.item == nil {
		return
	}
	th := m.ctx.Theme
	item := m.item
	label := th.DetailLabel
	var lines []string

	// State + Status line
	stateEmoji := stateIndicator(item.State)
	statusEmoji := statusIndicator(item.Status)
	stStyle := stateStyle(th, item.State)
	suStyle := statusStyle(th, item.Status)
	lines = append(lines,
		label.Render("State")+" "+stateEmoji+" "+stStyle.Render(item.State)+"    "+
			label.Render("Status")+" "+statusEmoji+" "+suStyle.Render(item.Status))

	lines = append(lines, "")

	// Repo
	lines = append(lines,
		label.Render("Repo")+th.Muted.Render(item.Repository))

	// Assignees
	assignees := "—"
	if len(item.Assignees) > 0 {
		var prefixed []string
		for _, a := range item.Assignees {
			prefixed = append(prefixed, "@"+a)
		}
		assignees = strings.Join(prefixed, ", ")
	}
	lines = append(lines,
		label.Render("Assignees")+assignees)

	// Labels
	labels := "—"
	if len(item.Labels) > 0 {
		labels = strings.Join(item.Labels, ", ")
	}
	lines = append(lines,
		label.Render("Labels")+labels)

	// Updated
	lines = append(lines,
		label.Render("Updated")+humanizeTime(item.UpdatedAt))

	// Type
	lines = append(lines,
		label.Render("Type")+item.ContentType)

	// Separator
	lines = append(lines, "")
	lines = append(lines, th.Dimmed.Render(strings.Repeat("─", 40)))
	lines = append(lines, "")

	// Body placeholder
	lines = append(lines,
		th.Muted.Render("Body not available in project view."))
	lines = append(lines,
		th.Muted.Render("Press o to open in browser for full details."))

	// Separator + help hints (muted)
	lines = append(lines, "")
	lines = append(lines, th.Dimmed.Render(strings.Repeat("─", 40)))
	lines = append(lines,
		th.Dimmed.Render("o open · j/k scroll · esc back"))

	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func stateIndicator(state string) string {
	switch strings.ToUpper(state) {
	case "OPEN":
		return "🟢"
	case "CLOSED":
		return "🔴"
	case "MERGED":
		return "🟣"
	default:
		return "⚪"
	}
}

func statusIndicator(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "done"):
		return "✅"
	case strings.Contains(lower, "progress"):
		return "🔵"
	case strings.Contains(lower, "review"):
		return "🟡"
	case strings.Contains(lower, "block"):
		return "🔴"
	default:
		return "⚪"
	}
}

func stateStyle(th *theme.Theme, state string) lipgloss.Style {
	switch strings.ToUpper(state) {
	case "OPEN":
		return th.Success
	case "CLOSED":
		return th.Danger
	case "MERGED":
		return th.Heading
	default:
		return th.Muted
	}
}

func statusStyle(th *theme.Theme, status string) lipgloss.Style {
	lower := strings.ToLower(status)
	switch {
	case strings.Contains(lower, "done"):
		return th.Success
	case strings.Contains(lower, "progress"):
		return th.Warning
	case strings.Contains(lower, "block"):
		return th.Danger
	default:
		return th.Muted
	}
}

func humanizeTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		days := int(math.Round(d.Hours() / 24))
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}
