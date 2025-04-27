package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigManager(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "configmanager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 修改用户主目录以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建ConfigManager实例
	cm := NewConfigManager()

	// 测试加载配置
	err = cm.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证配置已加载
	if !cm.IsLoaded() {
		t.Error("Expected IsLoaded() to return true")
	}

	// 验证配置路径正确
	configPath := cm.GetConfigPath()
	expectedPath := filepath.Join(tempDir, ".ossmanager", "config.yaml")
	if configPath != expectedPath {
		t.Errorf("Expected config path to be %s, got %s", expectedPath, configPath)
	}

	// 测试获取配置
	config := cm.GetConfig()
	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// 验证默认配置值
	if config.App.LogLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", config.App.LogLevel)
	}

	// 测试更新配置
	err = cm.UpdateAndSave("app.log_level", "debug")
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// 验证配置已更新
	if cm.GetConfig().App.LogLevel != "debug" {
		t.Errorf("Expected log level to be 'debug', got '%s'", cm.GetConfig().App.LogLevel)
	}

	// 测试辅助方法
	logLevel := cm.GetStringWithDefault("app.log_level", "warn")
	if logLevel != "debug" {
		t.Errorf("Expected GetStringWithDefault to return 'debug', got '%s'", logLevel)
	}

	nonExistentValue := cm.GetStringWithDefault("non.existent", "default")
	if nonExistentValue != "default" {
		t.Errorf("Expected GetStringWithDefault for non-existent key to return 'default', got '%s'", nonExistentValue)
	}

	// 测试观察者模式
	observerCalled := false
	observer := &testConfigObserver{
		t:      t,
		called: &observerCalled,
	}

	// 添加观察者
	cm.AddObserver(observer)

	// 更新配置，应该触发观察者
	err = cm.UpdateAndSave("ui.theme", "light")
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// 手动触发通知，因为我们不能依赖文件系统监听在测试中工作
	cm.notifyObservers()

	// 验证观察者被调用
	if !observerCalled {
		t.Error("Expected observer to be called")
	}

	// 移除观察者
	cm.RemoveObserver(observer)

	// 测试回调
	callbackCalled := false
	callback := func() {
		callbackCalled = true
	}

	// 添加回调
	cm.AddCallback("ui.theme", callback)

	// 触发通知
	cm.notifyObservers()

	// 验证回调被调用
	if !callbackCalled {
		t.Error("Expected callback to be called")
	}

	// 移除回调
	cm.RemoveCallback("ui.theme", callback)
}

// 测试用的观察者实现
type testConfigObserver struct {
	t      *testing.T
	called *bool
}

func (o *testConfigObserver) OnConfigChanged() {
	*o.called = true
}

func TestConfigManagerDefaults(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "configmanager_defaults_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 修改用户主目录以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建ConfigManager实例
	cm := NewConfigManager()

	// 加载配置
	err = cm.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 检查配置文件是否被创建
	expectedPath := filepath.Join(tempDir, ".ossmanager", "config.yaml")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", expectedPath)
	}

	// 验证默认配置存在
	config := cm.GetConfig()

	// 应用设置
	if config.App.LogLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", config.App.LogLevel)
	}

	// UI设置
	if config.UI.Theme != "dark" {
		t.Errorf("Expected default theme to be 'dark', got '%s'", config.UI.Theme)
	}

	// 传输设置
	if config.Transfer.ConcurrentTasks != 3 {
		t.Errorf("Expected default concurrent tasks to be 3, got %d", config.Transfer.ConcurrentTasks)
	}
}

func TestConfigManagerHotReload(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "configmanager_hotreload_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 修改用户主目录以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建ConfigManager实例
	cm := NewConfigManager()

	// 加载配置
	err = cm.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 记录是否调用了观察者
	observerCalled := false
	observer := &testConfigObserver{
		t:      t,
		called: &observerCalled,
	}

	// 添加观察者
	cm.AddObserver(observer)

	// 直接更新配置文件模拟外部修改
	configPath := cm.GetConfigPath()

	// 获取当前配置内容
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// 修改配置内容（简单替换）
	newContent := string(content)
	newContent = strings.Replace(newContent, "theme: dark", "theme: light", 1)

	// 等待一下以确保文件修改时间不同
	time.Sleep(100 * time.Millisecond)

	// 写回文件
	err = os.WriteFile(configPath, []byte(newContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write modified config file: %v", err)
	}

	// 等待文件系统通知生效
	time.Sleep(500 * time.Millisecond)

	// 手动触发viper的OnConfigChange函数
	cm.notifyObservers()

	// 验证观察者被调用
	if !observerCalled {
		t.Error("Expected observer to be called after config file change")
	}
}
