package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/maxbeizer/gh-planning/internal/tutorial"
	"github.com/spf13/cobra"
)

var (
	tutorialReset   bool
	tutorialList    bool
	tutorialHandsOn bool
	tutorialExplore bool
)

var tutorialCmd = &cobra.Command{
	Use:   "tutorial",
	Short: "Interactive tutorial that teaches gh-planning by doing",
	Long: `An end-to-end walkthrough that teaches gh-planning with real commands.

Modes:
  --hands-on    Full lifecycle: create an issue → claim → focus → log → complete
  --explore     Read-only tour: view your board, status, standup (no writes)

If neither is specified, you'll be prompted to choose.

Progress is saved automatically — resume where you left off.
Requires a configured project (run 'gh planning setup' first).

Use --reset to start over, or --list to see your progress.`,
	Aliases: []string{"learn", "teach"},
	RunE:    runTutorial,
}

func init() {
	tutorialCmd.Flags().BoolVar(&tutorialReset, "reset", false, "Reset tutorial progress")
	tutorialCmd.Flags().BoolVar(&tutorialList, "list", false, "Show lesson list and progress")
	tutorialCmd.Flags().BoolVar(&tutorialHandsOn, "hands-on", false, "Full lifecycle walkthrough (creates a real tutorial issue)")
	tutorialCmd.Flags().BoolVar(&tutorialExplore, "explore", false, "Read-only tour of your project (no writes)")
}

// tutorialRunner holds shared state for the tutorial walkthrough.
type tutorialRunner struct {
	ctx      context.Context
	cfg      *config.Config
	reader   *bufio.Reader
	progress *tutorial.Progress
	user     string

	// state accumulated during hands-on mode
	createdIssueRef string
	createdIssueNum int
	createdRepo     string
}

func runTutorial(cmd *cobra.Command, args []string) error {
	progress, err := tutorial.Load()
	if err != nil {
		return fmt.Errorf("loading tutorial progress: %w", err)
	}

	if tutorialReset {
		progress.Reset()
		if err := progress.Save(); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), tui.Success.Render("✓ Tutorial progress reset."))
		return nil
	}

	if tutorialList {
		return listTutorialProgress(cmd, progress)
	}

	// Check that a project is configured
	cfg, err := config.Load()
	if err != nil || cfg.DefaultOwner == "" || cfg.DefaultProject == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), tui.Warning.Render("⚠ No project configured yet."))
		fmt.Fprintln(cmd.OutOrStdout(), "Run "+tui.Command.Render("gh planning setup")+" first, then come back!")
		return nil
	}

	user, err := github.CurrentUser(cmd.Context())
	if err != nil {
		return fmt.Errorf("detecting GitHub user: %w", err)
	}

	r := &tutorialRunner{
		ctx:      cmd.Context(),
		cfg:      cfg,
		reader:   bufio.NewReader(os.Stdin),
		progress: progress,
		user:     user,
	}

	handsOn := tutorialHandsOn
	explore := tutorialExplore

	if !handsOn && !explore {
		fmt.Println()
		fmt.Println(tui.Title.Render("Welcome to the gh-planning tutorial!"))
		fmt.Println()
		fmt.Println("  " + tui.Command.Render("1") + "  " + tui.Subtitle.Render("Hands-on") + tui.Muted.Render(" — full lifecycle with a real tutorial issue (recommended)"))
		fmt.Println("  " + tui.Command.Render("2") + "  " + tui.Subtitle.Render("Explore") + tui.Muted.Render(" — read-only tour of your existing project data"))
		fmt.Println()
		choice := r.prompt("Choose a mode [1/2]: ")
		switch strings.TrimSpace(choice) {
		case "2", "explore":
			explore = true
		default:
			handsOn = true
		}
	}

	fmt.Println()

	if explore {
		return r.runExploreTour()
	}
	return r.runHandsOnTour()
}
