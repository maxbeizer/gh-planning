package cmd

import (
	"context"
	"fmt"
	"os"

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
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(unfocusCmd)
	rootCmd.AddCommand(standupCmd)
	rootCmd.AddCommand(catchupCmd)
	rootCmd.AddCommand(breakdownCmd)
	rootCmd.AddCommand(handoffCmd)
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
		return output.PrintJSON(payload, rootOpts)
	}

	if focus != nil {
		fmt.Printf("🎯 Focus: %s (%s)\n", focus.Issue, humanizeDuration(focus.Elapsed()))
	} else {
		fmt.Println("🎯 Focus: none")
	}
	if statusSummary != "" {
		fmt.Printf("📊 %s\n", statusSummary)
	} else if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(os.Stderr, "Run `gh planning init` to set a default project.")
	}
	return nil
}
