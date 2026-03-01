package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hiragram/agent-workspace/internal/pipeline"
	"github.com/hiragram/agent-workspace/internal/profile"
	"github.com/hiragram/agent-workspace/internal/stage"
	"github.com/hiragram/agent-workspace/internal/update"
	"github.com/hiragram/agent-workspace/internal/version"
)

// Run is the top-level entry point. Returns an exit code.
func Run(args []string) int {
	if hasVersionFlag(args) {
		fmt.Printf("aw %s\n", version.Version)
		return 0
	}

	if len(args) > 0 && args[0] == "update" {
		if err := update.Run(version.Version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	if len(args) > 0 && args[0] == "profiles" {
		return runProfiles()
	}

	// Determine profile name
	profileName := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		profileName = args[0]
	}

	// Load config
	cfg, err := profile.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Validate the whole config
	if err := profile.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// If no profile name given, use default or list profiles
	if profileName == "" {
		if cfg.Default != "" {
			profileName = cfg.Default
		} else {
			printAvailableProfiles(cfg)
			return 0
		}
	}

	p, ok := cfg.Profiles[profileName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: profile %q not found\n", profileName)
		printAvailableProfiles(cfg)
		return 1
	}

	// Validate the selected profile
	if err := profile.Validate(p); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid profile %q: %v\n", profileName, err)
		return 1
	}

	// Build execution context
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	ec := &pipeline.ExecutionContext{
		Profile:     p,
		ProfileName: profileName,
		HomeDir:     homeDir,
		OrigWorkDir: workDir,
		WorkDir:     workDir,
	}

	// Warn about on-end limitations
	if p.Worktree != nil && p.Worktree.OnEnd != "" &&
		p.Environment == profile.EnvironmentHost &&
		p.Launch != profile.LaunchZellij {
		fmt.Fprintf(os.Stderr, "Warning: on-end hook will not run with environment: host + launch: %s (process is replaced via exec)\n", p.Launch)
	}

	// Build pipeline stages
	stages := buildStages(p)
	pipe := pipeline.New(stages...)

	if err := pipe.Execute(context.Background(), ec); err != nil {
		runOnEndIfConfigured(ec)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	runOnEndIfConfigured(ec)
	return 0
}

func runOnEndIfConfigured(ec *pipeline.ExecutionContext) {
	if ec.Profile.Worktree == nil || ec.Profile.Worktree.OnEnd == "" {
		return
	}
	if ec.WorktreePath == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "Running on-end hook...\n")
	if err := stage.RunOnEndHook(ec); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: on-end hook failed: %v\n", err)
	}
}

// buildStages creates the pipeline stages based on the profile configuration.
func buildStages(p profile.Profile) []pipeline.Stage {
	var stages []pipeline.Stage

	// Stage 1: Worktree (conditional)
	if p.Worktree != nil {
		stages = append(stages, &stage.WorktreeStage{})
	}

	// Stage 2: Docker setup (conditional)
	if p.Environment == profile.EnvironmentDocker {
		stages = append(stages, stage.NewDockerStage())
	}

	// Stage 3: Env loading (conditional — only for Docker, where custom env vars are needed)
	if p.Environment == profile.EnvironmentDocker {
		stages = append(stages, &stage.EnvStage{})
	}

	// Stage 4: Launch (always)
	stages = append(stages, &stage.LaunchStage{})

	return stages
}

func runProfiles() int {
	cfg, err := profile.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Show config source
	if cfg.Source.IsBuiltin {
		fmt.Println("Source: built-in default (no .agent-workspace.yml found)")
	} else {
		fmt.Printf("Source: %s\n", cfg.Source.FilePath)
	}
	fmt.Println()

	printAvailableProfiles(cfg)
	return 0
}

func printAvailableProfiles(cfg *profile.Config) {
	fmt.Println("Available profiles:")
	for name, p := range cfg.Profiles {
		marker := "  "
		if name == cfg.Default {
			marker = "* "
		}

		desc := describeProfile(p)
		fmt.Printf("  %s%s  (%s)\n", marker, name, desc)
	}
	fmt.Println()
	fmt.Println("Usage: aw <profile-name>")
	if cfg.Default != "" {
		fmt.Printf("       aw              (runs default: %s)\n", cfg.Default)
	}
}

func describeProfile(p profile.Profile) string {
	parts := []string{}
	if p.Worktree != nil {
		parts = append(parts, "worktree")
	}
	parts = append(parts, string(p.Environment))
	parts = append(parts, string(p.Launch))
	return strings.Join(parts, " + ")
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
