package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/spf13/cobra"
)

var focusCmd = &cobra.Command{
	Use:   "focus [owner/repo#number]",
	Short: "Set or show current focus",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFocus,
}

var unfocusOpts struct {
	Comment string
}

var unfocusCmd = &cobra.Command{
	Use:   "unfocus",
	Short: "Clear current focus",
	RunE:  runUnfocus,
}

func init() {
	unfocusCmd.Flags().StringVar(&unfocusOpts.Comment, "comment", "", "Comment to add to the issue")
}

func runFocus(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		current, err := session.LoadCurrent()
		if err != nil {
			return err
		}
		if current == nil {
			fmt.Println("No active focus session.")
			return nil
		}
		fmt.Printf("Focused on %s (%s)\n", current.Issue, humanizeDuration(current.Elapsed()))
		return nil
	}
	issueRef := args[0]
	repo, number, err := parseIssueRef(issueRef)
	if err != nil {
		return err
	}
	// Normalize to full ref format for display and persistence
	fullRef := fmt.Sprintf("%s#%d", repo, number)
	sess := &session.FocusSession{
		Issue:       fullRef,
		IssueNumber: number,
		Repo:        repo,
		StartedAt:   time.Now().UTC(),
		SessionID:   fmt.Sprintf("%d", time.Now().UnixNano()),
	}
	if err := session.SaveCurrent(sess); err != nil {
		return err
	}
	fmt.Printf("Focused on %s\n", issueRef)
	return nil
}

func runUnfocus(cmd *cobra.Command, args []string) error {
	current, err := session.LoadCurrent()
	if err != nil {
		return err
	}
	if current == nil {
		fmt.Println("No active focus session.")
		return nil
	}
	elapsed := current.Elapsed()
	if unfocusOpts.Comment != "" {
		owner, repo, err := splitRepo(current.Repo)
		if err != nil {
			return err
		}
		if err := github.CreateIssueComment(cmd.Context(), owner, repo, current.IssueNumber, unfocusOpts.Comment); err != nil {
			return err
		}
	}
	if err := session.ClearCurrent(); err != nil {
		return err
	}
	fmt.Printf("Cleared focus (%s)\n", humanizeDuration(elapsed))
	return nil
}

func splitRepo(value string) (string, string, error) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo: %s", value)
	}
	return parts[0], parts[1], nil
}
