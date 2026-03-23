package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type standupData struct {
	User       string             `json:"user"`
	Since      time.Time          `json:"since"`
	Generated  time.Time          `json:"generatedAt"`
	Done       []standupItem      `json:"done"`
	InProgress []standupItem      `json:"inProgress"`
	Blocked    []standupItem      `json:"blocked"`
	InReview   []standupItem      `json:"inReview"`
}

type standupItem struct {
	Title     string    `json:"title"`
	Number    int       `json:"number"`
	Repo      string    `json:"repo"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

var standupOpts struct {
	Project int
	Owner   string
	Since   string
	Team    bool
}

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate a standup report",
	RunE:  runStandup,
}

func init() {
	standupCmd.Flags().IntVar(&standupOpts.Project, "project", 0, "Project number")
	standupCmd.Flags().StringVar(&standupOpts.Owner, "owner", "", "Project owner")
	standupCmd.Flags().StringVar(&standupOpts.Since, "since", "24h", "Lookback duration")
	standupCmd.Flags().BoolVar(&standupOpts.Team, "team", false, "Include team members")
}

func runStandup(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(standupOpts.Owner, standupOpts.Project)
	if err != nil {
		return err
	}

	sinceDuration, err := parseDuration(standupOpts.Since)
	if err != nil {
		return fmt.Errorf("invalid since duration: %w", err)
	}
	if sinceDuration == 0 {
		sinceDuration = 24 * time.Hour
	}
	sinceTime := time.Now().Add(-sinceDuration)

	users := []string{}
	if standupOpts.Team {
		users = append(users, pc.Cfg.Team...)
		if len(users) == 0 {
			return fmt.Errorf("no team configured")
		}
	} else {
		current, err := github.CurrentUser(cmd.Context())
		if err != nil {
			return err
		}
		users = append(users, current)
	}

	projectData, err := github.GetProject(cmd.Context(), pc.Owner, pc.Project)
	if err != nil {
		return err
	}

	results := []standupData{}
	for _, user := range users {
		data, err := buildStandup(cmd.Context(), user, sinceTime, projectData)
		if err != nil {
			return err
		}
		results = append(results, data)
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"since":     sinceTime,
			"generated": time.Now().UTC(),
			"team":      standupOpts.Team,
			"reports":   results,
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "📋 Standup — %s\n\n", time.Now().Format("Mon Jan 2, 2006"))
	for idx, report := range results {
		if standupOpts.Team {
			fmt.Fprintf(cmd.OutOrStdout(), "👤 @%s\n\n", report.User)
		}
		printStandupReport(cmd.OutOrStdout(), report, sinceDuration)
		if idx < len(results)-1 {
			fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	return nil
}

func buildStandup(ctx context.Context, user string, since time.Time, project *github.Project) (standupData, error) {
	queryDate := since.Format(time.RFC3339)

	type searchResult struct {
		items []github.SearchIssue
		err   error
	}

	mergedCh := make(chan searchResult, 1)
	closedCh := make(chan searchResult, 1)
	reviewCh := make(chan searchResult, 1)

	go func() {
		items, err := github.SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:merged merged:>%s", user, queryDate))
		mergedCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(ctx, fmt.Sprintf("author:%s type:issue is:closed closed:>%s", user, queryDate))
		closedCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(ctx, fmt.Sprintf("author:%s type:pr is:open review:required", user))
		reviewCh <- searchResult{items, err}
	}()

	merged := <-mergedCh
	if merged.err != nil {
		return standupData{}, merged.err
	}
	closed := <-closedCh
	if closed.err != nil {
		return standupData{}, closed.err
	}
	review := <-reviewCh
	if review.err != nil {
		return standupData{}, review.err
	}

	mergedPRs := merged.items
	closedIssues := closed.items
	inReview := review.items
	inProgress := filterProjectByStatus(project, user, "in progress")
	blocked := filterProjectByStatus(project, user, "blocked")

	data := standupData{User: user, Since: since, Generated: time.Now().UTC()}
	for _, item := range mergedPRs {
		data.Done = append(data.Done, standupItem{
			Title:  item.Title,
			Number: item.Number,
			Repo:   github.RepositoryNameFromURL(item.RepositoryURL),
			URL:    github.IssueURL(item),
		})
	}
	for _, item := range closedIssues {
		data.Done = append(data.Done, standupItem{
			Title:  item.Title,
			Number: item.Number,
			Repo:   github.RepositoryNameFromURL(item.RepositoryURL),
			URL:    github.IssueURL(item),
		})
	}
	for _, item := range inReview {
		data.InReview = append(data.InReview, standupItem{
			Title:  item.Title,
			Number: item.Number,
			Repo:   github.RepositoryNameFromURL(item.RepositoryURL),
			URL:    github.IssueURL(item),
		})
	}
	for _, item := range inProgress {
		data.InProgress = append(data.InProgress, standupItem{
			Title:     item.Title,
			Number:    item.Number,
			Repo:      item.Repository,
			URL:       item.URL,
			UpdatedAt: item.UpdatedAt,
		})
	}
	for _, item := range blocked {
		data.Blocked = append(data.Blocked, standupItem{
			Title:     item.Title,
			Number:    item.Number,
			Repo:      item.Repository,
			URL:       item.URL,
			UpdatedAt: item.UpdatedAt,
		})
	}

	sort.Slice(data.Done, func(i, j int) bool { return data.Done[i].Number < data.Done[j].Number })
	sort.Slice(data.InProgress, func(i, j int) bool { return data.InProgress[i].Number < data.InProgress[j].Number })
	sort.Slice(data.Blocked, func(i, j int) bool { return data.Blocked[i].Number < data.Blocked[j].Number })
	sort.Slice(data.InReview, func(i, j int) bool { return data.InReview[i].Number < data.InReview[j].Number })
	return data, nil
}

func filterProjectByStatus(project *github.Project, assignee string, status string) []github.ProjectItem {
	items := []github.ProjectItem{}
	for key, group := range project.Items {
		if !strings.Contains(strings.ToLower(key), status) {
			continue
		}
		for _, item := range group {
			if assignee == "" {
				items = append(items, item)
				continue
			}
			for _, a := range item.Assignees {
				if strings.EqualFold(a, assignee) {
					items = append(items, item)
					break
				}
			}
		}
	}
	return items
}

func printStandupReport(w io.Writer, report standupData, since time.Duration) {
	fmt.Fprintf(w, "✅ Done (since %s)\n", humanizeSinceLabel(since))
	if len(report.Done) == 0 {
		fmt.Fprintln(w, "  • None")
	} else {
		for _, item := range report.Done {
			label := "Closed"
			if strings.Contains(strings.ToLower(item.URL), "/pull/") || strings.Contains(strings.ToLower(item.URL), "pull") {
				label = "Merged PR"
			}
			fmt.Fprintf(w, "  • %s %s: %s (%s)\n", label, issueRef(item.Number, item.URL), item.Title, item.Repo)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "🔵 In Progress")
	if len(report.InProgress) == 0 {
		fmt.Fprintln(w, "  • None")
	} else {
		for _, item := range report.InProgress {
			fmt.Fprintf(w, "  • %s: %s (%s) — %s\n", issueRef(item.Number, item.URL), item.Title, item.Repo, activeLabel(item.UpdatedAt))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "🚫 Blocked")
	if len(report.Blocked) == 0 {
		fmt.Fprintln(w, "  • None")
	} else {
		for _, item := range report.Blocked {
			fmt.Fprintf(w, "  • %s: %s (%s) — %s\n", issueRef(item.Number, item.URL), item.Title, item.Repo, activeLabel(item.UpdatedAt))
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "🔍 In Review")
	if len(report.InReview) == 0 {
		fmt.Fprintln(w, "  • None")
	} else {
		for _, item := range report.InReview {
			fmt.Fprintf(w, "  • PR %s: %s (%s) — awaiting review\n", issueRef(item.Number, item.URL), item.Title, item.Repo)
		}
	}
}

func humanizeSinceLabel(d time.Duration) string {
	if d >= 48*time.Hour {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	return "yesterday"
}

func activeLabel(updatedAt time.Time) string {
	if updatedAt.IsZero() {
		return "updated recently"
	}
	elapsed := time.Since(updatedAt)
	if elapsed < 24*time.Hour {
		return "started today"
	}
	days := int(elapsed.Hours() / 24)
	return fmt.Sprintf("%dd active", days)
}
