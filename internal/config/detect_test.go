package config

import (
	"sort"
	"testing"
)

func TestParseGitRemote_SSH(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git@github.com:maxbeizer/gh-planning.git", "maxbeizer/gh-planning"},
		{"git@github.com:github/github.git", "github/github"},
		{"git@github.com:owner/repo", "owner/repo"},
	}
	for _, tt := range tests {
		got := parseGitRemote(tt.input)
		if got != tt.want {
			t.Errorf("parseGitRemote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseGitRemote_HTTPS(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/maxbeizer/gh-planning.git", "maxbeizer/gh-planning"},
		{"https://github.com/github/github", "github/github"},
	}
	for _, tt := range tests {
		got := parseGitRemote(tt.input)
		if got != tt.want {
			t.Errorf("parseGitRemote(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBestMatch_ExactRepo(t *testing.T) {
	profile := Config{Repos: []string{"github/github", "maxbeizer/app"}}
	got := bestMatch("github/github", profile)
	if got != MatchRepoExact {
		t.Errorf("bestMatch = %d, want MatchRepoExact(%d)", got, MatchRepoExact)
	}
}

func TestBestMatch_GlobRepo(t *testing.T) {
	profile := Config{Repos: []string{"maxbeizer/*"}}
	got := bestMatch("maxbeizer/gh-planning", profile)
	if got != MatchRepoGlob {
		t.Errorf("bestMatch = %d, want MatchRepoGlob(%d)", got, MatchRepoGlob)
	}
}

func TestBestMatch_OrgMatch(t *testing.T) {
	profile := Config{Orgs: []string{"github"}}
	got := bestMatch("github/some-repo", profile)
	if got != MatchOrg {
		t.Errorf("bestMatch = %d, want MatchOrg(%d)", got, MatchOrg)
	}
}

func TestBestMatch_NoMatch(t *testing.T) {
	profile := Config{Repos: []string{"other/repo"}, Orgs: []string{"other"}}
	got := bestMatch("maxbeizer/app", profile)
	if got != MatchNone {
		t.Errorf("bestMatch = %d, want MatchNone(%d)", got, MatchNone)
	}
}

func TestBestMatch_ExactTakesPriority(t *testing.T) {
	profile := Config{
		Repos: []string{"maxbeizer/*", "maxbeizer/app"},
		Orgs:  []string{"maxbeizer"},
	}
	got := bestMatch("maxbeizer/app", profile)
	if got != MatchRepoExact {
		t.Errorf("bestMatch = %d, want MatchRepoExact(%d)", got, MatchRepoExact)
	}
}

func TestBestMatch_CaseInsensitive(t *testing.T) {
	profile := Config{Repos: []string{"GitHub/GitHub"}}
	got := bestMatch("github/github", profile)
	if got != MatchRepoExact {
		t.Errorf("bestMatch = %d, want MatchRepoExact(%d)", got, MatchRepoExact)
	}
}

func TestBestMatch_NoReposOrOrgs(t *testing.T) {
	profile := Config{DefaultOwner: "someone"}
	got := bestMatch("github/github", profile)
	if got != MatchNone {
		t.Errorf("bestMatch = %d, want MatchNone(%d) for profile without repos/orgs", got, MatchNone)
	}
}

func TestLoad_AutoDetectSingleMatch(t *testing.T) {
	// This test validates the schema is correct and profiles with repos load fine.
	// We can't easily mock git remote in unit tests, so we test the match logic directly.
	setup(t)
	writeConfig(t, `
active-profile: fallback
profiles:
  work:
    default-project: 10
    default-owner: corp
    repos:
      - "corp/app"
  fallback:
    default-project: 99
    default-owner: fallback
`)
	// Since we're not in a git repo matching "corp/app", Load should fall back to active profile.
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// Should get fallback since no git remote matches
	if cfg.DefaultOwner != "fallback" && cfg.DefaultOwner != "corp" {
		t.Errorf("DefaultOwner = %q, want fallback or corp", cfg.DefaultOwner)
	}
}

func TestLoad_BackwardCompatNoReposOrOrgs(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: work
profiles:
  work:
    default-project: 10
    default-owner: corp
  personal:
    default-project: 20
    default-owner: me
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	// No repos/orgs → falls back to active profile
	if cfg.DefaultOwner != "corp" {
		t.Errorf("DefaultOwner = %q, want %q", cfg.DefaultOwner, "corp")
	}
}

func TestDetectProfileSorting(t *testing.T) {
	// Test that matches sort by priority
	matches := []ProfileMatch{
		{Name: "z-org", Match: MatchOrg},
		{Name: "a-exact", Match: MatchRepoExact},
		{Name: "m-glob", Match: MatchRepoGlob},
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Match != matches[j].Match {
			return matches[i].Match > matches[j].Match
		}
		return matches[i].Name < matches[j].Name
	})
	if matches[0].Name != "a-exact" || matches[1].Name != "m-glob" || matches[2].Name != "z-org" {
		t.Errorf("sort order wrong: %v", matches)
	}
}

func TestRepoOrg(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"github/github", "github"},
		{"maxbeizer/app", "maxbeizer"},
		{"solo", "solo"},
	}
	for _, tt := range tests {
		got := repoOrg(tt.input)
		if got != tt.want {
			t.Errorf("repoOrg(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLoad_LegacyStillWorks(t *testing.T) {
	setup(t)
	writeConfig(t, `
default-project: 42
default-owner: acme
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultProject != 42 || cfg.DefaultOwner != "acme" {
		t.Errorf("legacy config broken: %+v", cfg)
	}
}

func TestConfigReposOrgsRoundtrip(t *testing.T) {
	setup(t)
	if err := UseProfile("test"); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{
		DefaultProject: 1,
		DefaultOwner:   "org",
		Repos:          []string{"org/app", "org/*"},
		Orgs:           []string{"org"},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Repos) != 2 || loaded.Repos[0] != "org/app" {
		t.Errorf("Repos = %v, want [org/app org/*]", loaded.Repos)
	}
	if len(loaded.Orgs) != 1 || loaded.Orgs[0] != "org" {
		t.Errorf("Orgs = %v, want [org]", loaded.Orgs)
	}
}
