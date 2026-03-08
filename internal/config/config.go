package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultProject     int      `yaml:"default-project" json:"defaultProject"`
	DefaultOwner       string   `yaml:"default-owner" json:"defaultOwner"`
	Team               []string `yaml:"team,omitempty" json:"team,omitempty"`
	OneOnOneRepoPattern string   `yaml:"1-1-repo-pattern,omitempty" json:"oneOnOneRepoPattern,omitempty"`
	AgentMaxPerHour    int      `yaml:"agent.max-per-hour,omitempty" json:"agentMaxPerHour,omitempty"`
}

func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gh-planning", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
