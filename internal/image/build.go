package image

import (
	"fmt"
	"os"
	"path/filepath"
)

// PrepareBuildContext creates a temporary directory containing the Dockerfile
// and entrypoint.sh needed to build the Docker image.
// If customDockerfilePath is non-empty, the Dockerfile is read from that path
// instead of using the embedded default. The entrypoint.sh is always the
// embedded default (it is still copied to the build context so the custom
// Dockerfile can reference it with COPY if desired).
// The caller must call the returned cleanup function when done.
func PrepareBuildContext(customDockerfilePath string) (dir string, cleanup func(), err error) {
	tmpDir, err := os.MkdirTemp("", "aw-build-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}

	cleanupFn := func() { _ = os.RemoveAll(tmpDir) }

	var dockerfileContent []byte
	if customDockerfilePath != "" {
		dockerfileContent, err = os.ReadFile(customDockerfilePath)
		if err != nil {
			cleanupFn()
			return "", nil, fmt.Errorf("reading custom Dockerfile %q: %w", customDockerfilePath, err)
		}
	} else {
		dockerfileContent = dockerfile
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), dockerfileContent, 0644); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("writing Dockerfile: %w", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "entrypoint.sh"), entrypointSh, 0755); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("writing entrypoint.sh: %w", err)
	}

	return tmpDir, cleanupFn, nil
}
