package app

import (
	stdctx "context"
	"fmt"
	"time"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/board"
	"github.com/maxbeizer/gh-planning/internal/tui/components/footer"
	"github.com/maxbeizer/gh-planning/internal/tui/components/help"
	"github.com/maxbeizer/gh-planning/internal/tui/components/picker"
	"github.com/maxbeizer/gh-planning/internal/tui/components/search"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/maxbeizer/gh-planning/internal/tui/detail"
	"github.com/maxbeizer/gh-planning/internal/tui/depgraph"
	"github.com/maxbeizer/gh-planning/internal/tui/keys"
	"github.com/maxbeizer/gh-planning/internal/tui/listview"
	"github.com/maxbeizer/gh-planning/internal/tui/tree"
)

// projectDataMsg carries the result of a background project data fetch.
type projectDataMsg struct {
	project *github.Project
	err     error
}

// fetchProjectData returns a tea.Cmd that fetches project data in the background.
func fetchProjectData(owner string, number int) tea.Cmd {
	return func() tea.Msg {
		ctx := stdctx.Background()
		proj, err := github.GetProject(ctx, owner, number)
		if err == nil && proj != nil {
			// Best-effort: enrich with dependency data.
			_ = github.EnrichWithDependencies(ctx, proj)
		}
		return projectDataMsg{project: proj, err: err}
	}
}

// fetchProjectDataFresh returns a tea.Cmd that clears the cache before fetching.
func fetchProjectDataFresh(owner string, number int) tea.Cmd {
	return func() tea.Msg {
		ctx := stdctx.Background()
		// Remove cache file to force a fresh API call.
		if path, err := github.ProjectCachePath(owner, number); err == nil {
			os.Remove(path)
		}
		proj, err := github.GetProject(ctx, owner, number)
		if err == nil && proj != nil {
			_ = github.EnrichWithDependencies(ctx, proj)
		}
		return projectDataMsg{project: proj, err: err}
	}
}

// statusUpdatedMsg carries the result of a status mutation.
type statusUpdatedMsg struct {
	itemID    string
	newStatus string
	err       error
}

// statusInfoMsg carries project metadata needed for the status picker.
type statusInfoMsg struct {
	projectID     string
	statusFieldID string
	statusOptions map[string]string
	item          github.ProjectItem
	err           error
}

// fetchStatusInfo fetches project field metadata then opens the picker.
func fetchStatusInfo(owner string, number int, item github.ProjectItem) tea.Cmd {
	return func() tea.Msg {
		projectID, _, fieldID, opts, err := github.GetProjectInfo(stdctx.Background(), owner, number)
		return statusInfoMsg{
			projectID:     projectID,
			statusFieldID: fieldID,
			statusOptions: opts,
			item:          item,
			err:           err,
		}
	}
}

// updateItemStatus fires a GraphQL mutation to change the item's status.
func updateItemStatus(owner string, number int, projectID, itemID, fieldID, optionID, statusName string) tea.Cmd {
	return func() tea.Msg {
		err := github.UpdateItemStatus(stdctx.Background(), projectID, itemID, fieldID, optionID)
		return statusUpdatedMsg{
			itemID:    itemID,
			newStatus: statusName,
			err:       err,
		}
	}
}

// Model is the root Bubble Tea model for the TUI dashboard.
type Model struct {
	ctx        *tuictx.ProgramContext
	keys       keys.GlobalKeyMap
	actionKeys keys.ActionKeyMap
	activeTab  int
	tabs      []string
	ready     bool
	loading   bool
	err       error
	footer    footer.Model
	help      help.Model
	detail    detail.Model
	search    search.Model
	listview  listview.Model
	tree      tree.Model
	board     board.Model
	depgraph  depgraph.Model
	picker        picker.Model
	projectPicker picker.Model
	prompt        prompt.Model

	// Cached status mutation metadata.
	statusProjectID string
	statusFieldID   string
	statusOptions   map[string]string // name → option ID
	pendingItem     github.ProjectItem
	mutating            bool
	autoRefreshInterval time.Duration
}

// New creates a root Model wired to the given ProgramContext.
func New(ctx *tuictx.ProgramContext) Model {
	return Model{
		ctx:           ctx,
		keys:          keys.NewGlobalKeyMap(),
		tabs:          []string{"Board", "List", "Tree", "Deps"},
		footer:        footer.New(ctx),
		help:          help.New(ctx),
		detail:        detail.New(ctx),
		search:        search.New(ctx),
		listview:      listview.New(ctx),
		tree:          tree.New(ctx),
		board:         board.New(ctx),
		depgraph:      depgraph.New(ctx),
		picker:        picker.New(ctx, nil, nil),
		projectPicker: newProjectPicker(ctx),
	}
}

// projectSwitchedMsg is emitted when the user selects a different project.
type projectSwitchedMsg struct {
	owner  string
	number int
}

// newProjectPicker builds a picker pre-loaded with configured projects.
func newProjectPicker(ctx *tuictx.ProgramContext) picker.Model {
	opts := buildProjectOptions(ctx.Config, ctx.Owner, ctx.ProjectNumber)
	p := picker.New(ctx, opts, func(opt picker.Option) tea.Cmd {
		owner, number := parseProjectOptionID(opt.ID)
		return func() tea.Msg {
			return projectSwitchedMsg{owner: owner, number: number}
		}
	})
	return p
}

// buildProjectOptions returns picker options from the config's Projects list
// plus the current default project (de-duplicated).
func buildProjectOptions(cfg *config.Config, currentOwner string, currentNumber int) []picker.Option {
	type projKey struct {
		owner  string
		number int
	}
	seen := map[projKey]bool{}
	var opts []picker.Option

	// Always include the current/default project first.
	if currentOwner != "" && currentNumber > 0 {
		k := projKey{currentOwner, currentNumber}
		seen[k] = true
		label := fmt.Sprintf("%s/#%d (current)", currentOwner, currentNumber)
		// Check if there's a label in config Projects.
		if cfg != nil {
			for _, p := range cfg.Projects {
				if p.Owner == currentOwner && p.Number == currentNumber && p.Label != "" {
					label = p.Label + " (current)"
					break
				}
			}
		}
		opts = append(opts, picker.Option{
			Name: label,
			ID:   fmt.Sprintf("%s/%d", currentOwner, currentNumber),
		})
	}

	// Add explicitly configured projects.
	if cfg != nil {
		for _, p := range cfg.Projects {
			k := projKey{p.Owner, p.Number}
			if seen[k] {
				continue
			}
			seen[k] = true
			label := p.Label
			if label == "" {
				label = fmt.Sprintf("%s/#%d", p.Owner, p.Number)
			}
			opts = append(opts, picker.Option{
				Name: label,
				ID:   fmt.Sprintf("%s/%d", p.Owner, p.Number),
			})
		}

		// Include default project if different from current.
		if cfg.DefaultOwner != "" && cfg.DefaultProject > 0 {
			k := projKey{cfg.DefaultOwner, cfg.DefaultProject}
			if !seen[k] {
				opts = append(opts, picker.Option{
					Name: fmt.Sprintf("%s/#%d", cfg.DefaultOwner, cfg.DefaultProject),
					ID:   fmt.Sprintf("%s/%d", cfg.DefaultOwner, cfg.DefaultProject),
				})
			}
		}
	}

	return opts
}

// parseProjectOptionID splits "owner/number" back into components.
func parseProjectOptionID(id string) (string, int) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return id, 0
	}
	n, _ := strconv.Atoi(parts[1])
	return parts[0], n
}

func (m Model) Init() tea.Cmd {
	m.loading = true
	return fetchProjectData(m.ctx.Owner, m.ctx.ProjectNumber)
}
