package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type breakdownIssue struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Comments []struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Body string `json:"body"`
	} `json:"comments"`
}

type breakdownIssueView struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

var breakdownOpts struct {
	Repo      string
	Model     string
	DryRun    bool
	MaxIssues int
}

var breakdownCmd = &cobra.Command{
	Use:   "breakdown <issue-url-or-number>",
	Short: "Split an issue into sub-issues",
	Args:  cobra.ExactArgs(1),
	RunE:  runBreakdown,
}

func init() {
	breakdownCmd.Flags().StringVar(&breakdownOpts.Repo, "repo", "", "Repository (owner/repo)")
	breakdownCmd.Flags().StringVar(&breakdownOpts.Model, "model", "gpt-4o", "Model name")
	breakdownCmd.Flags().BoolVar(&breakdownOpts.DryRun, "dry-run", false, "Preview plan without creating issues")
	breakdownCmd.Flags().IntVar(&breakdownOpts.MaxIssues, "max-issues", 10, "Maximum sub-issues to create")
}

func runBreakdown(cmd *cobra.Command, args []string) error {
	repo, number, err := resolveIssueInput(args[0], breakdownOpts.Repo)
	if err != nil {
		return err
	}
	issue, err := fetchIssueForBreakdown(cmd.Context(), repo, number)
	if err != nil {
		return err
	}
	prompt := buildBreakdownPrompt(issue)

	items, err := github.ModelsBreakdown(cmd.Context(), breakdownOpts.Model, prompt)
	if err != nil {
		return err
	}
	if breakdownOpts.MaxIssues > 0 && len(items) > breakdownOpts.MaxIssues {
		items = items[:breakdownOpts.MaxIssues]
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"issue":  fmt.Sprintf("%s#%d", repo, number),
			"items":  items,
			"dryRun": breakdownOpts.DryRun,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Printf("🔨 Breakdown: #%d \"%s\" → %d sub-issues\n\n", number, issue.Title, len(items))
	printBreakdownItems(items)
	fmt.Println()
	if breakdownOpts.DryRun {
		fmt.Println("Create these sub-issues? [y/N]")
		return nil
	}

	created, err := createSubIssues(cmd.Context(), repo, number, items)
	if err != nil {
		return err
	}
	fmt.Printf("Created %d sub-issues.\n", len(created))
	return nil
}

func fetchIssueForBreakdown(ctx context.Context, repo string, number int) (*breakdownIssue, error) {
	payload, err := github.Run(ctx, "issue", "view", fmt.Sprintf("%s#%d", repo, number), "--json", "title,body,comments")
	if err != nil {
		return nil, err
	}
	var issue breakdownIssue
	if err := json.Unmarshal(payload, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func buildBreakdownPrompt(issue *breakdownIssue) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Title: %s\n\n", issue.Title))
	if issue.Body != "" {
		builder.WriteString("Body:\n")
		builder.WriteString(issue.Body)
		builder.WriteString("\n\n")
	}
	if len(issue.Comments) > 0 {
		builder.WriteString("Comments:\n")
		for _, comment := range issue.Comments {
			line := strings.TrimSpace(comment.Body)
			if line == "" {
				continue
			}
			builder.WriteString(fmt.Sprintf("- @%s: %s\n", comment.Author.Login, line))
		}
	}
	return builder.String()
}

func printBreakdownItems(items []github.BreakdownItem) {
	for idx, item := range items {
		fmt.Printf("  %d. %s\n", idx+1, item.Title)
		if len(item.Labels) > 0 {
			fmt.Printf("     Labels: %s\n", strings.Join(item.Labels, ", "))
		}
		if len(item.DependsOn) > 0 {
			deps := []string{}
			for _, dep := range item.DependsOn {
				deps = append(deps, fmt.Sprintf("#%d", dep))
			}
			fmt.Printf("     Depends on: %s\n", strings.Join(deps, ", "))
		}
		fmt.Println()
	}
}

func createSubIssues(ctx context.Context, repo string, parentNumber int, items []github.BreakdownItem) ([]breakdownIssueView, error) {
	parentView, err := fetchIssueView(ctx, repo, parentNumber)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	projectID := ""
	if cfg.DefaultOwner != "" && cfg.DefaultProject != 0 {
		id, _, _, _, err := github.GetProjectInfo(ctx, cfg.DefaultOwner, cfg.DefaultProject)
		if err == nil {
			projectID = id
		}
	}

	created := []breakdownIssueView{}
	for _, item := range items {
		createdIssue, err := createIssue(ctx, repo, item)
		if err != nil {
			return nil, err
		}
		view, err := fetchIssueView(ctx, repo, createdIssue.Number)
		if err != nil {
			return nil, err
		}
		if err := addSubIssue(ctx, parentView.ID, view.ID); err != nil {
			return nil, err
		}
		if projectID != "" {
			_, _ = github.AddItemToProject(ctx, projectID, view.ID)
		}
		created = append(created, *view)
	}
	return created, nil
}

type issueCreateResult struct {
	URL    string
	Number int
}

func createIssue(ctx context.Context, repo string, item github.BreakdownItem) (*issueCreateResult, error) {
	args := []string{"issue", "create", "--repo", repo, "--title", item.Title}
	body := item.Body
	if body == "" {
		body = "Generated from breakdown"
	}
	args = append(args, "--body", body)
	for _, label := range item.Labels {
		args = append(args, "--label", label)
	}
	payload, err := github.Run(ctx, args...)
	if err != nil {
		return nil, err
	}
	issueURL := strings.TrimSpace(string(payload))
	if issueURL == "" {
		return nil, fmt.Errorf("issue create returned empty URL")
	}
	_, number, err := parseIssueURL(issueURL)
	if err != nil {
		return nil, err
	}
	return &issueCreateResult{URL: issueURL, Number: number}, nil
}

func fetchIssueView(ctx context.Context, repo string, number int) (*breakdownIssueView, error) {
	payload, err := github.Run(ctx, "issue", "view", fmt.Sprintf("%s#%d", repo, number), "--json", "id,number")
	if err != nil {
		return nil, err
	}
	var view breakdownIssueView
	if err := json.Unmarshal(payload, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

func addSubIssue(ctx context.Context, parentID string, subID string) error {
	mutation := `mutation($parent: ID!, $sub: ID!) {
  addSubIssue(input: {issueId: $parent, subIssueId: $sub}) {
    subIssue { id number }
  }
}`
	_, err := github.GraphQL(ctx, mutation, map[string]interface{}{"parent": parentID, "sub": subID})
	return err
}

