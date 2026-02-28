package worktree

import (
	"testing"
)

func TestCheckRequiredDeps(t *testing.T) {
	err := CheckRequiredDeps()
	// We don't fail the test if zellij is missing, just check the function runs
	if err != nil {
		t.Logf("CheckRequiredDeps returned error (expected if zellij not installed): %v", err)
	}
}

func TestCheckOptionalDeps(t *testing.T) {
	warnings := CheckOptionalDeps()
	// Just verify it returns a slice (may or may not have warnings)
	t.Logf("Optional dep warnings: %v", warnings)
}
