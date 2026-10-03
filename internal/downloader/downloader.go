package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/hexlabelposition/gff/internal/progress"
)

type Downloader struct {
	client *http.Client
}

// New creates a Downloader. client must not be nil.
func New(client *http.Client) *Downloader {
	return &Downloader{client: client}
}

type DownloadRequest struct {
	URL         string
	Destination string
	OnProgress  func(written, total int64)
}

type DownloadResult struct {
	Path          string
	Size          int64
	Duration      time.Duration
	StatusCode    int
	ContentLength int64 // Expected body size; -1 means unknown.
}

func (d *Downloader) Download(
	ctx context.Context,
	request DownloadRequest,
) (result DownloadResult, returnErr error) {
	started := time.Now()
	rawURL := request.URL
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return DownloadResult{}, fmt.Errorf("URL must use http or https")
	}

	if parsed.Hostname() == "" {
		return DownloadResult{}, fmt.Errorf("URL must include a hostname")
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		parsed.String(),
		nil,
	)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("create HTTP request: %w", err)
	}

	resp, err := d.client.Do(httpRequest)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("error fetching URL: %w", err)
	}

	// Ensure the response body is closed when we're done with it
	defer func() {
		if err := resp.Body.Close(); err != nil && returnErr == nil {
			result = DownloadResult{}
			returnErr = fmt.Errorf("close response body: %w", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return DownloadResult{}, &HTTPStatusError{
			URL:        parsed.String(),
			StatusCode: resp.StatusCode,
		}
	}

	filename := request.Destination

	if filename == "" {
		filename = path.Base(parsed.Path)

		if filename == "." || filename == "/" {
			filename = "download"
		}
	}

	_, err = os.Lstat(filename)

	if err == nil {
		return DownloadResult{}, fmt.Errorf(
			"destination already exists: %w",
			os.ErrExist,
		)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return DownloadResult{}, fmt.Errorf(
			"check destination: %w",
			err,
		)
	}

	partialPath := filename + ".part"

	file, err := os.OpenFile(
		partialPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("create file: %w", err)
	}

	progressWriter := &progress.ProgressWriter{
		Writer:     file,
		Total:      resp.ContentLength,
		OnProgress: request.OnProgress,
	}

	written, copyErr := io.Copy(progressWriter, resp.Body)
	closeErr := file.Close()

	if copyErr != nil {
		return DownloadResult{}, fmt.Errorf("save response: %w", copyErr)
	}
	if closeErr != nil {
		return DownloadResult{}, fmt.Errorf("close file: %w", closeErr)
	}

	if err := os.Link(partialPath, filename); err != nil {
		return DownloadResult{}, fmt.Errorf(
			"publish downloaded file: %w",
			err,
		)
	}

	if err := os.Remove(partialPath); err != nil {
		return DownloadResult{}, fmt.Errorf(
			"remove partial file: %w",
			err,
		)
	}

	return DownloadResult{
		Path:          filename,
		Size:          written,
		Duration:      time.Since(started),
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
	}, nil
}

type HTTPStatusError struct {
	URL        string
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf(
		"unexpected HTTP status %d for %s",
		e.StatusCode,
		e.URL,
	)
}
