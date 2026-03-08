package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

type catchupSection struct {
	Title string        `json:"title"`
	Items []catchupItem `json:"items"`
}

type catchupItem struct {
	Title    string    `json:"title"`
	Number   int       `json:"number"`
	Repo     string    `json:"repo"`
	URL      string    `json:"url"`
	Author   string    `json:"author,omitempty"`
	Comments int       `json:"comments,omitempty"`
	Updated  time.Time `json:"updatedAt,omitempty"`
}

var catchupOpts struct {
	Since   string
	Project int
	Owner   string
}

var catchupCmd = &cobra.Command{
	Use:   "catch-up",
	Short: "Summarize updates since your last session",
	RunE:  runCatchup,
}

func init() {
	catchupCmd.Flags().StringVar(&catchupOpts.Since, "since", "", "Look back to a duration or date")
	catchupCmd.Flags().IntVar(&catchupOpts.Project, "project", 0, "Project number")
	catchupCmd.Flags().StringVar(&catchupOpts.Owner, "owner", "", "Project owner")
}

func runCatchup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := catchupOpts.Owner
	project := catchupOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}

	sinceTime, label, err := resolveSince(catchupOpts.Since)
	if err != nil {
		return err
	}
	if sinceTime.IsZero() {
		st, err := state.Load()
		if err != nil {
			return err
		}
		if !st.LastSeen.IsZero() {
			sinceTime = st.LastSeen
			label = formatSinceLabel(sinceTime)
		}
	}
	if sinceTime.IsZero() {
		sinceTime = time.Now().Add(-7 * 24 * time.Hour)
		label = formatSinceLabel(sinceTime)
	}

	// Parallel: CurrentUser + GetProject (independent)
	type userResult struct {
		user string
		err  error
	}
	type projectResult struct {
		data *github.Project
		err  error
	}

	userCh := make(chan userResult, 1)
	projectCh := make(chan projectResult, 1)

	go func() {
		user, err := github.CurrentUser(cmd.Context())
		userCh <- userResult{user, err}
	}()
	go func() {
		data, err := github.GetProject(cmd.Context(), owner, project)
		projectCh <- projectResult{data, err}
	}()

	ur := <-userCh
	if ur.err != nil {
		return ur.err
	}
	currentUser := ur.user

	pr := <-projectCh
	if pr.err != nil {
		return pr.err
	}
	projectData := pr.data

	repos := uniqueRepos(projectData)
	repoQuery := buildRepoQuery(repos)
	queryDate := sinceTime.Format(time.RFC3339)

	// Parallel: all 5 search queries (independent, only need currentUser + repoQuery + queryDate)
	type searchResult struct {
		items []github.SearchIssue
		err   error
	}

	newCh := make(chan searchResult, 1)
	mergedCh := make(chan searchResult, 1)
	closedCh := make(chan searchResult, 1)
	assignedCh := make(chan searchResult, 1)
	reviewCh := make(chan searchResult, 1)

	go func() {
		items, err := github.SearchIssues(cmd.Context(), composeQuery(repoQuery, fmt.Sprintf("is:issue created:>%s", queryDate)))
		newCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(cmd.Context(), composeQuery(repoQuery, fmt.Sprintf("is:pr is:merged merged:>%s", queryDate)))
		mergedCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(cmd.Context(), composeQuery(repoQuery, fmt.Sprintf("is:issue is:closed closed:>%s", queryDate)))
		closedCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(cmd.Context(), fmt.Sprintf("assignee:%s updated:>%s sort:updated", currentUser, queryDate))
		assignedCh <- searchResult{items, err}
	}()
	go func() {
		items, err := github.SearchIssues(cmd.Context(), fmt.Sprintf("review-requested:%s type:pr is:open updated:>%s", currentUser, queryDate))
		reviewCh <- searchResult{items, err}
	}()

	newResult := <-newCh
	if newResult.err != nil {
		return newResult.err
	}
	mergedResult := <-mergedCh
	if mergedResult.err != nil {
		return mergedResult.err
	}
	closedResult := <-closedCh
	if closedResult.err != nil {
		return closedResult.err
	}
	assignedResult := <-assignedCh
	if assignedResult.err != nil {
		return assignedResult.err
	}
	reviewResult := <-reviewCh
	if reviewResult.err != nil {
		return reviewResult.err
	}

	sections := []catchupSection{
		{Title: "📥 New", Items: formatCatchupItemsWithAuthor(newResult.items)},
		{Title: "✅ Merged", Items: formatCatchupItems(mergedResult.items)},
		{Title: "✅ Closed", Items: formatCatchupItems(closedResult.items)},
		{Title: "💬 Activity on your items", Items: filterCommentActivity(assignedResult.items)},
		{Title: "👀 Needs your review", Items: formatCatchupItems(reviewResult.items)},
	}

	if err := state.UpdateLastSeen(time.Now().UTC()); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"since":    sinceTime,
			"label":    label,
			"sections": sections,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Printf("📬 Catch-up — since %s\n\n", label)
	for _, section := range sections {
		if len(section.Items) == 0 {
			continue
		}
		fmt.Printf("%s (%d)\n", section.Title, len(section.Items))
		for _, item := range section.Items {
			line := fmt.Sprintf("  • #%d: %s (%s)", item.Number, item.Title, item.Repo)
			if item.Author != "" {
				line = fmt.Sprintf("%s — opened by @%s", line, item.Author)
			}
			if item.Comments > 0 {
				line = fmt.Sprintf("%s — %d new comments", line, item.Comments)
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	return nil
}

func resolveSince(value string) (time.Time, string, error) {
	if value == "" {
		return time.Time{}, "", nil
	}
	if dur, err := parseDuration(value); err == nil && dur > 0 {
		since := time.Now().Add(-dur)
		return since, formatSinceLabel(since), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, formatSinceLabel(t), nil
	}
	if day, ok := parseDayOfWeek(value); ok {
		return day, formatSinceLabel(day), nil
	}
	return time.Time{}, "", fmt.Errorf("invalid since value: %s", value)
}

func parseDayOfWeek(value string) (time.Time, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return time.Time{}, false
	}
	weekdays := map[string]time.Weekday{
		"sunday": time.Sunday,
		"monday": time.Monday,
		"tuesday": time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday": time.Thursday,
		"friday": time.Friday,
		"saturday": time.Saturday,
	}
	weekday, ok := weekdays[value]
	if !ok {
		return time.Time{}, false
	}
	now := time.Now()
	offset := int(now.Weekday()) - int(weekday)
	if offset <= 0 {
		offset += 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(-time.Duration(offset) * 24 * time.Hour), true
}

func formatSinceLabel(t time.Time) string {
	return t.Format("Mon Jan 2")
}

func uniqueRepos(project *github.Project) []string {
	repoSet := map[string]struct{}{}
	for _, items := range project.Items {
		for _, item := range items {
			if item.Repository != "" {
				repoSet[item.Repository] = struct{}{}
			}
		}
	}
	repos := make([]string, 0, len(repoSet))
	for repo := range repoSet {
		repos = append(repos, repo)
	}
	sort.Strings(repos)
	return repos
}

func buildRepoQuery(repos []string) string {
	if len(repos) == 0 {
		return ""
	}
	parts := make([]string, 0, len(repos))
	for _, repo := range repos {
		parts = append(parts, fmt.Sprintf("repo:%s", repo))
	}
	return strings.Join(parts, " ")
}

func composeQuery(prefix string, rest string) string {
	if prefix == "" {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(prefix + " " + rest)
}

func formatCatchupItems(items []github.SearchIssue) []catchupItem {
	output := []catchupItem{}
	for _, item := range items {
		output = append(output, catchupItem{
			Title:    item.Title,
			Number:   item.Number,
			Repo:     github.RepositoryNameFromURL(item.RepositoryURL),
			URL:      github.IssueURL(item),
			Comments: item.Comments,
			Updated:  item.UpdatedAt,
		})
	}
	return output
}

func formatCatchupItemsWithAuthor(items []github.SearchIssue) []catchupItem {
	output := []catchupItem{}
	for _, item := range items {
		output = append(output, catchupItem{
			Title:    item.Title,
			Number:   item.Number,
			Repo:     github.RepositoryNameFromURL(item.RepositoryURL),
			URL:      github.IssueURL(item),
			Author:   item.User.Login,
			Comments: item.Comments,
			Updated:  item.UpdatedAt,
		})
	}
	return output
}

func filterCommentActivity(items []github.SearchIssue) []catchupItem {
	filtered := []catchupItem{}
	for _, item := range items {
		if item.Comments == 0 {
			continue
		}
		filtered = append(filtered, catchupItem{
			Title:    item.Title,
			Number:   item.Number,
			Repo:     github.RepositoryNameFromURL(item.RepositoryURL),
			URL:      github.IssueURL(item),
			Comments: item.Comments,
			Updated:  item.UpdatedAt,
		})
	}
	return filtered
}
