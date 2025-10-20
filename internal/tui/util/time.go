package util

import (
	"fmt"
	"time"
)

// FormatTimeAgo formats a Unix timestamp (in seconds) as a relative time string.
func FormatTimeAgo(timestamp int64) string {
	if timestamp == 0 {
		return "never"
	}

	t := time.Unix(timestamp, 0)
	now := time.Now()
	duration := now.Sub(t)

	seconds := int(duration.Seconds())
	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30

	switch {
	case seconds < 60:
		return "just now"
	case minutes < 2:
		return "1 minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 2:
		return "1 hour ago"
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	case days < 2:
		return "1 day ago"
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case weeks < 2:
		return "1 week ago"
	case weeks < 4:
		return fmt.Sprintf("%d weeks ago", weeks)
	case months < 2:
		return "1 month ago"
	default:
		return t.Format("Jan 2")
	}
}
