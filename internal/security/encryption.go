package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor 加密器
// 用途: 对敏感数据进行 AES-256-GCM 加密
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
// 参数:
//
//	key: 32字节的加密密钥(AES-256)
//
// 返回:
//
//	*Encryptor: 加密器实例
//	error: 错误信息
func NewEncryptor(key string) (*Encryptor, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("密钥长度必须为32字节，当前为%d字节", len(keyBytes))
	}
	return &Encryptor{key: keyBytes}, nil
}

// Encrypt 加密敏感数据
// 用途: 使用 AES-256-GCM 算法加密数据，返回 Base64 编码的密文
// 参数:
//
//	plaintext: 明文数据
//
// 返回:
//
//	string: Base64 编码的密文
//	error: 错误信息
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密数据 (nonce + ciphertext + tag)
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密敏感数据
// 用途: 解密 Base64 编码的密文
// 参数:
//
//	ciphertext: Base64 编码的密文
//
// 返回:
//
//	string: 明文数据
//	error: 错误信息
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	// 创建 GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// 提取 nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("密文长度不足")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// EncryptFields 批量加密字段
// 用途: 对结构体中的多个字段进行加密
// 参数:
//
//	fields: 字段名到值的映射
//
// 返回:
//
//	map[string]string: 加密后的字段映射
//	error: 错误信息
func (e *Encryptor) EncryptFields(fields map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range fields {
		encrypted, err := e.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("加密字段%s失败: %w", key, err)
		}
		result[key] = encrypted
	}
	return result, nil
}

// DecryptFields 批量解密字段
// 用途: 对结构体中的多个字段进行解密
// 参数:
//
//	fields: 加密字段名到密文的映射
//
// 返回:
//
//	map[string]string: 解密后的字段映射
//	error: 错误信息
func (e *Encryptor) DecryptFields(fields map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range fields {
		decrypted, err := e.Decrypt(value)
		if err != nil {
			return nil, fmt.Errorf("解密字段%s失败: %w", key, err)
		}
		result[key] = decrypted
	}
	return result, nil
}

// MaskSensitiveData 脱敏敏感数据
// 用途: 对敏感数据进行脱敏处理（用于日志记录）
// 参数:
//
//	data: 原始数据
//	dataType: 数据类型 (phone/email/idcard/bankcard)
//
// 返回:
//
//	string: 脱敏后的数据
func MaskSensitiveData(data string, dataType string) string {
	if data == "" {
		return ""
	}

	switch dataType {
	case "phone":
		// 手机号: 138****5678
		if len(data) == 11 {
			return data[:3] + "****" + data[7:]
		}
	case "email":
		// 邮箱: u***@example.com
		parts := splitEmail(data)
		if len(parts) == 2 && len(parts[0]) > 1 {
			return parts[0][:1] + "***@" + parts[1]
		}
	case "idcard":
		// 身份证: 110***********1234
		if len(data) == 18 {
			return data[:3] + "***********" + data[14:]
		}
	case "bankcard":
		// 银行卡: 6222 **** **** 1234
		if len(data) >= 16 {
			return data[:4] + " **** **** " + data[len(data)-4:]
		}
	case "password":
		// 密码: ******
		return "******"
	}

	// 默认脱敏: 只显示前2位和后2位
	if len(data) > 4 {
		return data[:2] + "***" + data[len(data)-2:]
	}
	return "***"
}

// splitEmail 分割邮箱地址
func splitEmail(email string) []string {
	for i, c := range email {
		if c == '@' {
			return []string{email[:i], email[i+1:]}
		}
	}
	return []string{email}
}
