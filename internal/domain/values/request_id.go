package values

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// RequestIDGenerator 请求ID生成器
type RequestIDGenerator struct{}

// NewRequestIDGenerator 创建请求ID生成器
func NewRequestIDGenerator() *RequestIDGenerator {
	return &RequestIDGenerator{}
}

// Generate 生成请求ID
// 格式: req_<timestamp>_<random>
func (g *RequestIDGenerator) Generate() (string, error) {
	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli()
	
	// 生成4字节随机数
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// 转换为十六进制
	randomHex := hex.EncodeToString(randomBytes)
	
	// 构造请求ID
	requestID := fmt.Sprintf("req_%d_%s", timestamp, randomHex)
	
	return requestID, nil
}

// Validate 验证请求ID格式
func (g *RequestIDGenerator) Validate(requestID string) bool {
	if len(requestID) < 20 { // req_ + timestamp + _ + 8位hex
		return false
	}
	
	// 检查前缀
	if requestID[:4] != "req_" {
		return false
	}
	
	// 简单格式检查，实际使用中可以更严格
	return true
}

// ExtractTimestamp 从请求ID中提取时间戳
func (g *RequestIDGenerator) ExtractTimestamp(requestID string) (time.Time, error) {
	if len(requestID) < 20 {
		return time.Time{}, fmt.Errorf("invalid request ID format")
	}
	
	// 查找第二个下划线的位置
	parts := requestID[4:] // 去掉 "req_" 前缀
	underscorePos := -1
	for i, char := range parts {
		if char == '_' {
			underscorePos = i
			break
		}
	}
	
	if underscorePos == -1 {
		return time.Time{}, fmt.Errorf("invalid request ID format")
	}
	
	timestampStr := parts[:underscorePos]
	
	// 解析时间戳
	var timestamp int64
	if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	
	return time.UnixMilli(timestamp), nil
}
