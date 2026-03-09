package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/tui"
)

// ─── Read-only Steps ───────────────────────────────────────────────────────

func (r *tutorialRunner) stepDashboard() error {
	r.runAndShow("gh", "planning")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ This is your home base. It shows your current focus (if any),"))
	fmt.Println(tui.Muted.Render("    project status counts, and suggested next commands."))
	return nil
}

func (r *tutorialRunner) stepStatus() error {
	r.runAndShow("gh", "planning", "status")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ Items are grouped by their project status column."))
	fmt.Println(tui.Muted.Render("    Use --stale 7d to flag items that haven't moved recently."))
	return nil
}

func (r *tutorialRunner) stepBoard() error {
	r.runAndShow("gh", "planning", "board")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ A kanban board right in your terminal! Done items are hidden by default."))
	fmt.Println(tui.Muted.Render("    Try --swimlanes to see items grouped by assignee."))
	return nil
}

func (r *tutorialRunner) stepStandup() error {
	r.runAndShow("gh", "planning", "standup", "--since", "7d")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ Generated from your real GitHub activity — no manual writing needed."))
	fmt.Println(tui.Muted.Render("    Use --since 24h for a daily standup, --team to include teammates."))
	return nil
}

func (r *tutorialRunner) stepCatchUp() error {
	r.runAndShow("gh", "planning", "catch-up", "--since", "7d")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ Great for Monday mornings or after time off."))
	return nil
}

// ─── Write Steps (hands-on) ────────────────────────────────────────────────

func (r *tutorialRunner) stepTrackIssue() error {
	defaultRepo := r.cfg.DefaultOwner + "/" + r.cfg.DefaultOwner
	repo := r.promptDefault(
		"  Which repo should we create the tutorial issue in? ",
		defaultRepo,
	)
	repo = strings.TrimSpace(repo)

	fmt.Println()

	r.printCommandBox("gh planning track \"[Tutorial] Learning gh-planning\" --repo " + repo + " --status Backlog")
	r.promptConfirm("Press enter to create the issue")

	output, err := r.runCapture("gh", "planning", "track",
		"[Tutorial] Learning gh-planning",
		"--repo", repo,
		"--body", "This issue was created by `gh planning tutorial` to walk through the lifecycle.\n\nFeel free to close it when done! 🎓",
		"--status", "Backlog",
	)
	fmt.Print(output)
	if err != nil {
		return fmt.Errorf("creating tutorial issue: %w", err)
	}

	ref, num := parseTrackOutput(output)
	if ref != "" {
		r.createdIssueRef = ref
		r.createdIssueNum = num
		r.createdRepo = repo
		r.progress.HandsOnIssue = ref
		r.progress.HandsOnRepo = repo
		r.progress.HandsOnIssueNum = num
		_ = r.progress.Save()
	} else {
		return fmt.Errorf("couldn't detect the created issue from output — you may need to find it manually")
	}

	fmt.Println()
	fmt.Println(tui.Success.Render("  ✓ Created " + r.createdIssueRef))
	fmt.Println(tui.Muted.Render("  ↑ One command created the issue AND added it to your project board."))
	return nil
}

func (r *tutorialRunner) stepClaimIssue() error {
	if r.createdIssueRef == "" {
		return fmt.Errorf("no tutorial issue to claim — run the track step first")
	}

	r.printCommandBox("gh planning claim " + r.createdIssueRef)
	r.promptConfirm("Press enter to claim the issue")

	r.runAndShow("gh", "planning", "claim", r.createdIssueRef)
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ You're now assigned and the issue moved to In Progress."))
	fmt.Println(tui.Muted.Render("    Claiming = assign + status update in one command."))
	return nil
}

func (r *tutorialRunner) stepFocusIssue() error {
	if r.createdIssueRef == "" {
		return fmt.Errorf("no tutorial issue to focus on")
	}

	r.printCommandBox("gh planning focus " + r.createdIssueRef)
	r.promptConfirm("Press enter to start a focus session")

	r.runAndShow("gh", "planning", "focus", r.createdIssueRef)
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ Focus mode is now tracking your time on this issue."))
	fmt.Println(tui.Muted.Render("    Other commands (log, handoff, complete) are now context-aware."))
	return nil
}

func (r *tutorialRunner) stepLogProgress() error {
	r.printCommandBox(`gh planning log "Completed the gh-planning tutorial walkthrough"`)
	r.promptConfirm("Press enter to log a progress entry")

	r.runAndShow("gh", "planning", "log", "Completed the gh-planning tutorial walkthrough")

	fmt.Println()
	fmt.Println(tui.Muted.Render("  Let's also log a decision — these get categorized separately:"))
	fmt.Println()

	r.printCommandBox(`gh planning log --decision "gh-planning is the way to manage my project"`)
	r.promptConfirm("Press enter to log a decision")

	r.runAndShow("gh", "planning", "log", "--decision", "gh-planning is the way to manage my project")

	fmt.Println()
	fmt.Println(tui.Muted.Render("  Now let's see the log timeline:"))
	fmt.Println()

	r.runAndShow("gh", "planning", "logs")

	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ Logs create a breadcrumb trail of your work. They show up in"))
	fmt.Println(tui.Muted.Render("    standup reports and handoff summaries automatically."))
	return nil
}

func (r *tutorialRunner) stepCompleteIssue() error {
	if r.createdIssueRef == "" {
		return fmt.Errorf("no tutorial issue to complete")
	}

	r.printCommandBox("gh planning complete " + r.createdIssueRef + ` --done "Walked through the full gh-planning lifecycle"`)
	r.promptConfirm("Press enter to complete the issue")

	r.runAndShow("gh", "planning", "complete", r.createdIssueRef,
		"--done", "Walked through the full gh-planning lifecycle")
	fmt.Println()
	fmt.Println(tui.Muted.Render("  ↑ This posted a structured summary to the issue and moved it forward."))
	fmt.Println(tui.Muted.Render("    Focus session was also cleared automatically."))
	return nil
}

func (r *tutorialRunner) stepCleanup() error {
	if r.createdIssueRef == "" || r.createdRepo == "" {
		fmt.Println(tui.Muted.Render("  No tutorial issue to clean up."))
		return nil
	}

	choice := r.prompt("  Close the tutorial issue " + tui.Command.Render(r.createdIssueRef) + "? [Y/n]: ")
	if strings.TrimSpace(strings.ToLower(choice)) == "n" {
		fmt.Println(tui.Muted.Render("  Keeping the issue open. Close it manually when you're done."))
		return nil
	}

	issueNum := fmt.Sprintf("%d", r.createdIssueNum)
	output, _ := r.runCapture("gh", "issue", "close", issueNum,
		"--repo", r.createdRepo,
		"--comment", "Closed by gh-planning tutorial. 🎓")
	if output != "" {
		fmt.Print("  " + output)
	}

	r.progress.HandsOnIssue = ""
	r.progress.HandsOnRepo = ""
	r.progress.HandsOnIssueNum = 0
	_ = r.progress.Save()

	fmt.Println(tui.Success.Render("  ✓ Tutorial issue closed and board is tidy."))
	return nil
}
