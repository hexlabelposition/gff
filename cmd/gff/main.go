package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hexlabelposition/gff/internal/downloader"
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

	client := &http.Client{}
	d := downloader.New(client)

	result, err := d.Download(downloader.DownloadRequest{
		URL: os.Args[1],
	})

	if err != nil {
		return err
	}

	fmt.Printf("Saved %s (%d bytes)\n", result.Path, result.Size)

	// Always return nil at the end of the run function to indicate success
	return nil
}
