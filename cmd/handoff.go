package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

var handoffOpts struct {
	Repo      string
	Done      []string
	Remaining []string
	Decision  []string
	Uncertain []string
	SessionID string
}

var handoffCmd = &cobra.Command{
	Use:   "handoff <issue-url-or-number>",
	Short: "Post a structured handoff comment to an issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runHandoff,
}

func init() {
	handoffCmd.Flags().StringVar(&handoffOpts.Repo, "repo", "", "Repository (owner/repo)")
	handoffCmd.Flags().StringArrayVar(&handoffOpts.Done, "done", nil, "Completed work (repeatable)")
	handoffCmd.Flags().StringArrayVar(&handoffOpts.Remaining, "remaining", nil, "Remaining work (repeatable)")
	handoffCmd.Flags().StringArrayVar(&handoffOpts.Decision, "decision", nil, "Decisions made (repeatable)")
	handoffCmd.Flags().StringArrayVar(&handoffOpts.Uncertain, "uncertain", nil, "Open questions (repeatable)")
	handoffCmd.Flags().StringVar(&handoffOpts.SessionID, "session", "", "Session ID")
}

func runHandoff(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], handoffOpts.Repo)
	if err != nil {
		return err
	}
	sessionID := handoffOpts.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	comment, stampedAt := buildHandoffComment(sessionID)
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return err
	}
	if err := github.CreateIssueComment(cmd.Context(), owner, repoName, number, comment); err != nil {
		return err
	}
	if err := maybeClearFocus(repo, number); err != nil {
		return err
	}
	if err := state.AddHandoff(state.Handoff{
		Issue:       fmt.Sprintf("%s#%d", repo, number),
		IssueNumber: number,
		Repo:        repo,
		SessionID:   sessionID,
		Time:        stampedAt,
		Done:        handoffOpts.Done,
		Remaining:   handoffOpts.Remaining,
		Decisions:   handoffOpts.Decision,
		Uncertain:   handoffOpts.Uncertain,
	}); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"issue":    fmt.Sprintf("%s#%d", repo, number),
			"session":  sessionID,
			"comment":  comment,
			"postedAt": stampedAt,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	issueURL := fmt.Sprintf("https://github.com/%s/issues/%d", repo, number)
	fmt.Printf("Posted handoff to %s%s\n", repo, issueRef(number, issueURL))
	return nil
}

func buildHandoffComment(sessionID string) (string, time.Time) {
	stamp := time.Now().UTC()
	loc, err := time.LoadLocation("America/Chicago")
	if err == nil {
		stamp = time.Now().In(loc)
	}
	stampLabel := fmt.Sprintf("%s CT", stamp.Format("Mon Jan 2, 2006 3:04 PM"))
	var builder strings.Builder
	builder.WriteString("## 🔄 Session Handoff\n\n")
	builder.WriteString(fmt.Sprintf("**Session:** %s | **Time:** %s\n\n", sessionID, stampLabel))
	appendSection(&builder, "✅ Done", handoffOpts.Done)
	appendSection(&builder, "📋 Remaining", handoffOpts.Remaining)
	appendSection(&builder, "🎯 Decisions", handoffOpts.Decision)
	appendSection(&builder, "❓ Uncertain", handoffOpts.Uncertain)
	return builder.String(), stamp
}

func appendSection(builder *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	builder.WriteString(fmt.Sprintf("### %s\n", title))
	for _, item := range items {
		builder.WriteString(fmt.Sprintf("- %s\n", item))
	}
	builder.WriteString("\n")
}

func maybeClearFocus(repo string, number int) error {
	current, err := session.LoadCurrent()
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}
	if current.Repo == repo && current.IssueNumber == number {
		return session.ClearCurrent()
	}
	return nil
}
