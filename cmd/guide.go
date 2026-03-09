package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/spf13/cobra"
)

var guidePlain bool

var guideCmd = &cobra.Command{
	Use:   "guide [workflow]",
	Short: "Workflow walkthroughs for common scenarios",
	Long: `Step-by-step guides for common gh-planning workflows.

Available workflows:
  morning      Morning catch-up → standup → board review
  new-task     Track → claim → focus → log → complete
  one-on-one   Team activity → 1-1 prep → health metrics
  agent        Agent context → queue → breakdown → MCP
  breakdown    Preview → create sub-issues → estimate

Run with no arguments to see all available workflows.`,
	Aliases:          []string{"guides", "workflow"},
	RunE:             runGuide,
	ValidArgsFunction: completeGuideWorkflows,
}

func init() {
	guideCmd.Flags().BoolVar(&guidePlain, "plain", false, "Plain text output (no interactive UI)")
}

func completeGuideWorkflows(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var names []string
	for _, w := range availableWorkflows() {
		names = append(names, w.Name+"\t"+w.Description)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func runGuide(cmd *cobra.Command, args []string) error {
	workflows := availableWorkflows()

	if len(args) == 0 {
		return listWorkflows(cmd, workflows)
	}

	name := strings.ToLower(args[0])
	for _, w := range workflows {
		if w.Name == name {
			return runWorkflowGuide(cmd, w)
		}
	}

	return fmt.Errorf("unknown workflow %q — run `gh planning guide` to see available workflows", name)
}

func listWorkflows(cmd *cobra.Command, workflows []workflow) error {
	fmt.Fprintln(cmd.OutOrStdout(), tui.Title.Render("Available Workflow Guides"))
	fmt.Fprintln(cmd.OutOrStdout())
	for _, w := range workflows {
		fmt.Fprintf(cmd.OutOrStdout(), "  %s  %-14s %s\n",
			tui.Command.Render(w.Name),
			"",
			tui.Muted.Render(w.Description),
		)
	}
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), tui.Muted.Render("Usage: gh planning guide <workflow>"))
	return nil
}

func runWorkflowGuide(cmd *cobra.Command, w workflow) error {
	if guidePlain || !isTerminal() {
		return renderGuidePlain(cmd, w)
	}

	m := tui.NewStepModel(w.Steps)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return err
	}

	final := result.(tui.StepModel)
	if !final.Quit {
		return nil
	}

	// Count completed steps
	done := 0
	for _, s := range final.Steps {
		if s.Done {
			done++
		}
	}
	if done == len(final.Steps) {
		fmt.Fprintln(cmd.OutOrStdout(), tui.Success.Render("✓ Guide complete!"))
	}
	return nil
}

func renderGuidePlain(cmd *cobra.Command, w workflow) error {
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n%s\n\n", w.Title, w.Description)
	for i, step := range w.Steps {
		fmt.Fprint(cmd.OutOrStdout(), tui.RenderStepPlain(step, i, len(w.Steps)))
	}
	return nil
}
