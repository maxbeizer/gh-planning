package app

import (
	tea "charm.land/bubbletea/v2"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
)

// Model is the root Bubble Tea model for the TUI dashboard.
type Model struct {
	ctx       *tuictx.ProgramContext
	keys      keys.GlobalKeyMap
	activeTab int
	tabs      []string
	ready     bool
	err       error
}

// New creates a root Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{
		ctx:  ctx,
		keys: keys.NewGlobalKeyMap(),
		tabs: []string{"Board", "List"},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
