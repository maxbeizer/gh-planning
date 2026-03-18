package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

var completeOpts struct {
	Repo    string
	Project int
	Owner   string
	Done    []string
	PR      int
	Session string
}

var completeCmd = &cobra.Command{
	Use:   "complete <issue>",
	Short: "Mark an issue as complete and move it forward",
	Args:  cobra.ExactArgs(1),
	RunE:  runComplete,
}

func init() {
	completeCmd.Flags().StringVar(&completeOpts.Repo, "repo", "", "Repository (owner/repo)")
	completeCmd.Flags().IntVar(&completeOpts.Project, "project", 0, "Project number")
	completeCmd.Flags().StringVar(&completeOpts.Owner, "owner", "", "Project owner")
	completeCmd.Flags().StringArrayVar(&completeOpts.Done, "done", nil, "Completed work (repeatable)")
	completeCmd.Flags().IntVar(&completeOpts.PR, "pr", 0, "Pull request number")
	completeCmd.Flags().StringVar(&completeOpts.Session, "session", "", "Session ID")
}

func runComplete(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], completeOpts.Repo)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := completeOpts.Owner
	project := completeOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}

	sessionID := completeOpts.Session
	if sessionID == "" {
		sessionID = shortSessionID()
	}

	comment, stampedAt := buildCompletionComment(sessionID, completeOpts.Done, completeOpts.PR)
	issueOwner, issueRepo, err := splitRepo(repo)
	if err != nil {
		return err
	}
	if err := github.CreateIssueComment(cmd.Context(), issueOwner, issueRepo, number, comment); err != nil {
		return err
	}

	projectID, _, statusFieldID, statusOptions, err := github.GetProjectInfo(cmd.Context(), owner, project)
	if err != nil {
		return err
	}
	if statusFieldID == "" {
		return fmt.Errorf("status field not found on project")
	}
	statusLabel := "Done"
	optionID := ""
	var ok bool
	if completeOpts.PR != 0 {
		optionID, ok = findStatusOption(statusOptions, "Needs Review", "In Review", "Review")
		if ok {
			statusLabel = "Needs Review"
		}
	}
	if optionID == "" {
		optionID, ok = findStatusOption(statusOptions, "Done", "Complete", "Closed")
		if ok {
			statusLabel = "Done"
		}
	}
	if optionID == "" {
		return fmt.Errorf("no suitable status option found for completion")
	}
	itemID, err := findProjectItemID(cmd.Context(), owner, project, repo, number)
	if err != nil {
		// Item not in project — try to auto-add it
		contentNum := number
		if completeOpts.PR != 0 {
			contentNum = completeOpts.PR
		}
		contentID, contentErr := github.GetContentID(cmd.Context(), issueOwner, issueRepo, contentNum)
		if contentErr != nil {
			return fmt.Errorf("item not in project and could not look up content ID: %w", contentErr)
		}
		itemID, err = github.AddItemToProject(cmd.Context(), projectID, contentID)
		if err != nil {
			return fmt.Errorf("failed to add item to project: %w", err)
		}
	}
	if err := github.UpdateItemStatus(cmd.Context(), projectID, itemID, statusFieldID, optionID); err != nil {
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
		Done:        completeOpts.Done,
	}); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"issue":    fmt.Sprintf("%s#%d", repo, number),
			"session":  sessionID,
			"status":   statusLabel,
			"comment":  comment,
			"postedAt": stampedAt,
		}
		if completeOpts.PR != 0 {
			payload["pr"] = completeOpts.PR
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	issueURL := fmt.Sprintf("https://github.com/%s/issues/%d", repo, number)
	fmt.Printf("Completed %s%s (moved to %s)\n", repo, issueRef(number, issueURL), statusLabel)
	return nil
}

func buildCompletionComment(sessionID string, done []string, pr int) (string, time.Time) {
	stamp := time.Now().UTC()
	loc, err := time.LoadLocation("America/Chicago")
	if err == nil {
		stamp = time.Now().In(loc)
	}
	stampLabel := fmt.Sprintf("%s CT", stamp.Format("Mon Jan 2, 2006 3:04 PM"))
	if pr != 0 {
		done = append(done, fmt.Sprintf("PR #%d", pr))
	}
	var builder strings.Builder
	builder.WriteString("## 🔄 Session Handoff\n\n")
	builder.WriteString(fmt.Sprintf("**Session:** %s | **Time:** %s\n\n", sessionID, stampLabel))
	appendSection(&builder, "✅ Done", done)
	return builder.String(), stamp
}
