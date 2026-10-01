package downloader

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadUsesDestination(t *testing.T) {
	const want = "hello from downloader\n"
	destination := filepath.Join(t.TempDir(), "saved.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, want); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	d := New(server.Client())

	result, err := d.Download(DownloadRequest{
		URL:         server.URL + "/source.txt",
		Destination: destination,
	})

	if err != nil {
		t.Fatalf("download: %v", err)
	}

	if result.Path != destination {
		t.Errorf("path: got %q, want %q", result.Path, destination)
	}

	if result.Size != int64(len(want)) {
		t.Errorf("size: got %d, want %d", result.Size, len(want))
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("status code: got %d, want %d", result.StatusCode, http.StatusOK)
	}

	if result.Duration < 0 {
		t.Errorf("duration: got %v, want >= 0", result.Duration)
	}

	data, err := os.ReadFile(destination)

	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}

	if string(data) != want {
		t.Errorf("file content: got %q, want %q", string(data), want)
	}
}
