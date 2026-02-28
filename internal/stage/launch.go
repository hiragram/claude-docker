package stage

import (
	"context"
	"fmt"

	"github.com/hiragram/agent-workspace/internal/launcher"
	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

// LaunchStage selects and executes the appropriate launcher.
type LaunchStage struct {
	// LauncherFactory creates a Launcher for the given launch mode.
	// If nil, the default factory is used.
	LauncherFactory func(mode profile.LaunchMode) (launcher.Launcher, error)
}

func (s *LaunchStage) Name() string { return "launch" }

func (s *LaunchStage) Run(ctx context.Context, ec *pipeline.ExecutionContext) error {
	factory := s.LauncherFactory
	if factory == nil {
		factory = defaultLauncherFactory
	}

	l, err := factory(ec.Profile.Launch)
	if err != nil {
		return err
	}

	return l.Launch(ctx, ec)
}

func defaultLauncherFactory(mode profile.LaunchMode) (launcher.Launcher, error) {
	switch mode {
	case profile.LaunchShell:
		return &launcher.ShellLauncher{}, nil
	case profile.LaunchClaude:
		return &launcher.ClaudeLauncher{}, nil
	case profile.LaunchZellij:
		return &launcher.ZellijLauncher{}, nil
	default:
		return nil, fmt.Errorf("unknown launch mode: %q", mode)
	}
}
