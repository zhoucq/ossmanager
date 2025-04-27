package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ossmanager/cmd/ossmanager"
)

func main() {
	// Record initialization time for logging
	startTime := time.Now()

	// Initialize application
	fmt.Println("Initializing OSS Manager...")

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create initialization log file
	logFile, err := os.Create(fmt.Sprintf("logs/init_%s.log", time.Now().Format("20060102_150405")))
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Set up basic logger
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("OSS Manager initialization started at %s", startTime.Format(time.RFC3339))

	// Log Go version
	logger.Printf("Go Version: %s", os.Getenv("GOVERSION"))

	// Run the application
	if err := ossmanager.Run(); err != nil {
		logger.Printf("Error running application: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Log initialization completion
	logger.Printf("OSS Manager initialization completed in %v", time.Since(startTime))

	// Ensure all application logs are flushed
	// Note: This is a placeholder for future implementation
	// The logger.Shutdown() function will be called here once it's implemented
}
