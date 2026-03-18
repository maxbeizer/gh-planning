package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type pulseMember struct {
	User             string    `json:"user"`
	Since            time.Time `json:"since"`
	PRsMerged        int       `json:"prsMerged"`
	PRsPerWeek       float64   `json:"prsPerWeek"`
	CycleTimeDays    float64   `json:"cycleTimeDays"`
	ReviewTimeHours  float64   `json:"reviewTimeHours"`
	Blocked          int       `json:"blocked"`
	Flagged          bool      `json:"flagged"`
	VelocityBelowAvg bool      `json:"velocityBelowAvg"`
	ReviewAboveAvg   bool      `json:"reviewAboveAvg"`
}

type pulseSummary struct {
	TeamAveragePRsPerWeek      float64 `json:"teamAveragePrsPerWeek"`
	TeamAverageCycleTimeDays   float64 `json:"teamAverageCycleTimeDays"`
	TeamAverageReviewTimeHours float64 `json:"teamAverageReviewTimeHours"`
}

var pulseOpts struct {
	Team  string
	Since string
}

var pulseCmd = &cobra.Command{
	Use:   "pulse",
	Short: "Show team health metrics",
	RunE:  runPulse,
}

func init() {
	pulseCmd.Flags().StringVar(&pulseOpts.Team, "team", "", "Comma-separated team usernames")
	pulseCmd.Flags().StringVar(&pulseOpts.Since, "since", "30d", "Lookback duration")
}

func runPulse(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	users := []string{}
	if pulseOpts.Team != "" {
		users = splitTeam(pulseOpts.Team)
	} else {
		users = append(users, cfg.Team...)
	}
	if len(users) == 0 {
		return fmt.Errorf("no team configured")
	}

	sinceDuration, err := parseDuration(pulseOpts.Since)
	if err != nil {
		return fmt.Errorf("invalid since duration: %w", err)
	}
	if sinceDuration == 0 {
		sinceDuration = 30 * 24 * time.Hour
	}
	sinceTime := time.Now().Add(-sinceDuration)
	weeks := sinceDuration.Hours() / (24 * 7)
	if weeks < 1 {
		weeks = 1
	}

	members := []pulseMember{}
	for _, user := range users {
		activity, err := github.FetchActivity(cmd.Context(), user, sinceTime)
		if err != nil {
			return err
		}
		cycleTime := averageCycleTime(activity.PRsMerged)
		reviewTime := averageReviewTime(activity.Reviews)
		member := pulseMember{
			User:            user,
			Since:           sinceTime,
			PRsMerged:       len(activity.PRsMerged),
			PRsPerWeek:      float64(len(activity.PRsMerged)) / weeks,
			CycleTimeDays:   cycleTime.Hours() / 24,
			ReviewTimeHours: reviewTime.Hours(),
			Blocked:         len(activity.Blocked),
		}
		members = append(members, member)
	}

	summary := pulseSummary{}
	for _, member := range members {
		summary.TeamAveragePRsPerWeek += member.PRsPerWeek
		summary.TeamAverageCycleTimeDays += member.CycleTimeDays
		summary.TeamAverageReviewTimeHours += member.ReviewTimeHours
	}
	count := float64(len(members))
	if count > 0 {
		summary.TeamAveragePRsPerWeek /= count
		summary.TeamAverageCycleTimeDays /= count
		summary.TeamAverageReviewTimeHours /= count
	}

	for i := range members {
		velocityBelow := members[i].PRsPerWeek < summary.TeamAveragePRsPerWeek
		reviewAbove := false
		if summary.TeamAverageReviewTimeHours > 0 {
			reviewAbove = members[i].ReviewTimeHours > summary.TeamAverageReviewTimeHours
		}
		members[i].VelocityBelowAvg = velocityBelow
		members[i].ReviewAboveAvg = reviewAbove
		members[i].Flagged = velocityBelow || reviewAbove
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"since":    sinceTime,
			"members":  members,
			"summary":  summary,
			"duration": sinceDuration.String(),
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "💓 Team Pulse — last %s\n\n", pulseOpts.Since)
	fmt.Fprintf(w, "%-12s %-7s %-11s %-11s %-7s\n", "", "PRs/wk", "Cycle Time", "Review Time", "Blocked")
	for _, member := range members {
		flag := ""
		if member.Flagged {
			flag = " ⚠️"
		}
		fmt.Fprintf(w, "@%-11s %-7.1f %-11s %-11s %-7d%s\n", member.User, member.PRsPerWeek, formatDays(member.CycleTimeDays), formatHours(member.ReviewTimeHours), member.Blocked, flag)
	}

	fmt.Fprintln(w)
	fmt.Fprintf(w, "📊 Team Average: %.1f PRs/wk, %s cycle, %s reviews\n", summary.TeamAveragePRsPerWeek, formatDays(summary.TeamAverageCycleTimeDays), formatHours(summary.TeamAverageReviewTimeHours))
	flags := []string{}
	for _, member := range members {
		if member.Flagged {
			reasons := []string{}
			if member.VelocityBelowAvg {
				reasons = append(reasons, "velocity")
			}
			if member.ReviewAboveAvg {
				reasons = append(reasons, "reviews")
			}
			flags = append(flags, fmt.Sprintf("@%s below average on %s", member.User, strings.Join(reasons, " + ")))
		}
	}
	if len(flags) > 0 {
		fmt.Fprintf(w, "🚩 Flags: %s\n", strings.Join(flags, "; "))
	}
	return nil
}

func averageCycleTime(items []github.SearchIssue) time.Duration {
	if len(items) == 0 {
		return 0
	}
	var total time.Duration
	count := 0
	for _, item := range items {
		if item.CreatedAt.IsZero() || item.ClosedAt.IsZero() {
			continue
		}
		total += item.ClosedAt.Sub(item.CreatedAt)
		count++
	}
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

func averageReviewTime(items []github.SearchIssue) time.Duration {
	if len(items) == 0 {
		return 0
	}
	var total time.Duration
	count := 0
	for _, item := range items {
		if item.CreatedAt.IsZero() || item.UpdatedAt.IsZero() {
			continue
		}
		total += item.UpdatedAt.Sub(item.CreatedAt)
		count++
	}
	if count == 0 {
		return 0
	}
	return total / time.Duration(count)
}

func formatDays(days float64) string {
	if days == 0 {
		return "0d"
	}
	return fmt.Sprintf("%.1fd", days)
}

func formatHours(hours float64) string {
	if hours == 0 {
		return "0h"
	}
	return fmt.Sprintf("%.0fh", hours)
}
