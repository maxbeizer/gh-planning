package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

type hygieneFinding struct {
	Kind        string             `json:"kind"`
	Description string             `json:"description"`
	Item        github.ProjectItem `json:"item"`
	Detail      string             `json:"detail,omitempty"`
}

type hygieneReport struct {
	Title    string           `json:"title"`
	Owner    string           `json:"owner"`
	Project  int              `json:"project"`
	Findings []hygieneFinding `json:"findings"`
	Filters  []fieldFilter    `json:"filters,omitempty"`
}

var hygieneOpts struct {
	Project        int
	Owner          string
	Fields         []string
	Stale          string
	Format         string
	RequiredFields []string
	OwnerFields    []string
}

var hygieneCmd = &cobra.Command{
	Use:   "hygiene",
	Short: "Report actionable project board hygiene issues",
	RunE:  runHygiene,
}

func init() {
	hygieneCmd.Flags().IntVar(&hygieneOpts.Project, "project", 0, "Project number")
	hygieneCmd.Flags().StringVar(&hygieneOpts.Owner, "owner", "", "Project owner")
	hygieneCmd.Flags().StringArrayVar(&hygieneOpts.Fields, "field", nil, "Filter by project field value (repeatable, e.g. --field Manager=me)")
	hygieneCmd.Flags().StringVar(&hygieneOpts.Stale, "stale", "7d", "Flag active items stale for this duration")
	hygieneCmd.Flags().StringVar(&hygieneOpts.Format, "format", "text", "Output format: text or markdown")
	hygieneCmd.Flags().StringArrayVar(&hygieneOpts.RequiredFields, "required-field", []string{"Target Date", "Workstream"}, "Field that active items should have (repeatable)")
	hygieneCmd.Flags().StringArrayVar(&hygieneOpts.OwnerFields, "owner-field", []string{"Owner", "DRI", "Manager"}, "Ownership field to treat like an owner/DRI (repeatable)")
}

func runHygiene(cmd *cobra.Command, args []string) error {
	pc, err := resolveProjectConfig(hygieneOpts.Owner, hygieneOpts.Project)
	if err != nil {
		return err
	}
	staleDuration, err := parseDuration(hygieneOpts.Stale)
	if err != nil {
		return fmt.Errorf("invalid stale duration: %w", err)
	}
	fieldFilters, err := resolveFieldFilters(cmd.Context(), pc.Cfg.Filters, hygieneOpts.Fields)
	if err != nil {
		return err
	}
	projectData, err := github.GetProject(cmd.Context(), pc.Owner, pc.Project)
	if err != nil {
		return err
	}
	if len(fieldFilters) > 0 {
		projectData.Items = filterProjectItems(projectData, "", 0, nil, fieldFilters)
	}
	if err := github.EnrichWithDependencies(cmd.Context(), projectData); err != nil {
		return err
	}
	if err := github.EnrichWithSubIssues(cmd.Context(), projectData); err != nil {
		return err
	}

	report := hygieneReport{
		Title:    projectData.Title,
		Owner:    pc.Owner,
		Project:  pc.Project,
		Findings: buildHygieneFindings(projectData, staleDuration, hygieneOpts.RequiredFields, hygieneOpts.OwnerFields),
		Filters:  fieldFilters,
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		if report.Findings[i].Kind != report.Findings[j].Kind {
			return report.Findings[i].Kind < report.Findings[j].Kind
		}
		return report.Findings[i].Item.UpdatedAt.Before(report.Findings[j].Item.UpdatedAt)
	})

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(cmd.OutOrStdout(), report, OutputOptions())
	}
	switch strings.ToLower(hygieneOpts.Format) {
	case "text":
		printHygieneText(cmd.OutOrStdout(), report)
	case "markdown":
		printHygieneMarkdown(cmd.OutOrStdout(), report)
	default:
		return fmt.Errorf("invalid format %q (expected text or markdown)", hygieneOpts.Format)
	}
	return nil
}

func buildHygieneFindings(project *github.Project, stale time.Duration, requiredFields, ownerFields []string) []hygieneFinding {
	var findings []hygieneFinding
	childrenByParent := map[int][]github.ProjectItem{}
	for _, item := range flattenProjectItems(project) {
		if item.ParentIssue != nil {
			childrenByParent[item.ParentIssue.Number] = append(childrenByParent[item.ParentIssue.Number], item)
		}
	}

	for _, item := range flattenProjectItems(project) {
		active := isActiveProjectItem(item)
		if active && stale > 0 && time.Since(item.UpdatedAt) >= stale {
			findings = append(findings, finding("stale-active", "Active item has not been updated recently", item, humanizeDuration(time.Since(item.UpdatedAt))))
		}
		if active && !hasOwner(item, ownerFields) {
			findings = append(findings, finding("unowned-active", "Active item has no assignee or ownership field", item, ""))
		}
		if active {
			for _, field := range requiredFields {
				field = strings.TrimSpace(field)
				if field != "" && !hasProjectField(item, field) {
					findings = append(findings, finding("missing-field", "Active item is missing a required planning field", item, field))
				}
			}
		}
		if isClosedContent(item) && !isDoneStatus(item.Status) {
			findings = append(findings, finding("closed-not-done", "Closed issue or PR is not marked Done on the board", item, item.Status))
		}
		if isDoneStatus(item.Status) && isOpenContent(item) {
			findings = append(findings, finding("done-open", "Done board item is still open", item, item.State))
		}
		if active && isBlockedMarker(item) && !item.IsBlocked() {
			findings = append(findings, finding("blocked-without-link", "Blocked item has no open linked blocker", item, ""))
		}
		if active && strings.EqualFold(item.ContentType, "PullRequest") && strings.EqualFold(item.State, "MERGED") {
			findings = append(findings, finding("merged-pr-active", "Merged PR remains active on the board", item, item.Status))
		}
		if active {
			for _, child := range childrenByParent[item.Number] {
				if stale > 0 && time.Since(child.UpdatedAt) >= stale {
					findings = append(findings, finding("parent-child-hygiene", "Active parent has a stale child issue", item, fmt.Sprintf("child #%d stale %s", child.Number, humanizeDuration(time.Since(child.UpdatedAt)))))
				}
				if !hasOwner(child, ownerFields) {
					findings = append(findings, finding("parent-child-hygiene", "Active parent has an unowned child issue", item, fmt.Sprintf("child #%d unowned", child.Number)))
				}
			}
		}
	}
	return findings
}

func finding(kind, description string, item github.ProjectItem, detail string) hygieneFinding {
	return hygieneFinding{Kind: kind, Description: description, Item: item, Detail: detail}
}

func flattenProjectItems(project *github.Project) []github.ProjectItem {
	var items []github.ProjectItem
	for _, group := range project.Items {
		items = append(items, group...)
	}
	return items
}

func hasOwner(item github.ProjectItem, ownerFields []string) bool {
	if len(item.Assignees) > 0 {
		return true
	}
	for _, field := range ownerFields {
		if hasProjectField(item, field) {
			return true
		}
	}
	return false
}

func hasProjectField(item github.ProjectItem, name string) bool {
	value, ok := fieldValue(item, name)
	return ok && strings.TrimSpace(value) != ""
}

func isActiveProjectItem(item github.ProjectItem) bool {
	return !isDoneStatus(item.Status)
}

func isDoneStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "done", "closed", "complete", "completed":
		return true
	default:
		return false
	}
}

func isClosedContent(item github.ProjectItem) bool {
	return strings.EqualFold(item.State, "CLOSED") || strings.EqualFold(item.State, "MERGED")
}

func isOpenContent(item github.ProjectItem) bool {
	return strings.EqualFold(item.State, "OPEN")
}

func isBlockedMarker(item github.ProjectItem) bool {
	if strings.EqualFold(item.Status, "Blocked") {
		return true
	}
	for _, label := range item.Labels {
		if strings.EqualFold(label, "blocked") {
			return true
		}
	}
	return false
}

func printHygieneText(w io.Writer, report hygieneReport) {
	fmt.Fprintf(w, "🧹 Hygiene: %s (#%d)\n\n", report.Title, report.Project)
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "  No hygiene issues found.")
		return
	}
	for _, group := range groupFindings(report.Findings) {
		fmt.Fprintf(w, "%s (%d)\n", group.Kind, len(group.Findings))
		for _, finding := range group.Findings {
			fmt.Fprintf(w, "  • %s %s — %s", issueRef(finding.Item.Number, finding.Item.URL), finding.Item.Title, finding.Description)
			if finding.Detail != "" {
				fmt.Fprintf(w, " (%s)", finding.Detail)
			}
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w)
	}
}

func printHygieneMarkdown(w io.Writer, report hygieneReport) {
	fmt.Fprintf(w, "# Hygiene: %s (#%d)\n\n", report.Title, report.Project)
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "No hygiene issues found.")
		return
	}
	for _, group := range groupFindings(report.Findings) {
		fmt.Fprintf(w, "## %s (%d)\n\n", group.Kind, len(group.Findings))
		for _, finding := range group.Findings {
			ref := fmt.Sprintf("#%d", finding.Item.Number)
			if finding.Item.URL != "" {
				ref = fmt.Sprintf("[#%d](%s)", finding.Item.Number, finding.Item.URL)
			}
			fmt.Fprintf(w, "- %s %s — %s", ref, finding.Item.Title, finding.Description)
			if finding.Detail != "" {
				fmt.Fprintf(w, " (%s)", finding.Detail)
			}
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w)
	}
}

type hygieneFindingGroup struct {
	Kind     string
	Findings []hygieneFinding
}

func groupFindings(findings []hygieneFinding) []hygieneFindingGroup {
	grouped := map[string][]hygieneFinding{}
	for _, finding := range findings {
		grouped[finding.Kind] = append(grouped[finding.Kind], finding)
	}
	kinds := make([]string, 0, len(grouped))
	for kind := range grouped {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	result := make([]hygieneFindingGroup, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, hygieneFindingGroup{Kind: kind, Findings: grouped[kind]})
	}
	return result
}
