package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	App struct {
		LogLevel string `mapstructure:"log_level"`
		LogFile  string `mapstructure:"log_file"`
	} `mapstructure:"app"`

	UI struct {
		Theme       string `mapstructure:"theme"`
		KeyBindings struct {
			Upload   string `mapstructure:"upload"`
			Download string `mapstructure:"download"`
			Refresh  string `mapstructure:"refresh"`
			Delete   string `mapstructure:"delete"`
			Help     string `mapstructure:"help"`
			Quit     string `mapstructure:"quit"`
		} `mapstructure:"key_bindings"`
	} `mapstructure:"ui"`

	Transfer struct {
		ConcurrentTasks int   `mapstructure:"concurrent_tasks"`
		PartSize        int64 `mapstructure:"part_size"`
		RetryCount      int   `mapstructure:"retry_count"`
		AutoResume      bool  `mapstructure:"auto_resume"`
	} `mapstructure:"transfer"`

	// Accounts will be implemented in Task 5
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	// App defaults
	cfg.App.LogLevel = "info"
	cfg.App.LogFile = "logs/ossmanager.log"

	// UI defaults
	cfg.UI.Theme = "dark"
	cfg.UI.KeyBindings.Upload = "ctrl+u"
	cfg.UI.KeyBindings.Download = "ctrl+d"
	cfg.UI.KeyBindings.Refresh = "f5"
	cfg.UI.KeyBindings.Delete = "ctrl+x"
	cfg.UI.KeyBindings.Help = "f1"
	cfg.UI.KeyBindings.Quit = "ctrl+q"

	// Transfer defaults
	cfg.Transfer.ConcurrentTasks = 3
	cfg.Transfer.PartSize = 10 * 1024 * 1024 // 10MB
	cfg.Transfer.RetryCount = 3
	cfg.Transfer.AutoResume = true

	return cfg
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	// Set up viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add config paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".ossmanager"))
	}
	v.AddConfigPath(".")

	// Load default values
	cfg := DefaultConfig()

	// Set default values in viper
	v.SetDefault("app.log_level", cfg.App.LogLevel)
	v.SetDefault("app.log_file", cfg.App.LogFile)
	v.SetDefault("ui.theme", cfg.UI.Theme)
	v.SetDefault("ui.key_bindings.upload", cfg.UI.KeyBindings.Upload)
	v.SetDefault("ui.key_bindings.download", cfg.UI.KeyBindings.Download)
	v.SetDefault("ui.key_bindings.refresh", cfg.UI.KeyBindings.Refresh)
	v.SetDefault("ui.key_bindings.delete", cfg.UI.KeyBindings.Delete)
	v.SetDefault("ui.key_bindings.help", cfg.UI.KeyBindings.Help)
	v.SetDefault("ui.key_bindings.quit", cfg.UI.KeyBindings.Quit)
	v.SetDefault("transfer.concurrent_tasks", cfg.Transfer.ConcurrentTasks)
	v.SetDefault("transfer.part_size", cfg.Transfer.PartSize)
	v.SetDefault("transfer.retry_count", cfg.Transfer.RetryCount)
	v.SetDefault("transfer.auto_resume", cfg.Transfer.AutoResume)

	// Read config file (ignore error if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, using defaults
		fmt.Println("Config file not found, using defaults")
	}

	// Unmarshal config
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return cfg, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *Config) error {
	v := viper.New()

	// Set values from config
	v.Set("app.log_level", cfg.App.LogLevel)
	v.Set("app.log_file", cfg.App.LogFile)
	v.Set("ui.theme", cfg.UI.Theme)
	v.Set("ui.key_bindings.upload", cfg.UI.KeyBindings.Upload)
	v.Set("ui.key_bindings.download", cfg.UI.KeyBindings.Download)
	v.Set("ui.key_bindings.refresh", cfg.UI.KeyBindings.Refresh)
	v.Set("ui.key_bindings.delete", cfg.UI.KeyBindings.Delete)
	v.Set("ui.key_bindings.help", cfg.UI.KeyBindings.Help)
	v.Set("ui.key_bindings.quit", cfg.UI.KeyBindings.Quit)
	v.Set("transfer.concurrent_tasks", cfg.Transfer.ConcurrentTasks)
	v.Set("transfer.part_size", cfg.Transfer.PartSize)
	v.Set("transfer.retry_count", cfg.Transfer.RetryCount)
	v.Set("transfer.auto_resume", cfg.Transfer.AutoResume)

	// Ensure config directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ossmanager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}

	// Save config
	configPath := filepath.Join(configDir, "config.yaml")
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("unable to write config: %w", err)
	}

	fmt.Println("Configuration saved to", configPath)
	return nil
}
