package progress

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestProgressWriterWritesAndCounts(t *testing.T) {
	var destination bytes.Buffer
	writer := &ProgressWriter{Writer: &destination}

	n, err := writer.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("first write: %v", err)
	}
	if n != 5 {
		t.Errorf("first write count: got %d, want 5", n)
	}

	n, err = writer.Write([]byte(" world"))
	if err != nil {
		t.Fatalf("second write: %v", err)
	}
	if n != 6 {
		t.Errorf("second write count: got %d, want 6", n)
	}

	if writer.Written != 11 {
		t.Errorf("written: got %d, want 11", writer.Written)
	}

	if destination.String() != "hello world" {
		t.Errorf("content: got %q, want %q", destination.String(), "hello world")
	}
}

type partialWriter struct{}

func (partialWriter) Write(p []byte) (int, error) {
	return min(2, len(p)), io.ErrShortWrite
}

func TestProgressWriterCountsPartialWrite(t *testing.T) {
	writer := &ProgressWriter{Writer: partialWriter{}}
	n, err := writer.Write([]byte("hello"))

	if n != 2 {
		t.Errorf("write count: got %d, want 2", n)
	}
	if !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("error: got %v, want %v", err, io.ErrShortWrite)
	}
	if writer.Written != 2 {
		t.Errorf("written: got %d, want 2", writer.Written)
	}
}
