package services

import (
	"context"
	"ai-api-gateway/internal/domain/entities"
)

// AuthService 认证服务接口
type AuthService interface {
	// ValidateAPIKey 验证API密钥
	ValidateAPIKey(ctx context.Context, keyString string) (*entities.APIKey, *entities.User, error)
	
	// GenerateAPIKey 生成API密钥
	GenerateAPIKey(ctx context.Context, userID int64, name string, permissions *entities.APIKeyPermissions) (*entities.APIKey, string, error)
	
	// RevokeAPIKey 撤销API密钥
	RevokeAPIKey(ctx context.Context, keyID int64) error
	
	// RefreshAPIKey 刷新API密钥
	RefreshAPIKey(ctx context.Context, keyID int64) (*entities.APIKey, string, error)
	
	// CheckPermissions 检查权限
	CheckPermissions(ctx context.Context, apiKey *entities.APIKey, providerSlug, modelSlug string) error
	
	// UpdateLastUsed 更新最后使用时间
	UpdateLastUsed(ctx context.Context, keyID int64) error
}

// AuthResult 认证结果
type AuthResult struct {
	APIKey *entities.APIKey `json:"api_key"`
	User   *entities.User   `json:"user"`
	Valid  bool             `json:"valid"`
	Reason string           `json:"reason,omitempty"`
}
