package cmd

import (
	"context"
	"fmt"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/spf13/cobra"
)

var rootOpts output.Options

var rootCmd = &cobra.Command{
	Use:   "planning",
	Short: "The developer's command center for GitHub-native project management",
	RunE:  runRoot,
}

func ExecuteContext(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

func OutputOptions() output.Options {
	return rootOpts
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&rootOpts.JSON, "json", false, "Output JSON")
	rootCmd.PersistentFlags().StringVar(&rootOpts.JQ, "jq", "", "Filter JSON output with jq")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(boardCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(unfocusCmd)
	rootCmd.AddCommand(standupCmd)
	rootCmd.AddCommand(catchupCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(prepCmd)
	rootCmd.AddCommand(pulseCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(copilotCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(blockedCmd)
	rootCmd.AddCommand(unblockCmd)
	rootCmd.AddCommand(cheatsheetCmd)
	rootCmd.AddCommand(guideCmd)
	rootCmd.AddCommand(tutorialCmd)
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	focus, err := session.LoadCurrent()
	if err != nil {
		return err
	}
	statusSummary := ""
	if cfg.DefaultOwner != "" && cfg.DefaultProject != 0 {
		summary, err := buildStatusSummary(cmd.Context(), cfg.DefaultOwner, cfg.DefaultProject)
		if err == nil {
			statusSummary = summary
		}
	}

	if rootOpts.JSON || rootOpts.JQ != "" {
		payload := map[string]interface{}{
			"focus":         focus,
			"statusSummary": statusSummary,
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload, rootOpts)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "gh-planning — developer command center")
	fmt.Fprintln(cmd.OutOrStdout())

	if focus != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "🎯 Focus: %s (%s)\n", focus.Issue, humanizeDuration(focus.Elapsed()))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "🎯 Focus: none")
	}
	if statusSummary != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "📊 %s\n", statusSummary)
	} else if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "📊 No project configured")
	}

	fmt.Fprintln(cmd.OutOrStdout())
	if focus != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "  gh planning log \"message\"    — log progress")
		fmt.Fprintln(cmd.OutOrStdout(), "  gh planning unfocus          — clear focus")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "  gh planning status           — view your project board")
		fmt.Fprintln(cmd.OutOrStdout(), "  gh planning focus <issue>    — start focusing on an issue")
	}
	fmt.Fprintln(cmd.OutOrStdout(), "  gh planning standup          — generate a standup report")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Run `gh planning --help` for all commands.")
	if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "Run `gh planning setup` to get started.")
		fmt.Fprintln(cmd.OutOrStdout(), "Run `gh planning tutorial` for an interactive walkthrough.")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "Run `gh planning cheatsheet` to browse commands by scenario.")
	}
	return nil
}
