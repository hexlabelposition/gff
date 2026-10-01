package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: gff <url> | gff version")
		os.Exit(1)
	}

	if os.Args[1] == "version" {
		fmt.Println("gff", version)
		return
	}

	rawURL := os.Args[1]
	parsed, err := url.Parse(rawURL)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid URL:", err)
		os.Exit(1)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		fmt.Fprintln(os.Stderr, "URL must use http or https")
		os.Exit(1)
	}

	if parsed.Hostname() == "" {
		fmt.Fprintln(os.Stderr, "URL must include a hostname")
		os.Exit(1)
	}

	resp, err := http.Get(parsed.String())

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error fetching URL:", err)
		os.Exit(1)
	}

	// Ensure the response body is closed when we're done with it
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		// Close the response body before exiting
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "Error closing response body:", err)
		}

		fmt.Fprintln(os.Stderr, "Error: received non-OK HTTP status:", resp.Status)
		os.Exit(1)
	}

	fmt.Println("HTTP status:", resp.Status)
}
