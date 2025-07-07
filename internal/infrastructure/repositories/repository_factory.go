package repositories

import (
	"ai-api-gateway/internal/domain/repositories"

	"gorm.io/gorm"
)

// RepositoryFactory 仓储工厂（基于GORM）
type RepositoryFactory struct {
	gormDB *gorm.DB
}

// NewRepositoryFactory 创建GORM仓储工厂
func NewRepositoryFactory(gormDB *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{
		gormDB: gormDB,
	}
}

// UserRepository 获取用户仓储
func (f *RepositoryFactory) UserRepository() repositories.UserRepository {
	return NewUserRepositoryGorm(f.gormDB)
}

// APIKeyRepository 获取API密钥仓储
func (f *RepositoryFactory) APIKeyRepository() repositories.APIKeyRepository {
	return NewAPIKeyRepositoryGorm(f.gormDB)
}

// ProviderRepository 获取提供商仓储
func (f *RepositoryFactory) ProviderRepository() repositories.ProviderRepository {
	return NewProviderRepositoryGorm(f.gormDB)
}

// ModelRepository 获取模型仓储
func (f *RepositoryFactory) ModelRepository() repositories.ModelRepository {
	return NewModelRepositoryGorm(f.gormDB)
}

// ModelPricingRepository 获取模型定价仓储
func (f *RepositoryFactory) ModelPricingRepository() repositories.ModelPricingRepository {
	return NewModelPricingRepositoryGorm(f.gormDB)
}

// ProviderModelSupportRepository 获取提供商模型支持仓储
func (f *RepositoryFactory) ProviderModelSupportRepository() repositories.ProviderModelSupportRepository {
	return NewProviderModelSupportRepositoryGorm(f.gormDB)
}

// QuotaRepository 获取配额仓储
func (f *RepositoryFactory) QuotaRepository() repositories.QuotaRepository {
	return NewQuotaRepositoryGorm(f.gormDB)
}

// QuotaUsageRepository 获取配额使用仓储
func (f *RepositoryFactory) QuotaUsageRepository() repositories.QuotaUsageRepository {
	return NewQuotaUsageRepositoryGorm(f.gormDB)
}

// UsageLogRepository 获取使用日志仓储
func (f *RepositoryFactory) UsageLogRepository() repositories.UsageLogRepository {
	return NewUsageLogRepositoryGorm(f.gormDB)
}

// BillingRecordRepository 获取计费记录仓储
func (f *RepositoryFactory) BillingRecordRepository() repositories.BillingRecordRepository {
	return NewBillingRecordRepositoryGorm(f.gormDB)
}

// ToolRepository 获取工具仓储
func (f *RepositoryFactory) ToolRepository() repositories.ToolRepository {
	return NewToolRepositoryGorm(f.gormDB)
}

// GormDB 获取GORM数据库连接
func (f *RepositoryFactory) GormDB() *gorm.DB {
	return f.gormDB
}
