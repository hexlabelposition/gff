package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hexlabelposition/gff/internal/downloader"
	"github.com/hexlabelposition/gff/internal/progress"
)

const version = "0.2.0-dev"

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
	started := time.Now()
	progressShown := false
	d := downloader.New(client)

	result, err := d.Download(downloader.DownloadRequest{
		URL: os.Args[1],
		OnProgress: func(written, total int64) {
			fmt.Printf(
				// Use carriage return and clear line to overwrite the previous progress line
				"\r\x1b[2K%s",
				formatProgress(written, total, time.Since(started)),
			)
			progressShown = true
		},
	})

	if progressShown {
		fmt.Println()
	}

	if err != nil {
		return err
	}

	fmt.Printf("Saved %s (%d bytes)\n", result.Path, result.Size)

	// Always return nil at the end of the run function to indicate success
	return nil
}

func formatProgress(written, total int64, elapsed time.Duration) string {
	state := progress.ProgressWriter{
		Written: written,
		Total:   total,
	}

	percent, percentKnown := state.Percent()

	var text string

	if percentKnown {
		text = fmt.Sprintf(
			"%s %.1f%% | %s / %s",
			progress.FormatBar(percent, 20),
			percent,
			progress.FormatBytes(written),
			progress.FormatBytes(total),
		)
	} else {
		text = fmt.Sprintf("%s downloaded", progress.FormatBytes(written))
	}

	speed := state.Speed(elapsed)
	text += " | " + progress.FormatBytes(int64(speed)) + "/s"

	eta, etaKnown := state.ETA(elapsed)
	if etaKnown {
		text += " | ETA " + eta.Round(time.Second).String()
	}

	return text
}
