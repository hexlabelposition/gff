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
	"strconv"
	"strings"
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

	var offset int64

	info, err := os.Lstat(partialPath)
	if err == nil {
		if !info.Mode().IsRegular() {
			return DownloadResult{}, fmt.Errorf(
				"partial path must be a regular file",
			)
		}

		offset = info.Size()
	} else if !errors.Is(err, os.ErrNotExist) {
		return DownloadResult{}, fmt.Errorf(
			"check partial file: %w",
			err,
		)
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

	if offset > 0 {
		httpRequest.Header.Set(
			"Range",
			fmt.Sprintf("bytes=%d-", offset),
		)
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

	totalSize := resp.ContentLength

	if resp.StatusCode == http.StatusPartialContent {
		if offset == 0 {
			return DownloadResult{}, fmt.Errorf(
				"unexpected partial response without a Range request",
			)
		}

		start, end, total, err := parseContentRange(
			resp.Header.Get("Content-Range"),
		)
		if err != nil {
			return DownloadResult{}, fmt.Errorf(
				"validate Content-Range: %w",
				err,
			)
		}

		if start != offset {
			return DownloadResult{}, fmt.Errorf(
				"range start mismatch: got %d, want %d",
				start,
				offset,
			)
		}

		if end != total-1 {
			return DownloadResult{}, fmt.Errorf(
				"incomplete response range: ends at %d, want %d",
				end,
				total-1,
			)
		}

		expectedLength := end - start + 1

		if resp.ContentLength >= 0 && resp.ContentLength != expectedLength {
			return DownloadResult{}, fmt.Errorf(
				"range length mismatch: got %d, want %d",
				resp.ContentLength,
				expectedLength,
			)
		}

		totalSize = total
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusPartialContent {
		return DownloadResult{}, &HTTPStatusError{
			URL:        parsed.String(),
			StatusCode: resp.StatusCode,
		}
	}

	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL

	if resp.StatusCode == http.StatusPartialContent {
		flags = os.O_WRONLY | os.O_APPEND
	}

	file, err := os.OpenFile(
		partialPath,
		flags,
		0644,
	)

	if err != nil {
		return DownloadResult{}, fmt.Errorf("create file: %w", err)
	}

	progressWriter := &progress.ProgressWriter{
		Writer:     file,
		Total:      totalSize,
		Written:    offset,
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

	completedSize := offset + written

	if totalSize >= 0 && completedSize != totalSize {
		return DownloadResult{}, fmt.Errorf(
			"download size mismatch: got %d, want %d",
			completedSize,
			totalSize,
		)
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
		Size:          completedSize,
		ContentLength: totalSize,
		Duration:      time.Since(started),
		StatusCode:    resp.StatusCode,
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

func parseContentRange(value string) (
	start, end, total int64,
	err error,
) {
	unit, rest, ok := strings.Cut(value, " ")
	if !ok || unit != "bytes" {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range: %q", value)
	}

	bounds, totalText, ok := strings.Cut(rest, "/")
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range: %q", value)
	}

	total, err = strconv.ParseInt(totalText, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse range total: %w", err)
	}

	startText, endText, ok := strings.Cut(bounds, "-")
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range: %q", value)
	}

	end, err = strconv.ParseInt(endText, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse range end: %w", err)
	}

	start, err = strconv.ParseInt(startText, 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse range start: %w", err)
	}

	if start < 0 || end < start || total <= end {
		return 0, 0, 0, fmt.Errorf("invalid Content-Range bounds: %q", value)
	}

	return start, end, total, nil
}
