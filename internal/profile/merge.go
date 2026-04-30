package profile

// MergeProfile merges override into base.
// Non-zero values in override take precedence over base.
// Sub-structs (Worktree, Zellij) are merged field-by-field rather than replaced wholesale.
func MergeProfile(base, override Profile) Profile {
	merged := base

	if override.Environment != "" {
		merged.Environment = override.Environment
	}
	if override.Launch != "" {
		merged.Launch = override.Launch
	}
	merged.Worktree = mergeWorktree(merged.Worktree, override.Worktree)
	merged.Zellij = mergeZellij(merged.Zellij, override.Zellij)
	if override.Env != nil {
		envCopy := make(map[string]string, len(merged.Env)+len(override.Env))
		for k, v := range merged.Env {
			envCopy[k] = v
		}
		for k, v := range override.Env {
			envCopy[k] = v
		}
		merged.Env = envCopy
	}
	if override.Dockerfile != "" {
		merged.Dockerfile = override.Dockerfile
	}

	return merged
}

func mergeWorktree(base, override *WorktreeConfig) *WorktreeConfig {
	if override == nil {
		return base
	}
	if base == nil {
		v := *override
		return &v
	}
	merged := *base
	if override.Base != "" {
		merged.Base = override.Base
	}
	if override.Dir != "" {
		merged.Dir = override.Dir
	}
	if override.OnCreate != "" {
		merged.OnCreate = override.OnCreate
	}
	if override.OnEnd != "" {
		merged.OnEnd = override.OnEnd
	}
	return &merged
}

func mergeZellij(base, override *ZellijConfig) *ZellijConfig {
	if override == nil {
		return base
	}
	if base == nil {
		v := *override
		return &v
	}
	merged := *base
	if override.Layout != "" {
		merged.Layout = override.Layout
	}
	return &merged
}

// MergeConfig merges a user config on top of the builtin config.
//   - Builtin-only profiles are preserved as-is.
//   - User-only profiles are added as-is.
//   - Profiles in both are merged (builtin base + user overlay).
//   - User's Default takes precedence if non-empty.
//   - User's top-level Profile fields override the builtin top-level fields.
//
// This function does NOT apply the top-level Profile to each profile; that is
// done by ApplyTopLevel so that callers can inspect the raw per-profile data
// if they need to.
func MergeConfig(builtin, user Config) Config {
	merged := Config{
		Default:  builtin.Default,
		Profile:  MergeProfile(builtin.Profile, user.Profile),
		Profiles: make(map[string]Profile, len(builtin.Profiles)+len(user.Profiles)),
	}

	for name, p := range builtin.Profiles {
		merged.Profiles[name] = p
	}

	for name, userProfile := range user.Profiles {
		if base, ok := merged.Profiles[name]; ok {
			merged.Profiles[name] = MergeProfile(base, userProfile)
		} else {
			merged.Profiles[name] = userProfile
		}
	}

	if user.Default != "" {
		merged.Default = user.Default
	}

	return merged
}

// ApplyTopLevel returns a new Config in which each profile is the result of
// merging the top-level Profile defaults with the per-profile overrides.
// The returned Config's top-level Profile is left as-is (it's redundant once
// applied, but harmless and useful for round-tripping).
func ApplyTopLevel(cfg Config) Config {
	out := Config{
		Default:  cfg.Default,
		Profile:  cfg.Profile,
		Profiles: make(map[string]Profile, len(cfg.Profiles)),
		Source:   cfg.Source,
	}
	for name, p := range cfg.Profiles {
		out.Profiles[name] = MergeProfile(cfg.Profile, p)
	}
	return out
}
