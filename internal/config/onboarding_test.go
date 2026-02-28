package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureOnboardingState_CreatesWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agent-workspace.json")

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != "{}\n" {
		t.Errorf("content = %q, want %q", string(content), "{}\n")
	}
}

func TestEnsureOnboardingState_CreatesWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agent-workspace.json")

	// Create empty file
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatalf("writing empty file: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != "{}\n" {
		t.Errorf("content = %q, want %q", string(content), "{}\n")
	}
}

func TestEnsureOnboardingState_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".agent-workspace.json")

	existing := `{"hasCompletedOnboarding":true}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatalf("writing existing file: %v", err)
	}

	syncer := NewSyncer()
	if err := syncer.EnsureOnboardingState(path); err != nil {
		t.Fatalf("EnsureOnboardingState() error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != existing {
		t.Errorf("content = %q, want %q (should be preserved)", string(content), existing)
	}
}
