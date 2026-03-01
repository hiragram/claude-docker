package profile

import (
	"testing"
)

func TestMergeProfile_OverrideEnvironment(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
	}
	override := Profile{
		Environment: EnvironmentHost,
	}

	merged := MergeProfile(base, override)

	if merged.Environment != EnvironmentHost {
		t.Errorf("Environment = %q, want %q", merged.Environment, EnvironmentHost)
	}
	if merged.Launch != LaunchClaude {
		t.Errorf("Launch = %q, want %q (should be preserved from base)", merged.Launch, LaunchClaude)
	}
}

func TestMergeProfile_OverrideLaunch(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
	}
	override := Profile{
		Launch: LaunchShell,
	}

	merged := MergeProfile(base, override)

	if merged.Environment != EnvironmentDocker {
		t.Errorf("Environment = %q, want %q (should be preserved from base)", merged.Environment, EnvironmentDocker)
	}
	if merged.Launch != LaunchShell {
		t.Errorf("Launch = %q, want %q", merged.Launch, LaunchShell)
	}
}

func TestMergeProfile_AddWorktree(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
	}
	override := Profile{
		Worktree: &WorktreeConfig{Base: "origin/develop"},
	}

	merged := MergeProfile(base, override)

	if merged.Worktree == nil {
		t.Fatal("Worktree should not be nil")
	}
	if merged.Worktree.Base != "origin/develop" {
		t.Errorf("Worktree.Base = %q, want %q", merged.Worktree.Base, "origin/develop")
	}
}

func TestMergeProfile_OverrideWorktree(t *testing.T) {
	base := Profile{
		Worktree:    &WorktreeConfig{Base: "origin/main"},
		Environment: EnvironmentDocker,
		Launch:      LaunchZellij,
		Zellij:      &ZellijConfig{Layout: "default"},
	}
	override := Profile{
		Worktree: &WorktreeConfig{Base: "origin/develop", OnCreate: "./setup.sh"},
	}

	merged := MergeProfile(base, override)

	if merged.Worktree == nil {
		t.Fatal("Worktree should not be nil")
	}
	if merged.Worktree.Base != "origin/develop" {
		t.Errorf("Worktree.Base = %q, want %q", merged.Worktree.Base, "origin/develop")
	}
	if merged.Worktree.OnCreate != "./setup.sh" {
		t.Errorf("Worktree.OnCreate = %q, want %q", merged.Worktree.OnCreate, "./setup.sh")
	}
}

func TestMergeProfile_PreserveWorktreeFromBase(t *testing.T) {
	base := Profile{
		Worktree:    &WorktreeConfig{Base: "origin/main"},
		Environment: EnvironmentDocker,
		Launch:      LaunchZellij,
		Zellij:      &ZellijConfig{Layout: "default"},
	}
	override := Profile{
		Environment: EnvironmentHost,
	}

	merged := MergeProfile(base, override)

	if merged.Worktree == nil {
		t.Fatal("Worktree should be preserved from base")
	}
	if merged.Worktree.Base != "origin/main" {
		t.Errorf("Worktree.Base = %q, want %q", merged.Worktree.Base, "origin/main")
	}
}

func TestMergeProfile_AddZellij(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchZellij,
	}
	override := Profile{
		Zellij: &ZellijConfig{Layout: "custom"},
	}

	merged := MergeProfile(base, override)

	if merged.Zellij == nil {
		t.Fatal("Zellij should not be nil")
	}
	if merged.Zellij.Layout != "custom" {
		t.Errorf("Zellij.Layout = %q, want %q", merged.Zellij.Layout, "custom")
	}
}

func TestMergeProfile_EmptyOverride(t *testing.T) {
	base := Profile{
		Worktree:    &WorktreeConfig{},
		Environment: EnvironmentDocker,
		Launch:      LaunchZellij,
		Zellij:      &ZellijConfig{Layout: "default"},
	}
	override := Profile{}

	merged := MergeProfile(base, override)

	if merged.Environment != EnvironmentDocker {
		t.Errorf("Environment = %q, want %q", merged.Environment, EnvironmentDocker)
	}
	if merged.Launch != LaunchZellij {
		t.Errorf("Launch = %q, want %q", merged.Launch, LaunchZellij)
	}
	if merged.Worktree == nil {
		t.Fatal("Worktree should be preserved from base")
	}
	if merged.Zellij == nil {
		t.Fatal("Zellij should be preserved from base")
	}
}

func TestMergeConfig_BuiltinOnlyProfilesPreserved(t *testing.T) {
	builtin := Config{
		Default: "a",
		Profiles: map[string]Profile{
			"a": {Environment: EnvironmentDocker, Launch: LaunchClaude},
			"b": {Environment: EnvironmentHost, Launch: LaunchShell},
		},
	}
	user := Config{
		Profiles: map[string]Profile{
			"c": {Environment: EnvironmentHost, Launch: LaunchClaude},
		},
	}

	merged := MergeConfig(builtin, user)

	if _, ok := merged.Profiles["a"]; !ok {
		t.Error("builtin profile 'a' should be preserved")
	}
	if _, ok := merged.Profiles["b"]; !ok {
		t.Error("builtin profile 'b' should be preserved")
	}
	if _, ok := merged.Profiles["c"]; !ok {
		t.Error("user profile 'c' should be added")
	}
}

func TestMergeConfig_UserOnlyProfileAdded(t *testing.T) {
	builtin := Config{
		Profiles: map[string]Profile{
			"a": {Environment: EnvironmentDocker, Launch: LaunchClaude},
		},
	}
	user := Config{
		Profiles: map[string]Profile{
			"custom": {Environment: EnvironmentHost, Launch: LaunchShell},
		},
	}

	merged := MergeConfig(builtin, user)

	p, ok := merged.Profiles["custom"]
	if !ok {
		t.Fatal("user-only profile 'custom' should be present")
	}
	if p.Environment != EnvironmentHost {
		t.Errorf("custom.Environment = %q, want %q", p.Environment, EnvironmentHost)
	}
	if p.Launch != LaunchShell {
		t.Errorf("custom.Launch = %q, want %q", p.Launch, LaunchShell)
	}
}

func TestMergeConfig_SameNameProfileMerged(t *testing.T) {
	builtin := Config{
		Profiles: map[string]Profile{
			"claude": {Environment: EnvironmentDocker, Launch: LaunchClaude},
		},
	}
	user := Config{
		Profiles: map[string]Profile{
			"claude": {
				Worktree: &WorktreeConfig{Base: "origin/develop"},
			},
		},
	}

	merged := MergeConfig(builtin, user)

	p := merged.Profiles["claude"]
	if p.Environment != EnvironmentDocker {
		t.Errorf("Environment = %q, want %q (should be preserved from builtin)", p.Environment, EnvironmentDocker)
	}
	if p.Launch != LaunchClaude {
		t.Errorf("Launch = %q, want %q (should be preserved from builtin)", p.Launch, LaunchClaude)
	}
	if p.Worktree == nil {
		t.Fatal("Worktree should be added from user config")
	}
	if p.Worktree.Base != "origin/develop" {
		t.Errorf("Worktree.Base = %q, want %q", p.Worktree.Base, "origin/develop")
	}
}

func TestMergeConfig_DefaultOverride(t *testing.T) {
	builtin := Config{
		Default: "builtin-default",
		Profiles: map[string]Profile{
			"builtin-default": {Environment: EnvironmentDocker, Launch: LaunchClaude},
		},
	}
	user := Config{
		Default: "user-default",
		Profiles: map[string]Profile{
			"user-default": {Environment: EnvironmentHost, Launch: LaunchShell},
		},
	}

	merged := MergeConfig(builtin, user)

	if merged.Default != "user-default" {
		t.Errorf("Default = %q, want %q", merged.Default, "user-default")
	}
}

func TestMergeConfig_DefaultPreservedWhenUserEmpty(t *testing.T) {
	builtin := Config{
		Default: "builtin-default",
		Profiles: map[string]Profile{
			"builtin-default": {Environment: EnvironmentDocker, Launch: LaunchClaude},
		},
	}
	user := Config{
		Profiles: map[string]Profile{
			"custom": {Environment: EnvironmentHost, Launch: LaunchShell},
		},
	}

	merged := MergeConfig(builtin, user)

	if merged.Default != "builtin-default" {
		t.Errorf("Default = %q, want %q (should be preserved from builtin)", merged.Default, "builtin-default")
	}
}

func TestMergeProfile_EnvMerged(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
		Env:         map[string]string{"A": "1", "B": "2"},
	}
	override := Profile{
		Env: map[string]string{"B": "override", "C": "3"},
	}

	merged := MergeProfile(base, override)

	if merged.Env["A"] != "1" {
		t.Errorf("Env[A] = %q, want %q", merged.Env["A"], "1")
	}
	if merged.Env["B"] != "override" {
		t.Errorf("Env[B] = %q, want %q (override should win)", merged.Env["B"], "override")
	}
	if merged.Env["C"] != "3" {
		t.Errorf("Env[C] = %q, want %q", merged.Env["C"], "3")
	}
}

func TestMergeProfile_NilEnvOverridePreservesBase(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
		Env:         map[string]string{"A": "1"},
	}
	override := Profile{}

	merged := MergeProfile(base, override)

	if merged.Env["A"] != "1" {
		t.Errorf("Env[A] = %q, want %q (should be preserved from base)", merged.Env["A"], "1")
	}
}

func TestMergeProfile_EnvDoesNotMutateBase(t *testing.T) {
	base := Profile{
		Environment: EnvironmentDocker,
		Launch:      LaunchClaude,
		Env:         map[string]string{"A": "1"},
	}
	override := Profile{
		Env: map[string]string{"B": "2"},
	}

	MergeProfile(base, override)

	// base.Env should not have been mutated
	if _, ok := base.Env["B"]; ok {
		t.Error("base.Env should not have been mutated")
	}
}

func TestMergeConfig_WorktreeEmptyObjectEnablesWorktree(t *testing.T) {
	builtin := Config{
		Profiles: map[string]Profile{
			"claude": {Environment: EnvironmentDocker, Launch: LaunchClaude},
		},
	}
	user := Config{
		Profiles: map[string]Profile{
			"claude": {
				Worktree: &WorktreeConfig{},
			},
		},
	}

	merged := MergeConfig(builtin, user)

	p := merged.Profiles["claude"]
	if p.Worktree == nil {
		t.Fatal("Worktree should be enabled via empty object from user config")
	}
	if p.Environment != EnvironmentDocker {
		t.Errorf("Environment = %q, want %q (should be preserved)", p.Environment, EnvironmentDocker)
	}
}
