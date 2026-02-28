package cmd

import (
	"testing"

	"github.com/hiragram/agent-workspace/internal/profile"
)

func TestBuildStages_DockerClaude(t *testing.T) {
	p := profile.Profile{
		Environment: profile.EnvironmentDocker,
		Launch:      profile.LaunchClaude,
	}
	stages := buildStages(p)

	// Should have DockerStage + LaunchStage = 2 stages
	if len(stages) != 2 {
		t.Fatalf("got %d stages, want 2", len(stages))
	}
	if stages[0].Name() != "docker" {
		t.Errorf("stage[0] = %q, want 'docker'", stages[0].Name())
	}
	if stages[1].Name() != "launch" {
		t.Errorf("stage[1] = %q, want 'launch'", stages[1].Name())
	}
}

func TestBuildStages_WorktreeHostShell(t *testing.T) {
	p := profile.Profile{
		Worktree:    &profile.WorktreeConfig{},
		Environment: profile.EnvironmentHost,
		Launch:      profile.LaunchShell,
	}
	stages := buildStages(p)

	// Should have WorktreeStage + LaunchStage = 2 stages
	if len(stages) != 2 {
		t.Fatalf("got %d stages, want 2", len(stages))
	}
	if stages[0].Name() != "worktree" {
		t.Errorf("stage[0] = %q, want 'worktree'", stages[0].Name())
	}
	if stages[1].Name() != "launch" {
		t.Errorf("stage[1] = %q, want 'launch'", stages[1].Name())
	}
}

func TestBuildStages_WorktreeDockerZellij(t *testing.T) {
	p := profile.Profile{
		Worktree:    &profile.WorktreeConfig{},
		Environment: profile.EnvironmentDocker,
		Launch:      profile.LaunchZellij,
	}
	stages := buildStages(p)

	// Should have WorktreeStage + DockerStage + LaunchStage = 3 stages
	if len(stages) != 3 {
		t.Fatalf("got %d stages, want 3", len(stages))
	}
	if stages[0].Name() != "worktree" {
		t.Errorf("stage[0] = %q, want 'worktree'", stages[0].Name())
	}
	if stages[1].Name() != "docker" {
		t.Errorf("stage[1] = %q, want 'docker'", stages[1].Name())
	}
	if stages[2].Name() != "launch" {
		t.Errorf("stage[2] = %q, want 'launch'", stages[2].Name())
	}
}

func TestBuildStages_HostClaude(t *testing.T) {
	p := profile.Profile{
		Environment: profile.EnvironmentHost,
		Launch:      profile.LaunchClaude,
	}
	stages := buildStages(p)

	// Should have LaunchStage only = 1 stage
	if len(stages) != 1 {
		t.Fatalf("got %d stages, want 1", len(stages))
	}
	if stages[0].Name() != "launch" {
		t.Errorf("stage[0] = %q, want 'launch'", stages[0].Name())
	}
}

func TestDescribeProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile profile.Profile
		want    string
	}{
		{
			"docker claude",
			profile.Profile{Environment: profile.EnvironmentDocker, Launch: profile.LaunchClaude},
			"docker + claude",
		},
		{
			"worktree host shell",
			profile.Profile{Worktree: &profile.WorktreeConfig{}, Environment: profile.EnvironmentHost, Launch: profile.LaunchShell},
			"worktree + host + shell",
		},
		{
			"worktree docker zellij",
			profile.Profile{Worktree: &profile.WorktreeConfig{}, Environment: profile.EnvironmentDocker, Launch: profile.LaunchZellij},
			"worktree + docker + zellij",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describeProfile(tt.profile)
			if got != tt.want {
				t.Errorf("describeProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}
