package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileCredentialManager(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	masterPassword := "test-password-123"

	// 创建凭证管理器
	manager, err := NewFileCredentialManager(credFile, masterPassword)
	if err != nil {
		t.Fatalf("创建凭证管理器失败: %v", err)
	}

	// 测试初始状态
	isEmpty, err := manager.IsEmpty()
	if err != nil {
		t.Fatalf("检查存储是否为空失败: %v", err)
	}
	if !isEmpty {
		t.Error("期望初始存储为空")
	}

	// 测试存储和检索凭证
	credential := OSSCredential{
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-key-secret",
		SecurityToken:   "test-token",
	}

	// 存储凭证
	err = manager.StoreCredential("test-key", credential)
	if err != nil {
		t.Fatalf("存储凭证失败: %v", err)
	}

	// 获取凭证
	retrievedCred, err := manager.GetCredential("test-key")
	if err != nil {
		t.Fatalf("获取凭证失败: %v", err)
	}

	// 验证凭证内容
	if retrievedCred.AccessKeyID != credential.AccessKeyID {
		t.Errorf("AccessKeyID不匹配: 期望=%s, 实际=%s", credential.AccessKeyID, retrievedCred.AccessKeyID)
	}
	if retrievedCred.AccessKeySecret != credential.AccessKeySecret {
		t.Errorf("AccessKeySecret不匹配: 期望=%s, 实际=%s", credential.AccessKeySecret, retrievedCred.AccessKeySecret)
	}
	if retrievedCred.SecurityToken != credential.SecurityToken {
		t.Errorf("SecurityToken不匹配: 期望=%s, 实际=%s", credential.SecurityToken, retrievedCred.SecurityToken)
	}

	// 测试列出凭证
	keys, err := manager.ListCredentialKeys()
	if err != nil {
		t.Fatalf("列出凭证键失败: %v", err)
	}
	if len(keys) != 1 || keys[0] != "test-key" {
		t.Errorf("凭证键列表不匹配: 期望=[test-key], 实际=%v", keys)
	}

	// 测试删除凭证
	err = manager.DeleteCredential("test-key")
	if err != nil {
		t.Fatalf("删除凭证失败: %v", err)
	}

	// 验证删除结果
	isEmpty, err = manager.IsEmpty()
	if err != nil {
		t.Fatalf("检查存储是否为空失败: %v", err)
	}
	if !isEmpty {
		t.Error("期望删除后存储为空")
	}

	// 测试获取不存在的凭证
	_, err = manager.GetCredential("non-existent")
	if err != ErrCredentialNotFound {
		t.Errorf("期望错误为ErrCredentialNotFound，实际为: %v", err)
	}
}

func TestFileCredentialManagerPersistence(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_persistence_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	masterPassword := "test-password-456"

	// 创建并填充凭证管理器
	{
		manager, err := NewFileCredentialManager(credFile, masterPassword)
		if err != nil {
			t.Fatalf("创建凭证管理器失败: %v", err)
		}

		// 存储测试凭证
		cred1 := OSSCredential{AccessKeyID: "id1", AccessKeySecret: "secret1"}
		cred2 := OSSCredential{AccessKeyID: "id2", AccessKeySecret: "secret2"}

		if err := manager.StoreCredential("key1", cred1); err != nil {
			t.Fatalf("存储凭证1失败: %v", err)
		}
		if err := manager.StoreCredential("key2", cred2); err != nil {
			t.Fatalf("存储凭证2失败: %v", err)
		}
	}

	// 创建新的管理器实例并检查持久化
	{
		manager, err := NewFileCredentialManager(credFile, masterPassword)
		if err != nil {
			t.Fatalf("创建第二个凭证管理器失败: %v", err)
		}

		// 检查凭证数量
		keys, err := manager.ListCredentialKeys()
		if err != nil {
			t.Fatalf("列出凭证键失败: %v", err)
		}
		if len(keys) != 2 {
			t.Errorf("期望有2个凭证，实际有%d个", len(keys))
		}

		// 验证凭证内容
		cred1, err := manager.GetCredential("key1")
		if err != nil {
			t.Fatalf("获取凭证1失败: %v", err)
		}
		if cred1.AccessKeyID != "id1" || cred1.AccessKeySecret != "secret1" {
			t.Errorf("凭证1内容不匹配: %+v", cred1)
		}

		cred2, err := manager.GetCredential("key2")
		if err != nil {
			t.Fatalf("获取凭证2失败: %v", err)
		}
		if cred2.AccessKeyID != "id2" || cred2.AccessKeySecret != "secret2" {
			t.Errorf("凭证2内容不匹配: %+v", cred2)
		}
	}
}

func TestChangeMasterPassword(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_password_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Log("创建临时目录完成:", tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	initialPassword := "initial-password"
	newPassword := "new-stronger-password"

	t.Log("准备创建凭证管理器...")
	// 创建凭证管理器
	manager, err := NewFileCredentialManager(credFile, initialPassword)
	if err != nil {
		t.Fatalf("创建凭证管理器失败: %v", err)
	}
	t.Log("凭证管理器创建成功")

	// 存储测试凭证
	testCred := OSSCredential{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
	}
	t.Log("准备存储测试凭证...")
	if err := manager.StoreCredential("test-key", testCred); err != nil {
		t.Fatalf("存储凭证失败: %v", err)
	}
	t.Log("测试凭证存储成功")

	// 更改主密码
	t.Log("准备更改主密码...")
	if err := manager.ChangeMasterPassword(newPassword); err != nil {
		t.Fatalf("更改主密码失败: %v", err)
	}
	t.Log("主密码更改成功")

	// 使用新密码创建新的管理器
	t.Log("使用新密码创建管理器...")
	newManager, err := NewFileCredentialManager(credFile, newPassword)
	if err != nil {
		t.Fatalf("用新密码创建凭证管理器失败: %v", err)
	}
	t.Log("使用新密码的管理器创建成功")

	// 验证可以访问凭证
	t.Log("准备使用新密码获取凭证...")
	retrievedCred, err := newManager.GetCredential("test-key")
	if err != nil {
		t.Fatalf("用新密码获取凭证失败: %v", err)
	}
	t.Log("使用新密码获取凭证成功")
	if retrievedCred.AccessKeyID != testCred.AccessKeyID || retrievedCred.AccessKeySecret != testCred.AccessKeySecret {
		t.Errorf("凭证内容不匹配: 期望=%+v, 实际=%+v", testCred, retrievedCred)
	}

	// 使用旧密码尝试创建管理器，应该失败或无法解密
	t.Log("尝试使用旧密码创建管理器...")
	oldManager, err := NewFileCredentialManager(credFile, initialPassword)
	if err != nil {
		// 可能在创建时失败，这是可接受的
		t.Log("使用旧密码创建管理器失败，这是预期行为:", err)
		return
	}
	t.Log("使用旧密码创建管理器成功，将尝试获取凭证...")

	// 若创建成功，尝试获取凭证应该失败
	_, err = oldManager.GetCredential("test-key")
	if err == nil {
		t.Error("使用旧密码应该无法解密凭证")
	} else {
		t.Log("使用旧密码无法解密凭证，这是预期行为:", err)
	}
}

func TestIncorrectPassword(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_wrong_password_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	correctPassword := "correct-password"
	wrongPassword := "wrong-password"

	// 使用正确密码创建凭证管理器
	manager, err := NewFileCredentialManager(credFile, correctPassword)
	if err != nil {
		t.Fatalf("创建凭证管理器失败: %v", err)
	}

	// 存储测试凭证
	testCred := OSSCredential{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
	}
	if err := manager.StoreCredential("test-key", testCred); err != nil {
		t.Fatalf("存储凭证失败: %v", err)
	}

	// 使用错误密码创建新的管理器
	wrongManager, err := NewFileCredentialManager(credFile, wrongPassword)
	if err != nil {
		// 可能在创建时失败，这是可接受的
		return
	}

	// 使用错误密码尝试获取凭证应该失败
	_, err = wrongManager.GetCredential("test-key")
	if err == nil {
		t.Error("使用错误密码应该无法解密凭证")
	}
}

func TestFilePermissions(t *testing.T) {
	// 跳过Windows测试，因为Windows权限模型不同
	if os.Getenv("GOOS") == "windows" {
		t.Skip("跳过Windows权限测试")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_permissions_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	password := "test-password"

	// 创建凭证管理器
	_, err = NewFileCredentialManager(credFile, password)
	if err != nil {
		t.Fatalf("创建凭证管理器失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(credFile); os.IsNotExist(err) {
		t.Fatalf("凭证文件未创建: %v", err)
	}

	// 检查文件权限
	info, err := os.Stat(credFile)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	// 检查文件权限是否为600（仅所有者可读写）
	if info.Mode().Perm() != 0600 {
		t.Errorf("文件权限不正确: 期望=0600, 实际=%o", info.Mode().Perm())
	}
}

func TestMasterKeyVerification(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "master_key_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 测试主密钥文件路径
	keyFile := filepath.Join(tempDir, "master.key")
	correctPassword := "master-password"
	wrongPassword := "incorrect-password"

	// 生成主密钥文件
	err = GenerateMasterKeyFile(keyFile, correctPassword)
	if err != nil {
		t.Fatalf("生成主密钥文件失败: %v", err)
	}

	// 正确密码验证
	err = VerifyMasterKeyFile(keyFile, correctPassword)
	if err != nil {
		t.Errorf("正确密码验证失败: %v", err)
	}

	// 错误密码验证
	err = VerifyMasterKeyFile(keyFile, wrongPassword)
	if err != ErrInvalidMasterKey {
		t.Errorf("期望错误密码验证返回ErrInvalidMasterKey，实际返回: %v", err)
	}
}

func TestStoreCredentialOnly(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "credential_store_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Log("创建临时目录完成:", tempDir)

	// 测试凭证文件路径
	credFile := filepath.Join(tempDir, "credentials.json")
	password := "test-password"

	t.Log("准备创建凭证管理器...")
	// 创建凭证管理器
	manager, err := NewFileCredentialManager(credFile, password)
	if err != nil {
		t.Fatalf("创建凭证管理器失败: %v", err)
	}
	t.Log("凭证管理器创建成功")

	// 存储测试凭证
	testCred := OSSCredential{
		AccessKeyID:     "test-id",
		AccessKeySecret: "test-secret",
	}
	t.Log("准备存储测试凭证...")

	// 直接存储凭证，不使用goroutine和超时机制
	if err := manager.StoreCredential("test-key", testCred); err != nil {
		t.Fatalf("存储凭证失败: %v", err)
	}
	t.Log("测试凭证存储成功")
}
