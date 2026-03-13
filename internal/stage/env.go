package stage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiragram/agent-workspace/internal/envfile"
	"github.com/hiragram/agent-workspace/internal/pipeline"
)

const (
	envFileName        = ".aw-env"
	profileEnvFileName = ".aw-profile-env"
)

// EnvStage loads custom environment variables from the profile config,
// .aw-profile-env file (written by parent process), and .aw-env file,
// merging them into the execution context.
//
// Override priority (highest wins):
//  1. .aw-env (dynamic, from on-create hook)
//  2. profile.Env (static, from current profile's env field)
//  3. .aw-profile-env (static, written by parent process's profile env)
type EnvStage struct{}

func (s *EnvStage) Name() string { return "env" }

func (s *EnvStage) Run(_ context.Context, ec *pipeline.ExecutionContext) error {
	merged := make(map[string]string)

	// 1. Start with .aw-profile-env (lowest priority, written by parent process)
	profileEnvFilePath := filepath.Join(ec.WorkDir, profileEnvFileName)
	profileFileEnv, err := envfile.ParseFile(profileEnvFilePath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", profileEnvFileName, err)
	}
	for k, v := range profileFileEnv {
		merged[k] = v
	}

	// 2. Overlay with current profile's env vars
	for k, v := range ec.Profile.Env {
		merged[k] = v
	}

	// 3. Write current profile env to .aw-profile-env for child processes
	if len(ec.Profile.Env) > 0 {
		if err := envfile.WriteFile(profileEnvFilePath, ec.Profile.Env); err != nil {
			return fmt.Errorf("writing %s: %w", profileEnvFileName, err)
		}
	}

	// 4. Overlay with .aw-env file vars (highest priority, from on-create hook)
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
