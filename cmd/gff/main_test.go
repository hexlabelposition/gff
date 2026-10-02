package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/hexlabelposition/gff/internal/downloader"
)

func TestRunDownloadsFile(t *testing.T) {
	const want = "hello from server\n"

	t.Chdir(t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, want); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	originalArgs := os.Args

	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"gff", server.URL + "/hello.txt"}

	if err := run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile("hello.txt")

	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}

	if string(data) != want {
		t.Errorf("file content: got %q, want %q", string(data), want)
	}
}

func TestRunRejectsNotFound(t *testing.T) {
	t.Chdir(t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		},
	))

	defer server.Close()

	originalArgs := os.Args

	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"gff", server.URL + "/missing.txt"}

	err := run()
	var statusErr *downloader.HTTPStatusError

	if !errors.As(err, &statusErr) {
		t.Fatalf("expected HTTPStatusError, got: %v", err)
	}

	if err := run(); err == nil {
		t.Fatal("expected an error for HTTP 404")
	}

	if statusErr.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", statusErr.StatusCode)
	}

	_, err = os.Stat("missing.txt")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no destination file, got: %v", err)
	}
}

func TestRunFollowsRedirects(t *testing.T) {
	const want = "downloaded after redirect\n"

	t.Chdir(t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/start.txt":
				http.Redirect(w, r, "/actual.txt", http.StatusFound)
			case "/actual.txt":
				if _, err := io.WriteString(w, want); err != nil {
					t.Errorf("write response: %v", err)
				}
			default:
				http.NotFound(w, r)
			}
		},
	))

	defer server.Close()

	originalArgs := os.Args

	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"gff", server.URL + "/start.txt"}

	if err := run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile("start.txt")

	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}

	if string(data) != want {
		t.Errorf("file content: got %q, want %q", string(data), want)
	}
}

func TestRunPreservesExistingFile(t *testing.T) {
	t.Chdir(t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "new content\n"); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	originalArgs := os.Args

	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"gff", server.URL + "/existing.txt"}

	const original = "original content\n"

	if err := os.WriteFile("existing.txt", []byte(original), 0644); err != nil {
		t.Fatalf("prepare existing file: %v", err)
	}

	if err := run(); !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected file-exists error, got: %v", err)
	}

	data, err := os.ReadFile("existing.txt")

	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}

	if string(data) != original {
		t.Errorf("existing file content changed: got %q, want %q", string(data), original)
	}
}

func TestFormatProgress(t *testing.T) {
	tests := []struct {
		name     string
		written  int64
		total    int64
		elapsed  time.Duration
		expected string
	}{
		{"quarter complete", 250, 1000, time.Second,
			"█████░░░░░░░░░░░░░░░ 25.0% | 250 B / 1000 B | 250 B/s | ETA 3s"},

		{"complete", 1000, 1000, time.Second,
			"████████████████████ 100.0% | 1000 B / 1000 B | 1000 B/s | ETA 0s"},

		{"no bytes written", 0, 1000, 2 * time.Second,
			"░░░░░░░░░░░░░░░░░░░░ 0.0% | 0 B / 1000 B | 0 B/s"},
		{"unknown total", 250, -1, time.Second, "250 B downloaded | 250 B/s"},
		{"empty response", 0, 0, time.Second, "0 B downloaded | 0 B/s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatProgress(tt.written, tt.total, tt.elapsed)
			if got != tt.expected {
				t.Errorf("formatProgress() = %q, want %q", got, tt.expected)
			}
		})
	}
}
