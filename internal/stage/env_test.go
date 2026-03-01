package stage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

func TestEnvStage_Name(t *testing.T) {
	s := &EnvStage{}
	if s.Name() != "env" {
		t.Errorf("Name() = %q, want %q", s.Name(), "env")
	}
}

func TestEnvStage_NoEnvNoFile(t *testing.T) {
	dir := t.TempDir()
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars == nil {
		t.Fatal("EnvVars should not be nil")
	}
	if len(ec.EnvVars) != 0 {
		t.Errorf("got %d env vars, want 0", len(ec.EnvVars))
	}
}

func TestEnvStage_ProfileEnvOnly(t *testing.T) {
	dir := t.TempDir()
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Env: map[string]string{
				"FOO": "bar",
				"BAZ": "qux",
			},
		},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", ec.EnvVars["FOO"], "bar")
	}
	if ec.EnvVars["BAZ"] != "qux" {
		t.Errorf("BAZ = %q, want %q", ec.EnvVars["BAZ"], "qux")
	}
}

func TestEnvStage_FileOnly(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".aw-env")
	if err := os.WriteFile(envFile, []byte("KEY=value\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["KEY"] != "value" {
		t.Errorf("KEY = %q, want %q", ec.EnvVars["KEY"], "value")
	}
}

func TestEnvStage_FileOverridesProfile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".aw-env")
	if err := os.WriteFile(envFile, []byte("FOO=from-file\nFILE_ONLY=yes\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Env: map[string]string{
				"FOO":          "from-profile",
				"PROFILE_ONLY": "yes",
			},
		},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["FOO"] != "from-file" {
		t.Errorf("FOO = %q, want %q (file should override profile)", ec.EnvVars["FOO"], "from-file")
	}
	if ec.EnvVars["PROFILE_ONLY"] != "yes" {
		t.Errorf("PROFILE_ONLY = %q, want %q", ec.EnvVars["PROFILE_ONLY"], "yes")
	}
	if ec.EnvVars["FILE_ONLY"] != "yes" {
		t.Errorf("FILE_ONLY = %q, want %q", ec.EnvVars["FILE_ONLY"], "yes")
	}
}

func TestEnvStage_InvalidFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".aw-env")
	if err := os.WriteFile(envFile, []byte("INVALID_LINE_WITHOUT_EQUALS\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{},
		WorkDir: dir,
	}

	s := &EnvStage{}
	err := s.Run(context.Background(), ec)
	if err == nil {
		t.Fatal("expected error for invalid .aw-env file")
	}
}
