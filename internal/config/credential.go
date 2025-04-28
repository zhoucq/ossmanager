package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ossmanager/internal/logger"
)

// 常见错误定义
var (
	ErrCredentialNotFound = errors.New("凭证未找到")
	ErrInvalidMasterKey   = errors.New("无效的主密钥")
	ErrEncryptionFailed   = errors.New("加密失败")
	ErrDecryptionFailed   = errors.New("解密失败")
)

// OSSCredential 存储OSS访问凭证
type OSSCredential struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token,omitempty"` // 可选，用于STS临时凭证
}

// CredentialManager 定义凭证管理接口
type CredentialManager interface {
	// StoreCredential 存储凭证
	StoreCredential(key string, credential OSSCredential) error
	// GetCredential 获取凭证
	GetCredential(key string) (OSSCredential, error)
	// DeleteCredential 删除凭证
	DeleteCredential(key string) error
	// ListCredentialKeys 列出所有凭证标识
	ListCredentialKeys() ([]string, error)
	// IsEmpty 检查凭证存储是否为空
	IsEmpty() (bool, error)
}

// 存储在文件中的凭证数据
type credentialStore struct {
	Credentials map[string][]byte `json:"credentials"` // 加密后的凭证，key -> 加密数据
	Salt        []byte            `json:"salt"`        // 用于派生主密钥的盐值
}

// FileCredentialManager 实现基于文件的凭证管理
type FileCredentialManager struct {
	filePath  string           // 凭证文件路径
	masterKey []byte           // 主密钥，用于加密/解密
	store     *credentialStore // 凭证存储
	mutex     sync.RWMutex     // 保护并发访问
	loaded    bool             // 是否已加载
}

// NewFileCredentialManager 创建基于文件的凭证管理器
func NewFileCredentialManager(filePath string, masterPassword string) (*FileCredentialManager, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("创建凭证目录失败: %w", err)
	}

	manager := &FileCredentialManager{
		filePath: filePath,
		store:    nil,
		loaded:   false,
	}

	// 派生主密钥
	if err := manager.deriveMasterKey(masterPassword); err != nil {
		return nil, err
	}

	// 加载或初始化存储
	if err := manager.loadOrInitStore(); err != nil {
		return nil, err
	}

	return manager, nil
}

// deriveMasterKey 从密码派生主密钥
func (m *FileCredentialManager) deriveMasterKey(password string) error {
	// 如果已经加载，先尝试从文件获取盐值
	if fileExists(m.filePath) {
		store := &credentialStore{}
		if data, err := os.ReadFile(m.filePath); err == nil {
			if err = json.Unmarshal(data, store); err == nil && len(store.Salt) > 0 {
				// 使用存储的盐派生密钥
				m.masterKey = deriveKey(password, store.Salt)
				return nil
			}
		}
		// 文件读取错误，继续使用新盐值
	}

	// 生成新的盐值
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("生成盐值失败: %w", err)
	}

	// 派生密钥
	m.masterKey = deriveKey(password, salt)

	// 创建新存储并保存盐值
	m.store = &credentialStore{
		Credentials: make(map[string][]byte),
		Salt:        salt,
	}

	return nil
}

// deriveKey 使用PBKDF2或简单的SHA256派生密钥
// 注意：在实际产品中，应使用PBKDF2这样的标准密钥派生函数
func deriveKey(password string, salt []byte) []byte {
	// 简化实现：仅使用SHA256，实际应使用PBKDF2
	h := sha256.New()
	h.Write([]byte(password))
	h.Write(salt)
	return h.Sum(nil)
}

// loadOrInitStore 加载或初始化凭证存储
func (m *FileCredentialManager) loadOrInitStore() error {
	m.mutex.Lock()

	// 检查文件是否存在
	fileExists := fileExists(m.filePath)

	// 需要初始化新存储的情况
	if !fileExists {
		logger.Info("凭证文件不存在，需要创建新的存储")
		// 初始化新存储
		if m.store == nil {
			logger.Info("初始化新的凭证存储")
			m.store = &credentialStore{
				Credentials: make(map[string][]byte),
				Salt:        make([]byte, 16),
			}
			// 生成随机盐值
			if _, err := io.ReadFull(rand.Reader, m.store.Salt); err != nil {
				m.mutex.Unlock()
				logger.Error("生成盐值失败: %v", err)
				return fmt.Errorf("生成盐值失败: %w", err)
			}
			logger.Info("已生成随机盐值")
		}
		m.mutex.Unlock()

		// 保存空存储（在锁外执行IO操作）
		logger.Info("准备保存初始化的凭证存储")
		return m.saveStore()
	}

	// 从文件加载
	logger.Info("准备从文件加载凭证存储: %s", m.filePath)
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		m.mutex.Unlock()
		logger.Error("读取凭证文件失败: %v", err)
		return fmt.Errorf("读取凭证文件失败: %w", err)
	}
	logger.Info("凭证文件读取成功，大小: %d字节", len(data))

	// 解析JSON
	store := &credentialStore{}
	if err := json.Unmarshal(data, store); err != nil {
		m.mutex.Unlock()
		logger.Error("解析凭证文件失败: %v", err)
		return fmt.Errorf("解析凭证文件失败: %w", err)
	}
	logger.Info("凭证文件解析成功")

	m.store = store
	m.loaded = true
	m.mutex.Unlock()

	logger.Info("已从 %s 加载凭证存储", m.filePath)
	return nil
}

// saveStore 保存凭证存储到文件
func (m *FileCredentialManager) saveStore() error {
	// 获取写锁
	m.mutex.Lock()

	// 检查是否已初始化
	if m.store == nil {
		m.mutex.Unlock()
		logger.Error("凭证存储未初始化")
		return errors.New("凭证存储未初始化")
	}

	// 创建一个存储的副本用于序列化（在锁内复制数据）
	storeToSave := &credentialStore{
		Credentials: make(map[string][]byte, len(m.store.Credentials)),
		Salt:        make([]byte, len(m.store.Salt)),
	}

	// 复制盐值
	copy(storeToSave.Salt, m.store.Salt)

	// 复制凭证数据
	for k, v := range m.store.Credentials {
		storeToSave.Credentials[k] = make([]byte, len(v))
		copy(storeToSave.Credentials[k], v)
	}

	// 释放锁，剩余的IO操作在锁外执行
	m.mutex.Unlock()

	logger.Info("开始保存凭证存储到文件: %s", m.filePath)

	// 确保目录存在
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		logger.Error("创建凭证目录失败: %v", err)
		return fmt.Errorf("创建凭证目录失败: %w", err)
	}
	logger.Info("确保目录存在: %s", dir)

	// 将存储序列化为JSON
	logger.Info("正在序列化凭证存储...")
	data, err := json.MarshalIndent(storeToSave, "", "  ")
	if err != nil {
		logger.Error("序列化凭证存储失败: %v", err)
		return fmt.Errorf("序列化凭证存储失败: %w", err)
	}
	logger.Info("凭证存储序列化完成，数据长度: %d", len(data))

	// 写入临时文件，确保适当的权限
	logger.Info("正在写入临时文件...")
	tempFile := m.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		logger.Error("写入临时凭证文件失败: %v", err)
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("写入凭证文件失败: %w", err)
	}
	logger.Info("临时文件写入成功: %s", tempFile)

	// 重命名临时文件，确保原子性
	logger.Info("正在重命名临时文件至目标文件...")
	if err := os.Rename(tempFile, m.filePath); err != nil {
		logger.Error("重命名凭证文件失败: %v", err)
		os.Remove(tempFile) // 清理临时文件
		return fmt.Errorf("更新凭证文件失败: %w", err)
	}
	logger.Info("临时文件重命名成功")

	// 确保适当的权限
	if err := os.Chmod(m.filePath, 0600); err != nil {
		logger.Warn("设置凭证文件权限失败: %v", err)
	} else {
		logger.Info("凭证文件权限设置成功")
	}

	logger.Info("凭证存储已成功保存到: %s", m.filePath)
	return nil
}

// StoreCredential 存储加密的凭证
func (m *FileCredentialManager) StoreCredential(key string, credential OSSCredential) error {
	logger.Info("开始存储凭证 [%s]...", key)

	// 检查是否需要加载
	needLoad := false

	m.mutex.RLock()
	if !m.loaded {
		needLoad = true
	}
	m.mutex.RUnlock()

	// 需要先加载
	if needLoad {
		logger.Info("凭证存储未加载，正在加载...")
		if err := m.loadOrInitStore(); err != nil {
			logger.Error("加载凭证存储失败: %v", err)
			return err
		}
		logger.Info("凭证存储加载完成")
	}

	// 获取锁进行后续操作
	m.mutex.Lock()

	// 序列化凭证
	logger.Info("正在序列化凭证 [%s]...", key)
	plainData, err := json.Marshal(credential)
	if err != nil {
		m.mutex.Unlock()
		logger.Error("序列化凭证 [%s] 失败: %v", key, err)
		return fmt.Errorf("序列化凭证失败: %w", err)
	}
	logger.Info("凭证 [%s] 序列化完成，数据长度: %d", key, len(plainData))

	// 加密凭证
	logger.Info("正在加密凭证 [%s]...", key)
	encryptedData, err := m.encrypt(plainData)
	if err != nil {
		m.mutex.Unlock()
		logger.Error("加密凭证 [%s] 失败: %v", key, err)
		return fmt.Errorf("加密凭证失败: %w", err)
	}
	logger.Info("凭证 [%s] 加密完成，加密数据长度: %d", key, len(encryptedData))

	// 存储加密后的凭证
	m.store.Credentials[key] = encryptedData
	logger.Info("加密凭证 [%s] 已添加到内存存储", key)

	// 释放锁
	m.mutex.Unlock()

	// 保存到文件
	logger.Info("正在保存凭证存储到文件...")
	err = m.saveStore()
	if err != nil {
		logger.Error("保存凭证存储失败: %v", err)
		return err
	}
	logger.Info("凭证存储已保存到文件")

	logger.Info("已成功存储凭证 [%s]", key)
	return nil
}

// GetCredential 获取并解密凭证
func (m *FileCredentialManager) GetCredential(key string) (OSSCredential, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	logger.Debug("开始获取凭证: %s", key)

	if !m.loaded {
		logger.Debug("凭证存储未加载，需要先加载存储")
		// 这里需要释放读锁，然后获取写锁
		m.mutex.RUnlock()

		// 获取写锁加载存储
		m.mutex.Lock()
		logger.Debug("获取写锁来加载凭证存储")
		if !m.loaded { // 再次检查，因为可能在我们等待锁期间已被其他线程加载
			if err := m.loadOrInitStore(); err != nil {
				logger.Error("加载凭证存储失败: %v", err)
				m.mutex.Unlock()
				return OSSCredential{}, err
			}
			logger.Debug("凭证存储加载完成")
		} else {
			logger.Debug("凭证存储已被其他线程加载")
		}
		m.mutex.Unlock()

		// 重新获取读锁
		logger.Debug("重新获取读锁")
		m.mutex.RLock()
	}

	// 查找加密的凭证
	encryptedData, exists := m.store.Credentials[key]
	if !exists {
		logger.Debug("未找到凭证: %s", key)
		return OSSCredential{}, ErrCredentialNotFound
	}
	logger.Debug("找到加密的凭证: %s", key)

	// 解密凭证
	logger.Debug("开始解密凭证: %s", key)
	plainData, err := m.decrypt(encryptedData)
	if err != nil {
		logger.Error("解密凭证 [%s] 失败: %v", key, err)
		return OSSCredential{}, fmt.Errorf("解密凭证失败: %w", err)
	}
	logger.Debug("凭证 [%s] 解密成功", key)

	// 解析凭证
	var credential OSSCredential
	if err := json.Unmarshal(plainData, &credential); err != nil {
		logger.Error("解析凭证 [%s] 失败: %v", key, err)
		return OSSCredential{}, fmt.Errorf("解析凭证失败: %w", err)
	}

	logger.Debug("已获取凭证 [%s]", key)
	return credential, nil
}

// DeleteCredential 删除凭证
func (m *FileCredentialManager) DeleteCredential(key string) error {
	logger.Info("开始删除凭证 [%s]...", key)

	// 检查是否需要加载
	needLoad := false

	m.mutex.RLock()
	if !m.loaded {
		needLoad = true
	}
	m.mutex.RUnlock()

	// 需要先加载
	if needLoad {
		logger.Info("凭证存储未加载，正在加载...")
		if err := m.loadOrInitStore(); err != nil {
			logger.Error("加载凭证存储失败: %v", err)
			return err
		}
		logger.Info("凭证存储加载完成")
	}

	// 获取锁进行后续操作
	m.mutex.Lock()

	// 检查凭证是否存在
	if _, exists := m.store.Credentials[key]; !exists {
		m.mutex.Unlock()
		logger.Info("凭证 [%s] 不存在", key)
		return ErrCredentialNotFound
	}

	// 删除凭证
	delete(m.store.Credentials, key)
	logger.Info("已从内存中删除凭证 [%s]", key)

	// 释放锁
	m.mutex.Unlock()

	// 保存变更
	logger.Info("正在保存凭证存储到文件...")
	if err := m.saveStore(); err != nil {
		logger.Error("保存凭证存储失败: %v", err)
		return err
	}
	logger.Info("凭证存储已保存到文件")

	logger.Info("已成功删除凭证 [%s]", key)
	return nil
}

// ListCredentialKeys 列出所有凭证键
func (m *FileCredentialManager) ListCredentialKeys() ([]string, error) {
	logger.Debug("开始列出凭证键...")

	// 检查是否需要加载
	needLoad := false

	m.mutex.RLock()
	if !m.loaded {
		needLoad = true
	}
	m.mutex.RUnlock()

	// 需要先加载
	if needLoad {
		logger.Debug("凭证存储未加载，正在加载...")
		if err := m.loadOrInitStore(); err != nil {
			logger.Error("加载凭证存储失败: %v", err)
			return nil, err
		}
		logger.Debug("凭证存储加载完成")
	}

	// 获取读锁进行后续操作
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys := make([]string, 0, len(m.store.Credentials))
	for key := range m.store.Credentials {
		keys = append(keys, key)
	}

	logger.Debug("已列出 %d 个凭证键", len(keys))
	return keys, nil
}

// encrypt 使用AES-256-GCM加密数据
func (m *FileCredentialManager) encrypt(plainData []byte) ([]byte, error) {
	logger.Debug("开始加密数据，数据长度: %d", len(plainData))

	if len(m.masterKey) == 0 {
		logger.Error("主密钥无效，长度为0")
		return nil, ErrInvalidMasterKey
	}
	logger.Debug("主密钥有效，长度: %d", len(m.masterKey))

	// 创建AES-256加密器
	logger.Debug("正在创建AES-256加密器...")
	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		logger.Error("创建AES-256加密器失败: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}
	logger.Debug("AES-256加密器创建成功")

	// 创建GCM模式
	logger.Debug("正在创建GCM模式...")
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("创建GCM模式失败: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}
	logger.Debug("GCM模式创建成功")

	// 创建随机Nonce
	logger.Debug("正在生成随机Nonce...")
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		logger.Error("生成随机nonce失败: %v", err)
		return nil, fmt.Errorf("%w: 生成随机nonce失败: %v", ErrEncryptionFailed, err)
	}
	logger.Debug("随机Nonce生成成功，长度: %d", len(nonce))

	// 加密数据
	logger.Debug("正在加密数据...")
	ciphertext := gcm.Seal(nil, nonce, plainData, nil)
	logger.Debug("数据加密成功，密文长度: %d", len(ciphertext))

	// 将nonce添加到加密数据前
	result := append(nonce, ciphertext...)
	logger.Debug("加密完成，总数据长度: %d", len(result))

	return result, nil
}

// decrypt 使用AES-256-GCM解密数据
func (m *FileCredentialManager) decrypt(encryptedData []byte) ([]byte, error) {
	logger.Debug("开始解密数据，加密数据长度: %d", len(encryptedData))

	if len(m.masterKey) == 0 {
		logger.Error("主密钥无效，长度为0")
		return nil, ErrInvalidMasterKey
	}
	logger.Debug("主密钥有效，长度: %d", len(m.masterKey))

	// 创建AES-256解密器
	logger.Debug("正在创建AES-256解密器...")
	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		logger.Error("创建AES-256解密器失败: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	logger.Debug("AES-256解密器创建成功")

	// 创建GCM模式
	logger.Debug("正在创建GCM模式...")
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logger.Error("创建GCM模式失败: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	logger.Debug("GCM模式创建成功")

	// 数据长度检查
	nonceSize := gcm.NonceSize()
	logger.Debug("Nonce大小: %d", nonceSize)
	if len(encryptedData) < nonceSize {
		logger.Error("加密数据长度无效: %d，小于Nonce大小: %d", len(encryptedData), nonceSize)
		return nil, fmt.Errorf("%w: 加密数据长度无效", ErrDecryptionFailed)
	}

	// 分离nonce和密文
	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]
	logger.Debug("Nonce长度: %d, 密文长度: %d", len(nonce), len(ciphertext))

	// 解密数据
	logger.Debug("正在解密数据...")
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		logger.Error("解密数据失败: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	logger.Debug("数据解密成功，明文长度: %d", len(plaintext))

	return plaintext, nil
}

// ChangeMasterPassword 更改主密码
func (m *FileCredentialManager) ChangeMasterPassword(newPassword string) error {
	logger.Info("开始执行ChangeMasterPassword...")

	// 检查是否需要加载
	needLoad := false

	m.mutex.RLock()
	if !m.loaded {
		needLoad = true
	}
	m.mutex.RUnlock()

	// 需要先加载
	if needLoad {
		logger.Info("凭证存储未加载，正在加载...")
		if err := m.loadOrInitStore(); err != nil {
			logger.Error("加载凭证存储失败: %v", err)
			return err
		}
		logger.Info("凭证存储加载完成")
	}

	// 获取写锁
	m.mutex.Lock()

	// 备份当前凭证
	logger.Info("开始备份当前凭证...")
	allCredentials := make(map[string]OSSCredential)
	for key, encryptedData := range m.store.Credentials {
		logger.Info("正在备份凭证: %s", key)
		// 直接在当前锁内解密，避免递归调用GetCredential造成死锁
		plainData, err := m.decrypt(encryptedData)
		if err != nil {
			logger.Error("解密凭证 [%s] 失败: %v", key, err)
			m.mutex.Unlock()
			return fmt.Errorf("备份凭证失败: %w", err)
		}

		var credential OSSCredential
		if err := json.Unmarshal(plainData, &credential); err != nil {
			logger.Error("解析凭证 [%s] 失败: %v", key, err)
			m.mutex.Unlock()
			return fmt.Errorf("备份凭证失败: %w", err)
		}

		allCredentials[key] = credential
	}
	logger.Info("凭证备份完成，共 %d 个凭证", len(allCredentials))

	// 生成新盐值
	logger.Info("正在生成新盐值...")
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		logger.Error("生成盐值失败: %v", err)
		m.mutex.Unlock()
		return fmt.Errorf("生成盐值失败: %w", err)
	}
	logger.Info("新盐值生成完成")

	// 备份旧值
	logger.Info("正在备份旧密钥和盐值...")
	oldMasterKey := m.masterKey
	oldSalt := m.store.Salt

	// 设置新密钥和盐值
	logger.Info("正在设置新密钥和盐值...")
	m.masterKey = deriveKey(newPassword, salt)
	m.store.Salt = salt
	m.store.Credentials = make(map[string][]byte)
	logger.Info("新密钥和盐值设置完成")

	// 使用新密钥重新加密所有凭证
	logger.Info("开始使用新密钥重新加密凭证...")
	for key, credential := range allCredentials {
		logger.Info("正在重新加密凭证: %s", key)
		plainData, err := json.Marshal(credential)
		if err != nil {
			// 恢复旧密钥和盐值
			logger.Error("序列化凭证 [%s] 失败: %v", key, err)
			m.masterKey = oldMasterKey
			m.store.Salt = oldSalt
			m.mutex.Unlock()
			return fmt.Errorf("序列化凭证失败: %w", err)
		}

		encryptedData, err := m.encrypt(plainData)
		if err != nil {
			// 恢复旧密钥和盐值
			logger.Error("加密凭证 [%s] 失败: %v", key, err)
			m.masterKey = oldMasterKey
			m.store.Salt = oldSalt
			m.mutex.Unlock()
			return fmt.Errorf("加密凭证失败: %w", err)
		}

		m.store.Credentials[key] = encryptedData
		logger.Info("凭证 [%s] 重新加密完成", key)
	}
	logger.Info("所有凭证重新加密完成")

	// 释放锁
	m.mutex.Unlock()

	// 保存更新后的存储
	logger.Info("正在保存更新后的凭证存储...")
	if err := m.saveStore(); err != nil {
		// 获取锁恢复旧密钥和盐值
		m.mutex.Lock()
		m.masterKey = oldMasterKey
		m.store.Salt = oldSalt
		m.mutex.Unlock()

		logger.Error("保存凭证存储失败: %v", err)
		return err
	}
	logger.Info("凭证存储已保存")

	logger.Info("已成功更改主密码")
	return nil
}

// fileExists 检查文件是否存在
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// IsEmpty 检查凭证存储是否为空
func (m *FileCredentialManager) IsEmpty() (bool, error) {
	keys, err := m.ListCredentialKeys()
	if err != nil {
		return false, err
	}
	return len(keys) == 0, nil
}

// VerifyMasterPassword 检查凭证密钥是否正确（通过尝试解密）
func (m *FileCredentialManager) VerifyMasterPassword() error {
	logger.Debug("开始验证主密码...")

	// 检查是否需要加载
	needLoad := false

	m.mutex.RLock()
	if !m.loaded {
		needLoad = true
	}
	m.mutex.RUnlock()

	// 需要先加载
	if needLoad {
		logger.Debug("凭证存储未加载，正在加载...")
		if err := m.loadOrInitStore(); err != nil {
			logger.Error("加载凭证存储失败: %v", err)
			return err
		}
		logger.Debug("凭证存储加载完成")
	}

	// 如果存储为空，不需要验证
	isEmpty, err := m.IsEmpty()
	if err != nil {
		logger.Error("检查存储是否为空失败: %v", err)
		return err
	}
	if isEmpty {
		logger.Debug("凭证存储为空，无需验证")
		return nil
	}

	// 获取读锁
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 尝试解密任意一个凭证
	for key, encryptedData := range m.store.Credentials {
		logger.Debug("尝试解密凭证 [%s] 进行验证", key)
		_, err := m.decrypt(encryptedData)
		if err != nil {
			logger.Error("解密凭证 [%s] 失败: %v", key, err)
			return err
		}
		logger.Debug("成功解密凭证 [%s]，验证通过", key)
		return nil // 返回第一个尝试的结果
	}

	logger.Debug("没有可验证的凭证")
	return nil // 没有可验证的凭证
}

// 生成主密钥文件
func GenerateMasterKeyFile(filePath string, password string) error {
	// 生成盐值
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("生成盐值失败: %w", err)
	}

	// 派生密钥
	key := deriveKey(password, salt)

	// 生成验证数据
	validationData := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, validationData); err != nil {
		return fmt.Errorf("生成验证数据失败: %w", err)
	}

	// 加密验证数据
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("创建加密器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建GCM模式失败: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("生成随机nonce失败: %w", err)
	}

	encryptedData := gcm.Seal(nil, nonce, validationData, nil)

	// 组合数据
	data := struct {
		Salt      []byte `json:"salt"`
		Nonce     []byte `json:"nonce"`
		Encrypted []byte `json:"encrypted"`
	}{
		Salt:      salt,
		Nonce:     nonce,
		Encrypted: encryptedData,
	}

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 写入文件
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// 验证主密钥文件
func VerifyMasterKeyFile(filePath string, password string) error {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取主密钥文件失败: %w", err)
	}

	// 解析JSON
	var keyData struct {
		Salt      []byte `json:"salt"`
		Nonce     []byte `json:"nonce"`
		Encrypted []byte `json:"encrypted"`
	}

	if err := json.Unmarshal(data, &keyData); err != nil {
		return fmt.Errorf("解析主密钥文件失败: %w", err)
	}

	// 派生密钥
	key := deriveKey(password, keyData.Salt)

	// 解密验证数据
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("创建解密器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建GCM模式失败: %w", err)
	}

	// 尝试解密
	_, err = gcm.Open(nil, keyData.Nonce, keyData.Encrypted, nil)
	if err != nil {
		return ErrInvalidMasterKey
	}

	return nil
}
