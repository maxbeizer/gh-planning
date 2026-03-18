package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/config"
	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/spf13/cobra"
)

var setupOpts struct {
	Profile string
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive walkthrough to configure gh-planning",
	Long: `Walk through gh-planning configuration step-by-step.

This command guides you through setting up your default project,
team members, and other preferences. It explains what each option
does and saves your choices to the config file.

Use --profile to create or configure a named profile:
  gh planning setup --profile work
  gh planning setup --profile personal

Run this when you first install gh-planning or want to reconfigure.`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().StringVar(&setupOpts.Profile, "profile", "", "Create or configure a named profile")
}

func runSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Fprintln(cmd.OutOrStdout(), "🚀 Welcome to gh-planning!")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "gh-planning is your command center for GitHub-native project management.")
	fmt.Fprintln(cmd.OutOrStdout(), "It connects to GitHub Projects (V2) and gives you commands for:")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "  • status    — view project items grouped by status")
	fmt.Fprintln(cmd.OutOrStdout(), "  • track     — create issues and add them to your project")
	fmt.Fprintln(cmd.OutOrStdout(), "  • focus     — set a current working issue (like a pomodoro target)")
	fmt.Fprintln(cmd.OutOrStdout(), "  • standup   — generate a standup report from recent activity")
	fmt.Fprintln(cmd.OutOrStdout(), "  • breakdown — split a large issue into sub-issues using AI")
	fmt.Fprintln(cmd.OutOrStdout(), "  • team      — see what your teammates are working on")
	fmt.Fprintln(cmd.OutOrStdout(), "  • prep      — generate a 1-1 preparation document")
	fmt.Fprintln(cmd.OutOrStdout(), "  • queue     — show items ready for agent processing")
	fmt.Fprintln(cmd.OutOrStdout(), "  • review    — quick review summary for a pull request")
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Let's get you configured. Press Ctrl+C at any time to cancel.")
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 1: Detect current user
	fmt.Fprint(cmd.OutOrStdout(), "Detecting your GitHub username... ")
	currentUser, err := github.CurrentUser(cmd.Context())
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "❌")
		return fmt.Errorf("could not detect GitHub user (is `gh` authenticated?): %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✅ %s\n", currentUser)
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 2: Project owner
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 1: Project Owner ───")
	fmt.Fprintln(cmd.OutOrStdout(), "This is the GitHub user or org that owns the project you want to track.")
	fmt.Fprintf(cmd.OutOrStdout(), "Project owner [%s]: ", currentUser)
	owner, err := readLine(reader)
	if err != nil {
		return err
	}
	if owner == "" {
		owner = currentUser
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 3: Project number
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 2: Default Project ───")
	fmt.Fprintln(cmd.OutOrStdout(), "The project number from your GitHub Projects (V2) board.")
	fmt.Fprintln(cmd.OutOrStdout(), "You can find it in the URL: github.com/users/<owner>/projects/<NUMBER>")
	fmt.Fprint(cmd.OutOrStdout(), "Project number: ")
	projectStr, err := readLine(reader)
	if err != nil {
		return err
	}
	project, err := strconv.Atoi(strings.TrimSpace(projectStr))
	if err != nil || project <= 0 {
		return fmt.Errorf("invalid project number: %q", projectStr)
	}

	fmt.Fprint(cmd.OutOrStdout(), "Verifying project... ")
	title, err := github.VerifyProject(cmd.Context(), owner, project)
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), "❌")
		return fmt.Errorf("could not verify project: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✅ \"%s\"\n", title)
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 4: Team members
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 3: Team (optional) ───")
	fmt.Fprintln(cmd.OutOrStdout(), "Add GitHub usernames of teammates for `standup --team`, `team`, and `pulse`.")
	fmt.Fprintln(cmd.OutOrStdout(), "Comma-separated, or leave blank to skip.")
	fmt.Fprintf(cmd.OutOrStdout(), "Team members: ")
	teamInput, err := readLine(reader)
	if err != nil {
		return err
	}
	var team []string
	if teamInput != "" {
		for _, member := range strings.Split(teamInput, ",") {
			member = strings.TrimSpace(member)
			if member != "" {
				team = append(team, member)
			}
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 5: 1-1 repo pattern
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 4: 1-1 Repo Pattern (optional) ───")
	fmt.Fprintln(cmd.OutOrStdout(), "If you keep 1-1 notes in repos, set a pattern like: myorg/{handle}-1-1")
	fmt.Fprintln(cmd.OutOrStdout(), "The {handle} placeholder is replaced with each person's GitHub username.")
	fmt.Fprintln(cmd.OutOrStdout(), "Leave blank to skip.")
	fmt.Fprintf(cmd.OutOrStdout(), "1-1 repo pattern: ")
	repoPattern, err := readLine(reader)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 6: Agent rate limit
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 5: Agent Rate Limit (optional) ───")
	fmt.Fprintln(cmd.OutOrStdout(), "Max agent operations per hour for `queue` processing. 0 = unlimited.")
	fmt.Fprintf(cmd.OutOrStdout(), "Agent max per hour [0]: ")
	agentStr, err := readLine(reader)
	if err != nil {
		return err
	}
	agentMax := 0
	if agentStr != "" {
		agentMax, err = strconv.Atoi(strings.TrimSpace(agentStr))
		if err != nil {
			return fmt.Errorf("invalid number: %q", agentStr)
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 6: Repos mapping (for auto-detection)
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 6: Repository Mapping (optional) ───")
	fmt.Fprintln(cmd.OutOrStdout(), "List repos that should auto-activate this profile (e.g., github/github).")
	fmt.Fprintln(cmd.OutOrStdout(), "Supports globs like myorg/* to match all repos under an org.")
	fmt.Fprintln(cmd.OutOrStdout(), "Comma-separated, or leave blank to skip.")
	fmt.Fprintf(cmd.OutOrStdout(), "Repos: ")
	reposInput, err := readLine(reader)
	if err != nil {
		return err
	}
	var repos []string
	if reposInput != "" {
		for _, r := range strings.Split(reposInput, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				repos = append(repos, r)
			}
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Step 7: Org mapping
	fmt.Fprintln(cmd.OutOrStdout(), "─── Step 7: Org Mapping (optional) ───")
	fmt.Fprintln(cmd.OutOrStdout(), "List GitHub orgs that should auto-activate this profile.")
	fmt.Fprintln(cmd.OutOrStdout(), "Any repo under these orgs will match (lower priority than explicit repos).")
	fmt.Fprintln(cmd.OutOrStdout(), "Comma-separated, or leave blank to skip.")
	fmt.Fprintf(cmd.OutOrStdout(), "Orgs: ")
	orgsInput, err := readLine(reader)
	if err != nil {
		return err
	}
	var orgs []string
	if orgsInput != "" {
		for _, o := range strings.Split(orgsInput, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				orgs = append(orgs, o)
			}
		}
	}
	fmt.Fprintln(cmd.OutOrStdout())

	// Switch to profile if requested
	if setupOpts.Profile != "" {
		if err := config.UseProfile(setupOpts.Profile); err != nil {
			return err
		}
	}

	// Save config
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.DefaultOwner = owner
	cfg.DefaultProject = project
	if len(team) > 0 {
		cfg.Team = team
	}
	if repoPattern != "" {
		cfg.OneOnOneRepoPattern = repoPattern
	}
	if agentMax > 0 {
		cfg.AgentMaxPerHour = agentMax
	}
	if len(repos) > 0 {
		cfg.Repos = repos
	}
	if len(orgs) > 0 {
		cfg.Orgs = orgs
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	cfgPath, _ := config.Path()

	fmt.Fprintln(cmd.OutOrStdout(), "─── All set! ───")
	fmt.Fprintln(cmd.OutOrStdout())
	if setupOpts.Profile != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Profile: %s\n", setupOpts.Profile)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "  Owner:   %s\n", owner)
	fmt.Fprintf(cmd.OutOrStdout(), "  Project: %d (%s)\n", project, title)
	if len(team) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Team:    %s\n", strings.Join(team, ", "))
	}
	if repoPattern != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  1-1:     %s\n", repoPattern)
	}
	if agentMax > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Agent:   %d/hour\n", agentMax)
	}
	if len(repos) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Repos:   %s\n", strings.Join(repos, ", "))
	}
	if len(orgs) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Orgs:    %s\n", strings.Join(orgs, ", "))
	}
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintf(cmd.OutOrStdout(), "Config saved to %s\n", cfgPath)
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Try these next:")
	fmt.Fprintln(cmd.OutOrStdout(), "  gh planning status        — see your project board")
	fmt.Fprintln(cmd.OutOrStdout(), "  gh planning focus <issue>  — start focusing on an issue")
	fmt.Fprintln(cmd.OutOrStdout(), "  gh planning standup        — generate a standup report")
	fmt.Fprintln(cmd.OutOrStdout())
	return nil
}

func readLine(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
