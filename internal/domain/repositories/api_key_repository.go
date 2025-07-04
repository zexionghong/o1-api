package repositories

import (
	"context"
	"ai-api-gateway/internal/domain/entities"
)

// APIKeyRepository API密钥仓储接口
type APIKeyRepository interface {
	// Create 创建API密钥
	Create(ctx context.Context, apiKey *entities.APIKey) error
	
	// GetByID 根据ID获取API密钥
	GetByID(ctx context.Context, id int64) (*entities.APIKey, error)
	
	// GetByKeyHash 根据密钥哈希获取API密钥
	GetByKeyHash(ctx context.Context, keyHash string) (*entities.APIKey, error)
	
	// GetByUserID 根据用户ID获取API密钥列表
	GetByUserID(ctx context.Context, userID int64) ([]*entities.APIKey, error)
	
	// Update 更新API密钥
	Update(ctx context.Context, apiKey *entities.APIKey) error
	
	// UpdateLastUsed 更新最后使用时间
	UpdateLastUsed(ctx context.Context, id int64) error
	
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, id int64, status entities.APIKeyStatus) error
	
	// Delete 删除API密钥
	Delete(ctx context.Context, id int64) error
	
	// List 获取API密钥列表
	List(ctx context.Context, offset, limit int) ([]*entities.APIKey, error)
	
	// Count 获取API密钥总数
	Count(ctx context.Context) (int64, error)
	
	// GetActiveKeys 获取活跃的API密钥列表
	GetActiveKeys(ctx context.Context, userID int64) ([]*entities.APIKey, error)
	
	// GetExpiredKeys 获取过期的API密钥列表
	GetExpiredKeys(ctx context.Context, limit int) ([]*entities.APIKey, error)
}
