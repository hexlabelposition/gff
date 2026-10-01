package main

import (
	"fmt"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: gff <url>")
		os.Exit(1)
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

	fmt.Println("Fetching URL:", parsed.String())
}
