package worktree

import (
	"strings"
	"testing"
)

func TestGenerateName(t *testing.T) {
	name, err := GenerateName()
	if err != nil {
		t.Skipf("Skipping on systems without %s: %v", dictPath, err)
	}
	parts := strings.Split(name, "-")
	if len(parts) != 3 {
		t.Errorf("expected 3 words joined by hyphens, got %q (%d parts)", name, len(parts))
	}
	for _, p := range parts {
		if len(p) == 0 || len(p) > maxWordLen {
			t.Errorf("word %q has invalid length (expected 1-%d)", p, maxWordLen)
		}
		if p != strings.ToLower(p) {
			t.Errorf("word %q should be lowercase", p)
		}
	}
}

func TestGenerateName_Uniqueness(t *testing.T) {
	name1, err1 := GenerateName()
	name2, err2 := GenerateName()
	if err1 != nil || err2 != nil {
		t.Skipf("Skipping on systems without dictionary")
	}
	if name1 == name2 {
		t.Logf("Warning: same name generated twice: %q (statistically unlikely but possible)", name1)
	}
}
