package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type prReview struct {
	Title             string `json:"title"`
	State             string `json:"state"`
	Additions         int    `json:"additions"`
	Deletions         int    `json:"deletions"`
	Mergeable         string `json:"mergeable"`
	Files             []struct {
		Path string `json:"path"`
	} `json:"files"`
	Reviews []struct {
		State string `json:"state"`
	} `json:"reviews"`
	StatusCheckRollup []struct {
		Conclusion string `json:"conclusion"`
		State      string `json:"state"`
	} `json:"statusCheckRollup"`
}

type reviewSummary struct {
	Approvals        int      `json:"approvals"`
	ChangesRequested int      `json:"changesRequested"`
	ChecksPassing    bool     `json:"checksPassing"`
	ChecksPending    int      `json:"checksPending"`
	ChecksFailing    int      `json:"checksFailing"`
	KeyFiles         []string `json:"keyFiles"`
}

var reviewOpts struct {
	Repo string
}

var reviewCmd = &cobra.Command{
	Use:   "review <pr>",
	Short: "Summarize review status for a pull request",
	Args:  cobra.ExactArgs(1),
	RunE:  runReview,
}

func init() {
	reviewCmd.Flags().StringVar(&reviewOpts.Repo, "repo", "", "Repository (owner/repo)")
}

func runReview(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], reviewOpts.Repo)
	if err != nil {
		return err
	}
	payload, err := github.Run(cmd.Context(), "pr", "view", fmt.Sprintf("%d", number), "--repo", repo, "--json", "title,state,additions,deletions,files,reviews,statusCheckRollup,mergeable")
	if err != nil {
		return err
	}
	var pr prReview
	if err := json.Unmarshal(payload, &pr); err != nil {
		return err
	}

	summary := reviewSummary{}
	for _, review := range pr.Reviews {
		switch strings.ToUpper(review.State) {
		case "APPROVED":
			summary.Approvals++
		case "CHANGES_REQUESTED":
			summary.ChangesRequested++
		}
	}
	for _, check := range pr.StatusCheckRollup {
		state := strings.ToUpper(check.State)
		conclusion := strings.ToUpper(check.Conclusion)
		switch {
		case conclusion == "SUCCESS" || state == "SUCCESS":
			// Successful checks need no counter; ChecksPassing is derived
			// from the absence of failing/pending checks below.
		case conclusion == "FAILURE" || state == "FAILURE":
			summary.ChecksFailing++
		case conclusion == "NEUTRAL" || conclusion == "SKIPPED":
			// ignore
		default:
			summary.ChecksPending++
		}
	}
	summary.ChecksPassing = summary.ChecksFailing == 0 && summary.ChecksPending == 0 && len(pr.StatusCheckRollup) > 0
	for i, file := range pr.Files {
		if i >= 5 {
			break
		}
		summary.KeyFiles = append(summary.KeyFiles, file.Path)
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"repo":    repo,
			"number":  number,
			"details": pr,
			"summary": summary,
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload, OutputOptions())
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "🔍 Review: PR #%d \"%s\" (%s)\n\n", number, pr.Title, repo)
	fmt.Fprintf(w, "📊 Stats: +%d -%d across %d files\n", pr.Additions, pr.Deletions, len(pr.Files))
	if len(pr.StatusCheckRollup) == 0 {
		fmt.Fprintln(w, "✅ CI: no checks")
	} else if summary.ChecksPassing {
		fmt.Fprintln(w, "✅ CI: All checks passing")
	} else {
		parts := []string{}
		if summary.ChecksFailing > 0 {
			parts = append(parts, fmt.Sprintf("%d failing", summary.ChecksFailing))
		}
		if summary.ChecksPending > 0 {
			parts = append(parts, fmt.Sprintf("%d pending", summary.ChecksPending))
		}
		if len(parts) == 0 {
			parts = append(parts, "checks running")
		}
		fmt.Fprintf(w, "⚠️ CI: %s\n", strings.Join(parts, ", "))
	}
	fmt.Fprintf(w, "👥 Reviews: %d approved, %d changes requested\n", summary.Approvals, summary.ChangesRequested)
	conflicts := "none"
	if strings.EqualFold(pr.Mergeable, "CONFLICTING") {
		conflicts = "merge conflicts"
	}
	fmt.Fprintf(w, "⚠️ Conflicts: %s\n", conflicts)
	if len(summary.KeyFiles) > 0 {
		fmt.Fprintf(w, "📁 Key files: %s\n", strings.Join(summary.KeyFiles, ", "))
	}
	return nil
}
