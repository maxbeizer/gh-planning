package cmd

import (
	"fmt"
	neturl "net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/maxbeizer/gh-planning/internal/config"
)

// projectConfig holds the resolved project owner, number, and full config.
type projectConfig struct {
	Owner   string
	Project int
	Cfg     *config.Config
}

// resolveProjectConfig loads config and resolves owner/project from flag
// overrides with config defaults as fallback. Returns an error if either
// owner or project is still empty/zero after resolution.
func resolveProjectConfig(ownerFlag string, projectFlag int) (*projectConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	owner := ownerFlag
	project := projectFlag
	if owner == "" {
		owner = cfg.DefaultOwner
	}
	if project == 0 {
		project = cfg.DefaultProject
	}
	if owner == "" || project == 0 {
		return nil, fmt.Errorf("project owner and number are required (run `gh planning init`)")
	}
	return &projectConfig{Owner: owner, Project: project, Cfg: cfg}, nil
}

func parseDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	if strings.HasSuffix(value, "d") {
		days := strings.TrimSuffix(value, "d")
		count, err := strconv.Atoi(days)
		if err != nil {
			return 0, err
		}
		return time.Duration(count) * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func humanizeDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	weeks := days / 7
	return fmt.Sprintf("%dw ago", weeks)
}

func projectURL(owner string, project int) string {
return fmt.Sprintf("https://github.com/users/%s/projects/%d", owner, project)
}

func openURL(url string) error {
_, err := runCommand("open", url)
return err
}

func runCommand(name string, args ...string) ([]byte, error) {
cmd := exec.Command(name, args...)
return cmd.Output()
}

// hyperlink wraps text in an OSC 8 terminal hyperlink escape sequence.
// Terminals that support it (iTerm2, Windows Terminal, etc.) make the
// text clickable. Only emits escapes when stdout is a terminal.
func hyperlink(url, text string) string {
	if !isTerminal() {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x07%s\x1b]8;;\x07", url, text)
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// issueRef returns a possibly-hyperlinked "#N" string.
func issueRef(number int, url string) string {
ref := fmt.Sprintf("#%d", number)
if url != "" {
return hyperlink(url, ref)
}
return ref
}

// resolveIssueInput parses an issue argument in one of three formats:
//   - URL: https://github.com/owner/repo/issues/123
//   - Full ref: owner/repo#123
//   - Bare number: 123 (uses repoOverride, then git remote auto-detection)
func resolveIssueInput(input string, repoOverride string) (string, int, error) {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return parseIssueURL(input)
	}
	if strings.Contains(input, "#") {
		return parseIssueRef(input)
	}
	number, err := strconv.Atoi(input)
	if err != nil {
		return "", 0, fmt.Errorf("invalid issue reference: %q (use owner/repo#number, a URL, or a plain number)", input)
	}
	if repoOverride == "" {
		repoOverride = config.DetectGitRepo()
	}
	if repoOverride == "" {
		return "", 0, fmt.Errorf("--repo is required when not in a git repository")
	}
	return repoOverride, number, nil
}

// parseIssueRef parses "owner/repo#number" or a bare number (auto-detects repo).
func parseIssueRef(value string) (string, int, error) {
	if strings.Contains(value, "#") {
		parts := strings.SplitN(value, "#", 2)
		if len(parts) != 2 || parts[0] == "" {
			return "", 0, fmt.Errorf("issue must be in owner/repo#number format")
		}
		number, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid issue number in %q", value)
		}
		return parts[0], number, nil
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return "", 0, fmt.Errorf("issue must be in owner/repo#number format or a plain number")
	}
	repo := config.DetectGitRepo()
	if repo == "" {
		return "", 0, fmt.Errorf("could not detect repo — use owner/repo#number format")
	}
	return repo, number, nil
}

// parseIssueURL extracts owner/repo and issue number from a GitHub issue URL.
func parseIssueURL(rawURL string) (string, int, error) {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return "", 0, err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 {
		return "", 0, fmt.Errorf("invalid issue URL")
	}
	repo := fmt.Sprintf("%s/%s", parts[0], parts[1])
	number, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid issue number in URL")
	}
	return repo, number, nil
}
