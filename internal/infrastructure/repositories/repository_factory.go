package repositories

import (
	"database/sql"

	"ai-api-gateway/internal/domain/repositories"
)

// RepositoryFactory 仓储工厂
type RepositoryFactory struct {
	db *sql.DB
}

// NewRepositoryFactory 创建仓储工厂
func NewRepositoryFactory(db *sql.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,
	}
}

// UserRepository 获取用户仓储
func (f *RepositoryFactory) UserRepository() repositories.UserRepository {
	return NewUserRepository(f.db)
}

// APIKeyRepository 获取API密钥仓储
func (f *RepositoryFactory) APIKeyRepository() repositories.APIKeyRepository {
	return NewAPIKeyRepository(f.db)
}

// ProviderRepository 获取提供商仓储
func (f *RepositoryFactory) ProviderRepository() repositories.ProviderRepository {
	return NewProviderRepository(f.db)
}

// ModelRepository 获取模型仓储
func (f *RepositoryFactory) ModelRepository() repositories.ModelRepository {
	return NewModelRepository(f.db)
}

// ModelPricingRepository 获取模型定价仓储
func (f *RepositoryFactory) ModelPricingRepository() repositories.ModelPricingRepository {
	return NewModelPricingRepository(f.db)
}

// ProviderModelSupportRepository 获取提供商模型支持仓储
func (f *RepositoryFactory) ProviderModelSupportRepository() repositories.ProviderModelSupportRepository {
	return NewProviderModelSupportRepository(f.db)
}

// QuotaRepository 获取配额仓储
func (f *RepositoryFactory) QuotaRepository() repositories.QuotaRepository {
	return NewQuotaRepository(f.db)
}

// QuotaUsageRepository 获取配额使用仓储
func (f *RepositoryFactory) QuotaUsageRepository() repositories.QuotaUsageRepository {
	return NewQuotaUsageRepository(f.db)
}

// UsageLogRepository 获取使用日志仓储
func (f *RepositoryFactory) UsageLogRepository() repositories.UsageLogRepository {
	return NewUsageLogRepository(f.db)
}

// BillingRecordRepository 获取计费记录仓储
func (f *RepositoryFactory) BillingRecordRepository() repositories.BillingRecordRepository {
	return NewBillingRecordRepository(f.db)
}
