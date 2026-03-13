package envfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_WritesKeyValuePairs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-profile-env")

	env := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
	}

	if err := WriteFile(path, env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back with ParseFile and verify
	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if got["FOO"] != "bar" {
		t.Errorf("FOO = %q, want %q", got["FOO"], "bar")
	}
	if got["BAZ"] != "qux" {
		t.Errorf("BAZ = %q, want %q", got["BAZ"], "qux")
	}
}

func TestWriteFile_EmptyMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-profile-env")

	if err := WriteFile(path, map[string]string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not be created for empty map")
	}
}

func TestWriteFile_NilMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-profile-env")

	if err := WriteFile(path, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("file should not be created for nil map")
	}
}

func TestWriteFile_SortedKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-profile-env")

	env := map[string]string{
		"CHARLIE": "3",
		"ALPHA":   "1",
		"BRAVO":   "2",
	}

	if err := WriteFile(path, env); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	expected := "ALPHA=1\nBRAVO=2\nCHARLIE=3\n"
	if string(data) != expected {
		t.Errorf("file content = %q, want %q", string(data), expected)
	}
}

func TestWriteFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".aw-profile-env")

	// Use realistic values including JWT tokens
	env := map[string]string{
		"SUPABASE_ANON_KEY":         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0",
		"SUPABASE_URL":              "http://host.docker.internal:52296",
		"SUPABASE_SERVICE_ROLE_KEY": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU",
	}

	if err := WriteFile(path, env); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	got, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	for k, want := range env {
		if got[k] != want {
			t.Errorf("%s = %q, want %q", k, got[k], want)
		}
	}
	if len(got) != len(env) {
		t.Errorf("got %d entries, want %d", len(got), len(env))
	}
}
