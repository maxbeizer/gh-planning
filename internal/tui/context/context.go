package context

import (
	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/tui/theme"
)

// ProgramContext carries shared state through all TUI components.
// Passed by pointer so updates (e.g., terminal resize) propagate everywhere.
type ProgramContext struct {
	// Terminal dimensions, updated on WindowSizeMsg.
	Width  int
	Height int

	// Visual theme — initialized once at startup.
	Theme *theme.Theme

	// User config for the active profile.
	Config *config.Config

	// Project identification.
	Owner         string
	ProjectNumber int
	ProfileName   string
}

// New creates a ProgramContext with sensible defaults.
func New(cfg *config.Config, owner string, projectNumber int) *ProgramContext {
	return &ProgramContext{
		Width:         80,
		Height:        24,
		Theme:         theme.New(),
		Config:        cfg,
		Owner:         owner,
		ProjectNumber: projectNumber,
	}
}

// ContentHeight returns the usable height between the tab bar and status bar.
// Reserves 3 lines: 1 tab bar + 1 status bar + 1 help hints.
func (ctx *ProgramContext) ContentHeight() int {
	h := ctx.Height - 3
	if h < 1 {
		return 1
	}
	return h
}

// ContentWidth returns the usable width inside borders/padding.
// Reserves 2 characters for left/right border.
func (ctx *ProgramContext) ContentWidth() int {
	w := ctx.Width - 2
	if w < 20 {
		return 20
	}
	return w
}
