package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/ossmanager/internal/logger"
	"github.com/spf13/viper"
)

// ConfigObserver is an interface for objects that want to be notified of config changes
type ConfigObserver interface {
	OnConfigChanged()
}

// ConfigChangeCallback is a function type for config change callbacks
type ConfigChangeCallback func()

// ConfigManager manages application configuration with hot-reload support
type ConfigManager struct {
	viper         *viper.Viper
	config        *Config
	configPath    string
	observers     []ConfigObserver
	callbacks     map[string][]ConfigChangeCallback
	callbackMutex sync.RWMutex
	loaded        bool
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Default search paths
	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".ossmanager"))
	}
	v.AddConfigPath(".")

	cm := &ConfigManager{
		viper:     v,
		config:    DefaultConfig(),
		observers: make([]ConfigObserver, 0),
		callbacks: make(map[string][]ConfigChangeCallback),
		loaded:    false,
	}

	// Set up config change watching
	v.OnConfigChange(func(e fsnotify.Event) {
		logger.Info("Config file changed: %s", e.Name)

		// Reload the configuration
		if err := cm.reload(); err != nil {
			logger.Error("Failed to reload config: %v", err)
			return
		}

		// Notify observers
		cm.notifyObservers()
	})
	v.WatchConfig()

	return cm
}

// Load loads the configuration from file
func (cm *ConfigManager) Load() error {
	// Set default values in viper
	cm.setDefaults()

	// Try to read config file (ignore error if file doesn't exist)
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, create one with defaults
		logger.Info("Config file not found, using defaults")

		if err := cm.Save(); err != nil {
			logger.Warn("Failed to create default config file: %v", err)
		}
	} else {
		// Store the loaded config path
		cm.configPath = cm.viper.ConfigFileUsed()
		logger.Info("Config loaded from %s", cm.configPath)
	}

	// Unmarshal config
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}

	cm.loaded = true
	return nil
}

// reload reloads the configuration from the file
func (cm *ConfigManager) reload() error {
	// Reset viper to avoid conflicts
	cm.viper.ReadInConfig()

	// Unmarshal the new config data
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}

	return nil
}

// setDefaults sets the default values in viper
func (cm *ConfigManager) setDefaults() {
	// App defaults
	cm.viper.SetDefault("app.log_level", cm.config.App.LogLevel)
	cm.viper.SetDefault("app.log_file", cm.config.App.LogFile)

	// UI defaults
	cm.viper.SetDefault("ui.theme", cm.config.UI.Theme)
	cm.viper.SetDefault("ui.key_bindings.upload", cm.config.UI.KeyBindings.Upload)
	cm.viper.SetDefault("ui.key_bindings.download", cm.config.UI.KeyBindings.Download)
	cm.viper.SetDefault("ui.key_bindings.refresh", cm.config.UI.KeyBindings.Refresh)
	cm.viper.SetDefault("ui.key_bindings.delete", cm.config.UI.KeyBindings.Delete)
	cm.viper.SetDefault("ui.key_bindings.help", cm.config.UI.KeyBindings.Help)
	cm.viper.SetDefault("ui.key_bindings.quit", cm.config.UI.KeyBindings.Quit)

	// Transfer defaults
	cm.viper.SetDefault("transfer.concurrent_tasks", cm.config.Transfer.ConcurrentTasks)
	cm.viper.SetDefault("transfer.part_size", cm.config.Transfer.PartSize)
	cm.viper.SetDefault("transfer.retry_count", cm.config.Transfer.RetryCount)
	cm.viper.SetDefault("transfer.auto_resume", cm.config.Transfer.AutoResume)
}

// Save saves the current configuration to a file
func (cm *ConfigManager) Save() error {
	// Ensure config directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ossmanager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("unable to create config directory: %w", err)
	}

	// Update Viper with current config values
	cm.updateViperFromConfig()

	// Save config
	configPath := filepath.Join(configDir, "config.yaml")
	if err := cm.viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("unable to write config: %w", err)
	}

	cm.configPath = configPath
	logger.Info("Configuration saved to %s", configPath)
	return nil
}

// updateViperFromConfig updates viper values from the current config struct
func (cm *ConfigManager) updateViperFromConfig() {
	// App settings
	cm.viper.Set("app.log_level", cm.config.App.LogLevel)
	cm.viper.Set("app.log_file", cm.config.App.LogFile)

	// UI settings
	cm.viper.Set("ui.theme", cm.config.UI.Theme)
	cm.viper.Set("ui.key_bindings.upload", cm.config.UI.KeyBindings.Upload)
	cm.viper.Set("ui.key_bindings.download", cm.config.UI.KeyBindings.Download)
	cm.viper.Set("ui.key_bindings.refresh", cm.config.UI.KeyBindings.Refresh)
	cm.viper.Set("ui.key_bindings.delete", cm.config.UI.KeyBindings.Delete)
	cm.viper.Set("ui.key_bindings.help", cm.config.UI.KeyBindings.Help)
	cm.viper.Set("ui.key_bindings.quit", cm.config.UI.KeyBindings.Quit)

	// Transfer settings
	cm.viper.Set("transfer.concurrent_tasks", cm.config.Transfer.ConcurrentTasks)
	cm.viper.Set("transfer.part_size", cm.config.Transfer.PartSize)
	cm.viper.Set("transfer.retry_count", cm.config.Transfer.RetryCount)
	cm.viper.Set("transfer.auto_resume", cm.config.Transfer.AutoResume)
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// AddObserver adds an observer that will be notified of config changes
func (cm *ConfigManager) AddObserver(observer ConfigObserver) {
	cm.observers = append(cm.observers, observer)
}

// RemoveObserver removes an observer
func (cm *ConfigManager) RemoveObserver(observer ConfigObserver) {
	for i, obs := range cm.observers {
		if obs == observer {
			cm.observers = append(cm.observers[:i], cm.observers[i+1:]...)
			break
		}
	}
}

// AddCallback registers a callback for a specific config key
func (cm *ConfigManager) AddCallback(key string, callback ConfigChangeCallback) {
	cm.callbackMutex.Lock()
	defer cm.callbackMutex.Unlock()

	if _, exists := cm.callbacks[key]; !exists {
		cm.callbacks[key] = make([]ConfigChangeCallback, 0)
	}

	cm.callbacks[key] = append(cm.callbacks[key], callback)
}

// RemoveCallback removes a callback for a specific config key
func (cm *ConfigManager) RemoveCallback(key string, callback ConfigChangeCallback) {
	cm.callbackMutex.Lock()
	defer cm.callbackMutex.Unlock()

	if callbacks, exists := cm.callbacks[key]; exists {
		for i, cb := range callbacks {
			if fmt.Sprintf("%p", cb) == fmt.Sprintf("%p", callback) {
				cm.callbacks[key] = append(callbacks[:i], callbacks[i+1:]...)
				break
			}
		}
	}
}

// notifyObservers notifies all registered observers of config changes
func (cm *ConfigManager) notifyObservers() {
	// Notify general observers
	for _, observer := range cm.observers {
		observer.OnConfigChanged()
	}

	// Call specific callbacks
	cm.callbackMutex.RLock()
	defer cm.callbackMutex.RUnlock()

	for _, callbacks := range cm.callbacks {
		for _, callback := range callbacks {
			callback()
		}
	}
}

// UpdateAndSave updates a config value and saves the config to disk
func (cm *ConfigManager) UpdateAndSave(key string, value interface{}) error {
	cm.viper.Set(key, value)

	// Update our in-memory config
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// Save to disk
	return cm.Save()
}

// IsLoaded returns whether the config has been loaded
func (cm *ConfigManager) IsLoaded() bool {
	return cm.loaded
}

// GetConfigPath returns the path of the loaded config file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}

// GetStringWithDefault gets a string value from config with a default fallback
func (cm *ConfigManager) GetStringWithDefault(key string, defaultValue string) string {
	if !cm.viper.IsSet(key) {
		return defaultValue
	}
	return cm.viper.GetString(key)
}

// GetIntWithDefault gets an int value from config with a default fallback
func (cm *ConfigManager) GetIntWithDefault(key string, defaultValue int) int {
	if !cm.viper.IsSet(key) {
		return defaultValue
	}
	return cm.viper.GetInt(key)
}

// GetBoolWithDefault gets a boolean value from config with a default fallback
func (cm *ConfigManager) GetBoolWithDefault(key string, defaultValue bool) bool {
	if !cm.viper.IsSet(key) {
		return defaultValue
	}
	return cm.viper.GetBool(key)
}
