package worktree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareFiles(t *testing.T) {
	tmpDir, cleanup, err := prepareFiles()
	if err != nil {
		t.Fatalf("prepareFiles() error: %v", err)
	}
	defer cleanup()

	// Scripts directory should exist
	scriptsDir := filepath.Join(tmpDir, "scripts")
	info, err := os.Stat(scriptsDir)
	if err != nil || !info.IsDir() {
		t.Fatalf("scripts directory does not exist: %v", err)
	}

	// All 3 scripts should exist and be executable
	for _, name := range []string{"plans-watcher.sh", "git-diff-picker.sh", "pr-status.sh"} {
		path := filepath.Join(scriptsDir, name)
		fi, err := os.Stat(path)
		if err != nil {
			t.Errorf("script %s does not exist: %v", name, err)
			continue
		}
		if fi.Mode().Perm()&0111 == 0 {
			t.Errorf("script %s should be executable", name)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("reading script %s: %v", name, err)
			continue
		}
		if !strings.HasPrefix(string(content), "#!/bin/bash") {
			t.Errorf("script %s should start with shebang", name)
		}
	}

	// Layout file should exist and contain rendered template
	layoutPath := filepath.Join(tmpDir, "layout.kdl")
	content, err := os.ReadFile(layoutPath)
	if err != nil {
		t.Fatalf("layout.kdl does not exist: %v", err)
	}
	layoutStr := string(content)
	if !strings.Contains(layoutStr, scriptsDir) {
		t.Error("layout.kdl should contain the scripts directory path")
	}
	if !strings.Contains(layoutStr, "claude-docker") {
		t.Error("layout.kdl should contain 'claude-docker' command")
	}
	if !strings.Contains(layoutStr, "Plans") {
		t.Error("layout.kdl should contain Plans pane")
	}
}

func TestPrepareFiles_Cleanup(t *testing.T) {
	tmpDir, cleanup, err := prepareFiles()
	if err != nil {
		t.Fatalf("prepareFiles() error: %v", err)
	}

	// Directory should exist before cleanup
	if _, err := os.Stat(tmpDir); err != nil {
		t.Fatalf("dir should exist before cleanup: %v", err)
	}

	cleanup()

	// Directory should not exist after cleanup
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("dir should not exist after cleanup")
	}
}

func TestEmbeddedFilesNotEmpty(t *testing.T) {
	if len(plansWatcherSh) == 0 {
		t.Error("embedded plans-watcher.sh is empty")
	}
	if len(gitDiffPickerSh) == 0 {
		t.Error("embedded git-diff-picker.sh is empty")
	}
	if len(prStatusSh) == 0 {
		t.Error("embedded pr-status.sh is empty")
	}
	if len(layoutKdlTmpl) == 0 {
		t.Error("embedded layout.kdl.tmpl is empty")
	}
}
