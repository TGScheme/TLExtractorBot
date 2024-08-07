package utils

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	hours := int(d / time.Hour)
	minutes := int(d % time.Hour / time.Minute)
	seconds := int(d % time.Minute / time.Second)

	var result string
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 || hours > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 || minutes > 0 || hours > 0 {
		result += fmt.Sprintf("%ds", seconds)
	}
	return result
}
