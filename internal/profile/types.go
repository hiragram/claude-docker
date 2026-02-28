package profile

// Config represents the top-level .agent-workspace.yml file.
type Config struct {
	Default  string             `yaml:"default"`
	Profiles map[string]Profile `yaml:"profiles"`
}

// Profile describes a single named workspace profile.
type Profile struct {
	Worktree    *WorktreeConfig `yaml:"worktree,omitempty"`
	Environment Environment     `yaml:"environment"`
	Launch      LaunchMode      `yaml:"launch"`
	Zellij      *ZellijConfig   `yaml:"zellij,omitempty"`
}

// WorktreeConfig controls git worktree creation.
type WorktreeConfig struct {
	Base string `yaml:"base,omitempty"` // default: "origin/main"
}

// EffectiveBase returns the base ref, defaulting to "origin/main" if empty.
func (w *WorktreeConfig) EffectiveBase() string {
	if w.Base != "" {
		return w.Base
	}
	return "origin/main"
}

// ZellijConfig controls zellij session settings.
type ZellijConfig struct {
	Layout string `yaml:"layout,omitempty"` // "default" or custom path (future)
}

// Environment specifies where the main process runs.
type Environment string

const (
	EnvironmentHost   Environment = "host"
	EnvironmentDocker Environment = "docker"
)

// LaunchMode specifies what to launch.
type LaunchMode string

const (
	LaunchShell  LaunchMode = "shell"
	LaunchClaude LaunchMode = "claude"
	LaunchZellij LaunchMode = "zellij"
)
