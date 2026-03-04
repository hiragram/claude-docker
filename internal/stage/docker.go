package stage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hiragram/agent-workspace/internal/config"
	"github.com/hiragram/agent-workspace/internal/docker"
	"github.com/hiragram/agent-workspace/internal/image"
	"github.com/hiragram/agent-workspace/internal/mount"
	"github.com/hiragram/agent-workspace/internal/pipeline"
)

const (
	defaultImageName  = "claude-code-docker"
	defaultVolumeName = "claude-code-local"
)

// DockerStage builds the Docker image, creates volumes, syncs config, and builds mounts.
type DockerStage struct {
	DockerClient docker.Client
	ConfigSyncer config.Syncer
	MountBuilder mount.Builder
}

// NewDockerStage creates a DockerStage with default implementations.
func NewDockerStage() *DockerStage {
	return &DockerStage{
		DockerClient: docker.NewShellClient(),
		ConfigSyncer: config.NewSyncer(),
		MountBuilder: mount.NewBuilder(),
	}
}

func (s *DockerStage) Name() string { return "docker" }

func (s *DockerStage) Run(ctx context.Context, ec *pipeline.ExecutionContext) error {
	// 1. Check Docker availability
	if err := s.DockerClient.CheckAvailable(); err != nil {
		return fmt.Errorf("docker is not available: %w", err)
	}

	// 2. Resolve custom Dockerfile path
	customDockerfile := ""
	if ec.Profile.Dockerfile != "" {
		resolved, err := resolveDockerfilePath(ec.Profile.Dockerfile)
		if err != nil {
			return fmt.Errorf("resolving dockerfile path: %w", err)
		}
		customDockerfile = resolved
	}

	// 3. Build Docker image
	buildDir, cleanup, err := image.PrepareBuildContext(customDockerfile)
	if err != nil {
		return fmt.Errorf("preparing build context: %w", err)
	}
	defer cleanup()

	fmt.Fprintf(os.Stderr, "Building Docker image '%s'...\n", defaultImageName)
	if err := s.DockerClient.Build(ctx, defaultImageName, buildDir); err != nil {
		return fmt.Errorf("building image: %w", err)
	}

	// 3. Create Docker volume
	if err := s.DockerClient.VolumeCreate(ctx, defaultVolumeName); err != nil {
		return fmt.Errorf("creating volume: %w", err)
	}

	// 4. Sync host settings
	claudeHome := claudeHomePath(ec.HomeDir)
	containerClaudeHome := filepath.Join(ec.HomeDir, ".agent-workspace")
	containerClaudeJSON := filepath.Join(ec.HomeDir, ".agent-workspace.json")

	if err := s.ConfigSyncer.SyncSettings(claudeHome, containerClaudeHome); err != nil {
		return fmt.Errorf("syncing settings: %w", err)
	}

	// 5. Ensure onboarding state
	if err := s.ConfigSyncer.EnsureOnboardingState(containerClaudeJSON); err != nil {
		return fmt.Errorf("ensuring onboarding state: %w", err)
	}

	// 6. Build mounts
	mounts, err := s.MountBuilder.BuildMounts(mount.MountOptions{
		HomeDir:             ec.HomeDir,
		WorkDir:             ec.WorkDir,
		ClaudeHome:          claudeHome,
		ContainerClaudeHome: containerClaudeHome,
		ContainerClaudeJSON: containerClaudeJSON,
		VolumeName:          defaultVolumeName,
	})
	if err != nil {
		return fmt.Errorf("building mounts: %w", err)
	}

	// 7. Update execution context
	ec.DockerImage = defaultImageName
	ec.DockerMounts = mounts
	ec.DockerVolume = defaultVolumeName

	return nil
}

// resolveDockerfilePath resolves a Dockerfile path.
// If the path is absolute, it is returned as-is.
// If relative, it is resolved against the git repo root.
func resolveDockerfilePath(dockerfilePath string) (string, error) {
	if filepath.IsAbs(dockerfilePath) {
		return dockerfilePath, nil
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("finding git root to resolve dockerfile path: %w", err)
	}
	repoRoot := strings.TrimSpace(string(out))
	return filepath.Join(repoRoot, dockerfilePath), nil
}

func claudeHomePath(homeDir string) string {
	if v := os.Getenv("CLAUDE_HOME"); v != "" {
		return v
	}
	return filepath.Join(homeDir, ".claude")
}
