package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/tui"
)

func (r *tutorialRunner) runHandsOnTour() error {
	steps := []struct {
		id         string
		title      string
		explain    string
		needsIssue bool
		run        func() error
	}{
		{
			id:      "handson-dashboard",
			title:   "Your Command Center",
			explain: "First, let's see your dashboard — your home base for everything.",
			run:     r.stepDashboard,
		},
		{
			id:      "handson-board",
			title:   "Your Project Board",
			explain: "Now let's see your kanban board so you know what's already there.",
			run:     r.stepBoard,
		},
		{
			id:      "handson-track",
			title:   "Create a Tutorial Issue",
			explain: "Let's create a real issue and add it to your project board.\n  This issue is just for learning — we'll clean it up at the end.",
			run:     r.stepTrackIssue,
		},
		{
			id:         "handson-focus",
			title:      "Focus On It",
			needsIssue: true,
			explain:    "Focus mode tracks your time and makes other commands context-aware.\n  Let's focus on your new issue.",
			run:        r.stepFocusIssue,
		},
		{
			id:         "handson-log",
			title:      "Log Progress",
			needsIssue: true,
			explain:    "While focused, you can log progress, decisions, and blockers —\n  creating a breadcrumb trail of your work.",
			run:        r.stepLogProgress,
		},
		{
			id:      "handson-standup",
			title:   "Generate a Standup",
			explain: "Now let's generate a standup report — it pulls from your actual\n  GitHub activity.",
			run:     r.stepStandup,
		},
		{
			id:         "handson-cleanup",
			title:      "Clean Up",
			needsIssue: true,
			explain:    "Let's close the tutorial issue to keep your board tidy.",
			run:        r.stepCleanup,
		},
	}

	r.printBanner("Hands-On Tutorial", "Walk through the full lifecycle with a real tutorial issue.")

	// Resume state from a previous incomplete run
	if r.progress.HandsOnIssue != "" {
		fmt.Println(tui.Muted.Render("  Resuming with issue: ") + tui.Command.Render(r.progress.HandsOnIssue))
		r.createdIssueRef = r.progress.HandsOnIssue
		r.createdRepo = r.progress.HandsOnRepo
		r.createdIssueNum = r.progress.HandsOnIssueNum
		fmt.Println()
	}

	for i, step := range steps {
		if step.needsIssue && r.createdIssueRef == "" {
			continue
		}

		if r.progress.IsCompleted(step.id) {
			fmt.Printf("  %s %s %s\n",
				tui.Success.Render("✓"),
				tui.Subtitle.Render(step.title),
				tui.Muted.Render("(done — enter to re-run, 's' to skip)"))
			choice := r.prompt("  ")
			if strings.TrimSpace(strings.ToLower(choice)) == "s" {
				continue
			}
		}

		r.printStepHeader(i+1, len(steps), step.title)
		fmt.Println("  " + step.explain)
		fmt.Println()

		if err := step.run(); err != nil {
			fmt.Println(tui.Warning.Render("  ⚠ " + err.Error()))
			fmt.Println()
		}

		r.progress.MarkCompleted(step.id)
		_ = r.progress.Save()

		if i < len(steps)-1 {
			r.promptContinue()
		}
	}

	fmt.Println()
	fmt.Println(tui.Success.Render("🎉 Hands-on tutorial complete!"))
	fmt.Println()
	fmt.Println("  Here's what you learned:")
	fmt.Println("    " + tui.Command.Render("gh planning") + "           — your dashboard")
	fmt.Println("    " + tui.Command.Render("gh planning board") + "     — kanban view")
	fmt.Println("    " + tui.Command.Render("gh planning track") + "     — create & track issues")
	fmt.Println("    " + tui.Command.Render("gh planning focus") + "     — start a focus session")
	fmt.Println("    " + tui.Command.Render("gh planning log") + "       — log progress")
	fmt.Println("    " + tui.Command.Render("gh planning standup") + "   — generate reports")
	fmt.Println()
	fmt.Println("  Next: " + tui.Command.Render("gh planning cheatsheet") + " to explore all commands,")
	fmt.Println("  or " + tui.Command.Render("gh planning guide <workflow>") + " for specific workflows.")
	return nil
}
