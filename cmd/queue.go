package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type queueItem struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Repo      string    `json:"repo"`
	URL       string    `json:"url"`
	Priority  string    `json:"priority"`
	Labels    []string  `json:"labels"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

var queueOpts struct {
	Project int
	Owner   string
	Status  []string
	Label   string
	Limit   int
}

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Show items ready for agent processing",
	RunE:  runQueue,
}

func init() {
	queueCmd.Flags().IntVar(&queueOpts.Project, "project", 0, "Project number")
	queueCmd.Flags().StringVar(&queueOpts.Owner, "owner", "", "Project owner")
	queueCmd.Flags().StringSliceVar(&queueOpts.Status, "status", nil, "Status filter (repeatable)")
	queueCmd.Flags().StringVar(&queueOpts.Label, "label", "", "Filter by label")
	queueCmd.Flags().IntVar(&queueOpts.Limit, "limit", 10, "Max items")
}

func runQueue(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := queueOpts.Owner
	project := queueOpts.Project
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}
	statuses := queueOpts.Status
	if len(statuses) == 0 {
		statuses = []string{"Backlog", "Ready"}
	}

	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}

	readyItems := []queueItem{}
	statusLookup := map[string]bool{}
	for _, status := range statuses {
		statusLookup[strings.ToLower(status)] = true
	}
	for status, items := range projectData.Items {
		if !statusLookup[strings.ToLower(status)] {
			continue
		}
		for _, item := range items {
			if queueOpts.Label != "" && !hasLabel(item.Labels, queueOpts.Label) {
				continue
			}
			priority := ""
			if item.Fields != nil {
				priority = item.Fields["Priority"]
			}
			createdAt := item.CreatedAt
			if createdAt.IsZero() {
				createdAt = item.UpdatedAt
			}
			readyItems = append(readyItems, queueItem{
				Number:    item.Number,
				Title:     item.Title,
				Repo:      item.Repository,
				URL:       item.URL,
				Priority:  priority,
				Labels:    item.Labels,
				Status:    item.Status,
				CreatedAt: createdAt,
			})
		}
	}

	sort.Slice(readyItems, func(i, j int) bool {
		p1 := priorityRank(readyItems[i].Priority)
		p2 := priorityRank(readyItems[j].Priority)
		if p1 != p2 {
			return p1 < p2
		}
		return readyItems[i].CreatedAt.Before(readyItems[j].CreatedAt)
	})

	if queueOpts.Limit > 0 && len(readyItems) > queueOpts.Limit {
		readyItems = readyItems[:queueOpts.Limit]
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"project": project,
			"owner":   owner,
			"items":   readyItems,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Printf("🤖 Agent Queue — Project #%d\n\n", project)
	if len(readyItems) == 0 {
		fmt.Println("No items ready for processing.")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	for i, item := range readyItems {
		priority := item.Priority
		if priority == "" {
			priority = "—"
		}
		label := "—"
		if queueOpts.Label != "" {
			label = queueOpts.Label
		} else if len(item.Labels) > 0 {
			label = item.Labels[0]
		}
		fmt.Fprintf(w, "  %d.\t%s\t%s\t%s\t%s\t%s\n", i+1, issueRef(item.Number, item.URL), truncate(item.Title, 28), item.Repo, priority, label)
	}
	w.Flush()
	fmt.Println()
	fmt.Printf("%d items ready for processing\n", len(readyItems))
	return nil
}

func hasLabel(labels []string, target string) bool {
	for _, label := range labels {
		if strings.EqualFold(label, target) {
			return true
		}
	}
	return false
}

func priorityRank(value string) int {
	if value == "" {
		return 999
	}
	upper := strings.ToUpper(strings.TrimSpace(value))
	upper = strings.TrimPrefix(upper, "P")
	if upper == "HIGH" {
		return 1
	}
	if upper == "MEDIUM" {
		return 2
	}
	if upper == "LOW" {
		return 3
	}
	if num, err := strconv.Atoi(upper); err == nil {
		return num
	}
	return 999
}
