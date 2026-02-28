package image

import (
	"fmt"
	"os"
	"path/filepath"
)

// PrepareBuildContext creates a temporary directory containing the Dockerfile
// and entrypoint.sh needed to build the Docker image.
// The caller must call the returned cleanup function when done.
func PrepareBuildContext() (dir string, cleanup func(), err error) {
	tmpDir, err := os.MkdirTemp("", "aw-build-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp dir: %w", err)
	}

	cleanupFn := func() { _ = os.RemoveAll(tmpDir) }

	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), dockerfile, 0644); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("writing Dockerfile: %w", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "entrypoint.sh"), entrypointSh, 0755); err != nil {
		cleanupFn()
		return "", nil, fmt.Errorf("writing entrypoint.sh: %w", err)
	}

	return tmpDir, cleanupFn, nil
}
