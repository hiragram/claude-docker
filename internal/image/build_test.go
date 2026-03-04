package image

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmbeddedFilesNotEmpty(t *testing.T) {
	if len(dockerfile) == 0 {
		t.Error("embedded Dockerfile is empty")
	}
	if len(entrypointSh) == 0 {
		t.Error("embedded entrypoint.sh is empty")
	}
}

func TestEmbeddedDockerfileContent(t *testing.T) {
	content := string(dockerfile)
	if !strings.Contains(content, "FROM debian:bookworm-slim") {
		t.Error("Dockerfile should start with FROM debian:bookworm-slim")
	}
	if !strings.Contains(content, "ENTRYPOINT") {
		t.Error("Dockerfile should contain ENTRYPOINT")
	}
	if !strings.Contains(content, "useradd") {
		t.Error("Dockerfile should create claude user")
	}
}

func TestEmbeddedEntrypointContent(t *testing.T) {
	content := string(entrypointSh)
	if !strings.HasPrefix(content, "#!/bin/bash") {
		t.Error("entrypoint.sh should start with shebang")
	}
	if !strings.Contains(content, "setpriv") {
		t.Error("entrypoint.sh should use setpriv to switch user")
	}
	if !strings.Contains(content, "HOST_CLAUDE_HOME") {
		t.Error("entrypoint.sh should reference HOST_CLAUDE_HOME")
	}
}

func TestDefaultDockerfile(t *testing.T) {
	content := DefaultDockerfile()
	if len(content) == 0 {
		t.Error("DefaultDockerfile() returned empty content")
	}
	if string(content) != string(dockerfile) {
		t.Error("DefaultDockerfile() content does not match embedded dockerfile")
	}
}

func TestPrepareBuildContext(t *testing.T) {
	dir, cleanup, err := PrepareBuildContext("")
	if err != nil {
		t.Fatalf("PrepareBuildContext() error: %v", err)
	}
	defer cleanup()

	// Directory should exist
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("build context dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("build context path is not a directory")
	}

	// Dockerfile should exist with correct content
	dfContent, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}
	if string(dfContent) != string(dockerfile) {
		t.Error("Dockerfile content does not match embedded content")
	}

	// entrypoint.sh should exist with correct content and be executable
	epContent, err := os.ReadFile(filepath.Join(dir, "entrypoint.sh"))
	if err != nil {
		t.Fatalf("reading entrypoint.sh: %v", err)
	}
	if string(epContent) != string(entrypointSh) {
		t.Error("entrypoint.sh content does not match embedded content")
	}

	epInfo, err := os.Stat(filepath.Join(dir, "entrypoint.sh"))
	if err != nil {
		t.Fatalf("stat entrypoint.sh: %v", err)
	}
	if epInfo.Mode().Perm()&0111 == 0 {
		t.Error("entrypoint.sh should be executable")
	}
}

func TestPrepareBuildContext_CustomDockerfile(t *testing.T) {
	customDir := t.TempDir()
	customContent := []byte("FROM alpine:latest\nRUN echo custom\n")
	customPath := filepath.Join(customDir, "Dockerfile.custom")
	if err := os.WriteFile(customPath, customContent, 0644); err != nil {
		t.Fatal(err)
	}

	dir, cleanup, err := PrepareBuildContext(customPath)
	if err != nil {
		t.Fatalf("PrepareBuildContext() error: %v", err)
	}
	defer cleanup()

	// Dockerfile should contain custom content
	dfContent, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("reading Dockerfile: %v", err)
	}
	if string(dfContent) != string(customContent) {
		t.Errorf("Dockerfile content = %q, want %q", string(dfContent), string(customContent))
	}

	// entrypoint.sh should still be the embedded default
	epContent, err := os.ReadFile(filepath.Join(dir, "entrypoint.sh"))
	if err != nil {
		t.Fatalf("reading entrypoint.sh: %v", err)
	}
	if string(epContent) != string(entrypointSh) {
		t.Error("entrypoint.sh should be the embedded default even with custom Dockerfile")
	}
}

func TestPrepareBuildContext_CustomDockerfileNotFound(t *testing.T) {
	_, _, err := PrepareBuildContext("/nonexistent/Dockerfile")
	if err == nil {
		t.Fatal("expected error for nonexistent custom Dockerfile")
	}
	if !strings.Contains(err.Error(), "reading custom Dockerfile") {
		t.Errorf("error = %q, want containing 'reading custom Dockerfile'", err.Error())
	}
}

func TestPrepareBuildContextCleanup(t *testing.T) {
	dir, cleanup, err := PrepareBuildContext("")
	if err != nil {
		t.Fatalf("PrepareBuildContext() error: %v", err)
	}

	// Directory should exist before cleanup
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("dir should exist before cleanup: %v", err)
	}

	cleanup()

	// Directory should not exist after cleanup
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("dir should not exist after cleanup")
	}
}
