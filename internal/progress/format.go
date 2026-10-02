package progress

import "fmt"

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
