package envfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_ValidKeyValue(t *testing.T) {
	input := "FOO=bar\nBAZ=qux"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "bar")
	}
	if env["BAZ"] != "qux" {
		t.Errorf("BAZ = %q, want %q", env["BAZ"], "qux")
	}
}

func TestParse_SkipsComments(t *testing.T) {
	input := "# this is a comment\nFOO=bar\n# another comment"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 1 {
		t.Errorf("got %d entries, want 1", len(env))
	}
	if env["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "bar")
	}
}

func TestParse_SkipsEmptyLines(t *testing.T) {
	input := "\nFOO=bar\n\n\nBAZ=qux\n"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 {
		t.Errorf("got %d entries, want 2", len(env))
	}
}

func TestParse_StripsDoubleQuotes(t *testing.T) {
	input := `FOO="bar baz"`
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["FOO"] != "bar baz" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "bar baz")
	}
}

func TestParse_ErrorOnMissingEquals(t *testing.T) {
	input := "INVALID_LINE"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for line without =")
	}
}

func TestParse_ErrorOnEmptyKey(t *testing.T) {
	input := "=value"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestParse_ValueContainsEquals(t *testing.T) {
	input := "URL=http://host:1234?a=b"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["URL"] != "http://host:1234?a=b" {
		t.Errorf("URL = %q, want %q", env["URL"], "http://host:1234?a=b")
	}
}

func TestParse_EmptyValue(t *testing.T) {
	input := "EMPTY="
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["EMPTY"] != "" {
		t.Errorf("EMPTY = %q, want empty string", env["EMPTY"])
	}
}

func TestParse_WhitespaceAroundKeyValue(t *testing.T) {
	input := "  FOO  =  bar  "
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "bar")
	}
}

func TestParse_DuplicateKeyLastWins(t *testing.T) {
	input := "FOO=first\nFOO=second"
	env, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["FOO"] != "second" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "second")
	}
}

func TestParseFile_NonexistentFile(t *testing.T) {
	env, err := ParseFile("/tmp/nonexistent-aw-env-test-file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 0 {
		t.Errorf("got %d entries, want 0 for nonexistent file", len(env))
	}
}

func TestParseFile_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-env")
	content := "FOO=bar\nBAZ=qux\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	env, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", env["FOO"], "bar")
	}
	if env["BAZ"] != "qux" {
		t.Errorf("BAZ = %q, want %q", env["BAZ"], "qux")
	}
}
