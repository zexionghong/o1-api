package repositories

import (
	"context"
	"ai-api-gateway/internal/domain/entities"
)

// ProviderRepository 服务提供商仓储接口
type ProviderRepository interface {
	// Create 创建服务提供商
	Create(ctx context.Context, provider *entities.Provider) error
	
	// GetByID 根据ID获取服务提供商
	GetByID(ctx context.Context, id int64) (*entities.Provider, error)
	
	// GetBySlug 根据slug获取服务提供商
	GetBySlug(ctx context.Context, slug string) (*entities.Provider, error)
	
	// Update 更新服务提供商
	Update(ctx context.Context, provider *entities.Provider) error
	
	// UpdateHealthStatus 更新健康状态
	UpdateHealthStatus(ctx context.Context, id int64, status entities.HealthStatus) error
	
	// Delete 删除服务提供商
	Delete(ctx context.Context, id int64) error
	
	// List 获取服务提供商列表
	List(ctx context.Context, offset, limit int) ([]*entities.Provider, error)
	
	// Count 获取服务提供商总数
	Count(ctx context.Context) (int64, error)
	
	// GetActiveProviders 获取活跃的服务提供商列表
	GetActiveProviders(ctx context.Context) ([]*entities.Provider, error)
	
	// GetAvailableProviders 获取可用的服务提供商列表（活跃且健康）
	GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error)
	
	// GetProvidersByPriority 按优先级获取服务提供商列表
	GetProvidersByPriority(ctx context.Context) ([]*entities.Provider, error)
	
	// GetProvidersNeedingHealthCheck 获取需要健康检查的服务提供商列表
	GetProvidersNeedingHealthCheck(ctx context.Context) ([]*entities.Provider, error)
}
