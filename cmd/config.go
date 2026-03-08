package cmd

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gh-planning configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show config",
	RunE:  runConfigShow,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	key := args[0]
	value := args[1]
	switch key {
	case "default-project":
		var project int
		if _, err := fmt.Sscanf(value, "%d", &project); err != nil {
			return fmt.Errorf("invalid project number")
		}
		cfg.DefaultProject = project
	case "default-owner":
		cfg.DefaultOwner = value
	case "team":
		if value == "" {
			cfg.Team = nil
		} else {
			cfg.Team = strings.Split(value, ",")
			for i := range cfg.Team {
				cfg.Team[i] = strings.TrimSpace(cfg.Team[i])
			}
		}
	case "1-1-repo-pattern":
		cfg.OneOnOneRepoPattern = value
	case "agent.max-per-hour":
		var max int
		if _, err := fmt.Sscanf(value, "%d", &max); err != nil {
			return fmt.Errorf("invalid max value")
		}
		cfg.AgentMaxPerHour = max
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Updated %s\n", key)
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(cfg, OutputOptions())
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
