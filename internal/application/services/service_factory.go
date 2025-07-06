package services

import (
	"time"

	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/async"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/logger"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"
	infraRepos "ai-api-gateway/internal/infrastructure/repositories"
)

// ServiceFactory 服务工厂
type ServiceFactory struct {
	repoFactory  *infraRepos.RepositoryFactory
	redisFactory *redisInfra.RedisFactory
	config       *config.Config
	logger       logger.Logger
}

// NewServiceFactory 创建服务工厂
func NewServiceFactory(repoFactory *infraRepos.RepositoryFactory, redisFactory *redisInfra.RedisFactory, cfg *config.Config, log logger.Logger) *ServiceFactory {
	return &ServiceFactory{
		repoFactory:  repoFactory,
		redisFactory: redisFactory,
		config:       cfg,
		logger:       log,
	}
}

// UserService 获取用户服务
func (f *ServiceFactory) UserService() UserService {
	var cache *redisInfra.CacheService
	var lockService *redisInfra.DistributedLockService

	if f.redisFactory != nil {
		cache = f.redisFactory.GetCacheService()
		lockService = f.redisFactory.GetLockService()
	}

	return NewUserService(f.repoFactory.UserRepository(), cache, lockService)
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
	// 检查是否启用异步配额处理
	if f.isAsyncQuotaEnabled() && f.redisFactory != nil {
		// 创建异步配额服务
		asyncService, err := f.createAsyncQuotaService()
		if err != nil {
			f.logger.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Error("Failed to create async quota service, falling back to sync")
		} else {
			return asyncService
		}
	}

	// 如果有Redis工厂，创建带缓存的配额服务
	if f.redisFactory != nil {
		return NewQuotaServiceWithCache(
			f.repoFactory.QuotaRepository(),
			f.repoFactory.QuotaUsageRepository(),
			f.repoFactory.UserRepository(),
			f.redisFactory.GetCacheService(),
			f.redisFactory.GetInvalidationService(),
			f.logger,
		)
	}

	// 否则创建普通的配额服务
	return NewQuotaService(
		f.repoFactory.QuotaRepository(),
		f.repoFactory.QuotaUsageRepository(),
		f.repoFactory.UserRepository(),
		f.logger,
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

// JWTService 获取JWT服务
func (f *ServiceFactory) JWTService() JWTService {
	return NewJWTService(&f.config.JWT)
}

// AuthService 获取认证服务
func (f *ServiceFactory) AuthService() AuthService {
	return NewAuthService(
		f.repoFactory.UserRepository(),
		f.JWTService(),
	)
}

// UsageLogRepository 获取使用日志仓储
func (f *ServiceFactory) UsageLogRepository() repositories.UsageLogRepository {
	return f.repoFactory.UsageLogRepository()
}

// BillingRecordRepository 获取计费记录仓储
func (f *ServiceFactory) BillingRecordRepository() repositories.BillingRecordRepository {
	return f.repoFactory.BillingRecordRepository()
}

// UserRepository 获取用户仓储
func (f *ServiceFactory) UserRepository() repositories.UserRepository {
	return f.repoFactory.UserRepository()
}

// ModelRepository 获取模型仓储
func (f *ServiceFactory) ModelRepository() repositories.ModelRepository {
	return f.repoFactory.ModelRepository()
}

// isAsyncQuotaEnabled 检查是否启用异步配额处理
func (f *ServiceFactory) isAsyncQuotaEnabled() bool {
	// 暂时硬编码返回true来启用异步处理
	// 在实际项目中应该从配置文件读取: viper.GetBool("async_quota.enabled")
	return true
}

// createAsyncQuotaService 创建异步配额服务
func (f *ServiceFactory) createAsyncQuotaService() (QuotaService, error) {
	// 创建异步消费者配置
	config := f.getAsyncQuotaConfig()

	// 创建异步配额服务
	return NewAsyncQuotaService(
		f.repoFactory.QuotaRepository(),
		f.repoFactory.QuotaUsageRepository(),
		f.repoFactory.UserRepository(),
		f.redisFactory.GetCacheService(),
		f.redisFactory.GetInvalidationService(),
		config,
		f.logger,
	)
}

// getAsyncQuotaConfig 获取异步配额配置
func (f *ServiceFactory) getAsyncQuotaConfig() *async.QuotaConsumerConfig {
	// 暂时使用默认配置
	// 在实际项目中应该从配置文件读取
	return &async.QuotaConsumerConfig{
		WorkerCount:   3,                      // 3个工作协程
		ChannelSize:   1000,                   // 1000个事件缓冲
		BatchSize:     10,                     // 每批处理10个事件
		FlushInterval: 5 * time.Second,        // 5秒强制刷新
		RetryAttempts: 3,                      // 重试3次
		RetryDelay:    100 * time.Millisecond, // 100ms重试延迟
	}
}
