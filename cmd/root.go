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

	// Resolve active profile name (explicit or auto-detected)
	profileName := ""
	explicitProfile, _ := config.ActiveProfileName()
	detected := false
	if explicitProfile != "" {
		profileName = explicitProfile
	} else if matches, _ := config.DetectProfile(); len(matches) == 1 {
		profileName = matches[0].Name
		detected = true
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
		if profileName != "" {
			payload["profile"] = profileName
		}
		if cfg.DefaultOwner != "" && cfg.DefaultProject != 0 {
			payload["project"] = map[string]interface{}{
				"owner":  cfg.DefaultOwner,
				"number": cfg.DefaultProject,
			}
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload, rootOpts)
	}

	w := cmd.OutOrStdout()
	fmt.Fprintln(w, "gh-planning — developer command center")
	fmt.Fprintln(w)

	// Show active profile and project
	if profileName != "" {
		if detected && profileName != explicitProfile {
			fmt.Fprintf(w, "📂 Profile: %s (auto-detected)\n", profileName)
		} else {
			fmt.Fprintf(w, "📂 Profile: %s\n", profileName)
		}
	}
	if cfg.DefaultOwner != "" && cfg.DefaultProject != 0 {
		fmt.Fprintf(w, "📋 Project: %s #%d\n", cfg.DefaultOwner, cfg.DefaultProject)
	}

	if focus != nil {
		fmt.Fprintf(w, "🎯 Focus: %s (%s)\n", focus.Issue, humanizeDuration(focus.Elapsed()))
	} else {
		fmt.Fprintln(w, "🎯 Focus: none")
	}
	if statusSummary != "" {
		fmt.Fprintf(w, "📊 %s\n", statusSummary)
	} else if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(w, "📊 No project configured")
	}

	fmt.Fprintln(w)
	if focus != nil {
		fmt.Fprintln(w, "  gh planning log \"message\"    — log progress")
		fmt.Fprintln(w, "  gh planning unfocus          — clear focus")
	} else {
		fmt.Fprintln(w, "  gh planning status           — view your project board")
		fmt.Fprintln(w, "  gh planning focus <issue>    — start focusing on an issue")
	}
	fmt.Fprintln(w, "  gh planning standup          — generate a standup report")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run `gh planning --help` for all commands.")
	if cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(w, "Run `gh planning setup` to get started.")
		fmt.Fprintln(w, "Run `gh planning tutorial` for an interactive walkthrough.")
	} else {
		fmt.Fprintln(w, "Run `gh planning cheatsheet` to browse commands by scenario.")
	}
	return nil
}
