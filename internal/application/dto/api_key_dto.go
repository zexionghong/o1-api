package dto

import (
	"time"
	"ai-api-gateway/internal/domain/entities"
)

// CreateAPIKeyRequest 创建API密钥请求
type CreateAPIKeyRequest struct {
	UserID      int64                          `json:"user_id" validate:"required"`
	Name        string                         `json:"name" validate:"required,min=1,max=100"`
	Permissions *entities.APIKeyPermissions   `json:"permissions,omitempty"`
	ExpiresAt   *time.Time                     `json:"expires_at,omitempty"`
}

// UpdateAPIKeyRequest 更新API密钥请求
type UpdateAPIKeyRequest struct {
	Name        *string                        `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Status      *entities.APIKeyStatus         `json:"status,omitempty"`
	Permissions *entities.APIKeyPermissions   `json:"permissions,omitempty"`
	ExpiresAt   *time.Time                     `json:"expires_at,omitempty"`
}

// APIKeyResponse API密钥响应
type APIKeyResponse struct {
	ID          int64                        `json:"id"`
	UserID      int64                        `json:"user_id"`
	KeyPrefix   string                       `json:"key_prefix"`
	Name        *string                      `json:"name,omitempty"`
	Status      entities.APIKeyStatus        `json:"status"`
	Permissions *entities.APIKeyPermissions `json:"permissions,omitempty"`
	ExpiresAt   *time.Time                   `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time                   `json:"last_used_at,omitempty"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

// APIKeyCreateResponse API密钥创建响应（包含完整密钥）
type APIKeyCreateResponse struct {
	*APIKeyResponse
	Key string `json:"key"` // 完整的API密钥，只在创建时返回
}

// APIKeyListResponse API密钥列表响应
type APIKeyListResponse struct {
	APIKeys    []*APIKeyResponse `json:"api_keys"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// FromEntity 从实体转换
func (r *APIKeyResponse) FromEntity(apiKey *entities.APIKey) *APIKeyResponse {
	return &APIKeyResponse{
		ID:          apiKey.ID,
		UserID:      apiKey.UserID,
		KeyPrefix:   apiKey.KeyPrefix,
		Name:        apiKey.Name,
		Status:      apiKey.Status,
		Permissions: apiKey.Permissions,
		ExpiresAt:   apiKey.ExpiresAt,
		LastUsedAt:  apiKey.LastUsedAt,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}
}

// FromEntities 从实体列表转换
func FromAPIKeyEntities(apiKeys []*entities.APIKey) []*APIKeyResponse {
	responses := make([]*APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = (&APIKeyResponse{}).FromEntity(apiKey)
	}
	return responses
}
