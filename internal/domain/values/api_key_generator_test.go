package values

import (
	"strings"
	"testing"
)

func TestAPIKeyGenerator_Generate(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	// 测试生成API密钥
	key, hash, prefix, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}
	
	// 验证密钥格式
	if !strings.HasPrefix(key, APIKeyPrefix+"_") {
		t.Errorf("Generated key should start with %s_, got: %s", APIKeyPrefix, key)
	}
	
	// 验证密钥长度
	expectedLength := len(APIKeyPrefix) + 1 + APIKeyLength*2
	if len(key) != expectedLength {
		t.Errorf("Expected key length %d, got %d", expectedLength, len(key))
	}
	
	// 验证哈希不为空
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	// 验证前缀
	if len(prefix) != APIKeyPrefixLength {
		t.Errorf("Expected prefix length %d, got %d", APIKeyPrefixLength, len(prefix))
	}
	
	// 验证前缀是密钥的开头部分
	if !strings.HasPrefix(key, prefix) {
		t.Errorf("Key should start with prefix %s, got key: %s", prefix, key)
	}
}

func TestAPIKeyGenerator_ValidateFormat(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	// 生成一个有效的密钥
	key, _, _, err := generator.Generate()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}
	
	// 测试有效密钥
	if !generator.ValidateFormat(key) {
		t.Errorf("Valid key should pass validation: %s", key)
	}
	
	// 测试无效密钥
	invalidKeys := []string{
		"",                           // 空字符串
		"invalid",                    // 无前缀
		"wrong_prefix_123456789",     // 错误前缀
		"ak_",                        // 只有前缀
		"ak_short",                   // 太短
		"ak_" + strings.Repeat("g", APIKeyLength*2), // 无效十六进制
	}
	
	for _, invalidKey := range invalidKeys {
		if generator.ValidateFormat(invalidKey) {
			t.Errorf("Invalid key should fail validation: %s", invalidKey)
		}
	}
}

func TestAPIKeyGenerator_HashKey(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	testKey := "ak_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	// 测试哈希生成
	hash1 := generator.HashKey(testKey)
	hash2 := generator.HashKey(testKey)
	
	// 相同输入应该产生相同哈希
	if hash1 != hash2 {
		t.Error("Same input should produce same hash")
	}
	
	// 哈希不应该为空
	if hash1 == "" {
		t.Error("Hash should not be empty")
	}
	
	// 不同输入应该产生不同哈希
	differentKey := "ak_abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	hash3 := generator.HashKey(differentKey)
	
	if hash1 == hash3 {
		t.Error("Different inputs should produce different hashes")
	}
}

func TestAPIKeyGenerator_ExtractPrefix(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	testCases := []struct {
		key      string
		expected string
	}{
		{"ak_1234567890abcdef", "ak_12345"},
		{"short", "short"},
		{"", ""},
		{"ak_12345678901234567890", "ak_12345"},
	}
	
	for _, tc := range testCases {
		result := generator.ExtractPrefix(tc.key)
		if result != tc.expected {
			t.Errorf("ExtractPrefix(%s) = %s, expected %s", tc.key, result, tc.expected)
		}
	}
}

func TestAPIKeyGenerator_MaskKey(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	testCases := []struct {
		key      string
		expected string
	}{
		{"ak_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "ak_12345...cdef"},
		{"short", "***"},
		{"", "***"},
		{"ak_123456789012", "ak_12345...9012"},
	}
	
	for _, tc := range testCases {
		result := generator.MaskKey(tc.key)
		if result != tc.expected {
			t.Errorf("MaskKey(%s) = %s, expected %s", tc.key, result, tc.expected)
		}
	}
}

func TestAPIKeyGenerator_Uniqueness(t *testing.T) {
	generator := NewAPIKeyGenerator()
	
	// 生成多个密钥，确保它们是唯一的
	keys := make(map[string]bool)
	hashes := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		key, hash, _, err := generator.Generate()
		if err != nil {
			t.Fatalf("Failed to generate API key: %v", err)
		}
		
		// 检查密钥唯一性
		if keys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		keys[key] = true
		
		// 检查哈希唯一性
		if hashes[hash] {
			t.Errorf("Duplicate hash generated: %s", hash)
		}
		hashes[hash] = true
	}
}
