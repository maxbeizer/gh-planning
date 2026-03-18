package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var sprintOpts struct {
	Project  int
	Owner    string
	Duration string
	Limit    int
}

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Sprint/iteration management",
	Long:  "View and manage sprint iterations in your GitHub project.",
}

var sprintShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current sprint/iteration items",
	Long:  "Display items in the current iteration grouped by status. Falls back to showing In Progress and In Review items if no iteration field is configured.",
	RunE:  runSprintShow,
}

var sprintCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Plan a new sprint from backlog items",
	Long:  "List the top N highest-priority backlog items as candidates for the next sprint.",
	RunE:  runSprintCreate,
}

func init() {
	sprintCmd.PersistentFlags().IntVar(&sprintOpts.Project, "project", 0, "Project number")
	sprintCmd.PersistentFlags().StringVar(&sprintOpts.Owner, "owner", "", "Project owner")

	sprintCreateCmd.Flags().StringVar(&sprintOpts.Duration, "duration", "2w", "Sprint duration (e.g. 1w, 2w)")
	sprintCreateCmd.Flags().IntVar(&sprintOpts.Limit, "limit", 10, "Max items to pull from backlog")

	sprintCmd.AddCommand(sprintShowCmd)
	sprintCmd.AddCommand(sprintCreateCmd)
}

func resolveSprintProjectConfig() (string, int, error) {
	pc, err := resolveProjectConfig(sprintOpts.Owner, sprintOpts.Project)
	if err != nil {
		return "", 0, err
	}
	return pc.Owner, pc.Project, nil
}

// sprintIterationInfo holds metadata about a project's iteration field.
type sprintIterationInfo struct {
	FieldID    string
	Iterations []iterationValue
}

type iterationValue struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	StartDate string    `json:"startDate"`
	Duration  int       `json:"duration"`
	Current   bool      `json:"current"`
}

// getIterationField queries the project for an iteration field and its values.
func getIterationField(ctx context.Context, owner string, number int) (*sprintIterationInfo, error) {
	queryUser := `query($owner: String!, $number: Int!) {
  user(login: $owner) {
    projectV2(number: $number) {
      fields(first: 50) {
        nodes {
          ... on ProjectV2IterationField {
            id
            name
            configuration {
              iterations { id title startDate duration }
              completedIterations { id title startDate duration }
            }
          }
        }
      }
    }
  }
}`
	queryOrg := `query($owner: String!, $number: Int!) {
  organization(login: $owner) {
    projectV2(number: $number) {
      fields(first: 50) {
        nodes {
          ... on ProjectV2IterationField {
            id
            name
            configuration {
              iterations { id title startDate duration }
              completedIterations { id title startDate duration }
            }
          }
        }
      }
    }
  }
}`

	vars := map[string]interface{}{"owner": owner, "number": number}

	// Try user then org
	payload, err := github.GraphQL(ctx, queryUser, vars)
	if err != nil {
		payload, err = github.GraphQL(ctx, queryOrg, vars)
		if err != nil {
			return nil, err
		}
	}

	return parseIterationField(payload, "user", ctx, owner, number, queryOrg, vars)
}

func parseIterationField(payload []byte, source string, ctx context.Context, owner string, number int, orgQuery string, vars map[string]interface{}) (*sprintIterationInfo, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}

	data, _ := raw["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("no data in response")
	}

	// Try to find the project in user or organization
	var project map[string]interface{}
	for _, key := range []string{"user", "organization"} {
		if entry, ok := data[key].(map[string]interface{}); ok {
			if pv2, ok := entry["projectV2"].(map[string]interface{}); ok {
				project = pv2
				break
			}
		}
	}

	if project == nil {
		// If user query returned nil, try org
		if source == "user" {
			orgPayload, err := github.GraphQL(ctx, orgQuery, vars)
			if err != nil {
				return nil, nil
			}
			return parseIterationField(orgPayload, "organization", ctx, owner, number, orgQuery, vars)
		}
		return nil, nil
	}

	fields, _ := project["fields"].(map[string]interface{})
	if fields == nil {
		return nil, nil
	}
	nodes, _ := fields["nodes"].([]interface{})

	for _, node := range nodes {
		fieldMap, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		fieldID, _ := fieldMap["id"].(string)
		if fieldID == "" {
			continue
		}
		config, _ := fieldMap["configuration"].(map[string]interface{})
		if config == nil {
			continue
		}

		info := &sprintIterationInfo{FieldID: fieldID}
		now := time.Now()

		for _, key := range []string{"iterations", "completedIterations"} {
			iters, _ := config[key].([]interface{})
			for _, iter := range iters {
				iterMap, ok := iter.(map[string]interface{})
				if !ok {
					continue
				}
				iv := iterationValue{
					ID:        stringVal(iterMap, "id"),
					Title:     stringVal(iterMap, "title"),
					StartDate: stringVal(iterMap, "startDate"),
					Duration:  intVal(iterMap, "duration"),
				}
				if iv.StartDate != "" && iv.Duration > 0 {
					start, err := time.Parse("2006-01-02", iv.StartDate)
					if err == nil {
						end := start.AddDate(0, 0, iv.Duration)
						if !now.Before(start) && now.Before(end) {
							iv.Current = true
						}
					}
				}
				info.Iterations = append(info.Iterations, iv)
			}
		}
		return info, nil
	}
	return nil, nil
}

func stringVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intVal(m map[string]interface{}, key string) int {
	v, _ := m[key].(float64)
	return int(v)
}

func runSprintShow(cmd *cobra.Command, args []string) error {
	owner, project, err := resolveSprintProjectConfig()
	if err != nil {
		return err
	}

	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}

	iterInfo, _ := getIterationField(cmd.Context(), owner, project)

	var currentIteration *iterationValue
	hasIteration := false
	if iterInfo != nil {
		for i := range iterInfo.Iterations {
			if iterInfo.Iterations[i].Current {
				currentIteration = &iterInfo.Iterations[i]
				hasIteration = true
				break
			}
		}
	}

	// Collect items for the sprint view
	sprintItems := map[string][]github.ProjectItem{}

	if hasIteration && currentIteration != nil {
		// Filter items that have the iteration field set to the current iteration
		for status, items := range projectData.Items {
			for _, item := range items {
				if iterVal, ok := item.Fields["Iteration"]; ok {
					if iterVal == currentIteration.Title {
						sprintItems[status] = append(sprintItems[status], item)
					}
				}
			}
		}
	} else {
		// Fallback: use In Progress + In Review as a proxy for current sprint
		for status, items := range projectData.Items {
			lower := strings.ToLower(status)
			if lower == "in progress" || lower == "in review" || lower == "needs review" || lower == "needs my attention" {
				sprintItems[status] = append(sprintItems[status], items...)
			}
		}
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		iterTitle := ""
		if currentIteration != nil {
			iterTitle = currentIteration.Title
		}
		payload := map[string]interface{}{
			"title":          projectData.Title,
			"owner":          owner,
			"number":         project,
			"hasIteration":   hasIteration,
			"iteration":      iterTitle,
			"items":          sprintItems,
			"totalItems":     countItems(sprintItems),
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "🏃 Sprint: %s (#%d)\n", projectData.Title, project)
	if hasIteration && currentIteration != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "   Iteration: %s", currentIteration.Title)
		if currentIteration.StartDate != "" {
			start, err := time.Parse("2006-01-02", currentIteration.StartDate)
			if err == nil {
				end := start.AddDate(0, 0, currentIteration.Duration)
				remaining := time.Until(end)
				if remaining > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), " (%d days remaining)", int(remaining.Hours()/24))
				}
				fmt.Fprintf(cmd.OutOrStdout(), "\n   %s → %s", start.Format("Jan 2"), end.Format("Jan 2"))
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "   ℹ️  No iteration field found — showing active items as sprint proxy")
	}
	fmt.Fprintln(cmd.OutOrStdout())

	total := countItems(sprintItems)
	if total == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "  No items in the current sprint.")
		return nil
	}

	printStatusGroups(sprintItems, 0)

	fmt.Fprintf(cmd.OutOrStdout(), "📊 Total: %d items\n", total)
	return nil
}

func runSprintCreate(cmd *cobra.Command, args []string) error {
	owner, project, err := resolveSprintProjectConfig()
	if err != nil {
		return err
	}

	limit := sprintOpts.Limit
	duration := sprintOpts.Duration

	projectData, err := github.GetProject(cmd.Context(), owner, project)
	if err != nil {
		return err
	}

	// Find backlog items
	var backlogItems []github.ProjectItem
	for status, items := range projectData.Items {
		if strings.EqualFold(status, "Backlog") || strings.EqualFold(status, "Todo") || strings.EqualFold(status, "To Do") {
			backlogItems = append(backlogItems, items...)
		}
	}

	// Sort by creation date (oldest first = highest priority in typical backlogs)
	sort.Slice(backlogItems, func(i, j int) bool {
		return backlogItems[i].CreatedAt.Before(backlogItems[j].CreatedAt)
	})

	if limit > 0 && len(backlogItems) > limit {
		backlogItems = backlogItems[:limit]
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"title":      projectData.Title,
			"owner":      owner,
			"number":     project,
			"duration":   duration,
			"candidates": backlogItems,
			"total":      len(backlogItems),
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), "🆕 Sprint Planning: %s (#%d)\n", projectData.Title, project)
	fmt.Fprintf(cmd.OutOrStdout(), "   Duration: %s | Limit: %d items\n\n", duration, limit)

	if len(backlogItems) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "  No backlog items found. Your backlog is empty! 🎉")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "📋 Top %d backlog candidates:\n", len(backlogItems))
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	for i, item := range backlogItems {
		assignee := "—"
		if len(item.Assignees) > 0 {
			assignee = "@" + item.Assignees[0]
		}
		issueNum := fmt.Sprintf("#%d", item.Number)
		if item.URL != "" {
			issueNum = hyperlink(item.URL, issueNum)
		}
		fmt.Fprintf(w, "  %d.\t%s\t%s\t%s\t%s\n", i+1, issueNum, truncate(item.Title, 40), item.Repository, assignee)
	}
	w.Flush()

	fmt.Fprintf(cmd.OutOrStdout(), "\n💡 To move items into the sprint, update their status in your project board\n")
	fmt.Fprintf(cmd.OutOrStdout(), "   or set the iteration field via the GitHub UI.\n")
	return nil
}

func countItems(groups map[string][]github.ProjectItem) int {
	count := 0
	for _, items := range groups {
		count += len(items)
	}
	return count
}
