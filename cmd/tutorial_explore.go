package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/tui"
)

func (r *tutorialRunner) runExploreTour() error {
	steps := []struct {
		id      string
		title   string
		explain string
		run     func() error
	}{
		{
			id:      "explore-dashboard",
			title:   "Your Command Center",
			explain: "Let's start with your dashboard — this is what you see when you\nrun gh planning with no arguments.",
			run:     r.stepDashboard,
		},
		{
			id:      "explore-status",
			title:   "Project Status",
			explain: "Now let's look at your project items grouped by status column.",
			run:     r.stepStatus,
		},
		{
			id:      "explore-board",
			title:   "Kanban Board",
			explain: "Same data, different view — a kanban board with columns right in\nyour terminal.",
			run:     r.stepBoard,
		},
		{
			id:      "explore-standup",
			title:   "Standup Report",
			explain: "gh-planning can generate a standup report from your real GitHub\nactivity — commits, PRs, and issues.",
			run:     r.stepStandup,
		},
		{
			id:      "explore-catchup",
			title:   "Catch-Up Summary",
			explain: "If you've been away, catch-up summarizes what happened since your\nlast session.",
			run:     r.stepCatchUp,
		},
	}

	r.printBanner("Explore Tour", "A read-only tour of your project — no changes will be made.")

	for i, step := range steps {
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
	fmt.Println(tui.Success.Render("🎉 Explore tour complete!"))
	fmt.Println()
	fmt.Println("  Ready to go deeper? Try " + tui.Command.Render("gh planning tutorial --hands-on"))
	fmt.Println("  to create a real issue and walk through the full lifecycle:")
	fmt.Println("  track → claim → focus → log → complete.")
	fmt.Println()
	fmt.Println("  Or browse commands with " + tui.Command.Render("gh planning cheatsheet"))
	return nil
}
