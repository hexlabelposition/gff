package progress

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"negative size", -1, "unknown"},
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"just under a kilobyte", 1023, "1023 B"},
		{"one kilobyte", 1024, "1.0 KiB"},
		{"one and a half kilobytes", 1536, "1.5 KiB"},
		{"one megabyte", 1048576, "1.0 MiB"},
		{"one gigabyte", 1073741824, "1.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.size)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s; expected %s", tt.size, result, tt.expected)
			}
		})
	}
}

func TestFormatBar(t *testing.T) {
	tests := []struct {
		name     string
		percent  float64
		width    int
		expected string
	}{
		{"zero percent", 0, 10, "░░░░░░░░░░"},
		{"half complete", 50, 10, "█████░░░░░"},
		{"full percent", 100, 10, "██████████"},
		{"over full percent", 150, 10, "██████████"},
		{"negative width", 50, -1, ""},
		{"negative percent", -10, 10, "░░░░░░░░░░"},
		{"zero width", 50, 0, ""},
		{"fractional fill", 40, 4, "█░░░"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBar(tt.percent, tt.width)
			if result != tt.expected {
				t.Errorf("FormatBar(%f, %d) = %s; expected %s", tt.percent, tt.width, result, tt.expected)
			}
		})
	}
}
