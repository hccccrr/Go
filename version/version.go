package version

import (
	"runtime"
	"time"
)

// Version information
type Version struct {
	ShizuMusic string
	GoVersion  string
	Gogram     string
	NTgCalls   string
	StartTime  time.Time
}

var Info = &Version{
	ShizuMusic: "3.0.0",
	GoVersion:  runtime.Version(),
	Gogram:     "1.1.13",
	NTgCalls:   "1.0.0",
	StartTime:  time.Now(),
}

// GetUptime returns bot uptime
func GetUptime() time.Duration {
	return time.Since(Info.StartTime)
}

// GetUptimeString returns formatted uptime
func GetUptimeString() string {
	uptime := GetUptime()
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	
	return formatDuration(hours, minutes, seconds)
}

func formatDuration(h, m, s int) string {
	if h > 0 {
		return formatTime(h, "hour") + ", " + formatTime(m, "minute")
	}
	if m > 0 {
		return formatTime(m, "minute") + ", " + formatTime(s, "second")
	}
	return formatTime(s, "second")
}

func formatTime(val int, unit string) string {
	if val == 1 {
		return "1 " + unit
	}
	return formatInt(val) + " " + unit + "s"
}

func formatInt(n int) string {
	return string(rune(n + '0'))
}
