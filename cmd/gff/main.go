package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

const version = "0.1.0-dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: gff <url> | gff version")
	}

	if os.Args[1] == "version" {
		fmt.Println("gff", version)
		return nil
	}

	rawURL := os.Args[1]
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL must use http or https")
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("URL must include a hostname")
	}

	resp, err := http.Get(parsed.String())

	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	// Ensure the response body is closed when we're done with it
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	fmt.Println("HTTP status:", resp.Status)

	filename := path.Base(parsed.Path)

	if filename == "." || filename == "/" {
		filename = "download"
	}

	fmt.Println("Filename:", filename)

	file, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)

	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	written, copyErr := io.Copy(file, resp.Body)
	closeErr := file.Close()

	if copyErr != nil {
		return fmt.Errorf("save response: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close file: %w", closeErr)
	}

	fmt.Printf("Saved %s (%d bytes)\n", filename, written)

	// Always return nil at the end of the run function to indicate success
	return nil
}
