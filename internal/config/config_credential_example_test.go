package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ossmanager/internal/config"
)

func Example_configWithCredentials() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "config_credential_example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 修改HOME环境变量以便测试
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// 创建配置管理器
	cm := config.NewConfigManager()

	// 加载配置
	if err := cm.Load(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 初始化凭证管理器
	if err := cm.InitCredentialManager("example-password"); err != nil {
		fmt.Printf("初始化凭证管理器失败: %v\n", err)
		return
	}

	// 获取凭证管理器
	credManager := cm.GetCredentialManager()
	if credManager == nil {
		fmt.Println("凭证管理器未初始化")
		return
	}

	// 检查是否为空
	isEmpty, _ := credManager.IsEmpty()
	fmt.Printf("凭证存储是否为空: %v\n", isEmpty)

	// 创建并存储一个凭证
	aliyunCred := config.OSSCredential{
		AccessKeyID:     "sample-key-id",
		AccessKeySecret: "sample-key-secret",
	}

	// 存储凭证
	err = credManager.StoreCredential("aliyun-test", aliyunCred)
	if err != nil {
		fmt.Printf("存储凭证失败: %v\n", err)
		return
	}

	// 检查凭证是否已存储
	keys, _ := credManager.ListCredentialKeys()
	fmt.Printf("已存储的凭证: %v\n", keys)

	// 修改配置，添加账户引用凭证
	configFile := filepath.Join(tempDir, ".ossmanager", "config.yaml")
	accountsConfig := `
accounts:
  - id: "aliyun-test"
    name: "测试账户"
    credentials_key: "aliyun-test"
    default: true
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    default_bucket: "test-bucket"
`
	// 将账户配置追加到配置文件
	currentConfig, _ := os.ReadFile(configFile)
	newConfig := append(currentConfig, []byte(accountsConfig)...)
	os.WriteFile(configFile, newConfig, 0644)

	// 重新加载配置
	cm = config.NewConfigManager()
	cm.Load()
	cm.InitCredentialManager("example-password")

	// 通过配置管理器获取账户凭证
	retrievedCred, err := cm.GetAccountCredential("aliyun-test")
	if err != nil {
		fmt.Printf("获取账户凭证失败: %v\n", err)
		return
	}

	// 显示凭证信息（实际应用中应隐藏密钥）
	fmt.Printf("获取到的账户凭证 AccessKeyID: %s\n", retrievedCred.AccessKeyID)
	fmt.Printf("获取到的账户凭证 AccessKeySecret: %s (已隐藏部分内容)\n", retrievedCred.AccessKeySecret[:5]+"****")

	// Output:
	// 凭证存储是否为空: true
	// 已存储的凭证: [aliyun-test]
	// 获取到的账户凭证 AccessKeyID: sample-key-id
	// 获取到的账户凭证 AccessKeySecret: sampl**** (已隐藏部分内容)
}
