package config

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func setup(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	dirOverride = dir
	t.Cleanup(func() { dirOverride = "" })
}

func writeConfig(t *testing.T, content string) {
	t.Helper()
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_NoFile(t *testing.T) {
	setup(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultOwner != "" || cfg.DefaultProject != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoad_LegacyConfig(t *testing.T) {
	setup(t)
	writeConfig(t, `
default-project: 42
default-owner: acme
team:
  - alice
  - bob
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultProject != 42 {
		t.Errorf("DefaultProject = %d, want 42", cfg.DefaultProject)
	}
	if cfg.DefaultOwner != "acme" {
		t.Errorf("DefaultOwner = %q, want %q", cfg.DefaultOwner, "acme")
	}
	if len(cfg.Team) != 2 || cfg.Team[0] != "alice" {
		t.Errorf("Team = %v, want [alice bob]", cfg.Team)
	}
}

func TestLoad_ProfilesConfig(t *testing.T) {
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
	if cfg.DefaultProject != 10 {
		t.Errorf("DefaultProject = %d, want 10", cfg.DefaultProject)
	}
	if cfg.DefaultOwner != "corp" {
		t.Errorf("DefaultOwner = %q, want %q", cfg.DefaultOwner, "corp")
	}
}

func TestLoad_ProfilesDefaultFallback(t *testing.T) {
	setup(t)
	writeConfig(t, `
profiles:
  default:
    default-project: 99
    default-owner: fallback
`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultProject != 99 {
		t.Errorf("DefaultProject = %d, want 99", cfg.DefaultProject)
	}
}

func TestSave_Roundtrip(t *testing.T) {
	setup(t)
	original := &Config{
		DefaultProject: 7,
		DefaultOwner:   "test-org",
		Team:           []string{"charlie"},
	}
	if err := Save(original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.DefaultProject != original.DefaultProject {
		t.Errorf("DefaultProject = %d, want %d", loaded.DefaultProject, original.DefaultProject)
	}
	if loaded.DefaultOwner != original.DefaultOwner {
		t.Errorf("DefaultOwner = %q, want %q", loaded.DefaultOwner, original.DefaultOwner)
	}
	if len(loaded.Team) != 1 || loaded.Team[0] != "charlie" {
		t.Errorf("Team = %v, want [charlie]", loaded.Team)
	}
}

func TestSave_RoundtripWithProfiles(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: dev
profiles:
  dev:
    default-project: 1
    default-owner: old
`)
	cfg := &Config{DefaultProject: 2, DefaultOwner: "new"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.DefaultOwner != "new" {
		t.Errorf("DefaultOwner = %q, want %q", loaded.DefaultOwner, "new")
	}
}

func TestUseProfile_CreatesAndMigratesLegacy(t *testing.T) {
	setup(t)
	// Start with a legacy config
	writeConfig(t, `
default-project: 5
default-owner: legacy-org
`)
	if err := UseProfile("staging"); err != nil {
		t.Fatalf("UseProfile() error: %v", err)
	}
	// The legacy config should be migrated to "default" profile
	names, active, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error: %v", err)
	}
	if active != "staging" {
		t.Errorf("active = %q, want %q", active, "staging")
	}
	sort.Strings(names)
	if len(names) != 2 {
		t.Fatalf("expected 2 profiles, got %d: %v", len(names), names)
	}
	if names[0] != "default" || names[1] != "staging" {
		t.Errorf("names = %v, want [default staging]", names)
	}
	// Verify legacy values migrated to "default" profile
	if err := UseProfile("default"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProject != 5 || cfg.DefaultOwner != "legacy-org" {
		t.Errorf("migrated default profile = %+v, want project=5 owner=legacy-org", cfg)
	}
}

func TestUseProfile_SwitchesActive(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: a
profiles:
  a:
    default-owner: alpha
  b:
    default-owner: beta
`)
	if err := UseProfile("b"); err != nil {
		t.Fatalf("UseProfile() error: %v", err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultOwner != "beta" {
		t.Errorf("DefaultOwner = %q, want %q", cfg.DefaultOwner, "beta")
	}
}

func TestDeleteProfile_ErrorsOnActive(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: main
profiles:
  main:
    default-owner: x
  other:
    default-owner: y
`)
	err := DeleteProfile("main")
	if err == nil {
		t.Fatal("expected error deleting active profile, got nil")
	}
}

func TestDeleteProfile_RemovesInactive(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: main
profiles:
  main:
    default-owner: x
  other:
    default-owner: y
`)
	if err := DeleteProfile("other"); err != nil {
		t.Fatalf("DeleteProfile() error: %v", err)
	}
	names, _, err := ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "main" {
		t.Errorf("profiles = %v, want [main]", names)
	}
}

func TestDeleteProfile_NoProfiles(t *testing.T) {
	setup(t)
	err := DeleteProfile("anything")
	if err == nil {
		t.Fatal("expected error when no profiles, got nil")
	}
}

func TestListProfiles_NoProfiles(t *testing.T) {
	setup(t)
	names, active, err := ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if names != nil || active != "" {
		t.Errorf("expected nil names and empty active, got names=%v active=%q", names, active)
	}
}

func TestListProfiles_ReturnsNamesAndActive(t *testing.T) {
	setup(t)
	writeConfig(t, `
active-profile: second
profiles:
  first:
    default-owner: a
  second:
    default-owner: b
`)
	names, active, err := ListProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if active != "second" {
		t.Errorf("active = %q, want %q", active, "second")
	}
	sort.Strings(names)
	if len(names) != 2 || names[0] != "first" || names[1] != "second" {
		t.Errorf("names = %v, want [first second]", names)
	}
}
