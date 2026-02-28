package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncSettings_CopiesFiles(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	// Create source files
	if err := os.WriteFile(filepath.Join(claudeHome, "settings.json"), []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatalf("writing settings.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeHome, "CLAUDE.md"), []byte("# Instructions"), 0644); err != nil {
		t.Fatalf("writing CLAUDE.md: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	// Verify files were copied
	content, err := os.ReadFile(filepath.Join(containerHome, "settings.json"))
	if err != nil {
		t.Fatalf("reading settings.json: %v", err)
	}
	if string(content) != `{"key":"value"}` {
		t.Errorf("settings.json = %q, want %q", string(content), `{"key":"value"}`)
	}

	content, err = os.ReadFile(filepath.Join(containerHome, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	if string(content) != "# Instructions" {
		t.Errorf("CLAUDE.md = %q, want %q", string(content), "# Instructions")
	}
}

func TestSyncSettings_SkipsMissingFiles(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	// Don't create any source files
	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	// Verify no files were created
	if _, err := os.Stat(filepath.Join(containerHome, "settings.json")); !os.IsNotExist(err) {
		t.Error("settings.json should not exist when source is missing")
	}
}

func TestSyncSettings_CopiesDirectories(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	// Create source directory with files
	hooksDir := filepath.Join(claudeHome, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("creating hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit.sh"), []byte("#!/bin/bash\necho hi"), 0755); err != nil {
		t.Fatalf("writing pre-commit.sh: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	// Verify directory and file were copied
	content, err := os.ReadFile(filepath.Join(containerHome, "hooks", "pre-commit.sh"))
	if err != nil {
		t.Fatalf("reading hooks/pre-commit.sh: %v", err)
	}
	if string(content) != "#!/bin/bash\necho hi" {
		t.Errorf("hooks/pre-commit.sh = %q, want %q", string(content), "#!/bin/bash\necho hi")
	}
}

func TestSyncSettings_ReplacesExistingDirectories(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	// Create old content in container
	oldHooksDir := filepath.Join(containerHome, "hooks")
	if err := os.MkdirAll(oldHooksDir, 0755); err != nil {
		t.Fatalf("creating old hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(oldHooksDir, "old-hook.sh"), []byte("old"), 0644); err != nil {
		t.Fatalf("writing old-hook.sh: %v", err)
	}

	// Create new content in source
	newHooksDir := filepath.Join(claudeHome, "hooks")
	if err := os.MkdirAll(newHooksDir, 0755); err != nil {
		t.Fatalf("creating new hooks dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newHooksDir, "new-hook.sh"), []byte("new"), 0644); err != nil {
		t.Fatalf("writing new-hook.sh: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	// Old file should be gone
	if _, err := os.Stat(filepath.Join(containerHome, "hooks", "old-hook.sh")); !os.IsNotExist(err) {
		t.Error("old-hook.sh should have been removed")
	}

	// New file should exist
	content, err := os.ReadFile(filepath.Join(containerHome, "hooks", "new-hook.sh"))
	if err != nil {
		t.Fatalf("reading new-hook.sh: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("new-hook.sh = %q, want %q", string(content), "new")
	}
}

func TestSyncSettings_SkipsMissingDirectories(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	// Verify no directories were created
	for _, d := range syncDirs {
		if _, err := os.Stat(filepath.Join(containerHome, d)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist when source is missing", d)
		}
	}
}

func TestSyncSettings_CreatesContainerHome(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := filepath.Join(t.TempDir(), "nonexistent", "agent-workspace")

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	info, err := os.Stat(containerHome)
	if err != nil {
		t.Fatalf("container home should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("container home should be a directory")
	}
}

func TestSyncSettings_NestedDirectories(t *testing.T) {
	claudeHome := t.TempDir()
	containerHome := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(claudeHome, "plugins", "subdir")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("creating nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "plugin.json"), []byte(`{}`), 0644); err != nil {
		t.Fatalf("writing plugin.json: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.SyncSettings(claudeHome, containerHome); err != nil {
		t.Fatalf("SyncSettings() error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(containerHome, "plugins", "subdir", "plugin.json"))
	if err != nil {
		t.Fatalf("reading nested file: %v", err)
	}
	if string(content) != `{}` {
		t.Errorf("plugin.json = %q, want %q", string(content), `{}`)
	}
}
