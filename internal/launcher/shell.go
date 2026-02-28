package launcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/hiragram/agent-workspace/internal/docker"
	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

// ShellLauncher opens a shell in the workspace.
type ShellLauncher struct{}

func (l *ShellLauncher) Launch(ctx context.Context, ec *pipeline.ExecutionContext) error {
	switch ec.Profile.Environment {
	case profile.EnvironmentHost:
		return l.launchHostShell(ec)
	case profile.EnvironmentDocker:
		return l.launchDockerShell(ctx, ec)
	default:
		return fmt.Errorf("unsupported environment: %q", ec.Profile.Environment)
	}
}

func (l *ShellLauncher) launchHostShell(ec *pipeline.ExecutionContext) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	shellPath, err := exec.LookPath(shell)
	if err != nil {
		return fmt.Errorf("shell not found: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Opening shell in %s\n", ec.WorkDir)

	// Use syscall.Exec to replace the current process
	env := os.Environ()
	return syscall.Exec(shellPath, []string{shell}, env)
}

func (l *ShellLauncher) launchDockerShell(ctx context.Context, ec *pipeline.ExecutionContext) error {
	client := docker.NewShellClient()

	runConfig := docker.RunConfig{
		ImageName: ec.DockerImage,
		Mounts:    ec.DockerMounts,
		EnvVars: map[string]string{
			"HOST_CLAUDE_HOME": claudeHomePath(ec.HomeDir),
			"HOST_WORKSPACE":   ec.WorkDir,
		},
		WorkDir: ec.WorkDir,
		Command: []string{"/bin/bash"},
	}

	return client.Run(ctx, runConfig)
}
