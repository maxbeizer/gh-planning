package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/spf13/cobra"
)

var dashboardOpts struct {
	Project int
	Owner   string
	Plain   bool
}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Interactive planning dashboard with sprint, board, blockers, and focus views",
	Long: `Launch an interactive TUI dashboard that combines sprint items, kanban board,
blockers/critical-path, and focus session into a single tabbed view.

Use number keys (1-4) or Tab to switch panels, j/k to scroll, q to quit.
Use --plain for non-interactive output or --json for structured data.`,
	RunE: runDashboard,
}

func init() {
	dashboardCmd.Flags().IntVar(&dashboardOpts.Project, "project", 0, "Project number")
	dashboardCmd.Flags().StringVar(&dashboardOpts.Owner, "owner", "", "Project owner")
	dashboardCmd.Flags().BoolVar(&dashboardOpts.Plain, "plain", false, "Print all panels as plain text (no TUI)")
}

func runDashboard(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(dashboardOpts.Owner, dashboardOpts.Project)
	if err != nil {
		return err
	}

	data, err := loadDashboardData(cmd, pc.Owner, pc.Project)
	if err != nil {
		return err
	}

	// JSON output
	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(dashboardToJSON(data), OutputOptions())
	}

	// Plain text output
	if dashboardOpts.Plain || !isTerminal() {
		fmt.Fprint(cmd.OutOrStdout(), tui.RenderDashboardPlain(data))
		return nil
	}

	// Interactive TUI
	model := tui.NewDashboardModel(data)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func loadDashboardData(cmd *cobra.Command, owner string, project int) (tui.DashboardData, error) {
	data := tui.DashboardData{
		Owner:         owner,
		ProjectNumber: project,
	}

	// Load project items
	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return data, err
	}
	data.ProjectTitle = projectData.Title

	// Build board items (exclude done statuses)
	doneStatuses := map[string]bool{
		"done": true, "closed": true, "complete": true, "completed": true,
	}
	data.BoardItems = convertItems(projectData.Items, doneStatuses)

	// Build sprint items
	data.SprintItems, data.IterationTitle, data.IterationEnd, data.HasIteration = buildSprintData(cmd, owner, project, projectData)

	// Load blockers from state
	data.CriticalPath, data.BlockedItems, data.ImpactRanking = buildBlockerData()

	// Load focus session
	data.FocusIssue, data.FocusElapsed, data.FocusActive, data.RecentLogs = buildFocusData()

	return data, nil
}

func convertItems(groups map[string][]github.ProjectItem, excludeStatuses map[string]bool) map[string][]tui.DashboardItem {
	result := map[string][]tui.DashboardItem{}
	for status, items := range groups {
		if excludeStatuses != nil && excludeStatuses[strings.ToLower(status)] {
			continue
		}
		for _, item := range items {
			assignee := ""
			if len(item.Assignees) > 0 {
				assignee = item.Assignees[0]
			}
			result[status] = append(result[status], tui.DashboardItem{
				Number:   item.Number,
				Title:    item.Title,
				Status:   status,
				Assignee: assignee,
				URL:      item.URL,
				Updated:  item.UpdatedAt,
				Labels:   item.Labels,
			})
		}
	}
	return result
}

func buildSprintData(cmd *cobra.Command, owner string, project int, projectData *github.Project) (map[string][]tui.DashboardItem, string, time.Time, bool) {
	iterInfo, _ := getIterationField(cmd.Context(), owner, project)

	var currentIteration *iterationValue
	hasIteration := false
	if iterInfo != nil {
		for i := range iterInfo.Iterations {
			if iterInfo.Iterations[i].Current {
				currentIteration = &iterInfo.Iterations[i]
				hasIteration = true
				break
			}
		}
	}

	sprintItems := map[string][]github.ProjectItem{}
	if hasIteration && currentIteration != nil {
		for status, items := range projectData.Items {
			for _, item := range items {
				if iterVal, ok := item.Fields["Iteration"]; ok {
					if iterVal == currentIteration.Title {
						sprintItems[status] = append(sprintItems[status], item)
					}
				}
			}
		}
	} else {
		for status, items := range projectData.Items {
			lower := strings.ToLower(status)
			if lower == "in progress" || lower == "in review" || lower == "needs review" || lower == "needs my attention" {
				sprintItems[status] = append(sprintItems[status], items...)
			}
		}
	}

	converted := convertItems(sprintItems, nil)

	iterTitle := ""
	var iterEnd time.Time
	if currentIteration != nil {
		iterTitle = currentIteration.Title
		if currentIteration.StartDate != "" && currentIteration.Duration > 0 {
			start, err := time.Parse("2006-01-02", currentIteration.StartDate)
			if err == nil {
				iterEnd = start.AddDate(0, 0, currentIteration.Duration)
			}
		}
	}

	return converted, iterTitle, iterEnd, hasIteration
}

func buildBlockerData() ([]string, []tui.BlockerEntry, []tui.ImpactEntry) {
	st, err := state.Load()
	if err != nil || len(st.Dependencies) == 0 {
		return nil, nil, nil
	}

	// Build adjacency
	blocks := map[string][]string{}
	blockedBy := map[string]string{}
	allNodes := map[string]struct{}{}

	for _, dep := range st.Dependencies {
		blocks[dep.BlockedBy] = append(blocks[dep.BlockedBy], dep.Blocked)
		blockedBy[dep.Blocked] = dep.BlockedBy
		allNodes[dep.Blocked] = struct{}{}
		allNodes[dep.BlockedBy] = struct{}{}
	}

	// Find roots
	var roots []string
	for node := range allNodes {
		if _, isBlocked := blockedBy[node]; !isBlocked {
			if _, doesBlock := blocks[node]; doesBlock {
				roots = append(roots, node)
			}
		}
	}
	sort.Strings(roots)

	// DFS for critical path
	type chain struct{ path []string }
	var allChains []chain
	var dfs func(node string, current []string, visited map[string]bool)
	dfs = func(node string, current []string, visited map[string]bool) {
		current = append(current, node)
		children := blocks[node]
		if len(children) == 0 {
			ch := make([]string, len(current))
			copy(ch, current)
			allChains = append(allChains, chain{path: ch})
			return
		}
		sort.Strings(children)
		for _, child := range children {
			if visited[child] {
				continue
			}
			visited[child] = true
			dfs(child, current, visited)
			visited[child] = false
		}
	}
	for _, root := range roots {
		visited := map[string]bool{root: true}
		dfs(root, nil, visited)
	}

	var critical []string
	for _, ch := range allChains {
		if len(ch.path) > len(critical) {
			critical = ch.path
		}
	}

	// Build blocked entries
	var blockedEntries []tui.BlockerEntry
	for _, dep := range st.Dependencies {
		blockedEntries = append(blockedEntries, tui.BlockerEntry{
			Blocked:   dep.Blocked,
			BlockedBy: dep.BlockedBy,
		})
	}

	// Impact ranking
	impact := map[string]int{}
	var countBlockedFn func(node string) int
	countBlockedFn = func(node string) int {
		total := 0
		for _, child := range blocks[node] {
			total += 1 + countBlockedFn(child)
		}
		impact[node] = total
		return total
	}
	for _, root := range roots {
		countBlockedFn(root)
	}

	var ranking []tui.ImpactEntry
	for issue, count := range impact {
		if count > 0 {
			ranking = append(ranking, tui.ImpactEntry{Issue: issue, Unblocks: count})
		}
	}
	sort.Slice(ranking, func(i, j int) bool {
		if ranking[i].Unblocks != ranking[j].Unblocks {
			return ranking[i].Unblocks > ranking[j].Unblocks
		}
		return ranking[i].Issue < ranking[j].Issue
	})

	return critical, blockedEntries, ranking
}

func buildFocusData() (string, time.Duration, bool, []tui.LogLine) {
	focus, err := session.LoadCurrent()
	if err != nil || focus == nil {
		return "", 0, false, nil
	}

	// Load recent logs for this issue
	var logLines []tui.LogLine
	logs, err := state.GetLogs(focus.Issue, time.Time{})
	if err == nil {
		// Take last 10 entries
		start := 0
		if len(logs) > 10 {
			start = len(logs) - 10
		}
		for _, entry := range logs[start:] {
			logLines = append(logLines, tui.LogLine{
				Time:    entry.Time,
				Message: entry.Message,
				Kind:    entry.Kind,
			})
		}
	}

	return focus.Issue, focus.Elapsed(), true, logLines
}

func dashboardToJSON(data tui.DashboardData) map[string]interface{} {
	return map[string]interface{}{
		"project": map[string]interface{}{
			"title":  data.ProjectTitle,
			"number": data.ProjectNumber,
			"owner":  data.Owner,
		},
		"sprint": map[string]interface{}{
			"items":        data.SprintItems,
			"iteration":    data.IterationTitle,
			"hasIteration": data.HasIteration,
		},
		"board": map[string]interface{}{
			"items": data.BoardItems,
		},
		"blockers": map[string]interface{}{
			"criticalPath":  data.CriticalPath,
			"blocked":       data.BlockedItems,
			"impactRanking": data.ImpactRanking,
		},
		"focus": map[string]interface{}{
			"issue":   data.FocusIssue,
			"active":  data.FocusActive,
			"elapsed": data.FocusElapsed.String(),
			"logs":    data.RecentLogs,
		},
	}
}
