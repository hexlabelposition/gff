package progress

import (
	"io"
	"time"
)

type ProgressWriter struct {
	Writer     io.Writer
	Total      int64
	Written    int64
	OnProgress func(written, total int64)
}

func (w *ProgressWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.Written += int64(n)

	// Call the OnProgress callback if it's set
	if w.OnProgress != nil {
		w.OnProgress(w.Written, w.Total)
	}

	return n, err
}

func (w *ProgressWriter) Percent() (float64, bool) {
	if w.Total <= 0 {
		return 0, false
	}

	percent := float64(w.Written) / float64(w.Total) * 100

	return percent, true
}

func (w *ProgressWriter) Speed(elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}

	return float64(w.Written) / elapsed.Seconds()
}

func (w *ProgressWriter) ETA(elapsed time.Duration) (time.Duration, bool) {
	if w.Total <= 0 {
		return 0, false
	}

	if w.Written >= w.Total {
		return 0, true
	}

	speed := w.Speed(elapsed)

	if speed <= 0 {
		return 0, false
	}

	remaining := w.Total - w.Written
	seconds := float64(remaining) / speed
	eta := time.Duration(seconds * float64(time.Second))

	return eta, true
}
