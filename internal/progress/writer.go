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
