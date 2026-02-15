package utils

import "fmt"

// SecsToMins converts seconds to formatted time string
// Returns format: DD:HH:MM:SS, HH:MM:SS, MM:SS, or 00:SS
func SecsToMins(seconds int) string {
	if seconds <= 0 {
		return "-"
	}

	d := seconds / (3600 * 24)
	h := (seconds / 3600) % 24
	m := (seconds % 3600) / 60
	s := (seconds % 3600) % 60

	if d > 0 {
		return fmt.Sprintf("%02d:%02d:%02d:%02d", d, h, m, s)
	} else if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%02d:%02d", m, s)
	} else if s > 0 {
		return fmt.Sprintf("00:%02d", s)
	}

	return "-"
}

// MinsToSecs converts formatted time string to seconds
// Supports formats: DD:HH:MM:SS, HH:MM:SS, MM:SS
func MinsToSecs(timeStr string) int {
	if timeStr == "" || timeStr == "-" {
		return 0
	}

	var parts []int
	var current int
	
	for i := 0; i < len(timeStr); i++ {
		if timeStr[i] == ':' {
			parts = append(parts, current)
			current = 0
		} else if timeStr[i] >= '0' && timeStr[i] <= '9' {
			current = current*10 + int(timeStr[i]-'0')
		}
	}
	parts = append(parts, current)

	seconds := 0
	multiplier := 1

	// Process from right to left
	for i := len(parts) - 1; i >= 0; i-- {
		seconds += parts[i] * multiplier
		if multiplier == 1 {
			multiplier = 60 // seconds to minutes
		} else if multiplier == 60 {
			multiplier = 3600 // minutes to hours
		} else if multiplier == 3600 {
			multiplier = 86400 // hours to days
		}
	}

	return seconds
}

// FormatDuration formats seconds into human-readable duration
func FormatDuration(seconds int) string {
	if seconds <= 0 {
		return "0s"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// ParseDuration parses a duration string and returns seconds
func ParseDuration(duration string) int {
	// If it's in MM:SS or HH:MM:SS format
	if len(duration) > 0 && (duration[0] >= '0' && duration[0] <= '9') {
		for i := 0; i < len(duration); i++ {
			if duration[i] == ':' {
				return MinsToSecs(duration)
			}
		}
	}
	
	// Otherwise assume it's already in seconds or invalid
	return 0
}
