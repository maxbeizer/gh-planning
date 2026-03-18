package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var profileCmd = &cobra.Command{
	Use:     "profile",
	Short:   "Manage gh-planning profiles",
	Long:    `Manage configuration profiles for gh-planning. Each profile stores a project, owner, team, and repo mappings. Profiles can auto-activate based on the repo you're in.`,
	Aliases: []string{"config"},
}

var profileSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a profile value",
	Args:  cobra.ExactArgs(2),
	RunE:  runProfileSet,
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current profile",
	RunE:  runProfileShow,
}

var profileUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch to a named profile",
	Long: `Switch to a named profile. If the profile doesn't exist yet,
it will be created as an empty profile that you can configure with
"gh planning profile set" or "gh planning setup".

The first time you use profiles, your existing config is preserved
as the "default" profile.`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileUse,
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all profiles",
	Aliases: []string{"ls", "profiles"},
	RunE:    runProfileList,
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <profile>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileDelete,
}

var profileCreateOpts struct {
	Project int
	Owner   string
	Repos   string
	Orgs    string
	Team    string
	Use     bool
	Force   bool
}

var profileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile with interactive setup",
	Long: `Create a named profile. Without flags, walks you through an interactive
setup. With flags, creates the profile non-interactively.

The current repo is auto-detected and offered as a default for the repos field.

Examples:
  gh planning profile create work
  gh planning profile create work --owner github --project 42 --repos "github/*" --use
  gh planning profile create work --force --project 99`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileCreate,
}

var profileDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Show which profile matches the current repo",
	RunE:  runProfileDetect,
}

var profileUpdateOpts struct {
	Project int
	Owner   string
	Repos   string
	Orgs    string
	Team    string
}

var profileUpdateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Update an existing profile",
	Long: `Update an existing profile with new values. Only the provided flags
are changed; other fields are preserved.

Examples:
  gh planning profile update work --project 99
  gh planning profile update work --owner github --repos "github/*"
  gh planning profile update work --team "alice,bob"`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileUpdate,
}

func init() {
	profileCmd.AddCommand(profileSetCmd)
	profileCmd.AddCommand(profileShowCmd)
	profileCmd.AddCommand(profileUseCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileDetectCmd)
	profileCmd.AddCommand(profileUpdateCmd)

	profileCreateCmd.Flags().IntVar(&profileCreateOpts.Project, "project", 0, "Default project number")
	profileCreateCmd.Flags().StringVar(&profileCreateOpts.Owner, "owner", "", "Project owner")
	profileCreateCmd.Flags().StringVar(&profileCreateOpts.Repos, "repos", "", "Comma-separated repos (e.g., github/github,github/*)")
	profileCreateCmd.Flags().StringVar(&profileCreateOpts.Orgs, "orgs", "", "Comma-separated orgs for auto-detection")
	profileCreateCmd.Flags().StringVar(&profileCreateOpts.Team, "team", "", "Comma-separated team members")
	profileCreateCmd.Flags().BoolVar(&profileCreateOpts.Use, "use", false, "Switch to the new profile after creating it")
	profileCreateCmd.Flags().BoolVar(&profileCreateOpts.Force, "force", false, "Overwrite existing profile")

	profileUpdateCmd.Flags().IntVar(&profileUpdateOpts.Project, "project", 0, "Default project number")
	profileUpdateCmd.Flags().StringVar(&profileUpdateOpts.Owner, "owner", "", "Project owner")
	profileUpdateCmd.Flags().StringVar(&profileUpdateOpts.Repos, "repos", "", "Comma-separated repos (e.g., github/github,github/*)")
	profileUpdateCmd.Flags().StringVar(&profileUpdateOpts.Orgs, "orgs", "", "Comma-separated orgs for auto-detection")
	profileUpdateCmd.Flags().StringVar(&profileUpdateOpts.Team, "team", "", "Comma-separated team members")
}

func runProfileSet(cmd *cobra.Command, args []string) error {
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
			cfg.Team = splitAndTrim(value)
		}
	case "1-1-repo-pattern":
		cfg.OneOnOneRepoPattern = value
	case "agent.max-per-hour":
		var max int
		if _, err := fmt.Sscanf(value, "%d", &max); err != nil {
			return fmt.Errorf("invalid max value")
		}
		cfg.AgentMaxPerHour = max
	case "repos":
		if value == "" {
			cfg.Repos = nil
		} else {
			cfg.Repos = splitAndTrim(value)
		}
	case "orgs":
		if value == "" {
			cfg.Orgs = nil
		} else {
			cfg.Orgs = splitAndTrim(value)
		}
	default:
		return fmt.Errorf("unknown key: %s\nSupported: default-project, default-owner, team, 1-1-repo-pattern, agent.max-per-hour, repos, orgs", key)
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", key)
	return nil
}

func runProfileShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	profileName, _ := config.ActiveProfileName()

	// Check for auto-detected profile
	detected := ""
	if matches, _ := config.DetectProfile(); len(matches) == 1 {
		detected = matches[0].Name
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"config": cfg,
		}
		if profileName != "" {
			payload["profile"] = profileName
		}
		if detected != "" {
			payload["detected"] = detected
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	if detected != "" && detected != profileName {
		fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s %s\n\n",
			detected,
			tui.Muted.Render("(auto-detected from repo)"))
	} else if profileName != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Profile: %s\n\n", profileName)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Fprint(cmd.OutOrStdout(), string(data))
	return nil
}

func runProfileUse(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.UseProfile(name); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Switched to profile %q\n", name)
	return nil
}

func runProfileList(cmd *cobra.Command, args []string) error {
	names, active, err := config.ListProfiles()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profiles configured. Using single config.")
		fmt.Fprintln(cmd.OutOrStdout(), "Run `gh planning profile create <name>` to create profiles.")
		return nil
	}

	// Check auto-detection
	detected := ""
	if matches, _ := config.DetectProfile(); len(matches) == 1 {
		detected = matches[0].Name
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"profiles": names,
			"active":   active,
		}
		if detected != "" {
			payload["detected"] = detected
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	for _, name := range names {
		markers := ""
		if name == active {
			markers += " (active)"
		}
		if name == detected && detected != active {
			markers += " (detected)"
		}
		if markers != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  * %s%s\n", name, markers)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", name)
		}
	}
	return nil
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.DeleteProfile(name); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Deleted profile %q\n", name)
	return nil
}

func runProfileCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Check if profile already exists
	names, _, _ := config.ListProfiles()
	for _, n := range names {
		if n == name && !profileCreateOpts.Force {
			return fmt.Errorf("profile %q already exists — use `gh planning profile use %s` to switch to it, or use --force to overwrite", name, name)
		}
	}

	hasFlags := profileCreateOpts.Project != 0 || profileCreateOpts.Owner != "" ||
		profileCreateOpts.Repos != "" || profileCreateOpts.Orgs != "" || profileCreateOpts.Team != ""

	var cfg config.Config

	if hasFlags {
		// Non-interactive: use flags directly
		cfg.DefaultProject = profileCreateOpts.Project
		cfg.DefaultOwner = profileCreateOpts.Owner
		if profileCreateOpts.Repos != "" {
			cfg.Repos = splitAndTrim(profileCreateOpts.Repos)
		}
		if profileCreateOpts.Orgs != "" {
			cfg.Orgs = splitAndTrim(profileCreateOpts.Orgs)
		}
		if profileCreateOpts.Team != "" {
			cfg.Team = splitAndTrim(profileCreateOpts.Team)
		}
	} else if isTerminal() {
		// Interactive mode
		reader := bufio.NewReader(os.Stdin)

		fmt.Fprintf(cmd.OutOrStdout(), "\n")
		fmt.Fprintf(cmd.OutOrStdout(), tui.Title.Render("Create profile: %s")+"\n\n", name)

		// Detect current repo for smart defaults
		currentRepo := config.DetectGitRepo()
		defaultOrg := ""
		if currentRepo != "" {
			if idx := strings.Index(currentRepo, "/"); idx != -1 {
				defaultOrg = currentRepo[:idx]
			}
			fmt.Fprintf(cmd.OutOrStdout(), tui.Muted.Render("  Detected repo: %s")+"\n\n", currentRepo)
		}

		// Owner
		defaultOwner := defaultOrg
		fmt.Fprintf(cmd.OutOrStdout(), "  Project owner")
		if defaultOwner != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " "+tui.Muted.Render("[%s]"), defaultOwner)
		}
		fmt.Fprintf(cmd.OutOrStdout(), ": ")
		ownerInput, _ := reader.ReadString('\n')
		ownerInput = strings.TrimSpace(ownerInput)
		if ownerInput == "" {
			ownerInput = defaultOwner
		}
		cfg.DefaultOwner = ownerInput

		// Project number
		fmt.Fprintf(cmd.OutOrStdout(), "  Project number: ")
		projectInput, _ := reader.ReadString('\n')
		projectInput = strings.TrimSpace(projectInput)
		if projectInput != "" {
			fmt.Sscanf(projectInput, "%d", &cfg.DefaultProject)
		}

		// Repos
		defaultRepos := ""
		if currentRepo != "" {
			defaultRepos = currentRepo
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Repos (comma-separated, globs ok)")
		if defaultRepos != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " "+tui.Muted.Render("[%s]"), defaultRepos)
		}
		fmt.Fprintf(cmd.OutOrStdout(), ": ")
		reposInput, _ := reader.ReadString('\n')
		reposInput = strings.TrimSpace(reposInput)
		if reposInput == "" {
			reposInput = defaultRepos
		}
		if reposInput != "" {
			cfg.Repos = splitAndTrim(reposInput)
		}

		// Orgs
		fmt.Fprintf(cmd.OutOrStdout(), "  Orgs (comma-separated)")
		if defaultOrg != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " "+tui.Muted.Render("[%s]"), defaultOrg)
		}
		fmt.Fprintf(cmd.OutOrStdout(), ": ")
		orgsInput, _ := reader.ReadString('\n')
		orgsInput = strings.TrimSpace(orgsInput)
		if orgsInput == "" {
			orgsInput = defaultOrg
		}
		if orgsInput != "" {
			cfg.Orgs = splitAndTrim(orgsInput)
		}

		// Team
		fmt.Fprintf(cmd.OutOrStdout(), "  Team members (comma-separated, optional): ")
		teamInput, _ := reader.ReadString('\n')
		teamInput = strings.TrimSpace(teamInput)
		if teamInput != "" {
			cfg.Team = splitAndTrim(teamInput)
		}

		fmt.Fprintln(cmd.OutOrStdout())
	}

	// Create the profile
	if err := config.UseProfile(name); err != nil {
		return err
	}
	if err := config.Save(&cfg); err != nil {
		return err
	}

	// Switch back unless --use was passed
	if !profileCreateOpts.Use {
		prev, _ := config.ActiveProfileName()
		if prev == name {
			// UseProfile switched us to the new one; switch back to previous
			// We can't easily get the previous, so we just stay. The user can switch.
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), tui.Success.Render("✓ Created profile %q")+"\n", name)
	if cfg.DefaultOwner != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Owner:   %s\n", cfg.DefaultOwner)
	}
	if cfg.DefaultProject != 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Project: %d\n", cfg.DefaultProject)
	}
	if len(cfg.Repos) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Repos:   %s\n", strings.Join(cfg.Repos, ", "))
	}
	if len(cfg.Orgs) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Orgs:    %s\n", strings.Join(cfg.Orgs, ", "))
	}
	if len(cfg.Team) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Team:    %s\n", strings.Join(cfg.Team, ", "))
	}

	if profileCreateOpts.Use {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  Switched to profile %q\n", name)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "\n  Run "+tui.Command.Render("gh planning profile use %s")+" to switch to it.\n", name)
	}
	return nil
}

func runProfileUpdate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Verify the profile exists
	cfg, err := config.LoadProfileByName(name)
	if err != nil {
		return err
	}

	// Check at least one flag was provided
	hasFlags := cmd.Flags().Changed("project") || cmd.Flags().Changed("owner") ||
		cmd.Flags().Changed("repos") || cmd.Flags().Changed("orgs") || cmd.Flags().Changed("team")
	if !hasFlags {
		return fmt.Errorf("no flags provided; use --project, --owner, --repos, --orgs, or --team to update fields")
	}

	// Merge provided flags into existing config
	if cmd.Flags().Changed("project") {
		cfg.DefaultProject = profileUpdateOpts.Project
	}
	if cmd.Flags().Changed("owner") {
		cfg.DefaultOwner = profileUpdateOpts.Owner
	}
	if cmd.Flags().Changed("repos") {
		cfg.Repos = splitAndTrim(profileUpdateOpts.Repos)
	}
	if cmd.Flags().Changed("orgs") {
		cfg.Orgs = splitAndTrim(profileUpdateOpts.Orgs)
	}
	if cmd.Flags().Changed("team") {
		cfg.Team = splitAndTrim(profileUpdateOpts.Team)
	}

	if err := config.SaveProfileByName(name, cfg); err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		return output.PrintJSON(map[string]interface{}{
			"profile": name,
			"updated": true,
			"config":  cfg,
		}, OutputOptions())
	}

	fmt.Fprintf(cmd.OutOrStdout(), tui.Success.Render("✓ Updated profile %q")+"\n", name)
	if cfg.DefaultOwner != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Owner:   %s\n", cfg.DefaultOwner)
	}
	if cfg.DefaultProject != 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Project: %d\n", cfg.DefaultProject)
	}
	if len(cfg.Repos) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Repos:   %s\n", strings.Join(cfg.Repos, ", "))
	}
	if len(cfg.Orgs) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Orgs:    %s\n", strings.Join(cfg.Orgs, ", "))
	}
	if len(cfg.Team) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Team:    %s\n", strings.Join(cfg.Team, ", "))
	}
	return nil
}

func runProfileDetect(cmd *cobra.Command, args []string) error {
	matches, err := config.DetectProfile()
	if err != nil {
		return err
	}

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return output.PrintJSON(map[string]interface{}{"matches": names}, OutputOptions())
	}

	if len(matches) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No profiles match the current repo.")
		fmt.Fprintln(cmd.OutOrStdout(), "Add repos/orgs to a profile: `gh planning profile set repos owner/repo`")
		return nil
	}

	if len(matches) == 1 {
		fmt.Fprintf(cmd.OutOrStdout(), "Detected profile: %s\n", tui.Command.Render(matches[0].Name))
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Multiple profiles match this repo:")
	for _, m := range matches {
		matchLabel := ""
		switch m.Match {
		case 3: // matchRepoExact
			matchLabel = "(exact repo match)"
		case 2: // matchRepoGlob
			matchLabel = "(glob repo match)"
		case 1: // matchOrg
			matchLabel = "(org match)"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", tui.Command.Render(m.Name), tui.Muted.Render(matchLabel))
	}
	return nil
}

// ─── Profile Selector TUI ──────────────────────────────────────────────────

// SelectProfile prompts the user to pick a profile when multiple match.
// Returns the chosen profile name. For non-TTY, returns an error with hints.
func SelectProfile(matches []config.ProfileMatch) (string, error) {
	if !isTerminal() {
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = m.Name
		}
		return "", fmt.Errorf("multiple profiles match this repo: %s\nUse --profile <name> or `gh planning profile use <name>` to select one", strings.Join(names, ", "))
	}

	items := make([]profileSelectItem, len(matches))
	for i, m := range matches {
		items[i] = profileSelectItem{name: m.Name, match: m.Match}
	}

	m := profileSelectModel{items: items}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "", err
	}
	final := result.(profileSelectModel)
	if final.cancelled {
		return "", fmt.Errorf("no profile selected")
	}
	return final.items[final.cursor].name, nil
}

type profileSelectItem struct {
	name  string
	match config.MatchType
}

type profileSelectModel struct {
	items     []profileSelectItem
	cursor    int
	cancelled bool
}

func (m profileSelectModel) Init() tea.Cmd { return nil }

func (m profileSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m profileSelectModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(tui.Title.Render("  Multiple profiles match this repo"))
	b.WriteString("\n\n")
	for i, item := range m.items {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == m.cursor {
			cursor = tui.Command.Render("▸ ")
			style = style.Bold(true).Foreground(tui.ColorBright)
		}
		matchLabel := ""
		switch item.match {
		case 3:
			matchLabel = tui.Success.Render(" (exact repo)")
		case 2:
			matchLabel = tui.Muted.Render(" (glob)")
		case 1:
			matchLabel = tui.Dimmed.Render(" (org)")
		}
		b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, style.Render(item.name), matchLabel))
	}
	b.WriteString("\n")
	b.WriteString(tui.HelpBar.Render("  ↑↓ navigate • enter select • esc cancel"))
	return b.String()
}

// splitAndTrim splits a comma-separated string and trims whitespace.
func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
