package downloader

import (
	"fmt"
	"math"
	"time"
)

// formatDurationShort formats a duration in a short human-readable format.
// Returns format like "5s", "2m30s", or "1h15m" depending on duration length.
func formatDurationShort(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), math.Mod(d.Seconds(), 60))
	}
	return fmt.Sprintf("%.0fh%.0fm", d.Hours(), math.Mod(d.Minutes(), 60))
}
