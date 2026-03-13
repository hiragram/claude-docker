package profile

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = ".agent-workspace.yml"

// builtinConfig is used when no config file is found.
var builtinConfig = Config{
	Default: "worktree-zellij",
	Profiles: map[string]Profile{
		"claude": {
			Environment: EnvironmentDocker,
			Launch:      LaunchClaude,
		},
		"worktree-zellij": {
			Worktree:    &WorktreeConfig{},
			Environment: EnvironmentDocker,
			Launch:      LaunchZellij,
			Zellij:      &ZellijConfig{Layout: "default"},
		},
	},
}

// Load finds and loads the config file.
// It looks for .agent-workspace.yml at the git repository root.
// If no config file is found, it returns the built-in default config.
func Load() (*Config, error) {
	repoRoot, err := findGitRoot()
	if err != nil {
		// Not in a git repo — use built-in default
		cfg := builtinConfig
		cfg.Source = ConfigSource{IsBuiltin: true}
		return &cfg, nil
	}

	configPath := filepath.Join(repoRoot, configFileName)
	return LoadFile(configPath)
}

// LoadFile loads a config from the given file path.
// If the file does not exist, it returns the built-in default config.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := builtinConfig
			cfg.Source = ConfigSource{IsBuiltin: true}
			return &cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	userCfg, err := Parse(data)
	if err != nil {
		return nil, err
	}

	merged := MergeConfig(builtinConfig, *userCfg)
	merged.Source = ConfigSource{FilePath: path}
	return &merged, nil
}

// Parse parses YAML bytes into a Config.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return &cfg, nil
}

// findGitRoot returns the top-level directory of the current git repository.
var findGitRoot = func() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}
