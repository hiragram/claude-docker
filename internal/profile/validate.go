package profile

import (
	"fmt"
	"strings"
)

// Validate checks that a profile configuration is semantically valid.
func Validate(p Profile) error {
	// Validate environment
	switch p.Environment {
	case EnvironmentHost, EnvironmentDocker:
		// ok
	case "":
		return fmt.Errorf("environment is required (\"host\" or \"docker\")")
	default:
		return fmt.Errorf("unknown environment: %q (must be \"host\" or \"docker\")", p.Environment)
	}

	// Validate launch mode
	switch p.Launch {
	case LaunchShell, LaunchClaude, LaunchZellij:
		// ok
	case "":
		return fmt.Errorf("launch is required (\"shell\", \"claude\", or \"zellij\")")
	default:
		return fmt.Errorf("unknown launch mode: %q (must be \"shell\", \"claude\", or \"zellij\")", p.Launch)
	}

	// Validate zellij config is only used with launch: zellij
	if p.Zellij != nil && p.Launch != LaunchZellij {
		return fmt.Errorf("zellij config is only valid with launch: zellij")
	}

	return nil
}

// ValidateConfig checks the entire config for errors.
func ValidateConfig(cfg *Config) error {
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles defined")
	}

	// Check that default profile exists if specified
	if cfg.Default != "" {
		if _, ok := cfg.Profiles[cfg.Default]; !ok {
			return fmt.Errorf("default profile %q not found in profiles", cfg.Default)
		}
	}

	// Validate each profile
	var errs []string
	for name, p := range cfg.Profiles {
		if err := Validate(p); err != nil {
			errs = append(errs, fmt.Sprintf("profile %q: %v", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation errors:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}
