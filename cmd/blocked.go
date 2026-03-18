package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

var blockedOpts struct {
	By      string
	Repo    string
	Owner   string
	Project int
}

var blockedCmd = &cobra.Command{
	Use:   "blocked <issue>",
	Short: "Mark an issue as blocked by another issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runBlocked,
}

func init() {
	blockedCmd.Flags().StringVar(&blockedOpts.By, "by", "", "Issue that is blocking (required)")
	blockedCmd.Flags().StringVar(&blockedOpts.Repo, "repo", "", "Repository (owner/repo)")
	blockedCmd.Flags().StringVar(&blockedOpts.Owner, "owner", "", "Project owner")
	blockedCmd.Flags().IntVar(&blockedOpts.Project, "project", 0, "Project number")
	_ = blockedCmd.MarkFlagRequired("by")
}

var unblockCmd = &cobra.Command{
	Use:   "unblock <issue>",
	Short: "Remove a block from an issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnblock,
}

func init() {
	unblockCmd.Flags().StringVar(&blockedOpts.Repo, "repo", "", "Repository (owner/repo)")
}

func runBlocked(cmd *cobra.Command, args []string) error {
	blockedRepo, blockedNumber, err := resolveIssueInput(args[0], blockedOpts.Repo)
	if err != nil {
		return fmt.Errorf("invalid blocked issue: %w", err)
	}
	byRepo, byNumber, err := resolveIssueInput(blockedOpts.By, blockedOpts.Repo)
	if err != nil {
		return fmt.Errorf("invalid --by issue: %w", err)
	}

	blockedRef := fmt.Sprintf("%s#%d", blockedRepo, blockedNumber)
	byRef := fmt.Sprintf("%s#%d", byRepo, byNumber)

	// Post comment on the blocked issue.
	blockedOwner, blockedRepoName, err := splitRepo(blockedRepo)
	if err != nil {
		return err
	}
	comment := fmt.Sprintf("🚫 Blocked by #%d", byNumber)
	if byRepo != blockedRepo {
		comment = fmt.Sprintf("🚫 Blocked by %s#%d", byRepo, byNumber)
	}
	if err := github.CreateIssueComment(cmd.Context(), blockedOwner, blockedRepoName, blockedNumber, comment); err != nil {
		return err
	}

	// Store dependency in local state.
	now := time.Now()
	if err := state.AddDependency(state.Dependency{
		Blocked:   blockedRef,
		BlockedBy: byRef,
		Time:      now,
	}); err != nil {
		return err
	}

	// Optionally move to "Blocked" status.
	trySetBlockedStatus(cmd, blockedRepo, blockedNumber)

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"blocked":   blockedRef,
			"blockedBy": byRef,
			"comment":   comment,
			"time":      now,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Marked %s as blocked by %s\n", blockedRef, byRef)
	return nil
}

func runUnblock(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], blockedOpts.Repo)
	if err != nil {
		return fmt.Errorf("invalid issue: %w", err)
	}
	ref := fmt.Sprintf("%s#%d", repo, number)

	// Post comment.
	owner, repoName, err := splitRepo(repo)
	if err != nil {
		return err
	}
	comment := "✅ Unblocked"
	if err := github.CreateIssueComment(cmd.Context(), owner, repoName, number, comment); err != nil {
		return err
	}

	// Remove from local state.
	if err := state.RemoveDependency(ref); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"unblocked": ref,
			"comment":   comment,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Unblocked %s\n", ref)
	return nil
}

// trySetBlockedStatus attempts to move the issue to "Blocked" status in the project.
// It fails silently if the project or status is not configured.
func trySetBlockedStatus(cmd *cobra.Command, repo string, number int) {
	pc, err := resolveProjectConfig(blockedOpts.Owner, blockedOpts.Project)
	if err != nil {
		return
	}

	projectID, _, statusFieldID, statusOptions, err := github.GetProjectInfo(cmd.Context(), pc.Owner, pc.Project)
	if err != nil || statusFieldID == "" {
		return
	}
	optionID, ok := findStatusOption(statusOptions, "Blocked")
	if !ok {
		return
	}
	itemID, err := findProjectItemID(cmd.Context(), pc.Owner, pc.Project, repo, number)
	if err != nil {
		return
	}
	if err := github.UpdateItemStatus(cmd.Context(), projectID, itemID, statusFieldID, optionID); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to update project status: %v\n", err)
	}
}
