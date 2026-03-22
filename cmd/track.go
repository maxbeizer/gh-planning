package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
)

var trackOpts struct {
	Repo     string
	Project  int
	Body     string
	Labels   []string
	Assignee string
	Status   string
}

var trackCmd = &cobra.Command{
	Use:   "track <title>",
	Short: "Create an issue and add it to the project",
	Args:  cobra.ExactArgs(1),
	RunE:  runTrack,
}

func init() {
	trackCmd.Flags().StringVar(&trackOpts.Repo, "repo", "", "Target repo (owner/repo)")
	trackCmd.Flags().IntVar(&trackOpts.Project, "project", 0, "Project number")
	trackCmd.Flags().StringVar(&trackOpts.Body, "body", "", "Issue body")
	trackCmd.Flags().StringSliceVar(&trackOpts.Labels, "label", nil, "Labels (repeatable)")
	trackCmd.Flags().StringVar(&trackOpts.Assignee, "assignee", "", "Assignee")
	trackCmd.Flags().StringVar(&trackOpts.Status, "status", "", "Initial project status")
}

func runTrack(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	owner := cfg.DefaultOwner
	project := trackOpts.Project
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return fmt.Errorf("project owner and number are required (run `gh planning init`) ")
	}
	if trackOpts.Repo == "" {
		trackOpts.Repo = config.DetectGitRepo()
	}
	if trackOpts.Repo == "" {
		return fmt.Errorf("--repo is required when not in a git repository")
	}
	title := args[0]
	argsCreate := []string{"issue", "create", "--repo", trackOpts.Repo, "--title", title}
	if trackOpts.Body != "" {
		argsCreate = append(argsCreate, "--body", trackOpts.Body)
	}
	for _, label := range trackOpts.Labels {
		if strings.TrimSpace(label) == "" {
			continue
		}
		argsCreate = append(argsCreate, "--label", label)
	}
	if trackOpts.Assignee != "" {
		argsCreate = append(argsCreate, "--assignee", trackOpts.Assignee)
	}
	createOut, err := github.Run(cmd.Context(), argsCreate...)
	if err != nil {
		return err
	}
	issueURL := strings.TrimSpace(string(createOut))

	// gh issue create only returns the URL; fetch structured data with gh issue view
	issuePayload, err := github.Run(cmd.Context(), "issue", "view", issueURL, "--json", "id,number,url,repository")
	if err != nil {
		return err
	}
	issue, err := github.ParseIssueCreateOutput(issuePayload)
	if err != nil {
		return err
	}

	projectID, _, statusFieldID, statusOptions, err := github.GetProjectInfo(cmd.Context(), owner, project)
	if err != nil {
		return err
	}
	itemID, err := github.AddItemToProject(cmd.Context(), projectID, issue.ID)
	if err != nil {
		return err
	}
	if trackOpts.Status != "" {
		optionID, ok := statusOptions[trackOpts.Status]
		if !ok {
			return fmt.Errorf("status option not found: %s", trackOpts.Status)
		}
		if statusFieldID == "" {
			return fmt.Errorf("status field not found on project")
		}
		if err := github.UpdateItemStatus(cmd.Context(), projectID, itemID, statusFieldID, optionID); err != nil {
			return err
		}
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(issue, OutputOptions())
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", issue.URL)
	return nil
}
