package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestArchive creates a tar.gz archive containing a single file named "aw"
// with the given content.
func createTestArchive(t *testing.T, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	if err := tw.WriteHeader(&tar.Header{
		Name: "aw",
		Size: int64(len(content)),
		Mode: 0755,
	}); err != nil {
		t.Fatalf("writing tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("writing tar content: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("closing tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("closing gzip writer: %v", err)
	}
	return buf.Bytes()
}

func TestExecute_NewVersionAvailable(t *testing.T) {
	newBinary := []byte("#!/bin/sh\necho new-binary\n")
	archive := createTestArchive(t, newBinary)

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/hiragram/agent-workspace/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{
			"tag_name": "v0.2.0",
			"assets": [
				{
					"name": "aw_testOS_testArch.tar.gz",
					"browser_download_url": "%s/download/archive.tar.gz"
				}
			]
		}`, "http://"+r.Host)
	})
	mux.HandleFunc("/download/archive.tar.gz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archive)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	// Patch the API base URL by using a custom HTTP client that rewrites URLs
	client := &urlRewriteClient{
		inner:   server.Client(),
		baseURL: server.URL,
	}

	// Create a dummy binary to be replaced
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "aw")
	if err := os.WriteFile(targetPath, []byte("old-binary"), 0755); err != nil {
		t.Fatalf("writing dummy binary: %v", err)
	}

	var stderr bytes.Buffer
	u := &Updater{
		HTTPClient:     client,
		CurrentVersion: "0.1.0",
		GOOS:           "testOS",
		GOARCH:         "testArch",
		Stderr:         &stderr,
		ExecPath:       targetPath,
	}

	if err := u.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	// Verify the binary was replaced
	got, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("reading updated binary: %v", err)
	}
	if !bytes.Equal(got, newBinary) {
		t.Errorf("binary content = %q, want %q", string(got), string(newBinary))
	}

	// Verify output messages
	output := stderr.String()
	if !strings.Contains(output, "0.1.0 → 0.2.0") {
		t.Errorf("stderr missing version update message, got: %s", output)
	}
	if !strings.Contains(output, "Updated successfully") {
		t.Errorf("stderr missing success message, got: %s", output)
	}
}

func TestExecute_AlreadyUpToDate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/hiragram/agent-workspace/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name": "v0.1.0", "assets": []}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &urlRewriteClient{
		inner:   server.Client(),
		baseURL: server.URL,
	}

	var stderr bytes.Buffer
	u := &Updater{
		HTTPClient:     client,
		CurrentVersion: "0.1.0",
		GOOS:           "darwin",
		GOARCH:         "arm64",
		Stderr:         &stderr,
	}

	if err := u.Execute(); err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	output := stderr.String()
	if !strings.Contains(output, "already the latest version") {
		t.Errorf("stderr missing up-to-date message, got: %s", output)
	}
}

func TestExecute_NetworkError(t *testing.T) {
	// Server that immediately closes
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", 500)
	}))
	defer server.Close()

	client := &urlRewriteClient{
		inner:   server.Client(),
		baseURL: server.URL,
	}

	var stderr bytes.Buffer
	u := &Updater{
		HTTPClient:     client,
		CurrentVersion: "0.1.0",
		GOOS:           "darwin",
		GOARCH:         "arm64",
		Stderr:         &stderr,
	}

	err := u.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecute_NoMatchingAsset(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/hiragram/agent-workspace/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"tag_name": "v0.2.0",
			"assets": [
				{"name": "aw_linux_amd64.tar.gz", "browser_download_url": "https://example.com/linux.tar.gz"}
			]
		}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := &urlRewriteClient{
		inner:   server.Client(),
		baseURL: server.URL,
	}

	var stderr bytes.Buffer
	u := &Updater{
		HTTPClient:     client,
		CurrentVersion: "0.1.0",
		GOOS:           "windows",
		GOARCH:         "amd64",
		Stderr:         &stderr,
	}

	err := u.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no release asset found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExtractBinary(t *testing.T) {
	content := []byte("binary-content-here")
	archive := createTestArchive(t, content)

	got, err := extractBinary(archive)
	if err != nil {
		t.Fatalf("extractBinary() error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("got %q, want %q", string(got), string(content))
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	// Create archive with a different file name
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	data := []byte("some content")
	_ = tw.WriteHeader(&tar.Header{Name: "other-file", Size: int64(len(data)), Mode: 0644})
	_, _ = tw.Write(data)
	_ = tw.Close()
	_ = gw.Close()

	_, err := extractBinary(buf.Bytes())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found in archive") {
		t.Errorf("unexpected error: %v", err)
	}
}

// urlRewriteClient rewrites GitHub API URLs to point at the test server.
type urlRewriteClient struct {
	inner   *http.Client
	baseURL string
}

func (c *urlRewriteClient) Do(req *http.Request) (*http.Response, error) {
	// Rewrite github.com API URLs to test server
	url := req.URL.String()
	url = strings.Replace(url, "https://api.github.com", c.baseURL, 1)
	newReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header
	return c.inner.Do(newReq)
}
