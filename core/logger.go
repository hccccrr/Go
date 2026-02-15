package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger handles logging with file rotation
type Logger struct {
	file         *os.File
	filename     string
	maxBytes     int64
	backupCount  int
	currentSize  int64
	mutex        sync.Mutex
	writer       io.Writer
}

// Log levels
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// Global logger instance
var LOGS *Logger

func init() {
	// Initialize logger similar to Python's RotatingFileHandler
	// maxBytes: 5MB, backupCount: 10
	LOGS = NewLogger("ShizuMusic.log", 5*1024*1024, 10)
	
	// Set log format
	log.SetFlags(0) // We'll handle formatting ourselves
	log.SetOutput(LOGS)
}

// NewLogger creates a new rotating file logger
func NewLogger(filename string, maxBytes int64, backupCount int) *Logger {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Warning: Failed to open log file: %v", err)
		return &Logger{
			filename:    filename,
			maxBytes:    maxBytes,
			backupCount: backupCount,
			writer:      os.Stdout,
		}
	}

	// Get current file size
	info, _ := file.Stat()
	currentSize := int64(0)
	if info != nil {
		currentSize = info.Size()
	}

	logger := &Logger{
		file:        file,
		filename:    filename,
		maxBytes:    maxBytes,
		backupCount: backupCount,
		currentSize: currentSize,
		writer:      io.MultiWriter(os.Stdout, file),
	}

	return logger
}

// Write implements io.Writer interface
func (l *Logger) Write(p []byte) (n int, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if rotation is needed
	if l.file != nil && l.currentSize+int64(len(p)) > l.maxBytes {
		l.rotate()
	}

	// Write to both console and file
	n, err = l.writer.Write(p)
	l.currentSize += int64(n)

	return n, err
}

// rotate performs log file rotation
func (l *Logger) rotate() {
	if l.file != nil {
		l.file.Close()
	}

	// Rotate old files
	for i := l.backupCount; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", l.filename, i)
		newName := fmt.Sprintf("%s.%d", l.filename, i+1)
		
		// Remove oldest backup if it exists
		if i == l.backupCount {
			os.Remove(newName)
		}
		
		// Rename files
		os.Rename(oldName, newName)
	}

	// Rename current file to .1
	backupName := fmt.Sprintf("%s.1", l.filename)
	os.Rename(l.filename, backupName)

	// Create new log file
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("Failed to create new log file: %v", err)
		return
	}

	l.file = file
	l.writer = io.MultiWriter(os.Stdout, file)
	l.currentSize = 0
}

// Log formats and writes a log message
func (l *Logger) Log(level, name, message string) {
	timestamp := time.Now().Format("15:04:05")
	logMsg := fmt.Sprintf("[%s]:[%s]:[%s]:: %s\n", timestamp, level, name, message)
	
	l.Write([]byte(logMsg))
}

// Info logs info message
func (l *Logger) Info(message string) {
	l.Log(LevelInfo, "ShizuMusic", message)
}

// Error logs error message
func (l *Logger) Error(message string) {
	l.Log(LevelError, "ShizuMusic", message)
}

// Warn logs warning message
func (l *Logger) Warn(message string) {
	l.Log(LevelWarn, "ShizuMusic", message)
}

// Debug logs debug message (only in debug mode)
func (l *Logger) Debug(message string) {
	// Only log debug in development
	// Can be controlled via environment variable
	if os.Getenv("DEBUG") == "true" {
		l.Log(LevelDebug, "ShizuMusic", message)
	}
}

// Infof logs formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Errorf logs formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

// Warnf logs formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Close closes the log file
func (l *Logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file != nil {
		l.file.Close()
	}
}

// GetLogPath returns the full path of log file
func (l *Logger) GetLogPath() string {
	if l.filename == "" {
		return ""
	}
	absPath, _ := filepath.Abs(l.filename)
	return absPath
}

// ClearLogs removes all log files
func (l *Logger) ClearLogs() error {
	// Remove main log file
	if err := os.Remove(l.filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove backup files
	for i := 1; i <= l.backupCount; i++ {
		backupName := fmt.Sprintf("%s.%d", l.filename, i)
		os.Remove(backupName) // Ignore errors
	}

	// Recreate main log file
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	l.file = file
	l.writer = io.MultiWriter(os.Stdout, file)
	l.currentSize = 0

	return nil
}

// Helper function to suppress specific library logs
func SuppressLibraryLogs() {
	// In Go, we don't have direct equivalents to Python's logging levels per library
	// But we can create custom loggers if needed
	// For now, this is a placeholder
}
