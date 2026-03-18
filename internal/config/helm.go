package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// HelmProject holds the [project] section of helm.toml.
type HelmProject struct {
	Board int    `toml:"board"`
	Owner string `toml:"owner"`
}

// HelmConfig represents the subset of helm.toml we care about.
type HelmConfig struct {
	Project HelmProject `toml:"project"`
}

// LoadHelmToml reads helm.toml (or helm-manager.toml) from the git repo root.
// Returns nil without error if no file is found.
func LoadHelmToml() (*HelmConfig, error) {
	root, err := gitRepoRoot()
	if err != nil {
		return nil, nil
	}

	for _, name := range []string{"helm.toml", "helm-manager.toml"} {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg HelmConfig
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	return nil, nil
}
