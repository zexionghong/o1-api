package services

import (
	"context"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// 以下是其他服务的存根实现，用于确保项目能够编译
// 在实际开发中，这些应该被完整实现

// ProviderService 提供商服务接口
type ProviderService interface {
	// GetAvailableProviders 获取可用的提供商列表
	GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error)

	// SelectProvider 选择提供商（负载均衡）
	SelectProvider(ctx context.Context, modelSlug string) (*entities.Provider, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context, providerID int64) error
}

// providerServiceImpl 提供商服务存根实现
type providerServiceImpl struct {
	providerRepo repositories.ProviderRepository
	modelRepo    repositories.ModelRepository
}

func NewProviderService(providerRepo repositories.ProviderRepository, modelRepo repositories.ModelRepository) ProviderService {
	return &providerServiceImpl{
		providerRepo: providerRepo,
		modelRepo:    modelRepo,
	}
}

func (s *providerServiceImpl) GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error) {
	// TODO: 实现获取可用的提供商列表
	return s.providerRepo.GetAvailableProviders(ctx)
}

func (s *providerServiceImpl) SelectProvider(ctx context.Context, modelSlug string) (*entities.Provider, error) {
	// TODO: 实现负载均衡选择提供商
	providers, err := s.providerRepo.GetAvailableProviders(ctx)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, entities.ErrNoAvailableProvider
	}
	return providers[0], nil // 简单返回第一个可用的提供商
}

func (s *providerServiceImpl) HealthCheck(ctx context.Context, providerID int64) error {
	// TODO: 实现健康检查
	return nil
}

// ModelService 模型服务接口
type ModelService interface {
	// GetAvailableModels 获取可用的模型列表
	GetAvailableModels(ctx context.Context, providerID int64) ([]*entities.Model, error)

	// GetModelBySlug 根据slug获取模型
	GetModelBySlug(ctx context.Context, providerID int64, slug string) (*entities.Model, error)
}

// modelServiceImpl 模型服务存根实现
type modelServiceImpl struct {
	modelRepo        repositories.ModelRepository
	modelPricingRepo repositories.ModelPricingRepository
	providerRepo     repositories.ProviderRepository
}

func NewModelService(modelRepo repositories.ModelRepository, modelPricingRepo repositories.ModelPricingRepository, providerRepo repositories.ProviderRepository) ModelService {
	return &modelServiceImpl{
		modelRepo:        modelRepo,
		modelPricingRepo: modelPricingRepo,
		providerRepo:     providerRepo,
	}
}

func (s *modelServiceImpl) GetAvailableModels(ctx context.Context, providerID int64) ([]*entities.Model, error) {
	// TODO: 实现获取可用的模型列表
	// 注意：现在模型不再直接绑定到提供商，需要通过 provider_model_support 表查询
	// 这里暂时返回所有可用模型，实际应该根据 provider_model_support 过滤
	return s.modelRepo.GetAvailableModels(ctx)
}

func (s *modelServiceImpl) GetModelBySlug(ctx context.Context, providerID int64, slug string) (*entities.Model, error) {
	// TODO: 实现根据slug获取模型
	// 注意：现在模型不再绑定到特定提供商，直接通过 slug 查询
	return s.modelRepo.GetBySlug(ctx, slug)
}

// 配额服务实现已移至 quota_service_impl.go

// BillingService 计费服务接口
type BillingService interface {
	// CalculateCost 计算请求成本
	CalculateCost(ctx context.Context, modelID int64, inputTokens, outputTokens int) (float64, error)

	// ProcessBilling 处理计费
	ProcessBilling(ctx context.Context, usageLog *entities.UsageLog) error
}

// 计费服务实现已移至 billing_service_impl.go

// UsageLogService 使用日志服务接口
type UsageLogService interface {
	// CreateUsageLog 创建使用日志
	CreateUsageLog(ctx context.Context, log *entities.UsageLog) error

	// GetUsageStats 获取使用统计
	GetUsageStats(ctx context.Context, userID int64) (*repositories.UsageStats, error)
}

// usageLogServiceImpl 使用日志服务存根实现
type usageLogServiceImpl struct {
	usageLogRepo repositories.UsageLogRepository
	userRepo     repositories.UserRepository
	apiKeyRepo   repositories.APIKeyRepository
	providerRepo repositories.ProviderRepository
	modelRepo    repositories.ModelRepository
}

func NewUsageLogService(usageLogRepo repositories.UsageLogRepository, userRepo repositories.UserRepository, apiKeyRepo repositories.APIKeyRepository, providerRepo repositories.ProviderRepository, modelRepo repositories.ModelRepository) UsageLogService {
	return &usageLogServiceImpl{
		usageLogRepo: usageLogRepo,
		userRepo:     userRepo,
		apiKeyRepo:   apiKeyRepo,
		providerRepo: providerRepo,
		modelRepo:    modelRepo,
	}
}

func (s *usageLogServiceImpl) CreateUsageLog(ctx context.Context, log *entities.UsageLog) error {
	// TODO: 实现创建使用日志
	return s.usageLogRepo.Create(ctx, log)
}

func (s *usageLogServiceImpl) GetUsageStats(ctx context.Context, userID int64) (*repositories.UsageStats, error) {
	// TODO: 实现获取使用统计
	return nil, nil
}
