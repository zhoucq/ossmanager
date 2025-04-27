package logger_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ossmanager/internal/logger"
)

func Example() {
	// Create a temporary directory for example logs
	tempDir, err := os.MkdirTemp("", "logger_example")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Configure and initialize the logger
	config := &logger.LogConfig{
		Level:      logger.DEBUG,
		Format:     logger.TEXT,
		LogDir:     tempDir,
		LogFile:    "example.log",
		MaxSize:    10,    // 10MB
		MaxBackups: 5,     // Keep 5 backup files
		MaxAge:     30,    // 30 days
		Compress:   true,  // Compress old logs
		ToConsole:  false, // Don't log to console to avoid polluting test output
	}

	if err := logger.Initialize(config); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer logger.Shutdown()

	// Log messages at different levels
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	// Create a context logger for a specific operation
	ctxLogger := logger.WithContext(map[string]interface{}{
		"operation": "file_upload",
		"user_id":   "user123",
		"file_size": 1024 * 1024,
	})

	// Log with context
	ctxLogger.Info("Starting file upload")
	ctxLogger.Info("File upload completed")

	// Change log level at runtime
	logger.SetLevel(logger.INFO)
	logger.Debug("This debug message won't be logged now")
	logger.Info("But info messages will still be logged")

	// Output log file location for reference
	logFilePath := filepath.Join(tempDir, "example.log")
	fmt.Printf("Log file created at: %s\n", logFilePath)

	// Output: Log file created at: <temp_path>/example.log
}

func ExampleWithContext() {
	// Create a temporary directory for our logs
	tempDir, err := os.MkdirTemp("", "logger_example_context")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// Configure and initialize the logger
	config := &logger.LogConfig{
		Level:      logger.DEBUG,
		Format:     logger.TEXT,
		LogDir:     tempDir,
		LogFile:    "example-context.log",
		MaxSize:    10,    // 10MB
		MaxBackups: 5,     // Keep 5 backup files
		MaxAge:     30,    // 30 days
		Compress:   true,  // Compress old logs
		ToConsole:  false, // Don't log to console to avoid polluting test output
	}

	if err := logger.Initialize(config); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer logger.Shutdown()

	// Create logger with request context
	requestLogger := logger.WithContext(map[string]interface{}{
		"request_id": "req-123456",
		"client_ip":  "192.168.1.1",
	})

	// Log with request context
	requestLogger.Info("Request received")

	// Add more context for a specific operation
	operationLogger := requestLogger.WithContext(map[string]interface{}{
		"operation": "user_login",
		"username":  "john_doe",
	})

	// Log with combined context
	operationLogger.Info("User login attempt")
	operationLogger.Info("User login successful")

	// Original request logger still has only the request context
	requestLogger.Info("Request completed")

	// Output log file location for reference
	logFilePath := filepath.Join(tempDir, "example-context.log")
	fmt.Printf("Log file created at: %s\n", logFilePath)

	// Output: Log file created at: <temp_path>/example-context.log
}

func ExampleLogLevelFromString() {
	// Convert string to log level
	level := logger.LogLevelFromString("DEBUG")
	fmt.Println(level == logger.DEBUG) // true

	level = logger.LogLevelFromString("info") // Case insensitive
	fmt.Println(level == logger.INFO)         // true

	level = logger.LogLevelFromString("UNKNOWN") // Invalid level defaults to INFO
	fmt.Println(level == logger.INFO)            // true

	// Output:
	// true
	// true
	// true
}
