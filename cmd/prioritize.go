package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var prioritizeOpts struct {
	Project int
	Owner   string
	Status  string
}

var prioritizeCmd = &cobra.Command{
	Use:   "prioritize",
	Short: "Interactively reorder and prioritize backlog items",
	Long:  "Fetch project items filtered by status, display them as a numbered list, and interactively reorder or assign priority levels (P0–P4).",
	RunE:  runPrioritize,
}

func init() {
	prioritizeCmd.Flags().IntVar(&prioritizeOpts.Project, "project", 0, "Project number")
	prioritizeCmd.Flags().StringVar(&prioritizeOpts.Owner, "owner", "", "Project owner")
	prioritizeCmd.Flags().StringVar(&prioritizeOpts.Status, "status", "Backlog", "Status column to prioritize")
}

type prioritizeItem struct {
	ID       string   `json:"id"`
	Number   int      `json:"number"`
	Title    string   `json:"title"`
	Repo     string   `json:"repo"`
	Assignee string   `json:"assignee"`
	Priority string   `json:"priority,omitempty"`
	Labels   []string `json:"labels,omitempty"`
}

func runPrioritize(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(prioritizeOpts.Owner, prioritizeOpts.Project)
	if err != nil {
		return err
	}

	statusFilter := prioritizeOpts.Status

	// Fetch project items
	projectData, err := github.GetProject(cmd.Context(), pc.Owner, pc.Project)
	if err != nil {
		return err
	}

	// Find items matching the requested status
	var items []prioritizeItem
	for status, pItems := range projectData.Items {
		if !strings.EqualFold(status, statusFilter) {
			continue
		}
		for _, pi := range pItems {
			assignee := ""
			if len(pi.Assignees) > 0 {
				assignee = pi.Assignees[0]
			}
			priority := ""
			if pi.Fields != nil {
				priority = pi.Fields["Priority"]
			}
			items = append(items, prioritizeItem{
				ID:       pi.ID,
				Number:   pi.Number,
				Title:    pi.Title,
				Repo:     pi.Repository,
				Assignee: assignee,
				Priority: priority,
				Labels:   pi.Labels,
			})
		}
	}

	if len(items) == 0 {
		if OutputOptions().JSON || OutputOptions().JQ != "" {
			return output.PrintJSON(map[string]interface{}{
				"project": pc.Project,
				"owner":   pc.Owner,
				"status":  statusFilter,
				"items":   []prioritizeItem{},
				"message": "No items found",
			}, OutputOptions())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "No items found in %q status.\n", statusFilter)
		return nil
	}

	// Display numbered list
	fmt.Fprintf(cmd.OutOrStdout(), "📋 %s items in project #%d (%q status)\n\n", projectData.Title, pc.Project, statusFilter)
	printPrioritizeItems(items)

	reader := bufio.NewReader(os.Stdin)

	// Prompt for reorder
	fmt.Fprintf(cmd.OutOrStdout(), "\nEnter new order (e.g. %s), or press Enter to skip: ", exampleOrder(len(items)))
	orderInput, err := readLine(reader)
	if err != nil {
		return err
	}

	if orderInput != "" {
		reordered, err := parseOrder(orderInput, len(items))
		if err != nil {
			return fmt.Errorf("invalid order: %w", err)
		}
		newItems := make([]prioritizeItem, len(reordered))
		for i, idx := range reordered {
			newItems[i] = items[idx]
		}
		items = newItems
		fmt.Fprintln(cmd.OutOrStdout(), "\n✅ Reordered:")
		printPrioritizeItems(items)
	}

	// Get project ID for mutations
	projectID, _, _, _, infoErr := github.GetProjectInfo(cmd.Context(), pc.Owner, pc.Project)
	if infoErr != nil {
		return infoErr
	}

	// Check for Priority field on the project
	priorityFieldID, priorityOptions, err := github.GetProjectField(cmd.Context(), pc.Owner, pc.Project, "Priority")
	if err != nil {
		return err
	}

	if priorityFieldID != "" && len(priorityOptions) > 0 {
		optionNames := sortedPriorityOptions(priorityOptions)
		fmt.Fprintf(cmd.OutOrStdout(), "\n🏷️  Priority field found with options: %s\n", strings.Join(optionNames, ", "))
		fmt.Fprintf(cmd.OutOrStdout(), "Assign priorities to top items? Enter priorities for each item (e.g. P0,P1,P2), or press Enter to skip: ")
		priInput, err := readLine(reader)
		if err != nil {
			return err
		}

		if priInput != "" {
			assignments, err := parsePriorityAssignments(priInput, len(items), priorityOptions)
			if err != nil {
				return fmt.Errorf("invalid priority input: %w", err)
			}

			for i, optionID := range assignments {
				if optionID == "" {
					continue
				}
				if err := github.UpdateItemStatus(cmd.Context(), projectID, items[i].ID, priorityFieldID, optionID); err != nil {
					fmt.Fprintf(os.Stderr, "⚠️  Failed to set priority for #%d: %v\n", items[i].Number, err)
					continue
				}
				for name, id := range priorityOptions {
					if id == optionID {
						items[i].Priority = name
						break
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  ✓ #%d %s → %s\n", items[i].Number, truncate(items[i].Title, 30), items[i].Priority)
			}
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "\nℹ️  No Priority field found on this project. Showing reordered list as recommendation.")
	}

	// JSON output
	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"project": pc.Project,
			"owner":   pc.Owner,
			"status":  statusFilter,
			"items":   items,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\n📊 %d items prioritized in %q\n", len(items), statusFilter)
	return nil
}

func printPrioritizeItems(items []prioritizeItem) {
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	for i, item := range items {
		assignee := item.Assignee
		if assignee == "" {
			assignee = "—"
		}
		priority := item.Priority
		if priority == "" {
			priority = "—"
		}
		fmt.Fprintf(w, "  %d.\t#%d\t%s\t%s\t%s\t%s\n",
			i+1, item.Number, truncate(item.Title, 35), item.Repo, assignee, priority)
	}
	w.Flush()
}

func exampleOrder(n int) string {
	if n <= 1 {
		return "1"
	}
	if n == 2 {
		return "2,1"
	}
	parts := make([]string, 0, n)
	if n <= 5 {
		order := []int{n, 1}
		for i := 2; i < n; i++ {
			order = append(order, i)
		}
		for _, v := range order {
			parts = append(parts, strconv.Itoa(v))
		}
	} else {
		parts = append(parts, "3", "1", "5", "2", "4")
	}
	return strings.Join(parts, ",")
}

func parseOrder(input string, count int) ([]int, error) {
	parts := strings.Split(input, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty order")
	}

	seen := map[int]bool{}
	indices := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q", part)
		}
		if num < 1 || num > count {
			return nil, fmt.Errorf("number %d out of range (1-%d)", num, count)
		}
		if seen[num] {
			return nil, fmt.Errorf("duplicate number %d", num)
		}
		seen[num] = true
		indices = append(indices, num-1)
	}

	// Append remaining items in original order if partial
	if len(indices) < count {
		for i := 0; i < count; i++ {
			if !seen[i+1] {
				indices = append(indices, i)
			}
		}
	}

	return indices, nil
}

func sortedPriorityOptions(options map[string]string) []string {
	known := []string{"P0", "P1", "P2", "P3", "P4"}
	var result []string
	seen := map[string]bool{}
	for _, k := range known {
		for name := range options {
			if strings.EqualFold(name, k) {
				result = append(result, name)
				seen[name] = true
			}
		}
	}
	for name := range options {
		if !seen[name] {
			result = append(result, name)
		}
	}
	return result
}

func parsePriorityAssignments(input string, itemCount int, options map[string]string) ([]string, error) {
	parts := strings.Split(input, ",")
	assignments := make([]string, itemCount)

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "-" || part == "—" {
			continue
		}
		if i >= itemCount {
			break
		}
		found := false
		for name, id := range options {
			if strings.EqualFold(name, part) {
				assignments[i] = id
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown priority %q (available: %s)", part, strings.Join(sortedPriorityOptions(options), ", "))
		}
	}
	return assignments, nil
}
