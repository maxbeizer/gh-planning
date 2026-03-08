package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var statusOpts struct {
	Project   int
	Owner     string
	Assignee  string
	Stale     string
	Exclude   []string
	Board     bool
	Swimlanes bool
	Open      bool
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status grouped by status field",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().IntVar(&statusOpts.Project, "project", 0, "Project number")
	statusCmd.Flags().StringVar(&statusOpts.Owner, "owner", "", "Project owner")
	statusCmd.Flags().StringVar(&statusOpts.Assignee, "assignee", "", "Filter by assignee")
	statusCmd.Flags().StringVar(&statusOpts.Stale, "stale", "", "Only show items stale for this duration")
	statusCmd.Flags().StringSliceVar(&statusOpts.Exclude, "exclude", nil, "Exclude statuses (e.g. --exclude Done,Closed)")
	statusCmd.Flags().BoolVar(&statusOpts.Board, "board", false, "Show kanban board view")
	statusCmd.Flags().BoolVar(&statusOpts.Swimlanes, "swimlanes", false, "Add assignee swimlanes to board view (implies --board)")
	statusCmd.Flags().BoolVar(&statusOpts.Open, "open", false, "Open the project board in your browser")
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := statusOpts.Owner
	project := statusOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}
	staleDuration, err := parseDuration(statusOpts.Stale)
	if err != nil {
		return fmt.Errorf("invalid stale duration: %w", err)
	}
	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}
	filtered := filterProjectItems(projectData, statusOpts.Assignee, staleDuration, statusOpts.Exclude)

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"title":  projectData.Title,
			"owner":  owner,
			"number": project,
			"items":  filtered,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Printf("📊 Project: %s (#%d)\n", projectData.Title, project)
	fmt.Printf("   %s\n\n", projectURL(owner, project))

	if statusOpts.Open {
		_ = openURL(projectURL(owner, project))
	}

	if statusOpts.Board || statusOpts.Swimlanes {
		if statusOpts.Swimlanes {
			printSwimlaneBoardView(filtered)
		} else {
			printBoardView(filtered)
		}
	} else {
		printStatusGroups(filtered, staleDuration)
	}
	return nil
}

func filterProjectItems(project *github.Project, assignee string, stale time.Duration, exclude []string) map[string][]github.ProjectItem {
	excludeSet := map[string]bool{}
	for _, e := range exclude {
		excludeSet[strings.ToLower(strings.TrimSpace(e))] = true
	}
	filtered := map[string][]github.ProjectItem{}
	for status, items := range project.Items {
		if excludeSet[strings.ToLower(status)] {
			continue
		}
		for _, item := range items {
			if assignee != "" {
				found := false
				for _, a := range item.Assignees {
					if strings.EqualFold(a, assignee) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if stale > 0 && time.Since(item.UpdatedAt) < stale {
				continue
			}
			filtered[status] = append(filtered[status], item)
		}
	}
	return filtered
}

func printStatusGroups(groups map[string][]github.ProjectItem, stale time.Duration) {
	statuses := make([]string, 0, len(groups))
	for status := range groups {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		items := groups[status]
		if len(items) == 0 {
			continue
		}
		header := fmt.Sprintf("%s (%d)", status, len(items))
		fmt.Printf("%s\n", decorateStatus(status, header))
		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		for _, item := range items {
			assignee := "—"
			if len(item.Assignees) > 0 {
				assignee = "@" + item.Assignees[0]
			}
			staleMark := ""
			if stale > 0 && time.Since(item.UpdatedAt) >= stale {
				staleMark = " ⚠️ stale"
			}
			issueNum := fmt.Sprintf("#%d", item.Number)
			if item.URL != "" {
				issueNum = hyperlink(item.URL, issueNum)
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s%s\n", issueNum, truncate(item.Title, 28), item.Repository, assignee, humanizeDuration(time.Since(item.UpdatedAt)), staleMark)
		}
		w.Flush()
		fmt.Println()
	}
}

func decorateStatus(status, header string) string {
	switch strings.ToLower(status) {
	case "in progress":
		return "🔵 " + header
	case "backlog":
		return "📋 " + header
	case "done", "closed", "complete", "completed":
		return "✅ " + header
	case "needs my attention", "needs review", "in review":
		return "🔍 " + header
	default:
		return "• " + header
	}
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max-1] + "…"
}

func buildStatusSummary(ctx context.Context, owner string, project int) (string, error) {
	projectData, err := github.GetProject(ctx, owner, project)
	if err != nil {
		return "", err
	}
	counts := []string{}
	statuses := make([]string, 0, len(projectData.Items))
	for status := range projectData.Items {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		counts = append(counts, fmt.Sprintf("%s: %d", status, len(projectData.Items[status])))
	}
	return strings.Join(counts, " | "), nil
}
