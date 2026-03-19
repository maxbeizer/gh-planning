package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/maxbeizer/gh-planning/internal/tutorial"
	"github.com/spf13/cobra"
)

// ─── UI Helpers ────────────────────────────────────────────────────────────

func (r *tutorialRunner) printBanner(title, subtitle string) {
	fmt.Println()
	banner := tui.ActiveBox.Render(
		tui.Title.Render("🎓 "+title) + "\n" + tui.Muted.Render(subtitle))
	fmt.Println(banner)
	fmt.Println()
}

func (r *tutorialRunner) printStepHeader(current, total int, title string) {
	// Clear visual separator between steps
	fmt.Println()
	fmt.Println(tui.Dimmed.Render("  ─────────────────────────────────────────────────────────────"))
	fmt.Println()

	dots := tui.ProgressDots(current-1, total)
	stepLabel := fmt.Sprintf("Step %d of %d", current, total)

	header := tui.Box.Render(
		dots + "  " + tui.Muted.Render(stepLabel) + "\n" +
			tui.Title.Render(title))
	fmt.Println(header)
}

func (r *tutorialRunner) printCommandBox(command string) {
	box := tui.ActiveBox.Render("$ " + tui.Command.Render(command))
	fmt.Println("  " + box)
}

func (r *tutorialRunner) prompt(msg string) string {
	fmt.Print(msg)
	text, _ := r.reader.ReadString('\n')
	return strings.TrimRight(text, "\n\r")
}

func (r *tutorialRunner) promptDefault(msg, defaultVal string) string {
	fmt.Printf("%s"+tui.Muted.Render("[%s]")+": ", msg, defaultVal)
	text, _ := r.reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultVal
	}
	return text
}

func (r *tutorialRunner) promptContinue() {
	fmt.Println()
	fmt.Println()
	fmt.Print(tui.Muted.Render("  Press enter for next step (q to quit)... "))
	text, _ := r.reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(text)) == "q" {
		fmt.Println(tui.Muted.Render("\n  Progress saved. Run `gh planning tutorial` to resume."))
		os.Exit(0)
	}
}

func (r *tutorialRunner) promptConfirm(msg string) {
	fmt.Print(tui.Muted.Render("  "+msg) + " ")
	r.reader.ReadString('\n')
}

// runAndShow executes a command and prints its output inline.
func (r *tutorialRunner) runAndShow(name string, args ...string) {
	cmdStr := name
	for _, a := range args {
		if strings.Contains(a, " ") || strings.Contains(a, "\"") {
			cmdStr += ` "` + a + `"`
		} else {
			cmdStr += " " + a
		}
	}
	fmt.Println(tui.Dimmed.Render("  $ " + cmdStr))
	fmt.Println()

	cmd := exec.CommandContext(r.ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

// runCapture executes a command and returns its combined output.
func (r *tutorialRunner) runCapture(name string, args ...string) (string, error) {
	cmd := exec.CommandContext(r.ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// ─── List & Progress ───────────────────────────────────────────────────────

type progressStep struct {
	id, title string
}

type progressSection struct {
	name  string
	steps []progressStep
}

func tutorialSections() []progressSection {
	return []progressSection{
		{
			name: "Explore Tour",
			steps: []progressStep{
				{"explore-dashboard", "Your Command Center"},
				{"explore-status", "Project Status"},
				{"explore-board", "Kanban Board"},
				{"explore-standup", "Standup Report"},
				{"explore-catchup", "Catch-Up Summary"},
			},
		},
		{
			name: "Hands-On Tutorial",
			steps: []progressStep{
				{"handson-dashboard", "Your Command Center"},
				{"handson-board", "Your Project Board"},
				{"handson-track", "Create a Tutorial Issue"},
				{"handson-focus", "Focus On It"},
				{"handson-log", "Log Progress"},
				{"handson-standup", "Generate a Standup"},
				{"handson-cleanup", "Clean Up"},
			},
		},
	}
}

func listTutorialProgress(cmd *cobra.Command, progress *tutorial.Progress) error {
	fmt.Fprintln(cmd.OutOrStdout(), tui.Title.Render("Tutorial Progress"))
	fmt.Fprintln(cmd.OutOrStdout())

	totalDone := 0
	totalAll := 0
	for _, sec := range tutorialSections() {
		fmt.Fprintln(cmd.OutOrStdout(), "  "+tui.Heading.Render(sec.name))
		secDone := 0
		for _, step := range sec.steps {
			status := "○"
			style := tui.Dimmed
			if progress.IsCompleted(step.id) {
				status = "●"
				style = tui.Success
				secDone++
			}
			fmt.Fprintf(cmd.OutOrStdout(), "    %s %s\n", style.Render(status), step.title)
		}
		totalDone += secDone
		totalAll += len(sec.steps)
		fmt.Fprintf(cmd.OutOrStdout(), "    %s\n\n",
			tui.Muted.Render(fmt.Sprintf("%d of %d completed", secDone, len(sec.steps))))
	}

	fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", tui.ProgressDots(totalDone, totalAll))

	if progress.HandsOnIssue != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  %s %s\n",
			tui.Muted.Render("Active tutorial issue:"),
			tui.Command.Render(progress.HandsOnIssue))
	}
	return nil
}

// ─── Output Parsing ────────────────────────────────────────────────────────

// parseTrackOutput extracts the issue reference from `gh planning track` output.
func parseTrackOutput(output string) (string, int) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, "https://github.com/"); idx != -1 {
			url := strings.TrimSpace(line[idx:])
			url = strings.TrimRight(url, " \t\n\r.")
			parts := strings.Split(url, "/")
			if len(parts) >= 7 && parts[5] == "issues" {
				owner := parts[3]
				repo := parts[4]
				numStr := parts[6]
				var num int
				fmt.Sscanf(numStr, "%d", &num)
				if num > 0 {
					return owner + "/" + repo + "#" + numStr, num
				}
			}
		}
	}
	return "", 0
}
