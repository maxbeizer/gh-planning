package cmd

import (
	"fmt"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var boardOpts struct {
	Project     int
	Owner       string
	Assignee    string
	Stale       string
	Exclude     []string
	Swimlanes   bool
	IncludeDone bool
	Open        bool
}

// Default statuses to exclude when --include-done is not set.
var defaultDoneStatuses = []string{"Done", "Completed", "Closed"}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Show kanban board view of your project",
	RunE:  runBoard,
}

func init() {
	boardCmd.Flags().IntVar(&boardOpts.Project, "project", 0, "Project number")
	boardCmd.Flags().StringVar(&boardOpts.Owner, "owner", "", "Project owner")
	boardCmd.Flags().StringVar(&boardOpts.Assignee, "assignee", "", "Filter by assignee")
	boardCmd.Flags().StringVar(&boardOpts.Stale, "stale", "", "Only show items stale for this duration")
	boardCmd.Flags().StringSliceVar(&boardOpts.Exclude, "exclude", nil, "Exclude statuses (e.g. --exclude Done,Closed)")
	boardCmd.Flags().BoolVar(&boardOpts.Swimlanes, "swimlanes", false, "Add assignee swimlanes to board view")
	boardCmd.Flags().BoolVar(&boardOpts.IncludeDone, "include-done", false, "Include Done/Completed/Closed statuses (excluded by default)")
	boardCmd.Flags().BoolVar(&boardOpts.Open, "open", false, "Open the project board in your browser")
}

func runBoard(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := boardOpts.Owner
	project := boardOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}
	staleDuration, err := parseDuration(boardOpts.Stale)
	if err != nil {
		return fmt.Errorf("invalid stale duration: %w", err)
	}
	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}

	exclude := boardOpts.Exclude
	if !boardOpts.IncludeDone && !cmd.Flags().Changed("exclude") {
		exclude = defaultDoneStatuses
	}

	filtered := filterProjectItems(projectData, boardOpts.Assignee, staleDuration, exclude)

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

	if boardOpts.Open {
		_ = openURL(projectURL(owner, project))
	}

	if boardOpts.Swimlanes {
		printSwimlaneBoardView(filtered)
	} else {
		printBoardView(filtered)
	}
	return nil
}
