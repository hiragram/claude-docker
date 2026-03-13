package profile

// MergeProfile merges override into base.
// Non-zero values in override take precedence over base.
func MergeProfile(base, override Profile) Profile {
	merged := base

	if override.Environment != "" {
		merged.Environment = override.Environment
	}
	if override.Launch != "" {
		merged.Launch = override.Launch
	}
	if override.Worktree != nil {
		merged.Worktree = override.Worktree
	}
	if override.Zellij != nil {
		merged.Zellij = override.Zellij
	}
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

// MergeConfig merges a user config on top of the builtin config.
//   - Builtin-only profiles are preserved as-is.
//   - User-only profiles are added as-is.
//   - Profiles in both are merged (builtin base + user overlay).
//   - User's Default takes precedence if non-empty.
func MergeConfig(builtin, user Config) Config {
	merged := Config{
		Default:  builtin.Default,
		Profiles: make(map[string]Profile, len(builtin.Profiles)+len(user.Profiles)),
	}

	// Start with all builtin profiles
	for name, p := range builtin.Profiles {
		merged.Profiles[name] = p
	}

	// Overlay user profiles
	for name, userProfile := range user.Profiles {
		if base, ok := builtin.Profiles[name]; ok {
			merged.Profiles[name] = MergeProfile(base, userProfile)
		} else {
			merged.Profiles[name] = userProfile
		}
	}

	// User's Default takes precedence if non-empty
	if user.Default != "" {
		merged.Default = user.Default
	}

	return merged
}
