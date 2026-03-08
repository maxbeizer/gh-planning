package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

var logOpts struct {
	Decision   bool
	Blocker    bool
	Hypothesis bool
	Tried      bool
	Result     bool
}

var logCmd = &cobra.Command{
	Use:   "log <message>",
	Short: "Log progress on the current focus issue",
	Long: `Record a progress note, decision, blocker, or finding against the
current focus issue. Logs build a timeline that future sessions can
read to understand what happened and why.

Examples:
  gh planning log "OAuth callback working"
  gh planning log --decision "Using JWT for stateless auth"
  gh planning log --blocker "Unclear on refresh token rotation"
  gh planning log --tried "Session-based approach, too complex"
  gh planning log --result "Benchmarks show 2ms token validation"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLog,
}

func init() {
	logCmd.Flags().BoolVar(&logOpts.Decision, "decision", false, "Log a decision")
	logCmd.Flags().BoolVar(&logOpts.Blocker, "blocker", false, "Log a blocker")
	logCmd.Flags().BoolVar(&logOpts.Hypothesis, "hypothesis", false, "Log a hypothesis")
	logCmd.Flags().BoolVar(&logOpts.Tried, "tried", false, "Log something attempted")
	logCmd.Flags().BoolVar(&logOpts.Result, "result", false, "Log a result or finding")
}

func runLog(cmd *cobra.Command, args []string) error {
	focus, err := session.LoadCurrent()
	if err != nil {
		return err
	}
	if focus == nil {
		return fmt.Errorf("no focus issue set; run `gh planning focus <issue>` or `gh planning claim <issue>` first")
	}

	message := strings.Join(args, " ")
	kind := "progress"
	prefix := "📝"
	switch {
	case logOpts.Decision:
		kind = "decision"
		prefix = "🎯"
	case logOpts.Blocker:
		kind = "blocker"
		prefix = "🚫"
	case logOpts.Hypothesis:
		kind = "hypothesis"
		prefix = "💡"
	case logOpts.Tried:
		kind = "tried"
		prefix = "🔄"
	case logOpts.Result:
		kind = "result"
		prefix = "✅"
	}

	entry := state.LogEntry{
		Issue:     focus.Issue,
		SessionID: focus.SessionID,
		Time:      time.Now().UTC(),
		Message:   message,
		Kind:      kind,
	}

	if err := state.AddLog(entry); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(entry, OutputOptions())
	}

	fmt.Printf("%s [%s] %s (%s)\n", prefix, kind, message, focus.Issue)
	return nil
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show progress logs for the current focus issue",
	Long: `Display the timeline of progress notes, decisions, and findings
for the current focus issue (or all issues with --all).`,
	RunE: runLogs,
}

var logsOpts struct {
	All   bool
	Since string
}

func init() {
	logsCmd.Flags().BoolVar(&logsOpts.All, "all", false, "Show logs for all issues")
	logsCmd.Flags().StringVar(&logsOpts.Since, "since", "", "Show logs since duration (e.g. 24h, 7d)")
}

func runLogs(cmd *cobra.Command, args []string) error {
	issue := ""
	if !logsOpts.All {
		focus, err := session.LoadCurrent()
		if err != nil {
			return err
		}
		if focus != nil {
			issue = focus.Issue
		}
	}

	var since time.Time
	if logsOpts.Since != "" {
		d, err := parseDuration(logsOpts.Since)
		if err != nil {
			return fmt.Errorf("invalid since duration: %w", err)
		}
		since = time.Now().Add(-d)
	}

	entries, err := state.GetLogs(issue, since)
	if err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(entries, OutputOptions())
	}

	if len(entries) == 0 {
		if issue != "" {
			fmt.Printf("No logs for %s\n", issue)
		} else {
			fmt.Println("No logs found")
		}
		return nil
	}

	for _, entry := range entries {
		prefix := kindPrefix(entry.Kind)
		age := humanizeDuration(time.Since(entry.Time))
		if logsOpts.All {
			fmt.Printf("%s %s [%s] %s (%s)\n", prefix, entry.Issue, entry.Kind, entry.Message, age)
		} else {
			fmt.Printf("%s [%s] %s (%s)\n", prefix, entry.Kind, entry.Message, age)
		}
	}
	return nil
}

func kindPrefix(kind string) string {
	switch kind {
	case "decision":
		return "🎯"
	case "blocker":
		return "🚫"
	case "hypothesis":
		return "💡"
	case "tried":
		return "🔄"
	case "result":
		return "✅"
	default:
		return "📝"
	}
}
