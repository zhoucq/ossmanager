package config

import (
	"sync"
)

var (
	// globalConfigManager 存储全局配置管理器实例
	globalConfigManager *ConfigManager
	// globalConfigMutex 保护对全局配置管理器的访问
	globalConfigMutex sync.RWMutex
)

// SetGlobalConfigManager 设置全局配置管理器实例
func SetGlobalConfigManager(cm *ConfigManager) {
	globalConfigMutex.Lock()
	defer globalConfigMutex.Unlock()
	globalConfigManager = cm
}

// GetGlobalConfigManager 获取全局配置管理器实例
func GetGlobalConfigManager() *ConfigManager {
	globalConfigMutex.RLock()
	defer globalConfigMutex.RUnlock()
	return globalConfigManager
}
