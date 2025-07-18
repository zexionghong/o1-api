package entities

import (
	"encoding/json"
	"time"
)

// APIKeyStatus API密钥状态枚举
type APIKeyStatus string

const (
	APIKeyStatusActive    APIKeyStatus = "active"
	APIKeyStatusSuspended APIKeyStatus = "suspended"
	APIKeyStatusExpired   APIKeyStatus = "expired"
	APIKeyStatusRevoked   APIKeyStatus = "revoked"
)

// APIKeyPermissions API密钥权限
type APIKeyPermissions struct {
	AllowedModels        []string `json:"allowed_models,omitempty"`
	AllowedProviders     []string `json:"allowed_providers,omitempty"`
	MaxRequestsPerMinute int      `json:"max_requests_per_minute,omitempty"`
	MaxTokensPerRequest  int      `json:"max_tokens_per_request,omitempty"`
}

// APIKey API密钥实体
type APIKey struct {
	ID          int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      int64              `json:"user_id" gorm:"column:user_id;not null;index"`
	Key         string             `json:"key" gorm:"column:key;uniqueIndex;not null;size:255"` // 直接存储完整的API Key
	KeyPrefix   string             `json:"key_prefix" gorm:"column:key_prefix;not null;size:20;index"`
	Name        *string            `json:"name,omitempty" gorm:"size:100"`
	Status      APIKeyStatus       `json:"status" gorm:"not null;default:active;size:20;index"`
	Permissions *APIKeyPermissions `json:"permissions,omitempty" gorm:"type:jsonb"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time         `json:"last_used_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time          `json:"updated_at" gorm:"not null;autoUpdateTime"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}

// IsActive 检查API密钥是否处于活跃状态
func (ak *APIKey) IsActive() bool {
	if ak.Status != APIKeyStatusActive {
		return false
	}

	// 检查是否过期
	if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// IsExpired 检查API密钥是否已过期
func (ak *APIKey) IsExpired() bool {
	return ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now())
}

// UpdateLastUsed 更新最后使用时间
func (ak *APIKey) UpdateLastUsed() {
	now := time.Now()
	ak.LastUsedAt = &now
	ak.UpdatedAt = now
}

// HasPermissionForModel 检查是否有权限使用指定模型
func (ak *APIKey) HasPermissionForModel(modelSlug string) bool {
	if ak.Permissions == nil || len(ak.Permissions.AllowedModels) == 0 {
		return true // 没有限制则允许所有模型
	}

	for _, allowedModel := range ak.Permissions.AllowedModels {
		if allowedModel == modelSlug || allowedModel == "*" {
			return true
		}
	}
	return false
}

// HasPermissionForProvider 检查是否有权限使用指定提供商
func (ak *APIKey) HasPermissionForProvider(providerSlug string) bool {
	if ak.Permissions == nil || len(ak.Permissions.AllowedProviders) == 0 {
		return true // 没有限制则允许所有提供商
	}

	for _, allowedProvider := range ak.Permissions.AllowedProviders {
		if allowedProvider == providerSlug || allowedProvider == "*" {
			return true
		}
	}
	return false
}

// MarshalPermissions 序列化权限为JSON字符串
func (ak *APIKey) MarshalPermissions() (string, error) {
	if ak.Permissions == nil {
		return "", nil
	}

	data, err := json.Marshal(ak.Permissions)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalPermissions 从JSON字符串反序列化权限
func (ak *APIKey) UnmarshalPermissions(data string) error {
	if data == "" {
		ak.Permissions = nil
		return nil
	}

	var permissions APIKeyPermissions
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		return err
	}
	ak.Permissions = &permissions
	return nil
}
