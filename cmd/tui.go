package cmd

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/tui/app"
	tuictx "github.com/maxbeizer/gh-planning/internal/tui/context"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func newTUICmd() *cobra.Command {
	var (
		project     int
		owner       string
		assignee    string
		exclude     []string
		includeDone bool
		stale       string
	)

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive planning dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Fprintln(cmd.OutOrStdout(),
					"TUI requires an interactive terminal. Use 'gh planning board' or 'gh planning status' instead.")
				return nil
			}

			pc, err := resolveProjectConfig(owner, project)
			if err != nil {
				return err
			}

			profileName, _ := config.ActiveProfileName()

			ctx := tuictx.New(pc.Cfg, pc.Owner, pc.Project)
			ctx.ProfileName = profileName

			model := app.New(ctx)
			p := tea.NewProgram(model)
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().IntVar(&project, "project", 0, "Project number")
	cmd.Flags().StringVar(&owner, "owner", "", "Project owner")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Filter by assignee")
	cmd.Flags().StringSliceVar(&exclude, "exclude", nil, "Exclude statuses (e.g. --exclude Done,Closed)")
	cmd.Flags().BoolVar(&includeDone, "include-done", false, "Include Done/Completed/Closed statuses")
	cmd.Flags().StringVar(&stale, "stale", "", "Only show items stale for this duration")

	// Mark flags as used to avoid lint warnings while they're not yet
	// wired to the TUI filtering logic.
	_ = assignee
	_ = exclude
	_ = includeDone
	_ = stale

	return cmd
}
