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
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configProfilesCmd)
	configCmd.AddCommand(configDeleteCmd)
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

	profileName, _ := config.ActiveProfileName()

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"config": cfg,
		}
		if profileName != "" {
			payload["profile"] = profileName
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	if profileName != "" {
		fmt.Printf("Profile: %s\n\n", profileName)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

var configUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch to a named config profile",
	Long: `Switch to a named config profile. If the profile doesn't exist yet,
it will be created as an empty profile that you can configure with
"gh planning config set" or "gh planning setup".

The first time you use profiles, your existing config is preserved
as the "default" profile.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.UseProfile(name); err != nil {
			return err
		}
		fmt.Printf("Switched to profile %q\n", name)
		return nil
	},
}

var configProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List all config profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, active, err := config.ListProfiles()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Println("No profiles configured. Using single config.")
			fmt.Println("Run `gh planning config use <name>` to create profiles.")
			return nil
		}

		if OutputOptions().JSON || OutputOptions().JQ != "" {
			payload := map[string]interface{}{
				"profiles": names,
				"active":   active,
			}
			return output.PrintJSON(payload, OutputOptions())
		}

		for _, name := range names {
			if name == active {
				fmt.Printf("  * %s (active)\n", name)
			} else {
				fmt.Printf("    %s\n", name)
			}
		}
		return nil
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <profile>",
	Short: "Delete a config profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.DeleteProfile(name); err != nil {
			return err
		}
		fmt.Printf("Deleted profile %q\n", name)
		return nil
	},
}
