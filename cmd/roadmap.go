package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

const defaultLookback = "28d"
const maxItemsPerSection = 5

var roadmapOpts struct {
	Project int
	Owner   string
	Since   string
}

var roadmapCmd = &cobra.Command{
	Use:   "roadmap",
	Short: "Show a vertical timeline of project activity",
	Long:  "Render a terminal timeline of project items grouped by time period, showing completed, in-progress, and upcoming work.",
	RunE:  runRoadmap,
}

func init() {
	roadmapCmd.Flags().IntVar(&roadmapOpts.Project, "project", 0, "Project number")
	roadmapCmd.Flags().StringVar(&roadmapOpts.Owner, "owner", "", "Project owner")
	roadmapCmd.Flags().StringVar(&roadmapOpts.Since, "since", defaultLookback, "Lookback period (e.g. 30d, 8w)")
}

// timelineBucket holds items for a single time period in the roadmap.
type timelineBucket struct {
	Label string               `json:"label"`
	Start time.Time            `json:"start"`
	End   time.Time            `json:"end"`
	Kind  string               `json:"kind"` // "completed", "in_progress", "upcoming"
	Items []github.ProjectItem `json:"items"`
}

type roadmapOutput struct {
	Title   string           `json:"title"`
	Owner   string           `json:"owner"`
	Project int              `json:"project"`
	Buckets []timelineBucket `json:"buckets"`
}

func runRoadmap(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(roadmapOpts.Owner, roadmapOpts.Project)
	if err != nil {
		return err
	}

	lookback, err := parseDuration(roadmapOpts.Since)
	if err != nil {
		return fmt.Errorf("invalid --since value: %w", err)
	}

	projectData, err := github.GetProject(cmd.Context(), pc.Owner, pc.Project)
	if err != nil {
		return err
	}

	now := time.Now()
	cutoff := now.Add(-lookback)
	buckets := buildTimeline(projectData, now, cutoff)

	data := roadmapOutput{
		Title:   projectData.Title,
		Owner:   pc.Owner,
		Project: pc.Project,
		Buckets: buckets,
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(data, OutputOptions())
	}

	printTimeline(data)
	return nil
}

func buildTimeline(project *github.Project, now time.Time, cutoff time.Time) []timelineBucket {
	var inProgress []github.ProjectItem
	var upcoming []github.ProjectItem
	var done []github.ProjectItem

	inProgressStatuses := map[string]bool{
		"in progress": true, "in review": true, "needs review": true, "needs my attention": true,
	}
	doneStatuses := map[string]bool{
		"done": true, "closed": true, "complete": true, "completed": true,
	}

	for _, items := range project.Items {
		for _, item := range items {
			lower := strings.ToLower(item.Status)
			switch {
			case inProgressStatuses[lower]:
				inProgress = append(inProgress, item)
			case doneStatuses[lower]:
				if item.UpdatedAt.After(cutoff) {
					done = append(done, item)
				}
			default:
				upcoming = append(upcoming, item)
			}
		}
	}

	var buckets []timelineBucket

	// "Now" bucket for in-progress items
	if len(inProgress) > 0 {
		buckets = append(buckets, timelineBucket{
			Label: formatCurrentPeriod(now),
			Start: weekStart(now),
			End:   now,
			Kind:  "in_progress",
			Items: inProgress,
		})
	}

	// Group done items by week
	weekBuckets := groupByWeek(done, now)
	buckets = append(buckets, weekBuckets...)

	// Upcoming bucket
	if len(upcoming) > 0 {
		buckets = append(buckets, timelineBucket{
			Label: "Upcoming",
			Start: now,
			End:   now,
			Kind:  "upcoming",
			Items: upcoming,
		})
	}

	return buckets
}

func groupByWeek(items []github.ProjectItem, now time.Time) []timelineBucket {
	if len(items) == 0 {
		return nil
	}

	type weekKey struct {
		year int
		week int
	}

	groups := make(map[weekKey][]github.ProjectItem)
	for _, item := range items {
		y, w := item.UpdatedAt.ISOWeek()
		key := weekKey{year: y, week: w}
		groups[key] = append(groups[key], item)
	}

	// Sort week keys descending (most recent first)
	keys := make([]weekKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].year != keys[j].year {
			return keys[i].year > keys[j].year
		}
		return keys[i].week > keys[j].week
	})

	var buckets []timelineBucket
	for _, key := range keys {
		weekItems := groups[key]
		sort.Slice(weekItems, func(i, j int) bool {
			return weekItems[i].UpdatedAt.After(weekItems[j].UpdatedAt)
		})

		start := isoWeekStart(key.year, key.week)
		end := start.AddDate(0, 0, 6)
		label := formatWeekRange(start, end, now)

		buckets = append(buckets, timelineBucket{
			Label: label,
			Start: start,
			End:   end,
			Kind:  "completed",
			Items: weekItems,
		})
	}

	return buckets
}

func formatCurrentPeriod(now time.Time) string {
	return now.Format("Jan 2006") + " (This Week)"
}

func formatWeekRange(start, end, now time.Time) string {
	_, currentWeek := now.ISOWeek()
	_, startWeek := start.ISOWeek()
	if start.Year() == now.Year() && startWeek == currentWeek {
		return start.Format("Jan 2006") + " (This Week)"
	}
	if start.Month() == end.Month() {
		return fmt.Sprintf("%s %d – %d", start.Format("Jan"), start.Day(), end.Day())
	}
	return fmt.Sprintf("%s %d – %s %d", start.Format("Jan"), start.Day(), end.Format("Jan"), end.Day())
}

func weekStart(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-(weekday-1), 0, 0, 0, 0, t.Location())
}

// isoWeekStart returns the Monday of the given ISO week.
func isoWeekStart(year, week int) time.Time {
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	// Monday of ISO week 1
	isoWeek1Monday := jan4.AddDate(0, 0, -(weekday - 1))
	return isoWeek1Monday.AddDate(0, 0, (week-1)*7)
}

func printTimeline(data roadmapOutput) {
	fmt.Printf("\n🗺️  Roadmap: %s\n\n", data.Title)

	for i, bucket := range data.Buckets {
		printTimelineBucket(bucket, i == 0)
	}
	fmt.Println()
}

func printTimelineBucket(bucket timelineBucket, isFirst bool) {
	headerLine := fmt.Sprintf("── %s ", bucket.Label)
	padding := 40 - runewidth.StringWidth(headerLine)
	if padding < 0 {
		padding = 0
	}
	headerLine += strings.Repeat("─", padding)

	fmt.Println(headerLine)

	switch bucket.Kind {
	case "in_progress":
		fmt.Println("🔵 In Progress")
	case "completed":
		fmt.Printf("✅ Completed (%d)\n", len(bucket.Items))
	case "upcoming":
		fmt.Printf("📋 Backlog (%d items)\n", len(bucket.Items))
	}

	limit := maxItemsPerSection
	for i, item := range bucket.Items {
		if i >= limit {
			fmt.Printf("   ... +%d more\n", len(bucket.Items)-limit)
			break
		}
		line := formatTimelineItem(item)
		fmt.Printf("   %s\n", line)
	}
	fmt.Println()
}

func formatTimelineItem(item github.ProjectItem) string {
	var parts []string
	if item.Number > 0 {
		parts = append(parts, issueRef(item.Number, item.URL))
	}
	parts = append(parts, item.Title)

	line := strings.Join(parts, " ")

	if len(item.Assignees) > 0 {
		line += " (@" + strings.Join(item.Assignees, ", @") + ")"
	}

	// Truncate to reasonable terminal width
	maxWidth := 72
	if runewidth.StringWidth(line) > maxWidth {
		line = runewidth.Truncate(line, maxWidth, "…")
	}

	return line
}
