package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiragram/agent-workspace/internal/envfile"
	"github.com/hiragram/agent-workspace/internal/pipeline"
)

const envFileName = ".aw-env"

// EnvStage loads custom environment variables from the profile config
// and .aw-env file, merging them into the execution context.
// File values override profile values for the same key.
type EnvStage struct{}

func (s *EnvStage) Name() string { return "env" }

func (s *EnvStage) Run(_ context.Context, ec *pipeline.ExecutionContext) error {
	merged := make(map[string]string)

	// 1. Start with profile-level env vars (static, from .agent-workspace.yml)
	for k, v := range ec.Profile.Env {
		merged[k] = v
	}

	// 2. Overlay with .aw-env file vars (dynamic, from on-create hook)
	envFilePath := filepath.Join(ec.WorkDir, envFileName)
	fileEnv, err := envfile.ParseFile(envFilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", envFileName, err)
	}
	for k, v := range fileEnv {
		merged[k] = v
	}

	if len(merged) > 0 {
		fmt.Fprintf(os.Stderr, "Loaded %d custom env var(s)\n", len(merged))
	}

	ec.EnvVars = merged
	return nil
}
