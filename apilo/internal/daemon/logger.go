package daemon

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Logger provides structured logging with levels
type Logger struct {
	level       LogLevel
	file        *os.File
	logger      *log.Logger
	mu          sync.Mutex
	maxSizeMB   int64
	currentSize int64
}

// NewLogger creates a new logger instance
func NewLogger(logFile string, level LogLevel) (*Logger, error) {
	// Expand home directory
	if strings.HasPrefix(logFile, "~/") {
		home, _ := os.UserHomeDir()
		logFile = filepath.Join(home, logFile[2:])
	}

	// Create directory if needed
	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	stat, _ := file.Stat()
	currentSize := stat.Size()

	// Create multi-writer for console and file
	multiWriter := io.MultiWriter(os.Stdout, file)

	logger := &Logger{
		level:       level,
		file:        file,
		logger:      log.New(multiWriter, "", 0),
		maxSizeMB:   100, // 100MB default
		currentSize: currentSize,
	}

	return logger, nil
}

// Close closes the logger
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log writes a log message with the given level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Check file size and rotate if needed
	if l.currentSize > l.maxSizeMB*1024*1024 {
		l.rotate()
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := level.String()
	message := fmt.Sprintf(format, args...)

	logLine := fmt.Sprintf("[%s] %s - %s\n", timestamp, levelStr, message)

	l.logger.Print(logLine)
	l.currentSize += int64(len(logLine))
}

// rotate rotates the log file
func (l *Logger) rotate() {
	if l.file == nil {
		return
	}

	// Close current file
	l.file.Close()

	// Rename current file with timestamp
	oldPath := l.file.Name()
	newPath := fmt.Sprintf("%s.%s", oldPath, time.Now().Format("20060102-150405"))
	os.Rename(oldPath, newPath)

	// Open new file
	file, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to rotate log file: %v", err)
		return
	}

	l.file = file
	l.currentSize = 0

	// Update logger output
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.logger.SetOutput(multiWriter)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path string, statusCode int, latency time.Duration) {
	l.Info("HTTP %s %s - %d (%v)", method, path, statusCode, latency)
}

// LogOptimization logs an optimization request
func (l *Logger) LogOptimization(url string, cacheHit bool, latency time.Duration) {
	cacheStatus := "MISS"
	if cacheHit {
		cacheStatus = "HIT"
	}
	l.Info("Optimize %s - Cache: %s, Latency: %v", url, cacheStatus, latency)
}

// LogCacheOperation logs a cache operation
func (l *Logger) LogCacheOperation(operation, key string, success bool) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}
	l.Debug("Cache %s - Key: %s, Status: %s", operation, key[:min(16, len(key))], status)
}

// LogMetrics logs metrics snapshot
func (l *Logger) LogMetrics(stats *MetricsStats) {
	l.Info("Metrics - Requests: %d, Cache Hit Ratio: %.2f%%, Avg Latency: %v, Memory: %.2fMB",
		stats.TotalRequests,
		stats.CacheHitRatio*100,
		stats.AvgLatency,
		stats.MemoryUsageMB,
	)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
