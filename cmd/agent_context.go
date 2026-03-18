package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

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
	Project    int
	Owner      string
	Issue      string
	Repo       string
	NewSession bool
}

var agentContextCmd = &cobra.Command{
	Use:   "agent-context",
	Short: "Summarize project context for an AI agent",
	Long: `Display everything an agent needs to start or continue work.

Use --new-session at the start of each agent conversation to get full
context: open issues, recent logs, pending handoffs, and what to work on next.

Add this to your CLAUDE.md, system prompt, or agent instructions:
  Run ` + "`gh planning agent-context --new-session`" + ` at conversation start.`,
	RunE: runAgentContext,
}

func init() {
	agentContextCmd.Flags().IntVar(&agentContextOpts.Project, "project", 0, "Project number")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Owner, "owner", "", "Project owner")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Issue, "issue", "", "Issue URL, number, or owner/repo#number")
	agentContextCmd.Flags().StringVar(&agentContextOpts.Repo, "repo", "", "Repository (owner/repo) for --issue when using a number")
	agentContextCmd.Flags().BoolVar(&agentContextOpts.NewSession, "new-session", false, "Mark a new session start (updates last-seen timestamp)")
}

func runAgentContext(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(agentContextOpts.Owner, agentContextOpts.Project)
	if err != nil {
		return err
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

	// Parallel: GetProject + CurrentUser + fetchIssueTitle (all independent)
	type projectResult struct {
		data *github.Project
		err  error
	}
	type userResult struct {
		user string
		err  error
	}
	type titleResult struct {
		title string
	}

	projectCh := make(chan projectResult, 1)
	userCh := make(chan userResult, 1)
	titleCh := make(chan titleResult, 1)

	go func() {
		data, err := github.GetProject(cmd.Context(), pc.Owner, pc.Project)
		projectCh <- projectResult{data, err}
	}()
	go func() {
		user, err := github.CurrentUser(cmd.Context())
		userCh <- userResult{user, err}
	}()
	go func() {
		if focusRepo != "" && focusNumber != 0 {
			title, _ := fetchIssueTitle(cmd.Context(), focusRepo, focusNumber)
			titleCh <- titleResult{title}
		} else {
			titleCh <- titleResult{}
		}
	}()

	pr := <-projectCh
	if pr.err != nil {
		return pr.err
	}
	projectData := pr.data
	ur := <-userCh
	currentUser := ur.user
	tr := <-titleCh

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
		focusIssue.Title = tr.title
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

	// Fetch review requests scoped to repos in this project
	reviewRequests := []agentContextReview{}
	if currentUser != "" {
		repos := uniqueRepos(projectData)
		repoQuery := buildRepoQuery(repos)
		query := fmt.Sprintf("is:pr state:open review-requested:%s", currentUser)
		if repoQuery != "" {
			query = repoQuery + " " + query
		}
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
		"defaultOwner":   pc.Cfg.DefaultOwner,
		"defaultProject": pc.Cfg.DefaultProject,
		"team":           pc.Cfg.Team,
	}

	// Gather recent logs
	var logsSince time.Time
	if !st.LastSeen.IsZero() {
		logsSince = st.LastSeen
	} else {
		logsSince = time.Now().Add(-7 * 24 * time.Hour)
	}
	recentLogs, _ := state.GetLogs("", logsSince)
	if len(recentLogs) > 10 {
		recentLogs = recentLogs[len(recentLogs)-10:]
	}

	// Get next suggested item from queue
	var nextUp *github.ProjectItem
	for _, status := range []string{"Ready", "Backlog"} {
		if items, ok := projectData.Items[status]; ok && len(items) > 0 {
			nextUp = &items[0]
			break
		}
	}

	// Update last-seen if --new-session
	if agentContextOpts.NewSession {
		if err := state.UpdateLastSeen(time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to update last-seen: %v\n", err)
		}
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"focus":        focusIssue,
			"project":      pc.Project,
			"owner":        pc.Owner,
			"statusCounts": statusCounts,
			"handoffs":     focusHandoffs,
			"decisions":    decisions,
			"recentLogs":   recentLogs,
			"reviews":      reviewRequests,
			"blocked":      blockedItems,
			"nextUp":       nextUp,
			"config":       configPayload,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintln(cmd.OutOrStdout(), "🤖 Agent Context — gh-planning")
	if agentContextOpts.NewSession {
		fmt.Fprintln(cmd.OutOrStdout(), "   (new session started)")
	}
	fmt.Fprintln(cmd.OutOrStdout())

	if focusRepo != "" && focusNumber != 0 {
		title := focusIssue.Title
		if title == "" {
			title = "(title unavailable)"
		}
		focusRef := issueRef(focusNumber, "")
		if focusIssue.Title != "" {
			// Build URL from repo + number
			focusURL := fmt.Sprintf("https://github.com/%s/issues/%d", focusRepo, focusNumber)
			focusRef = issueRef(focusNumber, focusURL)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "📍 Focus: %s \"%s\" (%s)\n", focusRef, title, focusRepo)
		if len(focusHandoffs) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "   Last handoff (%s): %s\n", humanizeDuration(time.Since(focusHandoffs[0].Time)), focusIssue.LastHandoff)
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "📍 Focus: none")
	}

	if nextUp != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "📋 Next up: %s \"%s\" (%s)\n", issueRef(nextUp.Number, nextUp.URL), nextUp.Title, nextUp.Repository)
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintf(cmd.OutOrStdout(), "📊 Project #%d Status:\n", pc.Project)
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
		fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", strings.Join(statusParts, " | "))
	}

	if len(recentLogs) > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "📝 Recent Logs:")
		for _, entry := range recentLogs {
			age := humanizeDuration(time.Since(entry.Time))
			fmt.Fprintf(cmd.OutOrStdout(), "   • [%s] %s (%s, %s)\n", entry.Kind, entry.Message, entry.Issue, age)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "🔄 Recent Decisions:")
	if len(decisions) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "   • none")
	} else {
		for _, decision := range decisions {
			fmt.Fprintf(cmd.OutOrStdout(), "   • %s (%s)\n", decision.Decision, decision.Issue)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "👀 Needs Review:")
	if len(reviewRequests) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "   • none")
	} else {
		for _, review := range reviewRequests {
			fmt.Fprintf(cmd.OutOrStdout(), "   • PR %s: %s (@%s)\n", issueRef(review.Number, review.URL), review.Title, review.Author)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "🚫 Blocked:")
	if len(blockedItems) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "   • none")
	} else {
		for _, blocked := range blockedItems {
			fmt.Fprintf(cmd.OutOrStdout(), "   • %s: %s\n", issueRef(blocked.Number, blocked.URL), blocked.Title)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "⚙️ Config:")
	fmt.Fprintf(cmd.OutOrStdout(), "   • default-owner: %s\n", pc.Cfg.DefaultOwner)
	fmt.Fprintf(cmd.OutOrStdout(), "   • default-project: %d\n", pc.Cfg.DefaultProject)
	if len(pc.Cfg.Team) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "   • team: %s\n", strings.Join(pc.Cfg.Team, ", "))
	}

	if agentContextOpts.NewSession {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "💡 Commands available:")
		fmt.Fprintln(cmd.OutOrStdout(), "   gh planning claim <issue>     — claim and start work")
		fmt.Fprintln(cmd.OutOrStdout(), "   gh planning log \"message\"     — log progress")
		fmt.Fprintln(cmd.OutOrStdout(), "   gh planning handoff <issue>   — structured handoff")
		fmt.Fprintln(cmd.OutOrStdout(), "   gh planning complete <issue>  — mark work done")
		fmt.Fprintln(cmd.OutOrStdout(), "   gh planning queue             — find more work")
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
