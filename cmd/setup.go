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

	fmt.Println("🚀 Welcome to gh-planning!")
	fmt.Println()
	fmt.Println("gh-planning is your command center for GitHub-native project management.")
	fmt.Println("It connects to GitHub Projects (V2) and gives you commands for:")
	fmt.Println()
	fmt.Println("  • status    — view project items grouped by status")
	fmt.Println("  • track     — create issues and add them to your project")
	fmt.Println("  • focus     — set a current working issue (like a pomodoro target)")
	fmt.Println("  • standup   — generate a standup report from recent activity")
	fmt.Println("  • breakdown — split a large issue into sub-issues using AI")
	fmt.Println("  • team      — see what your teammates are working on")
	fmt.Println("  • prep      — generate a 1-1 preparation document")
	fmt.Println("  • queue     — show items ready for agent processing")
	fmt.Println("  • review    — quick review summary for a pull request")
	fmt.Println()
	fmt.Println("Let's get you configured. Press Ctrl+C at any time to cancel.")
	fmt.Println()

	// Step 1: Detect current user
	fmt.Print("Detecting your GitHub username... ")
	currentUser, err := github.CurrentUser(cmd.Context())
	if err != nil {
		fmt.Println("❌")
		return fmt.Errorf("could not detect GitHub user (is `gh` authenticated?): %w", err)
	}
	fmt.Printf("✅ %s\n", currentUser)
	fmt.Println()

	// Step 2: Project owner
	fmt.Println("─── Step 1: Project Owner ───")
	fmt.Println("This is the GitHub user or org that owns the project you want to track.")
	fmt.Printf("Project owner [%s]: ", currentUser)
	owner, err := readLine(reader)
	if err != nil {
		return err
	}
	if owner == "" {
		owner = currentUser
	}
	fmt.Println()

	// Step 3: Project number
	fmt.Println("─── Step 2: Default Project ───")
	fmt.Println("The project number from your GitHub Projects (V2) board.")
	fmt.Println("You can find it in the URL: github.com/users/<owner>/projects/<NUMBER>")
	fmt.Print("Project number: ")
	projectStr, err := readLine(reader)
	if err != nil {
		return err
	}
	project, err := strconv.Atoi(strings.TrimSpace(projectStr))
	if err != nil || project <= 0 {
		return fmt.Errorf("invalid project number: %q", projectStr)
	}

	fmt.Print("Verifying project... ")
	title, err := github.VerifyProject(cmd.Context(), owner, project)
	if err != nil {
		fmt.Println("❌")
		return fmt.Errorf("could not verify project: %w", err)
	}
	fmt.Printf("✅ \"%s\"\n", title)
	fmt.Println()

	// Step 4: Team members
	fmt.Println("─── Step 3: Team (optional) ───")
	fmt.Println("Add GitHub usernames of teammates for `standup --team`, `team`, and `pulse`.")
	fmt.Println("Comma-separated, or leave blank to skip.")
	fmt.Printf("Team members: ")
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
	fmt.Println()

	// Step 5: 1-1 repo pattern
	fmt.Println("─── Step 4: 1-1 Repo Pattern (optional) ───")
	fmt.Println("If you keep 1-1 notes in repos, set a pattern like: myorg/{handle}-1-1")
	fmt.Println("The {handle} placeholder is replaced with each person's GitHub username.")
	fmt.Println("Leave blank to skip.")
	fmt.Printf("1-1 repo pattern: ")
	repoPattern, err := readLine(reader)
	if err != nil {
		return err
	}
	fmt.Println()

	// Step 6: Agent rate limit
	fmt.Println("─── Step 5: Agent Rate Limit (optional) ───")
	fmt.Println("Max agent operations per hour for `queue` processing. 0 = unlimited.")
	fmt.Printf("Agent max per hour [0]: ")
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
	fmt.Println()

	// Step 6: Repos mapping (for auto-detection)
	fmt.Println("─── Step 6: Repository Mapping (optional) ───")
	fmt.Println("List repos that should auto-activate this profile (e.g., github/github).")
	fmt.Println("Supports globs like myorg/* to match all repos under an org.")
	fmt.Println("Comma-separated, or leave blank to skip.")
	fmt.Printf("Repos: ")
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
	fmt.Println()

	// Step 7: Org mapping
	fmt.Println("─── Step 7: Org Mapping (optional) ───")
	fmt.Println("List GitHub orgs that should auto-activate this profile.")
	fmt.Println("Any repo under these orgs will match (lower priority than explicit repos).")
	fmt.Println("Comma-separated, or leave blank to skip.")
	fmt.Printf("Orgs: ")
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
	fmt.Println()

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

	fmt.Println("─── All set! ───")
	fmt.Println()
	if setupOpts.Profile != "" {
		fmt.Printf("  Profile: %s\n", setupOpts.Profile)
	}
	fmt.Printf("  Owner:   %s\n", owner)
	fmt.Printf("  Project: %d (%s)\n", project, title)
	if len(team) > 0 {
		fmt.Printf("  Team:    %s\n", strings.Join(team, ", "))
	}
	if repoPattern != "" {
		fmt.Printf("  1-1:     %s\n", repoPattern)
	}
	if agentMax > 0 {
		fmt.Printf("  Agent:   %d/hour\n", agentMax)
	}
	if len(repos) > 0 {
		fmt.Printf("  Repos:   %s\n", strings.Join(repos, ", "))
	}
	if len(orgs) > 0 {
		fmt.Printf("  Orgs:    %s\n", strings.Join(orgs, ", "))
	}
	fmt.Println()
	fmt.Printf("Config saved to %s\n", cfgPath)
	fmt.Println()
	fmt.Println("Try these next:")
	fmt.Println("  gh planning status        — see your project board")
	fmt.Println("  gh planning focus <issue>  — start focusing on an issue")
	fmt.Println("  gh planning standup        — generate a standup report")
	fmt.Println()
	return nil
}

func readLine(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
