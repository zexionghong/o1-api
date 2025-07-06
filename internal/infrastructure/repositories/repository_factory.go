package repositories

import (
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/database"

	"github.com/jmoiron/sqlx"
)

// RepositoryFactory 仓储工厂
type RepositoryFactory struct {
	dbConn *database.Connection
}

// NewRepositoryFactory 创建仓储工厂
func NewRepositoryFactory(dbConn *database.Connection) *RepositoryFactory {
	return &RepositoryFactory{
		dbConn: dbConn,
	}
}

// UserRepository 获取用户仓储
func (f *RepositoryFactory) UserRepository() repositories.UserRepository {
	return NewUserRepository(f.dbConn.DB())
}

// APIKeyRepository 获取API密钥仓储
func (f *RepositoryFactory) APIKeyRepository() repositories.APIKeyRepository {
	return NewAPIKeyRepository(f.dbConn.DB())
}

// ProviderRepository 获取提供商仓储
func (f *RepositoryFactory) ProviderRepository() repositories.ProviderRepository {
	return NewProviderRepository(f.dbConn.DB())
}

// ModelRepository 获取模型仓储
func (f *RepositoryFactory) ModelRepository() repositories.ModelRepository {
	return NewModelRepository(f.dbConn.DB())
}

// ModelPricingRepository 获取模型定价仓储
func (f *RepositoryFactory) ModelPricingRepository() repositories.ModelPricingRepository {
	return NewModelPricingRepository(f.dbConn.DB())
}

// ProviderModelSupportRepository 获取提供商模型支持仓储
func (f *RepositoryFactory) ProviderModelSupportRepository() repositories.ProviderModelSupportRepository {
	return NewProviderModelSupportRepository(f.dbConn.DB())
}

// QuotaRepository 获取配额仓储
func (f *RepositoryFactory) QuotaRepository() repositories.QuotaRepository {
	return NewQuotaRepository(f.dbConn.DB())
}

// QuotaUsageRepository 获取配额使用仓储
func (f *RepositoryFactory) QuotaUsageRepository() repositories.QuotaUsageRepository {
	return NewQuotaUsageRepository(f.dbConn.DB())
}

// UsageLogRepository 获取使用日志仓储
func (f *RepositoryFactory) UsageLogRepository() repositories.UsageLogRepository {
	return NewUsageLogRepository(f.dbConn.DB())
}

// BillingRecordRepository 获取计费记录仓储
func (f *RepositoryFactory) BillingRecordRepository() repositories.BillingRecordRepository {
	return NewBillingRecordRepository(f.dbConn.DB())
}

// ToolRepository 获取工具仓储
func (f *RepositoryFactory) ToolRepository() repositories.ToolRepository {
	return NewToolRepository(f.dbConn.DBX())
}

// DB 获取数据库连接
func (f *RepositoryFactory) DB() *sqlx.DB {
	return f.dbConn.DBX()
}
