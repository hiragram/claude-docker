package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/hiragram/claude-docker/internal/docker"
	"github.com/hiragram/claude-docker/internal/mount"
	"github.com/hiragram/claude-docker/internal/version"
)

// --- Mock Docker Client ---

type mockDockerClient struct {
	checkErr      error
	buildErr      error
	volumeErr     error
	runErr        error
	buildCalls    []buildCall
	volumeCalls   []string
	runCalls      []docker.RunConfig
	checkCalled   bool
}

type buildCall struct {
	imageName  string
	contextDir string
}

func (m *mockDockerClient) CheckAvailable() error {
	m.checkCalled = true
	return m.checkErr
}

func (m *mockDockerClient) Build(_ context.Context, imageName, contextDir string) error {
	m.buildCalls = append(m.buildCalls, buildCall{imageName, contextDir})
	return m.buildErr
}

func (m *mockDockerClient) VolumeCreate(_ context.Context, volumeName string) error {
	m.volumeCalls = append(m.volumeCalls, volumeName)
	return m.volumeErr
}

func (m *mockDockerClient) Run(_ context.Context, config docker.RunConfig) error {
	m.runCalls = append(m.runCalls, config)
	return m.runErr
}

// --- Mock Config Syncer ---

type mockConfigSyncer struct {
	syncErr       error
	onboardingErr error
	syncCalls     []syncCall
	onboardCalls  []string
}

type syncCall struct {
	claudeHome    string
	containerHome string
}

func (m *mockConfigSyncer) SyncSettings(claudeHome, containerHome string) error {
	m.syncCalls = append(m.syncCalls, syncCall{claudeHome, containerHome})
	return m.syncErr
}

func (m *mockConfigSyncer) EnsureOnboardingState(path string) error {
	m.onboardCalls = append(m.onboardCalls, path)
	return m.onboardingErr
}

// --- Mock Mount Builder ---

type mockMountBuilder struct {
	mounts   []docker.Mount
	mountErr error
	calls    []mount.MountOptions
}

func (m *mockMountBuilder) BuildMounts(opts mount.MountOptions) ([]docker.Mount, error) {
	m.calls = append(m.calls, opts)
	return m.mounts, m.mountErr
}

// --- Tests ---

func TestExecute_FullFlow(t *testing.T) {
	dc := &mockDockerClient{}
	cs := &mockConfigSyncer{}
	mb := &mockMountBuilder{
		mounts: []docker.Mount{
			{Source: "vol", Target: "/data", IsVolume: true},
		},
	}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/home/testuser/project",
	}

	err := runner.Execute(context.Background(), []string{"-p", "hello"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Docker check was called
	if !dc.checkCalled {
		t.Error("Docker CheckAvailable was not called")
	}

	// Build was called with correct image name
	if len(dc.buildCalls) != 1 {
		t.Fatalf("expected 1 build call, got %d", len(dc.buildCalls))
	}
	if dc.buildCalls[0].imageName != imageName {
		t.Errorf("build image = %q, want %q", dc.buildCalls[0].imageName, imageName)
	}

	// Volume was created
	if len(dc.volumeCalls) != 1 || dc.volumeCalls[0] != volumeName {
		t.Errorf("volume calls = %v, want [%q]", dc.volumeCalls, volumeName)
	}

	// Settings were synced
	if len(cs.syncCalls) != 1 {
		t.Fatalf("expected 1 sync call, got %d", len(cs.syncCalls))
	}

	// Onboarding state was ensured
	if len(cs.onboardCalls) != 1 {
		t.Fatalf("expected 1 onboarding call, got %d", len(cs.onboardCalls))
	}

	// Mount builder was called
	if len(mb.calls) != 1 {
		t.Fatalf("expected 1 mount call, got %d", len(mb.calls))
	}

	// Container was run
	if len(dc.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(dc.runCalls))
	}

	rc := dc.runCalls[0]
	if rc.ImageName != imageName {
		t.Errorf("run image = %q, want %q", rc.ImageName, imageName)
	}
	if rc.WorkDir != "/home/testuser/project" {
		t.Errorf("run workdir = %q, want %q", rc.WorkDir, "/home/testuser/project")
	}

	// Command should contain user args + --allow-dangerously-skip-permissions
	expectedCmd := []string{"claude", "-p", "hello", "--allow-dangerously-skip-permissions"}
	if len(rc.Command) != len(expectedCmd) {
		t.Fatalf("command = %v, want %v", rc.Command, expectedCmd)
	}
	for i, arg := range expectedCmd {
		if rc.Command[i] != arg {
			t.Errorf("command[%d] = %q, want %q", i, rc.Command[i], arg)
		}
	}
}

func TestExecute_DockerCheckFails(t *testing.T) {
	dc := &mockDockerClient{checkErr: errors.New("docker not found")}
	cs := &mockConfigSyncer{}
	mb := &mockMountBuilder{}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/workspace",
	}

	err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when docker check fails")
	}
	if !errors.Is(err, dc.checkErr) {
		t.Errorf("error = %v, want wrapped %v", err, dc.checkErr)
	}

	// Nothing else should have been called
	if len(dc.buildCalls) > 0 {
		t.Error("build should not be called when check fails")
	}
}

func TestExecute_BuildFails(t *testing.T) {
	dc := &mockDockerClient{buildErr: errors.New("build failed")}
	cs := &mockConfigSyncer{}
	mb := &mockMountBuilder{}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/workspace",
	}

	err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when build fails")
	}

	// Volume should not be created if build fails
	if len(dc.volumeCalls) > 0 {
		t.Error("volume should not be created when build fails")
	}
}

func TestExecute_SyncFails(t *testing.T) {
	dc := &mockDockerClient{}
	cs := &mockConfigSyncer{syncErr: errors.New("sync failed")}
	mb := &mockMountBuilder{}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/workspace",
	}

	err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when sync fails")
	}

	// Container should not be run if sync fails
	if len(dc.runCalls) > 0 {
		t.Error("container should not be run when sync fails")
	}
}

func TestExecute_MountBuildFails(t *testing.T) {
	dc := &mockDockerClient{}
	cs := &mockConfigSyncer{}
	mb := &mockMountBuilder{mountErr: errors.New("mount error")}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/workspace",
	}

	err := runner.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when mount build fails")
	}

	if len(dc.runCalls) > 0 {
		t.Error("container should not be run when mount build fails")
	}
}

func TestExecute_EnvVars(t *testing.T) {
	dc := &mockDockerClient{}
	cs := &mockConfigSyncer{}
	mb := &mockMountBuilder{}

	runner := &Runner{
		DockerClient: dc,
		ConfigSyncer: cs,
		MountBuilder: mb,
		HomeDir:      "/home/testuser",
		WorkDir:      "/home/testuser/project",
	}

	if err := runner.Execute(context.Background(), nil); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if len(dc.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(dc.runCalls))
	}

	rc := dc.runCalls[0]
	if rc.EnvVars["HOST_WORKSPACE"] != "/home/testuser/project" {
		t.Errorf("HOST_WORKSPACE = %q, want %q", rc.EnvVars["HOST_WORKSPACE"], "/home/testuser/project")
	}
}

func TestHasVersionFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"--version flag", []string{"--version"}, true},
		{"-v flag", []string{"-v"}, true},
		{"--version among other args", []string{"-p", "hello", "--version"}, true},
		{"-v among other args", []string{"-v", "-p", "hello"}, true},
		{"no version flag", []string{"-p", "hello"}, false},
		{"empty args", nil, false},
		{"similar but not version", []string{"--verbose"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasVersionFlag(tt.args)
			if got != tt.want {
				t.Errorf("hasVersionFlag(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestHasWorktreeFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"--worktree flag", []string{"--worktree"}, true},
		{"--worktree among other args", []string{"-p", "hello", "--worktree"}, true},
		{"no worktree flag", []string{"-p", "hello"}, false},
		{"empty args", nil, false},
		{"similar but not worktree", []string{"--worktrees"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasWorktreeFlag(tt.args)
			if got != tt.want {
				t.Errorf("hasWorktreeFlag(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestVersionOutput(t *testing.T) {
	// Verify version constant is accessible and non-empty
	if version.Version == "" {
		t.Error("version.Version should not be empty")
	}
}
