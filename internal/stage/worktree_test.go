package stage

import (
	"os"
	"os/exec"
	"testing"

	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

func TestRunOnCreateHook_ShellInvocation(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return exec.Command("true")
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnCreate: "./setup.sh"},
			Environment: profile.EnvironmentDocker,
		},
		ProfileName:    "test-profile",
		WorktreePath:   t.TempDir(),
		WorktreeBranch: "test-branch",
	}

	err := runOnCreateHook(ec, "/fake/repo")
	if err != nil {
		t.Fatalf("runOnCreateHook() error: %v", err)
	}

	if capturedName != "sh" {
		t.Errorf("expected command 'sh', got %q", capturedName)
	}
	if len(capturedArgs) != 2 || capturedArgs[0] != "-c" || capturedArgs[1] != "./setup.sh" {
		t.Errorf("expected args [-c ./setup.sh], got %v", capturedArgs)
	}
}

func TestRunOnCreateHook_SetsEnvironmentAndDir(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()

	// Use a real command that prints env vars
	execCommand = func(name string, args ...string) *exec.Cmd {
		// Create a real command that will succeed and let us inspect env via the Cmd struct
		cmd := exec.Command("true")
		return cmd
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnCreate: "echo test"},
			Environment: profile.EnvironmentDocker,
		},
		ProfileName:    "my-profile",
		WorktreePath:   tmpDir,
		WorktreeBranch: "feature-branch",
	}

	// Instead of mocking, we test with a real shell command that verifies env vars
	execCommand = exec.Command
	// Use a command that checks env vars exist
	ec.Profile.Worktree.OnCreate = "test -n \"$AW_WORKTREE_PATH\" && test -n \"$AW_WORKTREE_BRANCH\" && test -n \"$AW_REPO_ROOT\" && test -n \"$AW_PROFILE_NAME\" && test -n \"$AW_ENVIRONMENT\""

	err := runOnCreateHook(ec, "/fake/repo")
	if err != nil {
		t.Fatalf("runOnCreateHook() error (env vars missing): %v", err)
	}
}

func TestRunOnCreateHook_EnvVarValues(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()
	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnCreate: "test \"$AW_PROFILE_NAME\" = \"special-profile\" && test \"$AW_ENVIRONMENT\" = \"host\" && test \"$AW_WORKTREE_BRANCH\" = \"my-branch\""},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "special-profile",
		WorktreePath:   tmpDir,
		WorktreeBranch: "my-branch",
	}

	err := runOnCreateHook(ec, "/some/repo")
	if err != nil {
		t.Fatalf("runOnCreateHook() env var values mismatch: %v", err)
	}
}

func TestRunOnCreateHook_WorkingDirectory(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()
	execCommand = exec.Command

	// pwd should match the worktree path
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnCreate: "test \"$(pwd)\" = \"" + tmpDir + "\""},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "test",
		WorktreePath:   tmpDir,
		WorktreeBranch: "branch",
	}

	err := runOnCreateHook(ec, "/repo")
	if err != nil {
		t.Fatalf("runOnCreateHook() working directory mismatch: %v", err)
	}
}

func TestRunOnCreateHook_FailureReturnsError(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnCreate: "exit 1"},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "test",
		WorktreePath:   t.TempDir(),
		WorktreeBranch: "branch",
	}

	err := runOnCreateHook(ec, "/repo")
	if err == nil {
		t.Fatal("expected error from failing hook, got nil")
	}
}

func TestRunOnEndHook_ShellInvocation(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return exec.Command("true")
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnEnd: "./cleanup.sh"},
			Environment: profile.EnvironmentDocker,
		},
		ProfileName:    "test-profile",
		WorktreePath:   t.TempDir(),
		WorktreeBranch: "test-branch",
		RepoRoot:       "/fake/repo",
	}

	err := RunOnEndHook(ec)
	if err != nil {
		t.Fatalf("RunOnEndHook() error: %v", err)
	}

	if capturedName != "sh" {
		t.Errorf("expected command 'sh', got %q", capturedName)
	}
	if len(capturedArgs) != 2 || capturedArgs[0] != "-c" || capturedArgs[1] != "./cleanup.sh" {
		t.Errorf("expected args [-c ./cleanup.sh], got %v", capturedArgs)
	}
}

func TestRunOnEndHook_SetsEnvironmentAndDir(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()
	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnEnd: "test -n \"$AW_WORKTREE_PATH\" && test -n \"$AW_WORKTREE_BRANCH\" && test -n \"$AW_REPO_ROOT\" && test -n \"$AW_PROFILE_NAME\" && test -n \"$AW_ENVIRONMENT\""},
			Environment: profile.EnvironmentDocker,
		},
		ProfileName:    "my-profile",
		WorktreePath:   tmpDir,
		WorktreeBranch: "feature-branch",
		RepoRoot:       "/fake/repo",
	}

	err := RunOnEndHook(ec)
	if err != nil {
		t.Fatalf("RunOnEndHook() error (env vars missing): %v", err)
	}
}

func TestRunOnEndHook_EnvVarValues(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()
	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnEnd: "test \"$AW_PROFILE_NAME\" = \"special-profile\" && test \"$AW_ENVIRONMENT\" = \"host\" && test \"$AW_WORKTREE_BRANCH\" = \"my-branch\""},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "special-profile",
		WorktreePath:   tmpDir,
		WorktreeBranch: "my-branch",
		RepoRoot:       "/some/repo",
	}

	err := RunOnEndHook(ec)
	if err != nil {
		t.Fatalf("RunOnEndHook() env var values mismatch: %v", err)
	}
}

func TestRunOnEndHook_WorkingDirectory(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	tmpDir := t.TempDir()
	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnEnd: "test \"$(pwd)\" = \"" + tmpDir + "\""},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "test",
		WorktreePath:   tmpDir,
		WorktreeBranch: "branch",
		RepoRoot:       "/repo",
	}

	err := RunOnEndHook(ec)
	if err != nil {
		t.Fatalf("RunOnEndHook() working directory mismatch: %v", err)
	}
}

func TestRunOnEndHook_FailureReturnsError(t *testing.T) {
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	execCommand = exec.Command

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree:    &profile.WorktreeConfig{OnEnd: "exit 1"},
			Environment: profile.EnvironmentHost,
		},
		ProfileName:    "test",
		WorktreePath:   t.TempDir(),
		WorktreeBranch: "branch",
		RepoRoot:       "/repo",
	}

	err := RunOnEndHook(ec)
	if err == nil {
		t.Fatal("expected error from failing hook, got nil")
	}
}

func TestResolveWorktreesDir_Default(t *testing.T) {
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{},
	}
	got, err := resolveWorktreesDir(ec, "/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/repo/worktrees" {
		t.Errorf("got %q, want %q", got, "/repo/worktrees")
	}
}

func TestResolveWorktreesDir_AbsoluteOverride(t *testing.T) {
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree: &profile.WorktreeConfig{Dir: "/abs/wt"},
		},
	}
	got, err := resolveWorktreesDir(ec, "/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/abs/wt" {
		t.Errorf("got %q, want %q", got, "/abs/wt")
	}
}

func TestResolveWorktreesDir_RelativeIsRepoRelative(t *testing.T) {
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree: &profile.WorktreeConfig{Dir: "../shared-worktrees"},
		},
	}
	got, err := resolveWorktreesDir(ec, "/repo/sub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/repo/shared-worktrees" {
		t.Errorf("got %q, want %q", got, "/repo/shared-worktrees")
	}
}

func TestResolveWorktreesDir_TildeExpansion(t *testing.T) {
	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{
			Worktree: &profile.WorktreeConfig{Dir: "~/aw-wt"},
		},
	}
	got, err := resolveWorktreesDir(ec, "/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	home, _ := os.UserHomeDir()
	want := home + "/aw-wt"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
