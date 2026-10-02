package progress

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"
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

func TestProgressWriterPercent(t *testing.T) {
	tests := []struct {
		name        string
		written     int64
		total       int64
		wantPercent float64
		wantKnown   bool
	}{
		{"zero written", 0, 100, 0, true},
		{"quarter written", 25, 100, 25, true},
		{"all written", 100, 100, 100, true},
		{"more than total", 150, 100, 150, true},
		{"zero total", 0, 0, 0, false},
		{"unknown total", 10, -1, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &ProgressWriter{Written: tt.written, Total: tt.total}
			gotPercent, gotKnown := writer.Percent()
			if gotPercent != tt.wantPercent || gotKnown != tt.wantKnown {
				t.Errorf("Percent() = (%v, %v), want (%v, %v)", gotPercent, gotKnown, tt.wantPercent, tt.wantKnown)
			}
		})
	}
}

func TestProgressWriterSpeed(t *testing.T) {
	tests := []struct {
		name      string
		written   int64
		elapsed   time.Duration
		wantSpeed float64
	}{
		{"no bytes written", 0, time.Second, 0},
		{"one second", 1000, time.Second, 1000},
		{"two seconds", 1000, 2 * time.Second, 500},
		{"half second", 1000, 500 * time.Millisecond, 2000},
		{"zero elapsed time", 1000, 0, 0},
		{"negative elapsed time", 1000, -time.Second, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &ProgressWriter{Written: tt.written}
			gotSpeed := writer.Speed(tt.elapsed)
			if gotSpeed != tt.wantSpeed {
				t.Errorf("Speed() = %v, want %v", gotSpeed, tt.wantSpeed)
			}
		})
	}
}
