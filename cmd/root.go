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
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(boardCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(unfocusCmd)
	rootCmd.AddCommand(standupCmd)
	rootCmd.AddCommand(catchupCmd)
	rootCmd.AddCommand(breakdownCmd)
	rootCmd.AddCommand(handoffCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(prepCmd)
	rootCmd.AddCommand(pulseCmd)
	rootCmd.AddCommand(agentContextCmd)
	rootCmd.AddCommand(claimCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(queueCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(copilotCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(logsCmd)
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

	fmt.Println("gh-planning — developer command center")
	fmt.Println()

	if focus != nil {
		fmt.Printf("🎯 Focus: %s (%s)\n", focus.Issue, humanizeDuration(focus.Elapsed()))
	} else {
		fmt.Println("🎯 Focus: none")
	}
	if statusSummary != "" {
		fmt.Printf("📊 %s\n", statusSummary)
	} else if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Println("📊 No project configured")
	}

	fmt.Println()
	if focus != nil {
		fmt.Println("  gh planning log \"message\"    — log progress")
		fmt.Println("  gh planning handoff <issue>  — hand off to next session")
		fmt.Println("  gh planning complete <issue> — mark done")
	} else {
		fmt.Println("  gh planning status           — view your project board")
		fmt.Println("  gh planning focus <issue>    — start focusing on an issue")
		fmt.Println("  gh planning queue            — find work to pick up")
	}
	fmt.Println("  gh planning standup          — generate a standup report")
	fmt.Println()
	fmt.Println("Run `gh planning --help` for all commands.")
	if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Println("Run `gh planning setup` to get started.")
	}
	return nil
}
