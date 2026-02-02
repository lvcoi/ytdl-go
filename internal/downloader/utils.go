package downloader

import (
	"fmt"
	"math"
	"time"
)

// formatDurationShort formats a duration in a short human-readable format.
// Returns format like "5s", "2m30s", or "1h15m" depending on duration length.
func formatDurationShort(d time.Duration) string {
	totalSeconds := int64(d.Seconds())

	if d < time.Minute {
		return fmt.Sprintf("%ds", totalSeconds)
	} else if d < time.Hour {
		mins := totalSeconds / 60
		secs := totalSeconds % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	}

	hours := totalSeconds / 3600
	mins := (totalSeconds % 3600) / 60
	return fmt.Sprintf("%dh%dm", hours, mins)
}
