package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Formatters provides utility functions for formatting data
type Formatters struct {
	timeZone *time.Location
}

// NewFormatters creates a new Formatters instance
func NewFormatters(timezone string) *Formatters {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return &Formatters{
		timeZone: loc,
	}
}

// CheckLimit checks if a value is within the configured limit
func (f *Formatters) CheckLimit(check, config int) bool {
	if config == 0 {
		return true
	}
	return check <= config
}

// MinsToSecs converts time string (MM:SS or HH:MM:SS) to seconds
func (f *Formatters) MinsToSecs(timeStr string) int {
	parts := strings.Split(timeStr, ":")
	seconds := 0
	multiplier := 1

	for i := len(parts) - 1; i >= 0; i-- {
		var val int
		fmt.Sscanf(parts[i], "%d", &val)
		seconds += val * multiplier
		multiplier *= 60
	}

	return seconds
}

// SecsToMins converts seconds to time string (MM:SS or HH:MM:SS)
func (f *Formatters) SecsToMins(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// GetReadableTime converts seconds to human-readable time format
func (f *Formatters) GetReadableTime(seconds int) string {
	if seconds == 0 {
		return "0s"
	}

	parts := []string{}
	timeUnits := []struct {
		name     string
		duration int
	}{
		{"days", 86400},
		{"h", 3600},
		{"m", 60},
		{"s", 1},
	}

	remaining := seconds
	for _, unit := range timeUnits {
		if remaining >= unit.duration {
			value := remaining / unit.duration
			remaining %= unit.duration
			parts = append(parts, fmt.Sprintf("%d%s", value, unit.name))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}

	// Join parts with colon for time units, comma for days
	if len(parts) > 0 && strings.HasSuffix(parts[0], "days") {
		return parts[0] + ", " + strings.Join(parts[1:], ":")
	}
	return strings.Join(parts, ":")
}

// BytesToMB converts bytes to megabytes
func (f *Formatters) BytesToMB(size int64) int {
	return int((size / 1024 / 1024))
}

// GenKey generates a random key with prefix
func (f *Formatters) GenKey(message string, length int) string {
	if length <= 0 {
		length = 5
	}

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return message + "_" + string(b)
}

// GroupTheList divides a slice into groups of specified size
func (f *Formatters) GroupTheList(collection []interface{}, groupSize int, returnLength bool) (interface{}, int) {
	if groupSize <= 0 {
		groupSize = 5
	}

	var groups [][]interface{}
	total := 0

	for i := 0; i < len(collection); i += groupSize {
		end := i + groupSize
		if end > len(collection) {
			end = len(collection)
		}
		group := collection[i:end]
		groups = append(groups, group)
		total += len(group)
	}

	if returnLength {
		return len(groups), total
	}
	return groups, total
}

// SystemStats returns current system statistics
func (f *Formatters) SystemStats(startTime time.Time) (map[string]interface{}, error) {
	// CPU usage
	cpuPercent, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// CPU cores
	coreCount, err := cpu.Counts(true)
	if err != nil {
		coreCount = 0
	}

	// Disk usage
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	// Memory usage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Uptime
	uptime := int(time.Since(startTime).Seconds())

	return map[string]interface{}{
		"cpu":    fmt.Sprintf("%.1f%%", cpuUsage),
		"core":   coreCount,
		"disk":   fmt.Sprintf("%.1f%%", diskStat.UsedPercent),
		"ram":    fmt.Sprintf("%.1f%%", memStat.UsedPercent),
		"uptime": f.GetReadableTime(uptime),
	}, nil
}

// ConvertTelegraphURL converts telegra.ph to te.legra.ph
func (f *Formatters) ConvertTelegraphURL(url string) string {
	pattern := regexp.MustCompile(`(https?://)(telegra\.ph)`)
	converted := pattern.ReplaceAllString(url, "${1}te.legra.ph")
	return converted
}

// TelegraphPaste creates a Telegraph page (simplified version)
// Note: Full Telegraph API implementation would require additional dependencies
func (f *Formatters) TelegraphPaste(title, text, author, authorURL string) (string, error) {
	// This is a placeholder. Full implementation would use Telegraph API
	// For now, we'll just return an error indicating it needs implementation
	return "", fmt.Errorf("telegraph paste not implemented - use bb_paste instead")
}

// Post makes an HTTP POST request
func (f *Formatters) Post(url string, data interface{}) (interface{}, error) {
	var body io.Reader

	switch v := data.(type) {
	case string:
		body = strings.NewReader(v)
	case []byte:
		body = bytes.NewReader(v)
	default:
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	}

	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var jsonResult interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResult); err != nil {
		// If not JSON, return as string
		return string(bodyBytes), nil
	}

	return jsonResult, nil
}

// BBPaste uploads text to batbin.me
func (f *Formatters) BBPaste(text string) (string, error) {
	const baseURL = "https://batbin.me/"

	resp, err := http.Post(baseURL+"api/v2/paste", "text/plain", strings.NewReader(text))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	success, ok := result["success"].(bool)
	if !ok || !success {
		return "", fmt.Errorf("paste upload failed")
	}

	message, ok := result["message"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	return baseURL + message, nil
}

// Global instance
var Formatter *Formatters

// InitFormatter initializes the global formatter with timezone
func InitFormatter(timezone string) {
	Formatter = NewFormatters(timezone)
}
