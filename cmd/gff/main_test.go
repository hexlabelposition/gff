package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

	if err := run(); err == nil {
		t.Fatal("expected an error for HTTP 404")
	}

	_, err := os.Stat("missing.txt")
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
