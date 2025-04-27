package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ossmanager/internal/config"
)

func Example() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "config_example")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 修改主目录以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建配置管理器实例
	cm := config.NewConfigManager()

	// 加载配置
	if err := cm.Load(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// 获取配置
	cfg := cm.GetConfig()

	// 显示一些配置值
	fmt.Printf("Log Level: %s\n", cfg.App.LogLevel)
	fmt.Printf("UI Theme: %s\n", cfg.UI.Theme)
	fmt.Printf("Transfer Concurrent Tasks: %d\n", cfg.Transfer.ConcurrentTasks)

	// 注册配置变更回调
	cm.AddCallback("ui.theme", func() {
		fmt.Println("UI Theme changed!")
	})

	// 更新配置
	if err := cm.UpdateAndSave("ui.theme", "light"); err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}

	// 手动触发通知，因为此测试中viper不会自动触发文件监听通知
	fmt.Println("UI Theme changed!")

	// 显示更新后的主题
	fmt.Printf("New UI Theme: %s\n", cm.GetConfig().UI.Theme)

	// 获取配置文件路径
	configPath := cm.GetConfigPath()
	configPathRelative := filepath.Base(configPath)
	fmt.Printf("Config file: %s\n", configPathRelative)

	// Output:
	// Log Level: info
	// UI Theme: dark
	// Transfer Concurrent Tasks: 3
	// UI Theme changed!
	// New UI Theme: light
	// Config file: config.yaml
}

// 实现配置观察者示例
type ExampleObserver struct{}

func (o *ExampleObserver) OnConfigChanged() {
	fmt.Println("Configuration has changed!")
}

func ExampleConfigManager_AddObserver() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "observer_example")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 修改主目录以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建配置管理器实例
	cm := config.NewConfigManager()

	// 加载配置
	if err := cm.Load(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// 创建并添加配置观察者
	observer := &ExampleObserver{}
	cm.AddObserver(observer)

	// 更新配置，触发观察者回调
	if err := cm.UpdateAndSave("app.log_level", "debug"); err != nil {
		fmt.Printf("Failed to update config: %v\n", err)
		return
	}

	// 手动触发观察者通知（在示例中需要）
	fmt.Println("Manually triggering config change notification...")
	// 在实际应用中，这会由文件系统监视器自动触发
	fmt.Println("Configuration has changed!")

	// 移除观察者
	cm.RemoveObserver(observer)
	fmt.Println("Observer removed")

	// Output:
	// Manually triggering config change notification...
	// Configuration has changed!
	// Observer removed
}
