package app

import (
	stdctx "context"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/board"
	"github.com/maxbeizer/gh-planning/internal/tui/components/footer"
	"github.com/maxbeizer/gh-planning/internal/tui/components/help"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/detail"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
	"github.com/maxbeizer/gh-planning/internal/tui/listview"
)

// projectDataMsg carries the result of a background project data fetch.
type projectDataMsg struct {
	project *github.Project
	err     error
}

// fetchProjectData returns a tea.Cmd that fetches project data in the background.
func fetchProjectData(owner string, number int) tea.Cmd {
	return func() tea.Msg {
		proj, err := github.GetProject(stdctx.Background(), owner, number)
		return projectDataMsg{project: proj, err: err}
	}
}

// fetchProjectDataFresh returns a tea.Cmd that clears the cache before fetching.
func fetchProjectDataFresh(owner string, number int) tea.Cmd {
	return func() tea.Msg {
		// Remove cache file to force a fresh API call.
		if path, err := github.ProjectCachePath(owner, number); err == nil {
			os.Remove(path)
		}
		proj, err := github.GetProject(stdctx.Background(), owner, number)
		return projectDataMsg{project: proj, err: err}
	}
}

// Model is the root Bubble Tea model for the TUI dashboard.
type Model struct {
	ctx       *tuictx.ProgramContext
	keys      keys.GlobalKeyMap
	activeTab int
	tabs      []string
	ready     bool
	loading   bool
	err       error
	footer    footer.Model
	help      help.Model
	detail    detail.Model
	listview  listview.Model
	board     board.Model
}

// New creates a root Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{
		ctx:    ctx,
		keys:   keys.NewGlobalKeyMap(),
		tabs:   []string{"Board", "List"},
		footer:   footer.New(ctx),
		help:     help.New(ctx),
		detail:   detail.New(ctx),
		listview: listview.New(ctx),
		board:    board.New(ctx),
	}
}

func (m Model) Init() tea.Cmd {
	m.loading = true
	return fetchProjectData(m.ctx.Owner, m.ctx.ProjectNumber)
}
