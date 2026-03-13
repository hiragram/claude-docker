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

func TestEnvStage_WritesProfileEnvFile(t *testing.T) {
	dir := t.TempDir()
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Env: map[string]string{
				"STATIC_KEY": "static_value",
			},
		},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// .aw-profile-env should have been written
	profileEnvFile := filepath.Join(dir, ".aw-profile-env")
	data, err := os.ReadFile(profileEnvFile)
	if err != nil {
		t.Fatalf("expected .aw-profile-env to be created: %v", err)
	}
	if string(data) != "STATIC_KEY=static_value\n" {
		t.Errorf("file content = %q, want %q", string(data), "STATIC_KEY=static_value\n")
	}
}

func TestEnvStage_ReadsProfileEnvFile(t *testing.T) {
	// Simulates child process: profile.Env is empty, but .aw-profile-env exists from parent
	dir := t.TempDir()
	profileEnvFile := filepath.Join(dir, ".aw-profile-env")
	if err := os.WriteFile(profileEnvFile, []byte("PARENT_KEY=parent_value\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{}, // empty, like builtin "claude" profile
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["PARENT_KEY"] != "parent_value" {
		t.Errorf("PARENT_KEY = %q, want %q", ec.EnvVars["PARENT_KEY"], "parent_value")
	}
}

func TestEnvStage_ProfileEnvOverridesProfileEnvFile(t *testing.T) {
	dir := t.TempDir()
	profileEnvFile := filepath.Join(dir, ".aw-profile-env")
	if err := os.WriteFile(profileEnvFile, []byte("SHARED=from-parent-file\nPARENT_ONLY=yes\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Env: map[string]string{
				"SHARED":       "from-current-profile",
				"CURRENT_ONLY": "yes",
			},
		},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["SHARED"] != "from-current-profile" {
		t.Errorf("SHARED = %q, want %q (current profile should override profile env file)", ec.EnvVars["SHARED"], "from-current-profile")
	}
	if ec.EnvVars["PARENT_ONLY"] != "yes" {
		t.Errorf("PARENT_ONLY = %q, want %q", ec.EnvVars["PARENT_ONLY"], "yes")
	}
	if ec.EnvVars["CURRENT_ONLY"] != "yes" {
		t.Errorf("CURRENT_ONLY = %q, want %q", ec.EnvVars["CURRENT_ONLY"], "yes")
	}
}

func TestEnvStage_AwEnvOverridesAll(t *testing.T) {
	dir := t.TempDir()

	// .aw-profile-env (lowest priority)
	profileEnvFile := filepath.Join(dir, ".aw-profile-env")
	if err := os.WriteFile(profileEnvFile, []byte("KEY=from-profile-env-file\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// .aw-env (highest priority)
	awEnvFile := filepath.Join(dir, ".aw-env")
	if err := os.WriteFile(awEnvFile, []byte("KEY=from-aw-env\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Env: map[string]string{
				"KEY": "from-current-profile",
			},
		},
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ec.EnvVars["KEY"] != "from-aw-env" {
		t.Errorf("KEY = %q, want %q (.aw-env should override everything)", ec.EnvVars["KEY"], "from-aw-env")
	}
}

func TestEnvStage_NoWriteWhenProfileEnvEmpty(t *testing.T) {
	dir := t.TempDir()
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{}, // no Env
		WorkDir: dir,
	}

	s := &EnvStage{}
	if err := s.Run(context.Background(), ec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profileEnvFile := filepath.Join(dir, ".aw-profile-env")
	if _, err := os.Stat(profileEnvFile); !os.IsNotExist(err) {
		t.Error(".aw-profile-env should not be created when profile.Env is empty")
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
