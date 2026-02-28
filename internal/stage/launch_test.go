package stage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hiragram/agent-workspace/internal/launcher"
	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

type mockLauncher struct {
	launched bool
	err      error
}

func (m *mockLauncher) Launch(_ context.Context, _ *pipeline.ExecutionContext) error {
	m.launched = true
	return m.err
}

func TestLaunchStage_Name(t *testing.T) {
	s := &LaunchStage{}
	if s.Name() != "launch" {
		t.Errorf("Name() = %q, want %q", s.Name(), "launch")
	}
}

func TestLaunchStage_Run(t *testing.T) {
	tests := []struct {
		name    string
		mode    profile.LaunchMode
		wantErr string
	}{
		{"shell launcher", profile.LaunchShell, ""},
		{"claude launcher", profile.LaunchClaude, ""},
		{"zellij launcher", profile.LaunchZellij, ""},
		{"unknown launcher", profile.LaunchMode("unknown"), "unknown launch mode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLauncher{}
			s := &LaunchStage{
				LauncherFactory: func(mode profile.LaunchMode) (launcher.Launcher, error) {
					if mode == tt.mode && tt.wantErr == "" {
						return mock, nil
					}
					if tt.wantErr != "" {
						return nil, fmt.Errorf("%s", tt.wantErr)
					}
					return mock, nil
				},
			}

			ec := &pipeline.ExecutionContext{
				Profile: profile.Profile{
					Launch: tt.mode,
				},
			}

			err := s.Run(context.Background(), ec)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Run() unexpected error: %v", err)
				}
				if !mock.launched {
					t.Error("launcher was not called")
				}
			} else {
				if err == nil {
					t.Fatal("Run() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestLaunchStage_LauncherError(t *testing.T) {
	mock := &mockLauncher{err: fmt.Errorf("launch failed")}
	s := &LaunchStage{
		LauncherFactory: func(_ profile.LaunchMode) (launcher.Launcher, error) {
			return mock, nil
		},
	}

	ec := &pipeline.ExecutionContext{
		Profile: profile.Profile{Launch: profile.LaunchShell},
	}

	err := s.Run(context.Background(), ec)
	if err == nil {
		t.Fatal("expected error from launcher")
	}
	if !strings.Contains(err.Error(), "launch failed") {
		t.Errorf("error = %q, want containing 'launch failed'", err.Error())
	}
}
