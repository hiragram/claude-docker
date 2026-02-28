package launcher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/hiragram/agent-workspace/internal/docker"
	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
)

// ClaudeLauncher runs Claude Code.
type ClaudeLauncher struct{}

func (l *ClaudeLauncher) Launch(ctx context.Context, ec *pipeline.ExecutionContext) error {
	switch ec.Profile.Environment {
	case profile.EnvironmentHost:
		return l.launchHostClaude(ec)
	case profile.EnvironmentDocker:
		return l.launchDockerClaude(ctx, ec)
	default:
		return fmt.Errorf("unsupported environment: %q", ec.Profile.Environment)
	}
}

func (l *ClaudeLauncher) launchHostClaude(ec *pipeline.ExecutionContext) error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude is not installed. Install Claude Code: https://claude.ai/install.sh")
	}

	fmt.Fprintf(os.Stderr, "Launching Claude in %s\n", ec.WorkDir)

	args := []string{"claude"}
	// Use syscall.Exec to replace the current process
	env := os.Environ()
	return syscall.Exec(claudePath, args, env)
}

func (l *ClaudeLauncher) launchDockerClaude(ctx context.Context, ec *pipeline.ExecutionContext) error {
	client := docker.NewShellClient()

	command := []string{"claude", "--dangerously-skip-permissions"}

	runConfig := docker.RunConfig{
		ImageName: ec.DockerImage,
		Mounts:    ec.DockerMounts,
		EnvVars: map[string]string{
			"HOST_CLAUDE_HOME": claudeHomePath(ec.HomeDir),
			"HOST_WORKSPACE":   ec.WorkDir,
		},
		WorkDir: ec.WorkDir,
		Command: command,
	}

	return client.Run(ctx, runConfig)
}

func claudeHomePath(homeDir string) string {
	if v := os.Getenv("CLAUDE_HOME"); v != "" {
		return v
	}
	return filepath.Join(homeDir, ".claude")
}
