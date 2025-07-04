package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
)

// ModelRepository 模型仓储接口
type ModelRepository interface {
	// Create 创建模型
	Create(ctx context.Context, model *entities.Model) error

	// GetByID 根据ID获取模型
	GetByID(ctx context.Context, id int64) (*entities.Model, error)

	// GetBySlug 根据slug获取模型
	GetBySlug(ctx context.Context, providerID int64, slug string) (*entities.Model, error)

	// GetByProviderID 根据提供商ID获取模型列表
	GetByProviderID(ctx context.Context, providerID int64) ([]*entities.Model, error)

	// Update 更新模型
	Update(ctx context.Context, model *entities.Model) error

	// Delete 删除模型
	Delete(ctx context.Context, id int64) error

	// List 获取模型列表
	List(ctx context.Context, offset, limit int) ([]*entities.Model, error)

	// Count 获取模型总数
	Count(ctx context.Context) (int64, error)

	// GetActiveModels 获取活跃的模型列表
	GetActiveModels(ctx context.Context) ([]*entities.Model, error)

	// GetModelsByType 根据类型获取模型列表
	GetModelsByType(ctx context.Context, modelType entities.ModelType) ([]*entities.Model, error)

	// GetAvailableModels 获取可用的模型列表（活跃且提供商可用）
	GetAvailableModels(ctx context.Context, providerID int64) ([]*entities.Model, error)
}

// ModelPricingRepository 模型定价仓储接口
type ModelPricingRepository interface {
	// Create 创建模型定价
	Create(ctx context.Context, pricing *entities.ModelPricing) error

	// GetByID 根据ID获取模型定价
	GetByID(ctx context.Context, id int64) (*entities.ModelPricing, error)

	// GetByModelID 根据模型ID获取定价列表
	GetByModelID(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error)

	// GetCurrentPricing 获取模型当前有效定价
	GetCurrentPricing(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error)

	// Update 更新模型定价
	Update(ctx context.Context, pricing *entities.ModelPricing) error

	// Delete 删除模型定价
	Delete(ctx context.Context, id int64) error

	// List 获取模型定价列表
	List(ctx context.Context, offset, limit int) ([]*entities.ModelPricing, error)

	// Count 获取模型定价总数
	Count(ctx context.Context) (int64, error)

	// GetPricingByType 根据定价类型获取定价
	GetPricingByType(ctx context.Context, modelID int64, pricingType entities.PricingType) (*entities.ModelPricing, error)
}

// ProviderModelSupportRepository 提供商模型支持仓储接口
type ProviderModelSupportRepository interface {
	// Create 创建提供商模型支持
	Create(ctx context.Context, support *entities.ProviderModelSupport) error

	// GetByID 根据ID获取提供商模型支持
	GetByID(ctx context.Context, id int64) (*entities.ProviderModelSupport, error)

	// GetByProviderAndModel 根据提供商和模型获取支持信息
	GetByProviderAndModel(ctx context.Context, providerID int64, modelSlug string) (*entities.ProviderModelSupport, error)

	// GetSupportingProviders 获取支持指定模型的提供商列表
	GetSupportingProviders(ctx context.Context, modelSlug string) ([]*entities.ModelSupportInfo, error)

	// GetProviderSupportedModels 获取提供商支持的模型列表
	GetProviderSupportedModels(ctx context.Context, providerID int64) ([]*entities.ProviderModelSupport, error)

	// Update 更新提供商模型支持
	Update(ctx context.Context, support *entities.ProviderModelSupport) error

	// Delete 删除提供商模型支持
	Delete(ctx context.Context, id int64) error

	// EnableSupport 启用模型支持
	EnableSupport(ctx context.Context, providerID int64, modelSlug string) error

	// DisableSupport 禁用模型支持
	DisableSupport(ctx context.Context, providerID int64, modelSlug string) error

	// List 获取提供商模型支持列表
	List(ctx context.Context, offset, limit int) ([]*entities.ProviderModelSupport, error)

	// Count 获取提供商模型支持总数
	Count(ctx context.Context) (int64, error)
}
