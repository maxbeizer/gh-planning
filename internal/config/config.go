package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProject      int      `yaml:"default-project" json:"defaultProject"`
	DefaultOwner        string   `yaml:"default-owner" json:"defaultOwner"`
	Team                []string `yaml:"team,omitempty" json:"team,omitempty"`
	OneOnOneRepoPattern string   `yaml:"1-1-repo-pattern,omitempty" json:"oneOnOneRepoPattern,omitempty"`
	AgentMaxPerHour     int      `yaml:"agent.max-per-hour,omitempty" json:"agentMaxPerHour,omitempty"`
}

// configFile is the on-disk structure that supports named profiles.
type configFile struct {
	// Legacy top-level fields (used when no profiles exist).
	Config `yaml:",inline"`

	ActiveProfile string            `yaml:"active-profile,omitempty"`
	Profiles      map[string]Config `yaml:"profiles,omitempty"`
}

// dirOverride, when non-empty, replaces the default config directory.
// Used by tests to redirect config I/O to a temp directory.
var dirOverride string

func Dir() (string, error) {
	if dirOverride != "" {
		return dirOverride, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gh-planning"), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func loadFile() (*configFile, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &configFile{}, nil
		}
		return nil, err
	}
	cf := &configFile{}
	if err := yaml.Unmarshal(data, cf); err != nil {
		return nil, err
	}
	return cf, nil
}

func saveFile(cf *configFile) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cf)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load returns the active profile's config. If no profiles exist, it
// falls back to legacy top-level fields for backward compatibility.
func Load() (*Config, error) {
	cf, err := loadFile()
	if err != nil {
		return nil, err
	}
	if len(cf.Profiles) == 0 {
		cfg := cf.Config
		return &cfg, nil
	}
	name := cf.ActiveProfile
	if name == "" {
		name = "default"
	}
	profile, ok := cf.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found (available: %s)", name, profileNames(cf))
	}
	return &profile, nil
}

// Save writes the config. When profiles exist it updates the active
// profile; otherwise it writes legacy top-level fields.
func Save(cfg *Config) error {
	cf, err := loadFile()
	if err != nil {
		return err
	}
	if len(cf.Profiles) == 0 {
		cf.Config = *cfg
		return saveFile(cf)
	}
	name := cf.ActiveProfile
	if name == "" {
		name = "default"
	}
	cf.Profiles[name] = *cfg
	return saveFile(cf)
}

// ActiveProfileName returns the name of the active profile, or empty
// string if profiles are not in use.
func ActiveProfileName() (string, error) {
	cf, err := loadFile()
	if err != nil {
		return "", err
	}
	if len(cf.Profiles) == 0 {
		return "", nil
	}
	if cf.ActiveProfile == "" {
		return "default", nil
	}
	return cf.ActiveProfile, nil
}

// ListProfiles returns profile names and the active one.
func ListProfiles() (names []string, active string, err error) {
	cf, err := loadFile()
	if err != nil {
		return nil, "", err
	}
	if len(cf.Profiles) == 0 {
		return nil, "", nil
	}
	active = cf.ActiveProfile
	if active == "" {
		active = "default"
	}
	for name := range cf.Profiles {
		names = append(names, name)
	}
	return names, active, nil
}

// UseProfile switches the active profile. If the profile doesn't exist
// yet it creates an empty one.
func UseProfile(name string) error {
	cf, err := loadFile()
	if err != nil {
		return err
	}
	// Migrate legacy config into "default" profile on first use.
	if len(cf.Profiles) == 0 {
		cf.Profiles = map[string]Config{
			"default": cf.Config,
		}
		cf.Config = Config{}
	}
	if _, ok := cf.Profiles[name]; !ok {
		cf.Profiles[name] = Config{}
	}
	cf.ActiveProfile = name
	return saveFile(cf)
}

// DeleteProfile removes a profile. Cannot delete the active profile.
func DeleteProfile(name string) error {
	cf, err := loadFile()
	if err != nil {
		return err
	}
	if len(cf.Profiles) == 0 {
		return fmt.Errorf("no profiles configured")
	}
	active := cf.ActiveProfile
	if active == "" {
		active = "default"
	}
	if name == active {
		return fmt.Errorf("cannot delete the active profile %q; switch first with `config use`", name)
	}
	if _, ok := cf.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(cf.Profiles, name)
	return saveFile(cf)
}

func profileNames(cf *configFile) string {
	names := []string{}
	for name := range cf.Profiles {
		names = append(names, name)
	}
	if len(names) == 0 {
		return "(none)"
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
