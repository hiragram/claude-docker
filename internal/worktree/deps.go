package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// requiredDeps are commands that must be present for --worktree to work.
var requiredDeps = []string{"git", "zellij"}

// optionalDep describes an optional command with an install hint.
type optionalDep struct {
	name string
	hint string
}

// optionalDeps are commands used by helper scripts, with install hints.
var optionalDeps = []optionalDep{
	{"fswatch", "brew install fswatch"},
	{"fzf", "brew install fzf"},
	{"delta", "brew install git-delta"},
	{"glow", "brew install glow"},
	{"gh", "brew install gh"},
}

// CheckRequiredDeps verifies that all required external commands are available.
func CheckRequiredDeps() error {
	var missing []string
	for _, cmd := range requiredDeps {
		if _, err := exec.LookPath(cmd); err != nil {
			missing = append(missing, cmd)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required commands not found: %s", strings.Join(missing, ", "))
	}
	return nil
}

// CheckOptionalDeps checks optional commands and returns warnings for missing ones.
func CheckOptionalDeps() []string {
	var warnings []string
	for _, dep := range optionalDeps {
		if _, err := exec.LookPath(dep.name); err != nil {
			warnings = append(warnings, fmt.Sprintf("  %s not found (install: %s)", dep.name, dep.hint))
		}
	}
	return warnings
}
