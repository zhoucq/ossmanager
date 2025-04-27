package ossmanager

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ossmanager/internal/config"
	"github.com/ossmanager/internal/logger"
	ossClient "github.com/ossmanager/internal/oss"
	"github.com/ossmanager/internal/ui"
)

// Version information
const (
	AppName    = "OSS Manager"
	AppVersion = "0.1.0"
)

// Run starts the OSS Manager application
func Run() error {
	// Log application version
	fmt.Printf("%s v%s\n", AppName, AppVersion)

	// Log dependency versions for debugging
	logDependencyVersions()

	// Initialize logger
	if err := initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize configuration (will be fully implemented in Task 3)
	if err := initConfig(); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Initialize OSS client (will be fully implemented in Task 7)
	if err := initOSS(); err != nil {
		return fmt.Errorf("failed to initialize OSS client: %w", err)
	}

	// Initialize UI (will be fully implemented in Task 17)
	if err := initUI(); err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}

	fmt.Println("OSS Manager initialized successfully")
	return nil
}

// logDependencyVersions logs the versions of dependencies
func logDependencyVersions() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: Failed to create logs directory: %v", err)
		return
	}

	// Create dependency log file
	logFile, err := os.Create(fmt.Sprintf("logs/dependencies_%s.log", time.Now().Format("20060102_150405")))
	if err != nil {
		log.Printf("Warning: Failed to create dependency log file: %v", err)
		return
	}
	defer logFile.Close()

	// Set up logger
	logger := log.New(logFile, "", log.LstdFlags)

	// Log system information
	logger.Printf("System Information:")
	logger.Printf("  OS: %s", runtime.GOOS)
	logger.Printf("  Architecture: %s", runtime.GOARCH)
	logger.Printf("  Go Version: %s", runtime.Version())
	logger.Printf("  NumCPU: %d", runtime.NumCPU())

	// Log application information
	logger.Printf("Application Information:")
	logger.Printf("  Name: %s", AppName)
	logger.Printf("  Version: %s", AppVersion)

	// Log dependency versions
	logger.Printf("Dependencies:")
	logger.Printf("  Bubble Tea: %s", "latest")     // Bubbletea doesn't expose version
	logger.Printf("  Viper: %s", "latest")          // Viper doesn't expose version
	logger.Printf("  Aliyun OSS SDK: %s", "latest") // OSS SDK doesn't expose version in a simple way

	fmt.Println("Dependency information logged to", logFile.Name())
}

// initConfig initializes the configuration system
func initConfig() error {
	// 创建配置管理器实例
	configManager := config.NewConfigManager()

	// 加载配置
	if err := configManager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// 获取配置文件路径
	configPath := configManager.GetConfigPath()
	if configPath != "" {
		logger.Info("Configuration loaded from %s", configPath)
	} else {
		logger.Info("Using default configuration (no config file found)")
	}

	// 注册配置变更回调，记录日志
	configManager.AddCallback("app.log_level", func() {
		newLevel := configManager.GetConfig().App.LogLevel
		logger.Info("Log level changed to %s", newLevel)
		// 设置新的日志级别
		logger.SetLevel(logger.LogLevelFromString(newLevel))
	})

	// 其他配置变更回调也可以在这里添加

	// 导出配置管理器，使其他模块可以访问它
	config.SetGlobalConfigManager(configManager)

	logger.Info("Configuration system initialized")
	return nil
}

// initLogger initializes the logging system
func initLogger() error {
	// Create a custom log configuration
	logConfig := &logger.LogConfig{
		Level:           logger.INFO,
		Format:          logger.TEXT,
		LogDir:          "logs",
		LogFile:         "ossmanager.log",
		MaxSize:         10,   // 10 MB
		MaxBackups:      5,    // 5 files
		MaxAge:          30,   // 30 days
		Compress:        true, // compress old log files
		ToConsole:       true, // also log to console
		TimestampFormat: "2006-01-02 15:04:05.000",
	}

	// Try to get log level from environment
	logLevelEnv := os.Getenv("OSSMANAGER_LOG_LEVEL")
	if logLevelEnv != "" {
		logConfig.Level = logger.LogLevelFromString(logLevelEnv)
	}

	// Initialize logger with the configuration
	err := logger.Initialize(logConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log system information
	sysLogger := logger.WithContext(map[string]interface{}{
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"go_version":  runtime.Version(),
		"num_cpu":     runtime.NumCPU(),
		"app_version": AppVersion,
	})

	// Log initialization success
	sysLogger.Info("OSS Manager %s starting up", AppVersion)
	logger.Info("Logger initialized successfully with level %s", logConfig.Level.String())
	logger.Debug("Debug logging is enabled")

	// Log different levels for testing
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	return nil
}

// initOSS initializes the OSS client
// This is a placeholder that will be fully implemented in Task 7
func initOSS() error {
	// Create a dummy OSS client config
	ossConfig := &ossClient.Config{
		Endpoint:        "oss-cn-beijing.aliyuncs.com",
		AccessKeyID:     "dummy-access-key-id",
		AccessKeySecret: "dummy-access-key-secret",
	}

	// Create OSS client (this will fail in real usage, but it's just a placeholder)
	_, err := ossClient.NewClient(ossConfig)
	if err != nil {
		// Just log the error for now, don't return it
		logger.Warn("OSS client initialization failed (expected in placeholder): %v", err)
	}

	logger.Info("OSS client placeholder initialized")
	return nil
}

// initUI initializes the terminal UI
// This is a placeholder that will be fully implemented in Task 17
func initUI() error {
	// Create a simple UI model to demonstrate the imports are working
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	fmt.Println(style.Render("Terminal UI initialized (placeholder)"))

	// Just to use the tea import
	_ = tea.Quit

	// Just to use the list import
	_ = list.NewDefaultDelegate()

	// Just to use the oss import
	_, _ = oss.New("endpoint", "accessKeyID", "accessKeySecret")

	// Just to use the ui import
	_ = ui.NewModel

	logger.Info("UI components initialized")
	return nil
}
