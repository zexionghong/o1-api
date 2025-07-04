package values

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// APIKeyPrefix API密钥前缀
	APIKeyPrefix = "ak"
	// APIKeyLength API密钥长度（不包括前缀）
	APIKeyLength = 32
	// APIKeyPrefixLength 前缀长度用于存储
	APIKeyPrefixLength = 8
)

// APIKeyGenerator API密钥生成器
type APIKeyGenerator struct{}

// NewAPIKeyGenerator 创建API密钥生成器
func NewAPIKeyGenerator() *APIKeyGenerator {
	return &APIKeyGenerator{}
}

// Generate 生成API密钥
func (g *APIKeyGenerator) Generate() (key, hash, prefix string, err error) {
	// 生成随机字节
	randomBytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// 转换为十六进制字符串
	randomHex := hex.EncodeToString(randomBytes)
	
	// 构造完整的API密钥
	key = fmt.Sprintf("%s_%s", APIKeyPrefix, randomHex)
	
	// 生成哈希用于存储
	hash = g.HashKey(key)
	
	// 提取前缀用于显示
	prefix = g.ExtractPrefix(key)
	
	return key, hash, prefix, nil
}

// HashKey 对API密钥进行哈希
func (g *APIKeyGenerator) HashKey(key string) string {
	hasher := sha256.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

// ExtractPrefix 提取API密钥前缀
func (g *APIKeyGenerator) ExtractPrefix(key string) string {
	if len(key) < APIKeyPrefixLength {
		return key
	}
	return key[:APIKeyPrefixLength]
}

// ValidateFormat 验证API密钥格式
func (g *APIKeyGenerator) ValidateFormat(key string) bool {
	// 检查前缀
	if !strings.HasPrefix(key, APIKeyPrefix+"_") {
		return false
	}
	
	// 检查长度
	expectedLength := len(APIKeyPrefix) + 1 + APIKeyLength*2 // prefix + _ + hex
	if len(key) != expectedLength {
		return false
	}
	
	// 检查十六进制部分
	hexPart := key[len(APIKeyPrefix)+1:]
	if _, err := hex.DecodeString(hexPart); err != nil {
		return false
	}
	
	return true
}

// MaskKey 掩码API密钥用于显示
func (g *APIKeyGenerator) MaskKey(key string) string {
	if len(key) < 12 {
		return "***"
	}
	
	prefix := key[:8]
	suffix := key[len(key)-4:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}
