package profile

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
		wantErr string
	}{
		{
			name: "valid docker + claude",
			profile: Profile{
				Environment: EnvironmentDocker,
				Launch:      LaunchClaude,
			},
		},
		{
			name: "valid host + shell with worktree",
			profile: Profile{
				Worktree:    &WorktreeConfig{Base: "origin/main"},
				Environment: EnvironmentHost,
				Launch:      LaunchShell,
			},
		},
		{
			name: "valid docker + zellij with config",
			profile: Profile{
				Worktree:    &WorktreeConfig{},
				Environment: EnvironmentDocker,
				Launch:      LaunchZellij,
				Zellij:      &ZellijConfig{Layout: "default"},
			},
		},
		{
			name: "missing environment",
			profile: Profile{
				Launch: LaunchClaude,
			},
			wantErr: "environment is required",
		},
		{
			name: "missing launch",
			profile: Profile{
				Environment: EnvironmentDocker,
			},
			wantErr: "launch is required",
		},
		{
			name: "unknown environment",
			profile: Profile{
				Environment: "kubernetes",
				Launch:      LaunchClaude,
			},
			wantErr: "unknown environment",
		},
		{
			name: "unknown launch mode",
			profile: Profile{
				Environment: EnvironmentHost,
				Launch:      "tmux",
			},
			wantErr: "unknown launch mode",
		},
		{
			name: "zellij config with non-zellij launch",
			profile: Profile{
				Environment: EnvironmentDocker,
				Launch:      LaunchClaude,
				Zellij:      &ZellijConfig{Layout: "default"},
			},
			wantErr: "zellij config is only valid with launch: zellij",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.profile)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr string
	}{
		{
			name: "valid config",
			config: Config{
				Default: "test",
				Profiles: map[string]Profile{
					"test": {
						Environment: EnvironmentDocker,
						Launch:      LaunchClaude,
					},
				},
			},
		},
		{
			name: "no profiles",
			config: Config{
				Profiles: map[string]Profile{},
			},
			wantErr: "no profiles defined",
		},
		{
			name: "default profile not found",
			config: Config{
				Default: "nonexistent",
				Profiles: map[string]Profile{
					"test": {
						Environment: EnvironmentDocker,
						Launch:      LaunchClaude,
					},
				},
			},
			wantErr: "default profile \"nonexistent\" not found",
		},
		{
			name: "invalid profile in config",
			config: Config{
				Profiles: map[string]Profile{
					"bad": {
						Environment: "invalid",
						Launch:      LaunchClaude,
					},
				},
			},
			wantErr: "config validation errors",
		},
		{
			name: "no default is ok",
			config: Config{
				Profiles: map[string]Profile{
					"test": {
						Environment: EnvironmentHost,
						Launch:      LaunchShell,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(&tt.config)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("ValidateConfig() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateConfig() error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}
