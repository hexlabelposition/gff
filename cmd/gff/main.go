package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/hexlabelposition/gff/internal/downloader"
	"github.com/hexlabelposition/gff/internal/progress"
)

const version = "0.3.0-dev"

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: gff <url> | gff version")
	}

	if os.Args[1] == "version" {
		if _, err := fmt.Fprintln(out, "gff", version); err != nil {
			return fmt.Errorf("print version: %w", err)
		}
		return nil
	}

	client := &http.Client{}
	started := time.Now()
	progressShown := false
	d := downloader.New(client)
	var outputErr error

	result, err := d.Download(ctx, downloader.DownloadRequest{
		URL: os.Args[1],
		OnProgress: func(written, total int64) {
			if outputErr != nil {
				return
			}

			progressShown = true
			_, outputErr = fmt.Fprintf(
				out,
				// Use carriage return and clear line to overwrite the previous progress line
				"\r\x1b[2K%s",
				formatProgress(written, total, time.Since(started)),
			)
		},
	})

	if progressShown {
		_, newlineErr := fmt.Fprintln(out)
		if outputErr == nil {
			outputErr = newlineErr
		}
	}

	if err != nil {
		return err
	}

	if outputErr != nil {
		return fmt.Errorf("print progress: %w", outputErr)
	}

	if _, err := fmt.Fprintf(
		out,
		"Saved %s (%d bytes)\n",
		result.Path,
		result.Size,
	); err != nil {
		return fmt.Errorf("print result: %w", err)
	}

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
