package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/preps"
	"github.com/spf13/cobra"
)

type prepReport struct {
	User            string             `json:"user"`
	Since           time.Time          `json:"since"`
	Generated       time.Time          `json:"generatedAt"`
	Wins            []prepItem         `json:"wins"`
	CurrentWork     []prepItem         `json:"currentWork"`
	ReviewRequests  []prepItem         `json:"reviewRequests"`
	Blockers        []prepItem         `json:"blockers"`
	FollowUps       []string           `json:"followUps"`
	SuggestedTopics []string           `json:"suggestedTopics"`
	NotesPath       string             `json:"notesPath,omitempty"`
}

type prepItem struct {
	Title      string    `json:"title"`
	Number     int       `json:"number"`
	Repo       string    `json:"repo"`
	URL        string    `json:"url"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
	StateLabel string    `json:"stateLabel,omitempty"`
}

var prepOpts struct {
	Since string
	Notes bool
}

var prepCmd = &cobra.Command{
	Use:   "prep <github-handle>",
	Short: "Generate a 1-1 prep document",
	Args:  cobra.ExactArgs(1),
	RunE:  runPrep,
}

func init() {
	prepCmd.Flags().StringVar(&prepOpts.Since, "since", "14d", "Lookback duration")
	prepCmd.Flags().BoolVar(&prepOpts.Notes, "notes", false, "Open or create meeting notes")
}

func runPrep(cmd *cobra.Command, args []string) error {
	handle := strings.TrimPrefix(args[0], "@")
	if handle == "" {
		return fmt.Errorf("github handle is required")
	}

	sinceDuration, err := parseDuration(prepOpts.Since)
	if err != nil {
		return fmt.Errorf("invalid since duration: %w", err)
	}
	if sinceDuration == 0 {
		sinceDuration = 14 * 24 * time.Hour
	}
	sinceTime := time.Now().Add(-sinceDuration)

	activity, err := github.FetchActivity(cmd.Context(), handle, sinceTime)
	if err != nil {
		return err
	}

	wins := prepWins(activity)
	currentWork := prepCurrentWork(activity)
	reviewRequests := prepReviewRequests(activity)
	blockers := prepBlockers(activity)

	previous, err := preps.Latest(handle)
	if err != nil {
		return err
	}
	followUps := []string{}
	if previous != nil {
		followUps = previous.FollowUps
	}
	suggestedTopics := prepSuggestedTopics(activity, blockers)

	notesPath := ""
	if prepOpts.Notes {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		path, err := prepNotesPath(cfg, handle)
		if err != nil {
			return err
		}
		notesPath = path
	}

	report := prepReport{
		User:            handle,
		Since:           sinceTime,
		Generated:       time.Now().UTC(),
		Wins:            wins,
		CurrentWork:     currentWork,
		ReviewRequests:  reviewRequests,
		Blockers:        blockers,
		FollowUps:       followUps,
		SuggestedTopics: suggestedTopics,
		NotesPath:       notesPath,
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(report, OutputOptions())
	}

	prepDoc := renderPrepDoc(report)
	prepPath, err := preps.Save(handle, time.Now(), prepDoc)
	if err != nil {
		return err
	}
	if prepOpts.Notes && notesPath != "" {
		if err := openNotes(notesPath); err != nil {
			return err
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), prepDoc)
	fmt.Fprintf(cmd.OutOrStdout(), "\nSaved prep to %s\n", prepPath)
	if previous != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Previous prep: %s\n", previous.Path)
	}
	return nil
}

func prepWins(activity github.Activity) []prepItem {
	items := []prepItem{}
	for _, item := range activity.PRsMerged {
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item)})
	}
	for _, item := range activity.IssuesClosed {
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item)})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Repo < items[j].Repo })
	return items
}

func prepCurrentWork(activity github.Activity) []prepItem {
	items := []prepItem{}
	draftSet := map[string]struct{}{}
	for _, item := range activity.PRsDraft {
		draftSet[github.IssueURL(item)] = struct{}{}
	}
	for _, item := range activity.PRsOpen {
		label := "open"
		if _, ok := draftSet[github.IssueURL(item)]; ok {
			label = "draft"
		}
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item), UpdatedAt: item.UpdatedAt, StateLabel: label})
	}
	for _, item := range activity.AssignedIssues {
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item), UpdatedAt: item.UpdatedAt, StateLabel: "assigned"})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items
}

func prepReviewRequests(activity github.Activity) []prepItem {
	items := []prepItem{}
	for _, item := range activity.ReviewRequests {
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item), UpdatedAt: item.UpdatedAt, StateLabel: "review"})
	}
	return items
}

func prepBlockers(activity github.Activity) []prepItem {
	items := []prepItem{}
	for _, item := range activity.Blocked {
		items = append(items, prepItem{Title: item.Title, Number: item.Number, Repo: github.RepositoryNameFromURL(item.RepositoryURL), URL: github.IssueURL(item), UpdatedAt: item.UpdatedAt, StateLabel: "blocked"})
	}
	stale := staleItems(activity, 7*24*time.Hour)
	for _, item := range stale {
		items = append(items, item)
	}
	return items
}

func staleItems(activity github.Activity, threshold time.Duration) []prepItem {
	items := []prepItem{}
	cutoff := time.Now().Add(-threshold)
	appendStale := func(issue github.SearchIssue) {
		if issue.UpdatedAt.IsZero() || issue.UpdatedAt.After(cutoff) {
			return
		}
		items = append(items, prepItem{Title: issue.Title, Number: issue.Number, Repo: github.RepositoryNameFromURL(issue.RepositoryURL), URL: github.IssueURL(issue), UpdatedAt: issue.UpdatedAt, StateLabel: "stale"})
	}
	for _, item := range activity.PRsOpen {
		appendStale(item)
	}
	for _, item := range activity.AssignedIssues {
		appendStale(item)
	}
	return items
}

func prepSuggestedTopics(activity github.Activity, blockers []prepItem) []string {
	topics := []string{}
	if len(activity.PRsMerged)+len(activity.IssuesClosed) > 0 {
		topics = append(topics, fmt.Sprintf("Acknowledge %d wins from the last %s", len(activity.PRsMerged)+len(activity.IssuesClosed), prepOpts.Since))
	}
	for _, item := range blockers {
		if item.StateLabel == "blocked" {
			topics = append(topics, fmt.Sprintf("Check on blocker: #%d (%s)", item.Number, item.Repo))
			break
		}
	}
	for _, item := range activity.PRsOpen {
		if time.Since(item.CreatedAt) > 7*24*time.Hour {
			age := strings.TrimSuffix(humanizeDuration(time.Since(item.CreatedAt)), " ago")
			topics = append(topics, fmt.Sprintf("PR #%d has been open %s — any blockers?", item.Number, age))
			break
		}
	}
	if len(topics) == 0 {
		topics = append(topics, "Quick check-in: anything you need help with this week?")
	}
	return topics
}

func renderPrepDoc(report prepReport) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("📋 1-1 Prep: @%s — %s\n\n", report.User, time.Now().Format("Mon Jan 2, 2006")))
	builder.WriteString(fmt.Sprintf("🏆 Wins (last %s)\n", prepOpts.Since))
	if len(report.Wins) == 0 {
		builder.WriteString("  • None\n\n")
	} else {
		for _, item := range report.Wins {
			builder.WriteString(fmt.Sprintf("  • %s #%d: %s (%s)\n", winLabel(item.URL), item.Number, item.Title, item.Repo))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("🔵 Current Work\n")
	if len(report.CurrentWork) == 0 {
		builder.WriteString("  • None\n\n")
	} else {
		for _, item := range report.CurrentWork {
			builder.WriteString(fmt.Sprintf("  • %s #%d: %s (%s) — %s\n", workLabel(item.URL), item.Number, item.Title, item.Repo, currentWorkDetail(item)))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("👀 Review Requests\n")
	if len(report.ReviewRequests) == 0 {
		builder.WriteString("  • None\n\n")
	} else {
		for _, item := range report.ReviewRequests {
			builder.WriteString(fmt.Sprintf("  • PR #%d: %s (%s)\n", item.Number, item.Title, item.Repo))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("🚫 Blockers\n")
	if len(report.Blockers) == 0 {
		builder.WriteString("  • None\n\n")
	} else {
		for _, item := range report.Blockers {
			builder.WriteString(fmt.Sprintf("  • %s #%d: %s (%s)\n", blockerLabel(item), item.Number, item.Title, item.Repo))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("🔄 Follow-ups from Last Time\n")
	if len(report.FollowUps) == 0 {
		builder.WriteString("  (none found — first prep)\n\n")
	} else {
		for _, item := range report.FollowUps {
			builder.WriteString(fmt.Sprintf("  • %s\n", item))
		}
		builder.WriteString("\n")
	}

	builder.WriteString("💬 Suggested Topics\n")
	for _, topic := range report.SuggestedTopics {
		builder.WriteString(fmt.Sprintf("  • %s\n", topic))
	}

	return builder.String()
}

func winLabel(url string) string {
	if strings.Contains(strings.ToLower(url), "/pull/") {
		return "Merged PR"
	}
	return "Closed issue"
}

func workLabel(url string) string {
	if strings.Contains(strings.ToLower(url), "/pull/") {
		return "PR"
	}
	return "Issue"
}

func blockerLabel(item prepItem) string {
	if item.StateLabel == "stale" {
		return "Stale"
	}
	return "Blocked"
}

func currentWorkDetail(item prepItem) string {
	if item.UpdatedAt.IsZero() {
		return "updated recently"
	}
	age := humanizeDuration(time.Since(item.UpdatedAt))
	if item.StateLabel == "draft" {
		return fmt.Sprintf("draft, updated %s", age)
	}
	return fmt.Sprintf("updated %s", age)
}

func prepNotesPath(cfg *config.Config, handle string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	folder := filepath.Join(base, "gh-planning", "notes")
	if cfg.OneOnOneRepoPattern != "" {
		repo := strings.ReplaceAll(cfg.OneOnOneRepoPattern, "{handle}", handle)
		safeRepo := strings.ReplaceAll(repo, "/", "_")
		folder = filepath.Join(base, "gh-planning", "notes", safeRepo)
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(folder, fmt.Sprintf("%s-%s.md", handle, time.Now().Format("2006-01-02")))
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(fmt.Sprintf("# 1-1 Notes: @%s\n\n", handle)), 0o644); err != nil {
				return "", err
			}
		}
	}
	return path, nil
}

func openNotes(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
