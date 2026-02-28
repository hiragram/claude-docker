package launcher

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

// layoutData holds template variables for the zellij layout.
type layoutData struct {
	ScriptsDir    string
	ClaudeCommand string
}

// ZellijLauncher launches a zellij session with multiple panes.
type ZellijLauncher struct{}

func (l *ZellijLauncher) Launch(_ context.Context, ec *pipeline.ExecutionContext) error {
	if _, err := exec.LookPath("zellij"); err != nil {
		return fmt.Errorf("zellij is not installed (brew install zellij)")
	}

	// Prepare temp directory with scripts and layout
	tmpDir, cleanup, err := l.prepareFiles(ec)
	if err != nil {
		return fmt.Errorf("preparing zellij files: %w", err)
	}
	defer cleanup()

	// Launch zellij
	sessionName := ec.WorktreeBranch
	if sessionName == "" {
		sessionName = ec.ProfileName
	}

	fmt.Fprintf(os.Stderr, "Launching zellij session: %s\n", sessionName)
	return l.launchZellij(ec.WorkDir, tmpDir, sessionName)
}

func (l *ZellijLauncher) prepareFiles(ec *pipeline.ExecutionContext) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "aw-zellij-*")
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

	// Build Claude command based on environment
	claudeCmd := l.buildClaudeCommand(ec)

	// Render and write layout template
	tmpl, err := template.New("layout").Parse(string(layoutKdlTmpl))
	if err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("parsing layout template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, layoutData{
		ScriptsDir:    scriptsDir,
		ClaudeCommand: claudeCmd,
	}); err != nil {
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

func (l *ZellijLauncher) buildClaudeCommand(ec *pipeline.ExecutionContext) string {
	switch ec.Profile.Environment {
	case profile.EnvironmentDocker:
		// Run aw with the docker-claude profile equivalent
		// This will cause it to build docker image and run claude inside
		self, err := os.Executable()
		if err != nil {
			self = "aw"
		}
		return fmt.Sprintf("%s docker-claude", self)
	default:
		// Host mode: just run claude directly
		return "claude"
	}
}

func (l *ZellijLauncher) launchZellij(workDir, tmpDir, sessionName string) error {
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
