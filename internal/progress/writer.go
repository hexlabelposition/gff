package progress

import "io"

type ProgressWriter struct {
	Writer  io.Writer
	Total   int64
	Written int64
}

func (w *ProgressWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.Written += int64(n)
	return n, err
}

func (w *ProgressWriter) Percent() (float64, bool) {
	if w.Total <= 0 {
		return 0, false
	}

	percent := float64(w.Written) / float64(w.Total) * 100

	return percent, true
}
