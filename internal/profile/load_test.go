package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	yaml := `
default: docker-claude

profiles:
  docker-claude:
    environment: docker
    launch: claude

  worktree-shell:
    worktree:
      base: origin/main
    environment: host
    launch: shell

  worktree-zellij:
    worktree: {}
    environment: docker
    launch: zellij
    zellij:
      layout: default
`

	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if cfg.Default != "docker-claude" {
		t.Errorf("Default = %q, want %q", cfg.Default, "docker-claude")
	}

	if len(cfg.Profiles) != 3 {
		t.Fatalf("got %d profiles, want 3", len(cfg.Profiles))
	}

	// Check docker-claude profile
	dc := cfg.Profiles["docker-claude"]
	if dc.Environment != EnvironmentDocker {
		t.Errorf("docker-claude.Environment = %q, want %q", dc.Environment, EnvironmentDocker)
	}
	if dc.Launch != LaunchClaude {
		t.Errorf("docker-claude.Launch = %q, want %q", dc.Launch, LaunchClaude)
	}
	if dc.Worktree != nil {
		t.Errorf("docker-claude.Worktree should be nil")
	}

	// Check worktree-shell profile
	ws := cfg.Profiles["worktree-shell"]
	if ws.Worktree == nil {
		t.Fatal("worktree-shell.Worktree should not be nil")
	}
	if ws.Worktree.Base != "origin/main" {
		t.Errorf("worktree-shell.Worktree.Base = %q, want %q", ws.Worktree.Base, "origin/main")
	}
	if ws.Environment != EnvironmentHost {
		t.Errorf("worktree-shell.Environment = %q, want %q", ws.Environment, EnvironmentHost)
	}
	if ws.Launch != LaunchShell {
		t.Errorf("worktree-shell.Launch = %q, want %q", ws.Launch, LaunchShell)
	}

	// Check worktree-zellij profile
	wz := cfg.Profiles["worktree-zellij"]
	if wz.Worktree == nil {
		t.Fatal("worktree-zellij.Worktree should not be nil")
	}
	if wz.Launch != LaunchZellij {
		t.Errorf("worktree-zellij.Launch = %q, want %q", wz.Launch, LaunchZellij)
	}
	if wz.Zellij == nil {
		t.Fatal("worktree-zellij.Zellij should not be nil")
	}
	if wz.Zellij.Layout != "default" {
		t.Errorf("worktree-zellij.Zellij.Layout = %q, want %q", wz.Zellij.Layout, "default")
	}
}

func TestParse_EmptyProfiles(t *testing.T) {
	yaml := `
default: ""
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if cfg.Profiles == nil {
		t.Fatal("Profiles should not be nil (should be empty map)")
	}
	if len(cfg.Profiles) != 0 {
		t.Errorf("got %d profiles, want 0", len(cfg.Profiles))
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	_, err := Parse([]byte("}{invalid"))
	if err == nil {
		t.Fatal("Parse() should return error for invalid YAML")
	}
}

func TestParse_WorktreeEmptyObject(t *testing.T) {
	yaml := `
profiles:
  test:
    worktree: {}
    environment: host
    launch: claude
`
	cfg, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	p := cfg.Profiles["test"]
	if p.Worktree == nil {
		t.Fatal("Worktree should not be nil for empty object")
	}
	if p.Worktree.EffectiveBase() != "origin/main" {
		t.Errorf("EffectiveBase() = %q, want %q", p.Worktree.EffectiveBase(), "origin/main")
	}
}

func TestLoadFile_NotFound(t *testing.T) {
	cfg, err := LoadFile("/nonexistent/path/.agent-workspace.yml")
	if err != nil {
		t.Fatalf("LoadFile() should not error for missing file, got: %v", err)
	}

	// Should return builtin default
	if cfg.Default != "docker-claude" {
		t.Errorf("Default = %q, want %q", cfg.Default, "docker-claude")
	}
	if _, ok := cfg.Profiles["docker-claude"]; !ok {
		t.Error("expected docker-claude profile in builtin default")
	}
}

func TestLoadFile_ValidFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".agent-workspace.yml")

	content := `
default: my-profile
profiles:
  my-profile:
    environment: host
    launch: shell
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(configPath)
	if err != nil {
		t.Fatalf("LoadFile() error: %v", err)
	}

	if cfg.Default != "my-profile" {
		t.Errorf("Default = %q, want %q", cfg.Default, "my-profile")
	}
}

func TestLoad_NoGitRepo(t *testing.T) {
	// Override findGitRoot to simulate not being in a git repo
	orig := findGitRoot
	findGitRoot = func() (string, error) {
		return "", fmt.Errorf("not in a git repository")
	}
	defer func() { findGitRoot = orig }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error when not in git repo, got: %v", err)
	}

	if cfg.Default != "docker-claude" {
		t.Errorf("Default = %q, want %q", cfg.Default, "docker-claude")
	}
}
