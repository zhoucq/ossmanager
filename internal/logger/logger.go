package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
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

// FromString converts a string to a LogLevel
func LogLevelFromString(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO // Default to INFO if unknown
	}
}

// LogFormat defines the format of log output
type LogFormat int

// Log formats
const (
	TEXT LogFormat = iota
	JSON
)

// String returns the string representation of a log format
func (f LogFormat) String() string {
	switch f {
	case TEXT:
		return "TEXT"
	case JSON:
		return "JSON"
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
	WithContext(ctx map[string]interface{}) Logger
	SetLevel(level LogLevel)
	GetLevel() LogLevel
	Close() error
}

// LogConfig holds configuration for the logger
type LogConfig struct {
	Level           LogLevel
	Format          LogFormat
	LogDir          string
	LogFile         string
	MaxSize         int    // maximum size in megabytes before log is rotated
	MaxBackups      int    // maximum number of old log files to retain
	MaxAge          int    // maximum number of days to retain old log files
	Compress        bool   // compress old log files
	ToConsole       bool   // also log to console
	TimestampFormat string // format for timestamps in logs
}

// DefaultLogConfig returns a default configuration for the logger
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           INFO,
		Format:          TEXT,
		LogDir:          "logs",
		LogFile:         "ossmanager.log",
		MaxSize:         10,   // 10 MB
		MaxBackups:      5,    // 5 files
		MaxAge:          30,   // 30 days
		Compress:        true, // compress old log files
		ToConsole:       true, // also log to console
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
}

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct {
	level      LogLevel
	format     LogFormat
	logger     *log.Logger
	fileWriter *lumberjack.Logger
	config     *LogConfig
	mutex      sync.RWMutex
	context    map[string]interface{}
}

// NewSimpleLogger creates a new SimpleLogger with the specified configuration
func NewSimpleLogger(config *LogConfig) (*SimpleLogger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Set up log rotation
	logFilename := filepath.Join(config.LogDir, config.LogFile)
	fileWriter := &lumberjack.Logger{
		Filename:   logFilename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// Create writer (file only or file+console)
	var writer io.Writer = fileWriter
	if config.ToConsole {
		writer = io.MultiWriter(fileWriter, os.Stdout)
	}

	// Create logger with file and line number
	logger := log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	return &SimpleLogger{
		level:      config.Level,
		format:     config.Format,
		logger:     logger,
		fileWriter: fileWriter,
		config:     config,
		context:    make(map[string]interface{}),
	}, nil
}

// getCallerInfo returns the file name and line number of the caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown:0"
	}
	// Get just the file name, not the full path
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

// formatMessage formats a log message based on the logger's configuration
func (l *SimpleLogger) formatMessage(level LogLevel, callerInfo string, format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)

	// Format based on the specified format
	if l.format == JSON {
		// Basic JSON format
		timestamp := time.Now().Format(l.config.TimestampFormat)
		jsonMsg := fmt.Sprintf(`{"timestamp":"%s","level":"%s","caller":"%s","message":"%s"`,
			timestamp, level.String(), callerInfo, message)

		// Add context if available
		if len(l.context) > 0 {
			for k, v := range l.context {
				jsonMsg += fmt.Sprintf(`,"ctx_%s":"%v"`, k, v)
			}
		}

		jsonMsg += "}"
		return jsonMsg
	} else {
		// Default TEXT format
		timestamp := time.Now().Format(l.config.TimestampFormat)
		txtMsg := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), callerInfo, message)

		// Add context if available
		if len(l.context) > 0 {
			contextStr := "{"
			first := true
			for k, v := range l.context {
				if !first {
					contextStr += ", "
				}
				contextStr += fmt.Sprintf("%s: %v", k, v)
				first = false
			}
			contextStr += "}"
			txtMsg += " " + contextStr
		}

		return txtMsg
	}
}

// Debug logs a debug message
func (l *SimpleLogger) Debug(format string, args ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.level <= DEBUG {
		caller := getCallerInfo(1)
		logMessage := l.formatMessage(DEBUG, caller, format, args...)
		l.logger.Println(logMessage)
	}
}

// Info logs an info message
func (l *SimpleLogger) Info(format string, args ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.level <= INFO {
		caller := getCallerInfo(1)
		logMessage := l.formatMessage(INFO, caller, format, args...)
		l.logger.Println(logMessage)
	}
}

// Warn logs a warning message
func (l *SimpleLogger) Warn(format string, args ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.level <= WARN {
		caller := getCallerInfo(1)
		logMessage := l.formatMessage(WARN, caller, format, args...)
		l.logger.Println(logMessage)
	}
}

// Error logs an error message
func (l *SimpleLogger) Error(format string, args ...interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.level <= ERROR {
		caller := getCallerInfo(1)
		logMessage := l.formatMessage(ERROR, caller, format, args...)
		l.logger.Println(logMessage)
	}
}

// WithContext returns a new logger with the given context
func (l *SimpleLogger) WithContext(ctx map[string]interface{}) Logger {
	newLogger := &SimpleLogger{
		level:      l.level,
		format:     l.format,
		logger:     l.logger,
		fileWriter: l.fileWriter,
		config:     l.config,
		context:    make(map[string]interface{}),
	}

	// Copy existing context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context
	for k, v := range ctx {
		newLogger.context[k] = v
	}

	return newLogger
}

// SetLevel changes the log level of the logger
func (l *SimpleLogger) SetLevel(level LogLevel) {
	var changed bool
	var oldLevel LogLevel

	l.mutex.Lock()
	if l.level != level {
		oldLevel = l.level
		l.level = level
		changed = true
	}
	l.mutex.Unlock()

	// 在锁外进行日志记录，避免死锁
	if changed {
		l.Info("Log level changed from %s to %s", oldLevel.String(), level.String())
	}
}

// GetLevel returns the current log level
func (l *SimpleLogger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.level
}

// Close closes the logger's file writer
func (l *SimpleLogger) Close() error {
	if l.fileWriter != nil {
		return l.fileWriter.Close()
	}
	return nil
}

// DefaultLogger is a global logger instance
var (
	DefaultLogger Logger
	loggerMutex   sync.RWMutex
)

// Initialize sets up the default logger with the specified configuration
func Initialize(config *LogConfig) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	// Close existing logger if it exists
	if DefaultLogger != nil {
		if closer, ok := DefaultLogger.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to close existing logger: %v\n", err)
			}
		}
	}

	if config == nil {
		config = DefaultLogConfig()
	}

	// Create the logger
	logger, err := NewSimpleLogger(config)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Set as default logger
	DefaultLogger = logger

	// Log initialization
	DefaultLogger.Info("Logger initialized with level %s, format %s", config.Level.String(), config.Format.String())
	DefaultLogger.Info("Logging to %s", filepath.Join(config.LogDir, config.LogFile))
	DefaultLogger.Info("Log rotation: maxSize=%dMB, maxBackups=%d, maxAge=%d days, compress=%v",
		config.MaxSize, config.MaxBackups, config.MaxAge, config.Compress)

	return nil
}

// InitializeWithDefaults sets up the default logger with default configuration
func InitializeWithDefaults() error {
	return Initialize(DefaultLogConfig())
}

// Shutdown closes the default logger
func Shutdown() error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if DefaultLogger == nil {
		return nil
	}

	// Close the logger
	if closer, ok := DefaultLogger.(io.Closer); ok {
		err := closer.Close()
		DefaultLogger = nil
		return err
	}

	DefaultLogger = nil
	return nil
}

// SetLevel changes the log level of the default logger
func SetLevel(level LogLevel) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		DefaultLogger.SetLevel(level)
	}
}

// GetLevel returns the current log level of the default logger
func GetLevel() LogLevel {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		return DefaultLogger.GetLevel()
	}
	return INFO // Default to INFO if logger not initialized
}

// WithContext returns a new logger with the given context
func WithContext(ctx map[string]interface{}) Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		return DefaultLogger.WithContext(ctx)
	}

	// If default logger is not initialized, initialize with defaults
	_ = InitializeWithDefaults()
	return DefaultLogger.WithContext(ctx)
}

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		DefaultLogger.Debug(format, args...)
	}
}

// Info logs an info message using the default logger
func Info(format string, args ...interface{}) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		DefaultLogger.Info(format, args...)
	}
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		DefaultLogger.Warn(format, args...)
	}
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()

	if DefaultLogger != nil {
		DefaultLogger.Error(format, args...)
	}
}
