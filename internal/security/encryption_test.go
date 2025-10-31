package security

import (
	"testing"
)

func TestEncryptor(t *testing.T) {
	// 创建加密器（32字节密钥）
	encryptor, err := NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("创建加密器失败: %v", err)
	}

	// 测试加密解密
	plaintext := "这是敏感数据"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("解密后数据不匹配: 期望 %s, 实际 %s", plaintext, decrypted)
	}
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		dataType string
		expected string
	}{
		{"手机号", "13800138000", "phone", "138****8000"},
		{"邮箱", "user@example.com", "email", "u***@example.com"},
		{"身份证", "110101199001011234", "idcard", "110***********1234"},
		{"银行卡", "6222021234567890123", "bankcard", "6222 **** **** 0123"},
		{"密码", "mypassword123", "password", "******"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskSensitiveData(tt.data, tt.dataType)
			if result != tt.expected {
				t.Errorf("期望 %s, 实际 %s", tt.expected, result)
			}
		})
	}
}
