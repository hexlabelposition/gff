package progress

import (
	"fmt"
	"strings"
)

func FormatBytes(size int64) string {
	if size < 0 {
		return "unknown"
	}

	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	floatSize := float64(size)
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	unitIndex := 0

	for floatSize >= 1024 && unitIndex < len(units)-1 {
		floatSize /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.1f %s", floatSize, units[unitIndex])
}

func FormatBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(width))

	return strings.Repeat("█", filled) +
		strings.Repeat("░", width-filled)
}
