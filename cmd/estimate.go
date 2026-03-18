package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var estimateOpts struct {
	Repo   string
	Model  string
	DryRun bool
}

var estimateCmd = &cobra.Command{
	Use:   "estimate <issue-url-or-number>",
	Short: "Suggest effort sizing for an issue based on similar closed issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runEstimate,
}

func init() {
	estimateCmd.Flags().StringVar(&estimateOpts.Repo, "repo", "", "Repository (owner/repo)")
	estimateCmd.Flags().StringVar(&estimateOpts.Model, "model", "gpt-4o", "Model name")
	estimateCmd.Flags().BoolVar(&estimateOpts.DryRun, "dry-run", false, "Show analysis without updating anything")
}

type estimateIssue struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Number int    `json:"number"`
}

type closedIssueSummary struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func runEstimate(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], estimateOpts.Repo)
	if err != nil {
		return err
	}

	issue, err := fetchIssueForEstimate(cmd.Context(), repo, number)
	if err != nil {
		return err
	}

	closed, err := fetchRecentClosedIssues(cmd.Context(), repo, number)
	if err != nil {
		return err
	}

	prompt := buildEstimatePrompt(issue, closed)

	result, err := github.ModelsEstimate(cmd.Context(), estimateOpts.Model, prompt)
	if err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"issue":    fmt.Sprintf("%s#%d", repo, number),
			"estimate": result,
			"dryRun":   estimateOpts.DryRun,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "📐 Estimate: #%d \"%s\"\n\n", number, issue.Title)
	fmt.Fprintf(cmd.OutOrStdout(), "  Size:       %s\n", result.Size)
	fmt.Fprintf(cmd.OutOrStdout(), "  Confidence: %s\n", result.Confidence)
	fmt.Fprintf(cmd.OutOrStdout(), "  Reasoning:  %s\n", result.Reasoning)

	if len(result.Similar) > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "  Similar closed issues:")
		for _, s := range result.Similar {
			fmt.Fprintf(cmd.OutOrStdout(), "    #%-5d %s (size: %s)\n", s.Number, s.Title, s.Size)
		}
	}

	if estimateOpts.DryRun {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "  (dry-run: no changes made)")
	}

	return nil
}

func fetchIssueForEstimate(ctx context.Context, repo string, number int) (*estimateIssue, error) {
	payload, err := github.Run(ctx, "issue", "view", fmt.Sprintf("%s#%d", repo, number), "--json", "title,body,number")
	if err != nil {
		return nil, err
	}
	var issue estimateIssue
	if err := json.Unmarshal(payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func fetchRecentClosedIssues(ctx context.Context, repo string, excludeNumber int) ([]closedIssueSummary, error) {
	since := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	query := fmt.Sprintf("repo:%s is:issue is:closed closed:>=%s", repo, since)

	payload, err := github.Run(ctx, "search", "issues", "--json", "number,title", "--limit", "10", "--", query)
	if err != nil {
		return nil, err
	}
	var issues []closedIssueSummary
	if err := json.Unmarshal(payload, &issues); err != nil {
		return nil, err
	}

	// Filter out the target issue if it appears in results.
	filtered := make([]closedIssueSummary, 0, len(issues))
	for _, issue := range issues {
		if issue.Number != excludeNumber {
			filtered = append(filtered, issue)
		}
	}
	return filtered, nil
}

func buildEstimatePrompt(issue *estimateIssue, closed []closedIssueSummary) string {
	var b strings.Builder

	b.WriteString("Estimate the effort for the following GitHub issue.\n\n")
	b.WriteString(fmt.Sprintf("## Target Issue #%d\n", issue.Number))
	b.WriteString(fmt.Sprintf("Title: %s\n", issue.Title))
	if issue.Body != "" {
		b.WriteString(fmt.Sprintf("Body:\n%s\n", issue.Body))
	}

	if len(closed) > 0 {
		b.WriteString("\n## Recently Closed Issues in the Same Repo\n")
		for _, c := range closed {
			b.WriteString(fmt.Sprintf("- #%d: %s\n", c.Number, c.Title))
		}
	}

	b.WriteString("\nEstimate the size using these categories:\n")
	b.WriteString("- XS: < 2 hours\n")
	b.WriteString("- S: half day\n")
	b.WriteString("- M: 1-2 days\n")
	b.WriteString("- L: 3-5 days\n")
	b.WriteString("- XL: 1+ week\n")

	b.WriteString("\nCompare the target issue against the closed issues to calibrate.\n")
	b.WriteString("Respond with ONLY a JSON object (no markdown fences):\n")
	b.WriteString(`{"size": "M", "confidence": "high|medium|low", "reasoning": "...", "similar": [{"number": 42, "title": "...", "size": "M"}]}`)
	b.WriteString("\n")

	return b.String()
}
