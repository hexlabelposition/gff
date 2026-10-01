package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
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
}

type DownloadResult struct {
	Path       string
	Size       int64
	Duration   time.Duration
	StatusCode int
}

func (d *Downloader) Download(
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

	resp, err := d.client.Get(parsed.String())

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

	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("create file: %w", err)
	}

	written, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()

	if copyErr != nil {
		return DownloadResult{}, fmt.Errorf("save response: %w", copyErr)
	}
	if closeErr != nil {
		return DownloadResult{}, fmt.Errorf("close file: %w", closeErr)
	}

	return DownloadResult{
		Path:       filename,
		Size:       written,
		Duration:   time.Since(started),
		StatusCode: resp.StatusCode,
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
