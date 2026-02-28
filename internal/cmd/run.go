package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiragram/claude-docker/internal/config"
	"github.com/hiragram/claude-docker/internal/container"
	"github.com/hiragram/claude-docker/internal/docker"
	"github.com/hiragram/claude-docker/internal/image"
	"github.com/hiragram/claude-docker/internal/mount"
	"github.com/hiragram/claude-docker/internal/version"
	"github.com/hiragram/claude-docker/internal/worktree"
)

const (
	imageName  = "claude-code-docker"
	volumeName = "claude-code-local"
)

// Runner orchestrates the full claude-docker workflow.
type Runner struct {
	DockerClient docker.Client
	ConfigSyncer config.Syncer
	MountBuilder mount.Builder
	HomeDir      string
	WorkDir      string
}

// NewRunner creates a Runner with default implementations and auto-detected paths.
func NewRunner() (*Runner, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("detecting home directory: %w", err)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("detecting working directory: %w", err)
	}

	return &Runner{
		DockerClient: docker.NewShellClient(),
		ConfigSyncer: config.NewSyncer(),
		MountBuilder: mount.NewBuilder(),
		HomeDir:      homeDir,
		WorkDir:      workDir,
	}, nil
}

// claudeHome returns the CLAUDE_HOME path (defaults to ~/.claude).
func (r *Runner) claudeHome() string {
	if v := os.Getenv("CLAUDE_HOME"); v != "" {
		return v
	}
	return filepath.Join(r.HomeDir, ".claude")
}

// containerClaudeHome returns the container-side claude config path.
func (r *Runner) containerClaudeHome() string {
	return filepath.Join(r.HomeDir, ".claude-docker")
}

// containerClaudeJSON returns the onboarding state file path.
func (r *Runner) containerClaudeJSON() string {
	return filepath.Join(r.HomeDir, ".claude-docker.json")
}

// Execute runs the full workflow with the given CLI arguments.
func (r *Runner) Execute(ctx context.Context, args []string) error {
	// 1. Check Docker availability
	if err := r.DockerClient.CheckAvailable(); err != nil {
		return fmt.Errorf("docker check: %w", err)
	}

	// 2. Build Docker image
	buildDir, cleanup, err := image.PrepareBuildContext()
	if err != nil {
		return fmt.Errorf("preparing build context: %w", err)
	}
	defer cleanup()

	fmt.Fprintf(os.Stderr, "Building Docker image '%s'...\n", imageName)
	if err := r.DockerClient.Build(ctx, imageName, buildDir); err != nil {
		return fmt.Errorf("building image: %w", err)
	}

	// 3. Create Docker volume
	if err := r.DockerClient.VolumeCreate(ctx, volumeName); err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}

	// 4. Sync host settings
	if err := r.ConfigSyncer.SyncSettings(r.claudeHome(), r.containerClaudeHome()); err != nil {
		return fmt.Errorf("syncing settings: %w", err)
	}

	// 5. Ensure onboarding state
	if err := r.ConfigSyncer.EnsureOnboardingState(r.containerClaudeJSON()); err != nil {
		return fmt.Errorf("ensuring onboarding state: %w", err)
	}

	// 6. Build mounts
	mounts, err := r.MountBuilder.BuildMounts(mount.MountOptions{
		HomeDir:             r.HomeDir,
		WorkDir:             r.WorkDir,
		ClaudeHome:          r.claudeHome(),
		ContainerClaudeHome: r.containerClaudeHome(),
		ContainerClaudeJSON: r.containerClaudeJSON(),
		VolumeName:          volumeName,
	})
	if err != nil {
		return fmt.Errorf("building mounts: %w", err)
	}

	// 7. Build and run container
	runConfig := container.BuildRunConfig(container.RunOptions{
		ImageName:  imageName,
		Mounts:     mounts,
		ClaudeHome: r.claudeHome(),
		WorkDir:    r.WorkDir,
		CLIArgs:    args,
	})

	return r.DockerClient.Run(ctx, runConfig)
}

// Run is the top-level entry point. Returns an exit code.
func Run(args []string) int {
	if hasVersionFlag(args) {
		fmt.Printf("claude-docker %s\n", version.Version)
		return 0
	}

	if hasWorktreeFlag(args) {
		if err := worktree.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	runner, err := NewRunner()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if err := runner.Execute(context.Background(), args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

// hasVersionFlag checks if the args contain --version or -v.
func hasVersionFlag(args []string) bool {
	for _, a := range args {
		if a == "--version" || a == "-v" {
			return true
		}
	}
	return false
}

// hasWorktreeFlag checks if the args contain --worktree.
func hasWorktreeFlag(args []string) bool {
	for _, a := range args {
		if a == "--worktree" {
			return true
		}
	}
	return false
}
