package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ossmanager/internal/config"
)

func Example_credentialManager() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "credential_example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 凭证文件路径
	credPath := filepath.Join(tempDir, "credentials.json")
	masterPassword := "example-master-password"

	// 创建凭证管理器
	credManager, err := config.NewFileCredentialManager(credPath, masterPassword)
	if err != nil {
		fmt.Printf("创建凭证管理器失败: %v\n", err)
		return
	}

	// 检查是否为空
	isEmpty, _ := credManager.IsEmpty()
	fmt.Printf("凭证存储是否为空: %v\n", isEmpty)

	// 创建凭证
	aliyunCred := config.OSSCredential{
		AccessKeyID:     "LTAI4FjmgXXXXXXXXXXXXXXX",
		AccessKeySecret: "PkT44XXXXXXXXXXXXXXXXXX",
	}

	// 存储凭证
	err = credManager.StoreCredential("aliyun-prod", aliyunCred)
	if err != nil {
		fmt.Printf("存储凭证失败: %v\n", err)
		return
	}

	// 列出所有凭证键
	keys, _ := credManager.ListCredentialKeys()
	fmt.Printf("已存储的凭证: %v\n", keys)

	// 获取凭证
	retrievedCred, err := credManager.GetCredential("aliyun-prod")
	if err != nil {
		fmt.Printf("获取凭证失败: %v\n", err)
		return
	}

	// 显示凭证信息（实际应用中应隐藏密钥）
	fmt.Printf("AccessKeyID: %s\n", retrievedCred.AccessKeyID)
	fmt.Printf("AccessKeySecret: %s (已隐藏部分内容)\n", retrievedCred.AccessKeySecret[:4]+"****")

	// 删除凭证
	err = credManager.DeleteCredential("aliyun-prod")
	if err != nil {
		fmt.Printf("删除凭证失败: %v\n", err)
		return
	}

	// 确认删除
	isEmpty, _ = credManager.IsEmpty()
	fmt.Printf("删除后凭证存储是否为空: %v\n", isEmpty)

	// Output:
	// 凭证存储是否为空: true
	// 已存储的凭证: [aliyun-prod]
	// AccessKeyID: LTAI4FjmgXXXXXXXXXXXXXXX
	// AccessKeySecret: PkT4**** (已隐藏部分内容)
	// 删除后凭证存储是否为空: true
}

func Example_changeMasterPassword() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "change_password_example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 凭证文件路径
	credPath := filepath.Join(tempDir, "credentials.json")
	oldPassword := "old-password"
	newPassword := "new-stronger-password"

	// 创建凭证管理器
	credManager, err := config.NewFileCredentialManager(credPath, oldPassword)
	if err != nil {
		fmt.Printf("创建凭证管理器失败: %v\n", err)
		return
	}

	// 存储凭证
	credential := config.OSSCredential{
		AccessKeyID:     "example-key-id",
		AccessKeySecret: "example-key-secret",
	}
	err = credManager.StoreCredential("example", credential)
	if err != nil {
		fmt.Printf("存储凭证失败: %v\n", err)
		return
	}

	// 验证凭证已存储
	_, err = credManager.GetCredential("example")
	if err != nil {
		fmt.Printf("获取凭证失败: %v\n", err)
		return
	}
	fmt.Println("使用旧密码可以访问凭证")

	// 更改主密码
	err = credManager.ChangeMasterPassword(newPassword)
	if err != nil {
		fmt.Printf("更改主密码失败: %v\n", err)
		return
	}
	fmt.Println("已更改主密码")

	// 尝试使用新密码创建新的管理器并访问凭证
	newManager, err := config.NewFileCredentialManager(credPath, newPassword)
	if err != nil {
		fmt.Printf("用新密码创建管理器失败: %v\n", err)
		return
	}

	// 使用新密码验证凭证访问
	_, err = newManager.GetCredential("example")
	if err != nil {
		fmt.Printf("用新密码获取凭证失败: %v\n", err)
		return
	}
	fmt.Println("使用新密码可以访问凭证")

	// Output:
	// 使用旧密码可以访问凭证
	// 已更改主密码
	// 使用新密码可以访问凭证
}

func Example_masterKeyVerification() {
	// 创建临时目录用于示例
	tempDir, err := os.MkdirTemp("", "master_key_example")
	if err != nil {
		fmt.Printf("创建临时目录失败: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	// 主密钥文件路径
	keyPath := filepath.Join(tempDir, "master.key")
	password := "secure-master-password"

	// 生成主密钥文件
	err = config.GenerateMasterKeyFile(keyPath, password)
	if err != nil {
		fmt.Printf("生成主密钥文件失败: %v\n", err)
		return
	}
	fmt.Println("已生成主密钥文件")

	// 使用正确密码验证
	err = config.VerifyMasterKeyFile(keyPath, password)
	if err != nil {
		fmt.Printf("主密钥验证失败: %v\n", err)
		return
	}
	fmt.Println("主密钥验证成功")

	// 使用错误密码验证
	err = config.VerifyMasterKeyFile(keyPath, "wrong-password")
	if err == config.ErrInvalidMasterKey {
		fmt.Println("错误密码验证失败，如预期")
	} else if err != nil {
		fmt.Printf("验证发生意外错误: %v\n", err)
	} else {
		fmt.Println("错误密码验证意外通过")
	}

	// Output:
	// 已生成主密钥文件
	// 主密钥验证成功
	// 错误密码验证失败，如预期
}
