package update

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestFetchLatestRelease(t *testing.T) {
	body := `{
		"tag_name": "v0.2.0",
		"assets": [
			{
				"name": "aw_darwin_arm64.tar.gz",
				"browser_download_url": "https://example.com/darwin_arm64.tar.gz"
			},
			{
				"name": "aw_linux_amd64.tar.gz",
				"browser_download_url": "https://example.com/linux_amd64.tar.gz"
			}
		]
	}`

	client := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		},
	}

	release, err := FetchLatestRelease(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release.TagName != "v0.2.0" {
		t.Errorf("tag_name = %q, want %q", release.TagName, "v0.2.0")
	}
	if len(release.Assets) != 2 {
		t.Errorf("got %d assets, want 2", len(release.Assets))
	}
}

func TestFetchLatestRelease_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Not Found"}`)),
		},
	}

	_, err := FetchLatestRelease(client)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFetchLatestRelease_BadJSON(t *testing.T) {
	client := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`not json`)),
		},
	}

	_, err := FetchLatestRelease(client)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindAssetURL(t *testing.T) {
	release := &ReleaseInfo{
		Assets: []Asset{
			{Name: "aw_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64.tar.gz"},
			{Name: "aw_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64.tar.gz"},
		},
	}

	url, err := FindAssetURL(release, "darwin", "arm64")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com/darwin_arm64.tar.gz" {
		t.Errorf("url = %q, want %q", url, "https://example.com/darwin_arm64.tar.gz")
	}
}

func TestFindAssetURL_NotFound(t *testing.T) {
	release := &ReleaseInfo{
		Assets: []Asset{
			{Name: "aw_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64.tar.gz"},
		},
	}

	_, err := FindAssetURL(release, "windows", "amd64")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
