package stage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hiragram/agent-workspace/internal/docker"
	"github.com/hiragram/agent-workspace/internal/mount"
	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

type mockDockerClient struct {
	available    bool
	buildCalled  bool
	volumeCalled bool
	runCalled    bool
	runConfig    docker.RunConfig
}

func (m *mockDockerClient) CheckAvailable() error {
	if !m.available {
		return fmt.Errorf("docker not available")
	}
	return nil
}

func (m *mockDockerClient) Build(_ context.Context, _, _ string) error {
	m.buildCalled = true
	return nil
}

func (m *mockDockerClient) VolumeCreate(_ context.Context, _ string) error {
	m.volumeCalled = true
	return nil
}

func (m *mockDockerClient) Run(_ context.Context, config docker.RunConfig) error {
	m.runCalled = true
	m.runConfig = config
	return nil
}

type mockConfigSyncer struct {
	syncCalled      bool
	onboardCalled   bool
	syncErr         error
	onboardErr      error
}

func (m *mockConfigSyncer) SyncSettings(_, _ string) error {
	m.syncCalled = true
	return m.syncErr
}

func (m *mockConfigSyncer) EnsureOnboardingState(_ string) error {
	m.onboardCalled = true
	return m.onboardErr
}

type mockMountBuilder struct {
	mounts []docker.Mount
	err    error
}

func (m *mockMountBuilder) BuildMounts(_ mount.MountOptions) ([]docker.Mount, error) {
	return m.mounts, m.err
}

func TestDockerStage_Name(t *testing.T) {
	s := &DockerStage{}
	if s.Name() != "docker" {
		t.Errorf("Name() = %q, want %q", s.Name(), "docker")
	}
}

func TestDockerStage_DockerNotAvailable(t *testing.T) {
	s := &DockerStage{
		DockerClient: &mockDockerClient{available: false},
		ConfigSyncer: &mockConfigSyncer{},
		MountBuilder: &mockMountBuilder{},
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{Environment: profile.EnvironmentDocker},
		HomeDir: "/home/test",
		WorkDir: "/workspace",
	}

	err := s.Run(context.Background(), ec)
	if err == nil {
		t.Fatal("expected error when docker not available")
	}
	if !strings.Contains(err.Error(), "docker is not available") {
		t.Errorf("error = %q, want containing 'docker is not available'", err.Error())
	}
}

func TestDockerStage_NewDockerStage(t *testing.T) {
	s := NewDockerStage()
	if s.DockerClient == nil {
		t.Error("DockerClient should not be nil")
	}
	if s.ConfigSyncer == nil {
		t.Error("ConfigSyncer should not be nil")
	}
	if s.MountBuilder == nil {
		t.Error("MountBuilder should not be nil")
	}
}
