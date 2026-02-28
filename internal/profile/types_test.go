package profile

import "testing"

func TestWorktreeConfig_EffectiveBase(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{"empty defaults to origin/main", "", "origin/main"},
		{"custom base", "origin/develop", "origin/develop"},
		{"specific commit", "abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorktreeConfig{Base: tt.base}
			if got := w.EffectiveBase(); got != tt.want {
				t.Errorf("EffectiveBase() = %q, want %q", got, tt.want)
			}
		})
	}
}
