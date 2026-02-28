package stage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/worktree"
)

// WorktreeStage creates a git worktree for the workspace.
type WorktreeStage struct{}

func (s *WorktreeStage) Name() string { return "worktree" }

func (s *WorktreeStage) Run(_ context.Context, ec *pipeline.ExecutionContext) error {
	// Find git repository root
	repoRoot, err := gitRepoRoot()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Generate random branch name
	name, err := worktree.GenerateName()
	if err != nil {
		return fmt.Errorf("generating branch name: %w", err)
	}

	// Determine base ref
	base := "origin/main"
	if ec.Profile.Worktree != nil {
		base = ec.Profile.Worktree.EffectiveBase()
	}

	// Fetch the base ref
	refParts := strings.SplitN(base, "/", 2)
	if len(refParts) == 2 {
		fmt.Fprintf(os.Stderr, "Fetching %s...\n", base)
		if err := gitFetch(repoRoot, refParts[0], refParts[1]); err != nil {
			return fmt.Errorf("fetching %s: %w", base, err)
		}
	}

	// Create worktrees directory
	worktreesDir := filepath.Join(repoRoot, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("creating worktrees directory: %w", err)
	}

	// Create worktree
	worktreePath := filepath.Join(worktreesDir, name)
	fmt.Fprintf(os.Stderr, "Creating worktree: worktrees/%s\n", name)
	if err := gitWorktreeAdd(repoRoot, name, worktreePath, base); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	// Update execution context
	ec.WorkDir = worktreePath
	ec.WorktreePath = worktreePath
	ec.WorktreeBranch = name

	return nil
}

func gitRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func gitFetch(repoRoot, remote, ref string) error {
	cmd := exec.Command("git", "-C", repoRoot, "fetch", remote, ref)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitWorktreeAdd(repoRoot, branchName, worktreePath, base string) error {
	cmd := exec.Command("git", "-C", repoRoot, "worktree", "add",
		"-b", branchName, worktreePath, base)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
