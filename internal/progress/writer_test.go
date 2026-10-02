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
	var updates []update

	writer := &ProgressWriter{
		Writer: partialWriter{},
		Total:  5,
		OnProgress: func(written, total int64) {
			updates = append(updates, update{
				written: written,
				total:   total,
			})
		},
	}

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

	if len(updates) != 1 {
		t.Fatalf("update count: got %d, want 1", len(updates))
	}

	want := update{written: 2, total: 5}

	if updates[0] != want {
		t.Errorf("update: got %+v, want %+v", updates[0], want)
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

func TestProgressWriterETA(t *testing.T) {
	tests := []struct {
		name      string
		written   int64
		total     int64
		elapsed   time.Duration
		wantETA   time.Duration
		wantKnown bool
	}{
		{"quarter complete", 250, 1000, time.Second, 3 * time.Second, true},
		{"half complete", 500, 1000, 2 * time.Second, 2 * time.Second, true},
		{"complete", 1000, 1000, time.Second, 0, true},
		{"written exceeds total", 1200, 1000, time.Second, 0, true},
		{"no bytes written", 0, 1000, time.Second, 0, false},
		{"zero elapsed time", 250, 1000, 0, 0, false},
		{"negative elapsed time", 250, 1000, -time.Second, 0, false},
		{"unknown total", 10, -1, time.Second, 0, false},
		{"zero total", 0, 0, time.Second, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &ProgressWriter{Written: tt.written, Total: tt.total}
			gotETA, gotKnown := writer.ETA(tt.elapsed)
			if gotETA != tt.wantETA || gotKnown != tt.wantKnown {
				t.Errorf("ETA() = (%v, %v), want (%v, %v)", gotETA, gotKnown, tt.wantETA, tt.wantKnown)
			}
		})
	}
}

type update struct {
	written int64
	total   int64
}

func TestProgressWriterReportsProgress(t *testing.T) {
	var updates []update

	writer := &ProgressWriter{
		Writer: io.Discard,
		Total:  11,
		OnProgress: func(written, total int64) {
			updates = append(updates, update{
				written: written,
				total:   total,
			})
		},
	}

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

	if len(updates) != 2 {
		t.Fatalf("update count: got %d, want 2", len(updates))
	}

	if updates[0] != (update{written: 5, total: 11}) {
		t.Errorf("first update: got %+v, want %+v", updates[0], update{written: 5, total: 11})
	}

	if updates[1] != (update{written: 11, total: 11}) {
		t.Errorf("second update: got %+v, want %+v", updates[1], update{written: 11, total: 11})
	}
}
