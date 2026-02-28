package worktree

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// layoutData holds template variables for the zellij layout.
type layoutData struct {
	ScriptsDir string
}

// Run executes the full worktree creation and zellij launch flow.
func Run() error {
	// 1. Check required dependencies
	if err := CheckRequiredDeps(); err != nil {
		return err
	}

	// 2. Warn about optional dependencies
	if warnings := CheckOptionalDeps(); len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Warning: some optional tools are missing:")
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, w)
		}
		fmt.Fprintln(os.Stderr)
	}

	// 3. Find git repository root
	repoRoot, err := gitRepoRoot()
	if err != nil {
		return fmt.Errorf("finding git repository root: %w", err)
	}

	// 4. Generate random branch name
	name, err := GenerateName()
	if err != nil {
		return fmt.Errorf("generating branch name: %w", err)
	}

	// 5. Fetch latest main
	fmt.Fprintln(os.Stderr, "Fetching origin/main...")
	if err := gitFetch(repoRoot); err != nil {
		return fmt.Errorf("fetching origin/main: %w", err)
	}

	// 6. Create worktrees directory
	worktreesDir := filepath.Join(repoRoot, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return fmt.Errorf("creating worktrees directory: %w", err)
	}

	// 7. Create worktree
	worktreePath := filepath.Join(worktreesDir, name)
	fmt.Fprintf(os.Stderr, "Creating worktree: worktrees/%s\n", name)
	if err := gitWorktreeAdd(repoRoot, name, worktreePath); err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}

	// 8. Prepare temp directory with scripts and layout
	tmpDir, cleanup, err := prepareFiles()
	if err != nil {
		return fmt.Errorf("preparing worktree files: %w", err)
	}
	defer cleanup()

	// 9. Launch zellij
	fmt.Fprintf(os.Stderr, "Launching zellij session: %s\n", name)
	return launchZellij(worktreePath, tmpDir, name)
}

// gitRepoRoot returns the top-level directory of the current git repository.
func gitRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// gitFetch runs git fetch origin main.
func gitFetch(repoRoot string) error {
	cmd := exec.Command("git", "-C", repoRoot, "fetch", "origin", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitWorktreeAdd creates a new git worktree with a new branch.
func gitWorktreeAdd(repoRoot, branchName, worktreePath string) error {
	cmd := exec.Command("git", "-C", repoRoot, "worktree", "add",
		"-b", branchName, worktreePath, "origin/main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// prepareFiles creates a temp directory with scripts and layout file.
// Returns the temp dir path and a cleanup function.
func prepareFiles() (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "claude-docker-worktree-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}
	cleanupFn := func() { _ = os.RemoveAll(tmpDir) }

	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("creating scripts dir: %w", err)
	}

	// Write shell scripts
	scripts := map[string][]byte{
		"plans-watcher.sh":   plansWatcherSh,
		"git-diff-picker.sh": gitDiffPickerSh,
		"pr-status.sh":       prStatusSh,
	}
	for name, content := range scripts {
		path := filepath.Join(scriptsDir, name)
		if err := os.WriteFile(path, content, 0755); err != nil {
			cleanupFn()
			return "", nil, fmt.Errorf("writing %s: %w", name, err)
		}
	}

	// Render and write layout template
	tmpl, err := template.New("layout").Parse(string(layoutKdlTmpl))
	if err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("parsing layout template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, layoutData{ScriptsDir: scriptsDir}); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("rendering layout template: %w", err)
	}

	layoutPath := filepath.Join(tmpDir, "layout.kdl")
	if err := os.WriteFile(layoutPath, buf.Bytes(), 0644); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("writing layout file: %w", err)
	}

	return tmpDir, cleanupFn, nil
}

// launchZellij starts a zellij session with the given layout.
func launchZellij(workDir, tmpDir, sessionName string) error {
	layoutPath := filepath.Join(tmpDir, "layout.kdl")
	cmd := exec.Command("zellij",
		"--new-session-with-layout", layoutPath,
		"-s", sessionName)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
