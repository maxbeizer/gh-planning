package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

type agentContextIssue struct {
	Repo        string `json:"repo"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	LastHandoff string `json:"lastHandoff"`
}

type agentContextReview struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Repo   string `json:"repo"`
	Author string `json:"author"`
	URL    string `json:"url"`
}

type agentContextDecision struct {
	Decision string    `json:"decision"`
	Issue    string    `json:"issue"`
	Time     time.Time `json:"time"`
}

type agentContextBlocked struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Repo   string `json:"repo"`
	URL    string `json:"url"`
}

var agentContextOpts struct {
	Project int
	Owner   string
	Issue   string
	Repo    string
}

var agentContextCmd = &cobra.Command{
	Use:   "agent-context",
	Short: "Summarize project context for an AI agent",
	RunE:  runAgentContext,
}

func init() {
	agentContextCmd.Flags().IntVar(&agentContextOpts.Project, "project", 0, "Project number")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Owner, "owner", "", "Project owner")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Issue, "issue", "", "Issue URL, number, or owner/repo#number")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Repo, "repo", "", "Repository (owner/repo) for --issue when using a number")
}

func runAgentContext(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := agentContextOpts.Owner
	project := agentContextOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}

	var focusRepo string
	var focusNumber int
	if agentContextOpts.Issue != "" {
		repo, number, err := resolveIssueInput(agentContextOpts.Issue, agentContextOpts.Repo)
		if err != nil {
			return err
		}
		focusRepo = repo
		focusNumber = number
	} else {
		current, err := session.LoadCurrent()
		if err != nil {
			return err
		}
		if current != nil {
			focusRepo = current.Repo
			focusNumber = current.IssueNumber
		}
	}

	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}

	statusCounts := map[string]int{}
	for status, items := range projectData.Items {
		statusCounts[status] = len(items)
	}

	blockedItems := []agentContextBlocked{}
	for status, items := range projectData.Items {
		if !strings.EqualFold(status, "blocked") {
			continue
		}
		for _, item := range items {
			blockedItems = append(blockedItems, agentContextBlocked{
				Number: item.Number,
				Title:  item.Title,
				Repo:   item.Repository,
				URL:    item.URL,
			})
		}
	}

	st, err := state.Load()
	if err != nil {
		return err
	}

	focusIssue := agentContextIssue{}
	focusHandoffs := []state.Handoff{}
	if focusRepo != "" && focusNumber != 0 {
		focusIssue.Repo = focusRepo
		focusIssue.Number = focusNumber
		issueTitle, _ := fetchIssueTitle(cmd.Context(), focusRepo, focusNumber)
		focusIssue.Title = issueTitle
		target := fmt.Sprintf("%s#%d", focusRepo, focusNumber)
		for _, h := range st.Handoffs {
			if h.Issue == target {
				focusHandoffs = append(focusHandoffs, h)
			}
		}
		sort.Slice(focusHandoffs, func(i, j int) bool { return focusHandoffs[i].Time.After(focusHandoffs[j].Time) })
		if len(focusHandoffs) > 0 {
			focusIssue.LastHandoff = summarizeHandoff(focusHandoffs[0])
		}
	}

	decisions := []agentContextDecision{}
	for _, h := range st.Handoffs {
		if len(h.Decisions) == 0 {
			continue
		}
		for _, decision := range h.Decisions {
			decisions = append(decisions, agentContextDecision{
				Decision: decision,
				Issue:    h.Issue,
				Time:     h.Time,
			})
		}
	}
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].Time.After(decisions[j].Time) })
	if len(decisions) > 5 {
		decisions = decisions[:5]
	}

	reviewRequests := []agentContextReview{}
	currentUser, err := github.CurrentUser(cmd.Context())
	if err == nil && currentUser != "" {
		query := fmt.Sprintf("is:pr state:open review-requested:%s", currentUser)
		results, err := github.SearchIssues(cmd.Context(), query)
		if err != nil {
			return err
		}
		limit := 5
		for i, item := range results {
			if i >= limit {
				break
			}
			reviewRequests = append(reviewRequests, agentContextReview{
				Number: item.Number,
				Title:  item.Title,
				Repo:   github.RepositoryNameFromURL(item.RepositoryURL),
				Author: item.User.Login,
				URL:    github.IssueURL(item),
			})
		}
	}

	configPayload := map[string]interface{}{
		"defaultOwner":   cfg.DefaultOwner,
		"defaultProject": cfg.DefaultProject,
		"team":           cfg.Team,
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"focus":        focusIssue,
			"project":      project,
			"owner":        owner,
			"statusCounts": statusCounts,
			"handoffs":     focusHandoffs,
			"decisions":    decisions,
			"reviews":      reviewRequests,
			"blocked":      blockedItems,
			"config":       configPayload,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Println("🤖 Agent Context — gh-planning")
	fmt.Println()

	if focusRepo != "" && focusNumber != 0 {
		title := focusIssue.Title
		if title == "" {
			title = "(title unavailable)"
		}
		fmt.Printf("📍 Focus: #%d \"%s\" (%s)\n", focusNumber, title, focusRepo)
		if len(focusHandoffs) > 0 {
			fmt.Printf("   Last handoff (%s): %s\n", humanizeDuration(time.Since(focusHandoffs[0].Time)), focusIssue.LastHandoff)
		}
	} else {
		fmt.Println("📍 Focus: none")
	}

	fmt.Println()
	fmt.Printf("📊 Project #%d Status:\n", project)
	statusKeys := make([]string, 0, len(statusCounts))
	for status := range statusCounts {
		statusKeys = append(statusKeys, status)
	}
	sort.Strings(statusKeys)
	statusParts := []string{}
	for _, status := range statusKeys {
		statusParts = append(statusParts, fmt.Sprintf("%s: %d", status, statusCounts[status]))
	}
	if len(statusParts) > 0 {
		fmt.Printf("   %s\n", strings.Join(statusParts, " | "))
	}

	fmt.Println()
	fmt.Println("🔄 Recent Decisions:")
	if len(decisions) == 0 {
		fmt.Println("   • none")
	} else {
		for _, decision := range decisions {
			fmt.Printf("   • %s (%s)\n", decision.Decision, decision.Issue)
		}
	}

	fmt.Println()
	fmt.Println("👀 Needs Review:")
	if len(reviewRequests) == 0 {
		fmt.Println("   • none")
	} else {
		for _, review := range reviewRequests {
			fmt.Printf("   • PR #%d: %s (@%s)\n", review.Number, review.Title, review.Author)
		}
	}

	fmt.Println()
	fmt.Println("🚫 Blocked:")
	if len(blockedItems) == 0 {
		fmt.Println("   • none")
	} else {
		for _, blocked := range blockedItems {
			fmt.Printf("   • #%d: %s\n", blocked.Number, blocked.Title)
		}
	}

	fmt.Println()
	fmt.Println("⚙️ Config:")
	fmt.Printf("   • default-owner: %s\n", cfg.DefaultOwner)
	fmt.Printf("   • default-project: %d\n", cfg.DefaultProject)
	if len(cfg.Team) > 0 {
		fmt.Printf("   • team: %s\n", strings.Join(cfg.Team, ", "))
	}
	return nil
}

func fetchIssueTitle(ctx context.Context, repo string, number int) (string, error) {
	payload, err := github.Run(ctx, "issue", "view", fmt.Sprintf("%s#%d", repo, number), "--json", "title")
	if err != nil {
		return "", err
	}
	var resp struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", err
	}
	return resp.Title, nil
}

func summarizeHandoff(h state.Handoff) string {
	done := ""
	remaining := ""
	if len(h.Done) > 0 {
		done = h.Done[0]
	}
	if len(h.Remaining) > 0 {
		remaining = h.Remaining[0]
	}
	if done != "" && remaining != "" {
		return fmt.Sprintf("%s done, %s remaining", done, remaining)
	}
	if done != "" {
		return fmt.Sprintf("%s done", done)
	}
	if remaining != "" {
		return fmt.Sprintf("%s remaining", remaining)
	}
	if len(h.Decisions) > 0 {
		return fmt.Sprintf("Decision: %s", h.Decisions[0])
	}
	return "No details provided"
}
