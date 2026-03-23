package cmd

import (
	"fmt"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type teamSummary struct {
	User         string    `json:"user"`
	Since        time.Time `json:"since"`
	PRsMerged    int       `json:"prsMerged"`
	IssuesClosed int       `json:"issuesClosed"`
	PRsOpen      int       `json:"prsOpen"`
	PRsInReview  int       `json:"prsInReview"`
	PRsDraft     int       `json:"prsDraft"`
	Reviews      int       `json:"reviews"`
	LastActive   time.Time `json:"lastActive"`
}

var teamOpts struct {
	Team  string
	Since string
	Quiet bool
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Show recent activity for your team",
	RunE:  runTeam,
}

func init() {
	teamCmd.Flags().StringVar(&teamOpts.Team, "team", "", "Comma-separated team usernames")
	teamCmd.Flags().StringVar(&teamOpts.Since, "since", "7d", "Lookback duration")
	teamCmd.Flags().BoolVar(&teamOpts.Quiet, "quiet", false, "Only show inactive teammates")
}

func runTeam(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	users := []string{}
	if teamOpts.Team != "" {
		users = splitAndTrim(teamOpts.Team)
	} else {
		users = append(users, cfg.Team...)
	}
	if len(users) == 0 {
		return fmt.Errorf("no team configured")
	}

	sinceDuration, err := parseDuration(teamOpts.Since)
	if err != nil {
		return fmt.Errorf("invalid since duration: %w", err)
	}
	if sinceDuration == 0 {
		sinceDuration = 7 * 24 * time.Hour
	}
	sinceTime := time.Now().Add(-sinceDuration)

	summaries := []teamSummary{}
	for _, user := range users {
		activity, err := github.FetchActivity(cmd.Context(), user, sinceTime)
		if err != nil {
			return err
		}
		summary := teamSummary{
			User:         user,
			Since:        sinceTime,
			PRsMerged:    len(activity.PRsMerged),
			IssuesClosed: len(activity.IssuesClosed),
			PRsOpen:      len(activity.PRsOpen),
			PRsInReview:  len(activity.PRsInReview),
			PRsDraft:     len(activity.PRsDraft),
			Reviews:      len(activity.Reviews),
			LastActive:   activity.LastActivity,
		}
		if teamOpts.Quiet && isRecentlyActive(summary.LastActive, 72*time.Hour) {
			continue
		}
		summaries = append(summaries, summary)
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"since":   sinceTime,
			"reports": summaries,
		}
		return output.PrintJSON(cmd.OutOrStdout(), payload, OutputOptions())
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "👥 Team Dashboard — last %s\n\n", teamOpts.Since)
	for idx, summary := range summaries {
		fmt.Fprintf(w, "@%s\n", summary.User)
		if summary.PRsMerged == 0 && summary.IssuesClosed == 0 && summary.PRsOpen == 0 && summary.PRsInReview == 0 && summary.Reviews == 0 {
			inactiveLabel := inactiveDuration(summary.LastActive)
			fmt.Fprintf(w, "  ⚠️ No activity in %s\n", inactiveLabel)
		} else {
			fmt.Fprintf(w, "  ✅ %d PRs merged, %d issues closed\n", summary.PRsMerged, summary.IssuesClosed)
			line := fmt.Sprintf("  🔵 %d PRs open", summary.PRsOpen)
			if summary.PRsInReview > 0 {
				line = fmt.Sprintf("%s, %d in review", line, summary.PRsInReview)
			}
			if summary.PRsDraft > 0 {
				line = fmt.Sprintf("%s (%d draft)", line, summary.PRsDraft)
			}
			fmt.Fprintln(w, line)
		}
		fmt.Fprintf(w, "  📅 Last active: %s\n", formatLastActive(summary.LastActive))
		if idx < len(summaries)-1 {
			fmt.Fprintln(w)
		}
	}
	return nil
}

func formatLastActive(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	elapsed := time.Since(t)
	if elapsed > 72*time.Hour {
		return t.Format("Mon Jan 2")
	}
	return humanizeDuration(elapsed)
}

func inactiveDuration(t time.Time) string {
	if t.IsZero() {
		return "the last 7 days"
	}
	elapsed := time.Since(t)
	if elapsed < 24*time.Hour {
		return "the last day"
	}
	days := int(elapsed.Hours() / 24)
	return fmt.Sprintf("%d days", days)
}

func isRecentlyActive(t time.Time, window time.Duration) bool {
	if t.IsZero() {
		return false
	}
	return time.Since(t) <= window
}
