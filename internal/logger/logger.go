package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

// Log levels
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of a log level
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

// Logger is a simple logging interface
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewSimpleLogger creates a new SimpleLogger with the specified level and output
func NewSimpleLogger(level LogLevel, out io.Writer) *SimpleLogger {
	return &SimpleLogger{
		level:  level,
		logger: log.New(out, "", log.LstdFlags),
	}
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Error logs an error message
func (l *SimpleLogger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// DefaultLogger is a global logger instance
var DefaultLogger Logger

// Initialize sets up the default logger
func Initialize(level LogLevel, logDir string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	logFile := filepath.Join(logDir, fmt.Sprintf("ossmanager_%s.log", time.Now().Format("20060102")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create a multi-writer to log to both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Initialize the default logger
	DefaultLogger = NewSimpleLogger(level, multiWriter)

	// Log initialization
	DefaultLogger.Info("Logger initialized with level %s", level.String())
	DefaultLogger.Info("Logging to file: %s", logFile)

	return nil
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Error(format, args...)
	}
}
