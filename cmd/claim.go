package cmd

import (
	"fmt"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/spf13/cobra"
)

var claimOpts struct {
	Repo    string
	Project int
	Owner   string
	Session string
}

var claimCmd = &cobra.Command{
	Use:   "claim <issue>",
	Short: "Claim an issue and move it to In Progress",
	Args:  cobra.ExactArgs(1),
	RunE:  runClaim,
}

func init() {
	claimCmd.Flags().StringVar(&claimOpts.Repo, "repo", "", "Repository (owner/repo)")
	claimCmd.Flags().IntVar(&claimOpts.Project, "project", 0, "Project number")
	claimCmd.Flags().StringVar(&claimOpts.Owner, "owner", "", "Project owner")
	claimCmd.Flags().StringVar(&claimOpts.Session, "session", "", "Session ID")
}

func runClaim(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], claimOpts.Repo)
	if err != nil {
		return err
	}
	pc, err := resolveProjectConfig(claimOpts.Owner, claimOpts.Project)
	if err != nil {
		return err
	}

	sessionID := claimOpts.Session
	if sessionID == "" {
		sessionID = shortSessionID()
	}

	stamp := formatTimestamp(time.Now())
	comment := fmt.Sprintf("🤖 Claimed by agent session `%s` at %s", sessionID, stamp)
	issueOwner, issueRepo, err := splitRepo(repo)
	if err != nil {
		return err
	}
	if err := github.CreateIssueComment(cmd.Context(), issueOwner, issueRepo, number, comment); err != nil {
		return err
	}

	projectID, _, statusFieldID, statusOptions, err := github.GetProjectInfo(cmd.Context(), pc.Owner, pc.Project)
	if err != nil {
		return err
	}
	if statusFieldID == "" {
		return fmt.Errorf("status field not found on project")
	}
	optionID, ok := findStatusOption(statusOptions, "In Progress")
	if !ok {
		return fmt.Errorf("status option not found: In Progress")
	}
	itemID, err := findProjectItemID(cmd.Context(), pc.Owner, pc.Project, repo, number)
	if err != nil {
		return err
	}
	if err := github.UpdateItemStatus(cmd.Context(), projectID, itemID, statusFieldID, optionID); err != nil {
		return err
	}

	focus := &session.FocusSession{
		Issue:       fmt.Sprintf("%s#%d", repo, number),
		IssueNumber: number,
		Repo:        repo,
		StartedAt:   time.Now().UTC(),
		SessionID:   sessionID,
	}
	if err := session.SaveCurrent(focus); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"issue":   fmt.Sprintf("%s#%d", repo, number),
			"session": sessionID,
			"status":  "In Progress",
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	issueURL := fmt.Sprintf("https://github.com/%s/issues/%d", repo, number)
	fmt.Fprintf(cmd.OutOrStdout(), "Claimed %s%s (session %s)\n", repo, issueRef(number, issueURL), sessionID)
	return nil
}
