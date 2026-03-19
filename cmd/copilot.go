package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/mcp"
	"github.com/spf13/cobra"
)

var copilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Copilot integration utilities",
}

var copilotServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server over stdio",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcp.Serve(os.Stdin, os.Stdout, os.Stderr)
	},
}

var copilotSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List available Copilot skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		skills, err := listCopilotSkills()
		if err != nil {
			return err
		}
		if OutputOptions().JSON || OutputOptions().JQ != "" {
			return output.PrintJSON(map[string]interface{}{"skills": skills}, OutputOptions())
		}
		for _, skill := range skills {
			fmt.Fprintln(cmd.OutOrStdout(), skill)
		}
		return nil
	},
}

var copilotTestCmd = &cobra.Command{
	Use:   "test <query>",
	Short: "Test a Copilot skill selection",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.TrimSpace(strings.Join(args, " "))
		suggestion := suggestSkill(query)
		payload := map[string]interface{}{
			"query":   query,
			"skill":   suggestion.Skill,
			"command": suggestion.Command,
			"reason":  suggestion.Reason,
		}
		if OutputOptions().JSON || OutputOptions().JQ != "" {
			return output.PrintJSON(payload, OutputOptions())
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Skill: %s\n", suggestion.Skill)
		fmt.Fprintf(cmd.OutOrStdout(), "Command: %s\n", suggestion.Command)
		if suggestion.Reason != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Reason: %s\n", suggestion.Reason)
		}
		return nil
	},
}

type skillSuggestion struct {
	Skill   string
	Command string
	Reason  string
}

func init() {
	copilotCmd.AddCommand(copilotServeCmd)
	copilotCmd.AddCommand(copilotSkillsCmd)
	copilotCmd.AddCommand(copilotTestCmd)
}

func listCopilotSkills() ([]string, error) {
	skillDir := filepath.Join("copilot-skills")
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, err
	}
	skills := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") {
			skills = append(skills, strings.TrimSuffix(name, ".md"))
		}
	}
	sort.Strings(skills)
	return skills, nil
}

func suggestSkill(query string) skillSuggestion {
	q := strings.ToLower(query)
	switch {
	case strings.Contains(q, "standup") || strings.Contains(q, "yesterday"):
		return skillSuggestion{Skill: "standup", Command: "gh planning standup --json", Reason: "standup-related query"}
	case strings.Contains(q, "catch") || strings.Contains(q, "miss"):
		return skillSuggestion{Skill: "catch-up", Command: "gh planning catch-up --json", Reason: "catch-up summary query"}
	case strings.Contains(q, "1-1") || strings.Contains(q, "one on one") || strings.Contains(q, "prep"):
		return skillSuggestion{Skill: "team-prep", Command: "gh planning prep {handle} --json", Reason: "1-1 prep query"}
	case strings.Contains(q, "pulse") || strings.Contains(q, "team"):
		return skillSuggestion{Skill: "team-prep", Command: "gh planning team --json", Reason: "team dashboard query"}
	case strings.Contains(q, "status") || strings.Contains(q, "blocked") || strings.Contains(q, "stale") || strings.Contains(q, "backlog"):
		return skillSuggestion{Skill: "project-status", Command: "gh planning status --json", Reason: "project status query"}
	default:
		return skillSuggestion{Skill: "project-status", Command: "gh planning status --json", Reason: "default project status"}
	}
}
