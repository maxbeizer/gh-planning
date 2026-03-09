package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/spf13/cobra"
)

var cheatsheetPlain bool

var cheatsheetCmd = &cobra.Command{
	Use:   "cheatsheet",
	Short: "Interactive quick-reference for gh-planning commands",
	Long: `Browse gh-planning commands organized by scenario.

Use the interactive browser to search and explore commands,
or use --plain for a static text version.

Categories include Morning Routine, Starting a Task, During Work,
Collaboration, AI & Agents, Wrapping Up, and Configuration.`,
	Aliases: []string{"cheat", "recipes"},
	RunE:    runCheatsheet,
}

func init() {
	cheatsheetCmd.Flags().BoolVar(&cheatsheetPlain, "plain", false, "Plain text output (no interactive UI)")
}

func runCheatsheet(cmd *cobra.Command, args []string) error {
	items := cheatsheetItems()

	if cheatsheetPlain || !isTerminal() {
		fmt.Fprint(cmd.OutOrStdout(), tui.RenderListPlain(items))
		return nil
	}

	m := tui.NewListModel(items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
