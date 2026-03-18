package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/session"
	"github.com/spf13/cobra"
)

var daemonOpts struct {
	Interval time.Duration
	DryRun   bool
	Once     bool
	Label    string
	Status   []string
	Project  int
	Owner    string
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Autonomously poll the queue, claim issues, and dispatch work",
	Long: `Run a daemon that polls the agent queue on a regular interval,
claims the highest-priority item, outputs agent context, and waits
for the next cycle. Use --dry-run to preview without claiming.`,
	RunE: runDaemon,
}

func init() {
	daemonCmd.Flags().DurationVar(&daemonOpts.Interval, "interval", 5*time.Minute, "Poll interval (e.g. 30s, 2m, 5m)")
	daemonCmd.Flags().BoolVar(&daemonOpts.DryRun, "dry-run", false, "Show what would be claimed without acting")
	daemonCmd.Flags().BoolVar(&daemonOpts.Once, "once", false, "Process one item and exit")
	daemonCmd.Flags().StringVar(&daemonOpts.Label, "label", "", "Filter queue by label")
	daemonCmd.Flags().StringSliceVar(&daemonOpts.Status, "status", nil, "Filter queue by status (repeatable)")
	daemonCmd.Flags().IntVar(&daemonOpts.Project, "project", 0, "Project number")
	daemonCmd.Flags().StringVar(&daemonOpts.Owner, "owner", "", "Project owner")
}

// rateLimiter tracks claims-per-hour with a sliding window.
type rateLimiter struct {
	mu        sync.Mutex
	max       int
	timestamps []time.Time
}

func newRateLimiter(max int) *rateLimiter {
	return &rateLimiter{max: max}
}

func (r *rateLimiter) allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)
	filtered := r.timestamps[:0]
	for _, t := range r.timestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	r.timestamps = filtered
	return len(r.timestamps) < r.max
}

func (r *rateLimiter) record() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timestamps = append(r.timestamps, time.Now())
}

func (r *rateLimiter) remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)
	count := 0
	for _, t := range r.timestamps {
		if t.After(cutoff) {
			count++
		}
	}
	rem := r.max - count
	if rem < 0 {
		return 0
	}
	return rem
}

func runDaemon(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(daemonOpts.Owner, daemonOpts.Project)
	if err != nil {
		return err
	}

	maxPerHour := pc.Cfg.AgentMaxPerHour
	if maxPerHour <= 0 {
		maxPerHour = 3
	}

	statuses := daemonOpts.Status
	if len(statuses) == 0 {
		statuses = []string{"Backlog", "Ready"}
	}

	limiter := newRateLimiter(maxPerHour)

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	interval := daemonOpts.Interval
	modeLabel := "loop"
	if daemonOpts.Once {
		modeLabel = "once"
	}
	if daemonOpts.DryRun {
		modeLabel += ", dry-run"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "🤖 Daemon started (interval: %s, max: %d/hour, mode: %s)\n", interval, maxPerHour, modeLabel)

	for {
		fmt.Fprintf(cmd.OutOrStdout(), "⏳ Polling queue...\n")

		items, err := fetchQueue(ctx, pc.Owner, pc.Project, statuses, daemonOpts.Label)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Queue poll failed: %v\n", err)
			if daemonOpts.Once {
				return err
			}
			if sleepOrDone(ctx, interval) {
				break
			}
			continue
		}

		if len(items) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "😴 No items ready\n")
			if daemonOpts.Once {
				break
			}
			fmt.Fprintf(cmd.OutOrStdout(), "⏳ Next poll in %s...\n", interval)
			if sleepOrDone(ctx, interval) {
				break
			}
			continue
		}

		fmt.Fprintf(cmd.OutOrStdout(), "📋 Found %d item(s) ready\n", len(items))

		if !limiter.allow() {
			fmt.Fprintf(cmd.OutOrStdout(), "⏸️  Rate limit reached (%d/hour) — waiting\n", maxPerHour)
			if daemonOpts.Once {
				break
			}
			fmt.Fprintf(cmd.OutOrStdout(), "⏳ Next poll in %s...\n", interval)
			if sleepOrDone(ctx, interval) {
				break
			}
			continue
		}

		item := items[0]
		issueRef := fmt.Sprintf("%s#%d", item.Repo, item.Number)

		if daemonOpts.DryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "🔍 Would claim #%d: %s (%s)\n", item.Number, item.Title, item.Repo)
			if OutputOptions().JSON || OutputOptions().JQ != "" {
				payload := map[string]interface{}{
					"action": "dry-run",
					"issue":  issueRef,
					"number": item.Number,
					"title":  item.Title,
					"repo":   item.Repo,
				}
				if err := output.PrintJSON(payload, OutputOptions()); err != nil {
					return err
				}
			}
			if daemonOpts.Once {
				break
			}
			fmt.Fprintf(cmd.OutOrStdout(), "⏳ Next poll in %s...\n", interval)
			if sleepOrDone(ctx, interval) {
				break
			}
			continue
		}

		// Claim the issue
		fmt.Fprintf(cmd.OutOrStdout(), "🎯 Claiming #%d: %s (%s)\n", item.Number, item.Title, item.Repo)
		sessionID, err := claimIssue(ctx, cmd, pc.Owner, pc.Project, item)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Claim failed for #%d: %v\n", item.Number, err)
			if daemonOpts.Once {
				return err
			}
			if sleepOrDone(ctx, interval) {
				break
			}
			continue
		}
		limiter.record()

		// Dump agent context
		fmt.Fprintf(cmd.OutOrStdout(), "🤖 Agent context for #%d (session %s):\n", item.Number, sessionID)
		printAgentContextSummary(cmd, item)

		if OutputOptions().JSON || OutputOptions().JQ != "" {
			payload := map[string]interface{}{
				"action":    "claimed",
				"issue":     issueRef,
				"number":    item.Number,
				"title":     item.Title,
				"repo":      item.Repo,
				"session":   sessionID,
				"remaining": limiter.remaining(),
			}
			if err := output.PrintJSON(payload, OutputOptions()); err != nil {
				return err
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "✅ Ready for processing — %d claim(s) remaining this hour\n", limiter.remaining())

		if daemonOpts.Once {
			break
		}
		fmt.Fprintf(cmd.OutOrStdout(), "⏳ Next poll in %s...\n", interval)
		if sleepOrDone(ctx, interval) {
			break
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "👋 Daemon stopped\n")
	return nil
}

// fetchQueue retrieves queue items from the project, mirroring queue.go logic.
func fetchQueue(ctx context.Context, owner string, project int, statuses []string, label string) ([]queueItem, error) {
	projectData, err := github.GetProject(ctx, owner, project)
	if err != nil {
		return nil, err
	}

	var items []queueItem
	statusLookup := map[string]bool{}
	for _, s := range statuses {
		statusLookup[strings.ToLower(s)] = true
	}
	for status, projectItems := range projectData.Items {
		if !statusLookup[strings.ToLower(status)] {
			continue
		}
		for _, pi := range projectItems {
			if label != "" && !hasLabel(pi.Labels, label) {
				continue
			}
			priority := ""
			if pi.Fields != nil {
				priority = pi.Fields["Priority"]
			}
			createdAt := pi.CreatedAt
			if createdAt.IsZero() {
				createdAt = pi.UpdatedAt
			}
			items = append(items, queueItem{
				Number:    pi.Number,
				Title:     pi.Title,
				Repo:      pi.Repository,
				Priority:  priority,
				Labels:    pi.Labels,
				Status:    pi.Status,
				CreatedAt: createdAt,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		p1 := priorityRank(items[i].Priority)
		p2 := priorityRank(items[j].Priority)
		if p1 != p2 {
			return p1 < p2
		}
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

// claimIssue claims an issue and returns the session ID.
func claimIssue(ctx context.Context, cmd *cobra.Command, owner string, project int, item queueItem) (string, error) {
	repo := item.Repo
	number := item.Number

	issueOwner, issueRepo, err := splitRepo(repo)
	if err != nil {
		return "", err
	}

	sessionID := shortSessionID()
	stamp := formatTimestamp(time.Now())
	comment := fmt.Sprintf("🤖 Claimed by daemon session `%s` at %s", sessionID, stamp)

	if err := github.CreateIssueComment(ctx, issueOwner, issueRepo, number, comment); err != nil {
		return "", fmt.Errorf("comment: %w", err)
	}

	projectID, _, statusFieldID, statusOptions, err := github.GetProjectInfo(ctx, owner, project)
	if err != nil {
		return "", fmt.Errorf("project info: %w", err)
	}
	if statusFieldID == "" {
		return "", fmt.Errorf("status field not found on project")
	}
	optionID, ok := findStatusOption(statusOptions, "In Progress")
	if !ok {
		return "", fmt.Errorf("status option not found: In Progress")
	}
	itemID, err := findProjectItemID(ctx, owner, project, repo, number)
	if err != nil {
		return "", fmt.Errorf("item lookup: %w", err)
	}
	if err := github.UpdateItemStatus(ctx, projectID, itemID, statusFieldID, optionID); err != nil {
		return "", fmt.Errorf("status update: %w", err)
	}

	focus := &session.FocusSession{
		Issue:       fmt.Sprintf("%s#%d", repo, number),
		IssueNumber: number,
		Repo:        repo,
		StartedAt:   time.Now().UTC(),
		SessionID:   sessionID,
	}
	if err := session.SaveCurrent(focus); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to save focus session: %v\n", err)
	}

	return sessionID, nil
}

// printAgentContextSummary outputs a brief context summary for the claimed item.
func printAgentContextSummary(cmd *cobra.Command, item queueItem) {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "  Issue:    %s#%d\n", item.Repo, item.Number)
	fmt.Fprintf(w, "  Title:    %s\n", item.Title)
	fmt.Fprintf(w, "  Priority: %s\n", item.Priority)
	if len(item.Labels) > 0 {
		fmt.Fprintf(w, "  Labels:   %s\n", strings.Join(item.Labels, ", "))
	}
}

// sleepOrDone waits for the given duration or until the context is cancelled.
// Returns true if the context was cancelled (caller should exit the loop).
func sleepOrDone(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return true
	case <-time.After(d):
		return false
	}
}
