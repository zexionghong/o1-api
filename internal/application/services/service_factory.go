package services

import (
	"ai-api-gateway/internal/infrastructure/repositories"
)

// ServiceFactory 服务工厂
type ServiceFactory struct {
	repoFactory *repositories.RepositoryFactory
}

// NewServiceFactory 创建服务工厂
func NewServiceFactory(repoFactory *repositories.RepositoryFactory) *ServiceFactory {
	return &ServiceFactory{
		repoFactory: repoFactory,
	}
}

// UserService 获取用户服务
func (f *ServiceFactory) UserService() UserService {
	return NewUserService(f.repoFactory.UserRepository())
}

// APIKeyService 获取API密钥服务
func (f *ServiceFactory) APIKeyService() APIKeyService {
	return NewAPIKeyService(
		f.repoFactory.APIKeyRepository(),
		f.repoFactory.UserRepository(),
	)
}

// ProviderService 获取提供商服务
func (f *ServiceFactory) ProviderService() ProviderService {
	return NewProviderService(
		f.repoFactory.ProviderRepository(),
		f.repoFactory.ModelRepository(),
	)
}

// ModelService 获取模型服务
func (f *ServiceFactory) ModelService() ModelService {
	return NewModelService(
		f.repoFactory.ModelRepository(),
		f.repoFactory.ModelPricingRepository(),
		f.repoFactory.ProviderRepository(),
	)
}

// QuotaService 获取配额服务
func (f *ServiceFactory) QuotaService() QuotaService {
	return NewQuotaService(
		f.repoFactory.QuotaRepository(),
		f.repoFactory.QuotaUsageRepository(),
		f.repoFactory.UserRepository(),
	)
}

// BillingService 获取计费服务
func (f *ServiceFactory) BillingService() BillingService {
	return NewBillingService(
		f.repoFactory.BillingRecordRepository(),
		f.repoFactory.UsageLogRepository(),
		f.repoFactory.ModelPricingRepository(),
		f.repoFactory.UserRepository(),
	)
}

// UsageLogService 获取使用日志服务
func (f *ServiceFactory) UsageLogService() UsageLogService {
	return NewUsageLogService(
		f.repoFactory.UsageLogRepository(),
		f.repoFactory.UserRepository(),
		f.repoFactory.APIKeyRepository(),
		f.repoFactory.ProviderRepository(),
		f.repoFactory.ModelRepository(),
	)
}
