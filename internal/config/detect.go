package config

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// MatchType represents the priority of a profile match.
type MatchType int

const (
	MatchNone     MatchType = iota
	MatchOrg                // org field matched
	MatchRepoGlob           // repos field matched via glob
	MatchRepoExact          // repos field matched exactly
)

// ProfileMatch describes a profile that matched the current repo.
type ProfileMatch struct {
	Name  string
	Match MatchType
}

// DetectProfile inspects the current directory's git remote origin and returns
// profiles whose repos or orgs fields match. Results are sorted by match
// priority (exact repo > glob repo > org) then alphabetically.
func DetectProfile() ([]ProfileMatch, error) {
	repo := DetectGitRepo()
	if repo == "" {
		return nil, nil
	}

	cf, err := loadFile()
	if err != nil {
		return nil, err
	}
	if len(cf.Profiles) == 0 {
		return nil, nil
	}

	var matches []ProfileMatch
	for name, profile := range cf.Profiles {
		best := bestMatch(repo, profile)
		if best > MatchNone {
			matches = append(matches, ProfileMatch{Name: name, Match: best})
		}
	}

	// Sort: highest match priority first, then alphabetical
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Match != matches[j].Match {
			return matches[i].Match > matches[j].Match
		}
		return matches[i].Name < matches[j].Name
	})

	return matches, nil
}

// bestMatch returns the best match type for a repo against a profile's repos and orgs.
func bestMatch(repo string, profile Config) MatchType {
	best := MatchNone

	// Check repos (exact then glob)
	for _, pattern := range profile.Repos {
		if strings.EqualFold(pattern, repo) {
			return MatchRepoExact
		}
		if matched, _ := filepath.Match(strings.ToLower(pattern), strings.ToLower(repo)); matched {
			if MatchRepoGlob > best {
				best = MatchRepoGlob
			}
		}
	}

	// Check orgs
	if best < MatchOrg {
		repoOwner := repoOrg(repo)
		for _, org := range profile.Orgs {
			if strings.EqualFold(org, repoOwner) {
				best = MatchOrg
				break
			}
		}
	}

	return best
}

// DetectGitRepo returns "owner/repo" from the current directory's git remote origin.
// Returns empty string if not in a git repo or remote can't be parsed.
func DetectGitRepo() string {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return ""
	}
	return parseGitRemote(strings.TrimSpace(string(out)))
}

// parseGitRemote extracts "owner/repo" from various git remote URL formats.
func parseGitRemote(remote string) string {
	// SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(remote, "git@") {
		if idx := strings.Index(remote, ":"); idx != -1 {
			path := remote[idx+1:]
			path = strings.TrimSuffix(path, ".git")
			return path
		}
	}

	// HTTPS: https://github.com/owner/repo.git
	remote = strings.TrimSuffix(remote, ".git")
	parts := strings.Split(remote, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return ""
}

// repoOrg extracts the owner/org from "owner/repo".
func repoOrg(repo string) string {
	if idx := strings.Index(repo, "/"); idx != -1 {
		return repo[:idx]
	}
	return repo
}

// gitRepoRoot returns the top-level directory of the current git repository.
func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
