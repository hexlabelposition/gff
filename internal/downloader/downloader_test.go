package downloader

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDownloadUsesDestination(t *testing.T) {
	const want = "hello from downloader\n"
	destination := filepath.Join(t.TempDir(), "saved.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", strconv.Itoa(len(want)))

			if _, err := io.WriteString(w, want); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	d := New(server.Client())

	var updates int
	var lastWritten, lastTotal int64

	result, err := d.Download(context.Background(), DownloadRequest{
		URL:         server.URL + "/source.txt",
		Destination: destination,
		OnProgress: func(written, total int64) {
			if total != int64(len(want)) {
				t.Errorf("progress total: got %d, want %d", total, len(want))
			}

			updates++
			lastWritten = written
			lastTotal = total
		},
	})

	if err != nil {
		t.Fatalf("download: %v", err)
	}

	if updates == 0 {
		t.Fatal("expected progress updates")
	}

	if lastWritten != int64(len(want)) {
		t.Errorf("progress written: got %d, want %d", lastWritten, len(want))
	}

	if lastTotal != int64(len(want)) {
		t.Errorf("progress total: got %d, want %d", lastTotal, len(want))
	}

	if result.ContentLength != int64(len(want)) {
		t.Errorf(
			"content length: got %d, want %d",
			result.ContentLength,
			len(want),
		)
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

	_, err = os.Stat(destination + ".part")

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no partial file after success, got %v", err)
	}
}

func TestDownloadUnknownContentLength(t *testing.T) {
	const want = "hello from downloader\n"
	destination := filepath.Join(t.TempDir(), "saved.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)

			if err := http.NewResponseController(w).Flush(); err != nil {
				t.Errorf("flush response: %v", err)
				return
			}

			if _, err := io.WriteString(w, want); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	d := New(server.Client())

	var updates int
	var lastWritten, lastTotal int64

	result, err := d.Download(context.Background(), DownloadRequest{
		URL:         server.URL + "/source.txt",
		Destination: destination,
		OnProgress: func(written, total int64) {
			updates++
			lastWritten = written
			lastTotal = total
		},
	})

	if err != nil {
		t.Fatalf("download: %v", err)
	}

	if updates == 0 {
		t.Fatal("expected progress updates")
	}

	if lastWritten != int64(len(want)) {
		t.Errorf("progress written: got %d, want %d", lastWritten, len(want))
	}

	if lastTotal != -1 {
		t.Errorf("progress total: got %d, want -1", lastTotal)
	}

	if result.ContentLength != -1 {
		t.Errorf("content length: got %d, want -1", result.ContentLength)
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

func TestDownloadEmptyResponse(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "saved.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusOK)
		},
	))

	defer server.Close()

	d := New(server.Client())

	result, err := d.Download(context.Background(), DownloadRequest{
		URL:         server.URL + "/source.txt",
		Destination: destination,
		OnProgress: func(written, total int64) {
			if written != 0 {
				t.Errorf("progress written: got %d, want 0", written)
			}
			if total != 0 {
				t.Errorf("progress total: got %d, want 0", total)
			}
		},
	})

	if err != nil {
		t.Fatalf("download: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("status code: got %d, want %d", result.StatusCode, http.StatusOK)
	}

	if result.ContentLength != 0 {
		t.Errorf("content length: got %d, want 0", result.ContentLength)
	}

	if result.Size != 0 {
		t.Errorf("size: got %d, want 0", result.Size)
	}

	data, err := os.ReadFile(destination)

	if err != nil {
		t.Fatalf("read empty file: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("file size: got %d, want 0", len(data))
	}
}

func TestDownloadCanceledContext(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "canceled.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "hello"); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))

	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	d := New(server.Client())

	_, err := d.Download(ctx, DownloadRequest{
		URL:         server.URL,
		Destination: destination,
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	_, err = os.Stat(destination)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no destination file, got %v", err)
	}
}

func TestDownloadCanceledDuringTransfer(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "canceled.txt")

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "100")

			if _, err := io.WriteString(w, "hello"); err != nil {
				t.Errorf("write response: %v", err)
				return
			}

			if err := http.NewResponseController(w).Flush(); err != nil {
				t.Errorf("flush response: %v", err)
				return
			}

			<-r.Context().Done()
		},
	))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := New(server.Client())

	_, err := d.Download(ctx, DownloadRequest{
		URL:         server.URL,
		Destination: destination,
		OnProgress: func(written, total int64) {
			if written > 0 {
				cancel()
			}
		},
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	data, err := os.ReadFile(destination + ".part")

	if err != nil {
		t.Fatalf("read partial file: %v", err)
	}

	if string(data) != "hello" {
		t.Errorf("file content: got %q, want %q", string(data), "hello")
	}

	_, err = os.Stat(destination)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no completed file after cancellation, got %v", err)
	}
}

func TestDownloadPreservesExistingPartialFile(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "existing.txt")
	partialPath := destination + ".part"

	const original = "previous download"

	if err := os.WriteFile(partialPath, []byte(original), 0644); err != nil {
		t.Fatalf("prepare partial file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if _, err := io.WriteString(w, "new content\n"); err != nil {
				t.Errorf("write response: %v", err)
			}
		},
	))
	defer server.Close()

	d := New(server.Client())

	_, err := d.Download(context.Background(), DownloadRequest{
		URL:         server.URL,
		Destination: destination,
	})

	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected file-exists error, got %v", err)
	}

	data, err := os.ReadFile(partialPath)

	if err != nil {
		t.Fatalf("read existing partial file: %v", err)
	}

	if string(data) != original {
		t.Errorf("partial file content: got %q, want %q", string(data), original)
	}

	_, err = os.Stat(destination)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no completed file after error, got %v", err)
	}
}
