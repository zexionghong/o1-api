package redis

import (
	"ai-api-gateway/internal/infrastructure/logger"
)

// RedisFactory Redis工厂
type RedisFactory struct {
	client              *RedisClient
	cache               *CacheService
	lockService         *DistributedLockService
	invalidationService *CacheInvalidationService
	cacheManager        *CacheManager
	logger              logger.Logger
}

// NewRedisFactory 创建Redis工厂
func NewRedisFactory(log logger.Logger) (*RedisFactory, error) {
	// 创建Redis客户端
	client, err := NewRedisClient(log)
	if err != nil {
		return nil, err
	}

	// 创建缓存服务
	cache := NewCacheService(client, log)

	// 创建分布式锁服务
	lockService := NewDistributedLockService(client, log)

	// 创建缓存失效服务
	invalidationService := NewCacheInvalidationService(cache, client, log)

	// 创建缓存管理器
	cacheManager := NewCacheManager(cache, client, invalidationService, log)

	return &RedisFactory{
		client:              client,
		cache:               cache,
		lockService:         lockService,
		invalidationService: invalidationService,
		cacheManager:        cacheManager,
		logger:              log,
	}, nil
}

// GetClient 获取Redis客户端
func (f *RedisFactory) GetClient() *RedisClient {
	return f.client
}

// GetCacheService 获取缓存服务
func (f *RedisFactory) GetCacheService() *CacheService {
	return f.cache
}

// GetLockService 获取分布式锁服务
func (f *RedisFactory) GetLockService() *DistributedLockService {
	return f.lockService
}

// GetInvalidationService 获取缓存失效服务
func (f *RedisFactory) GetInvalidationService() *CacheInvalidationService {
	return f.invalidationService
}

// GetCacheManager 获取缓存管理器
func (f *RedisFactory) GetCacheManager() *CacheManager {
	return f.cacheManager
}

// Close 关闭Redis连接
func (f *RedisFactory) Close() error {
	if f.client != nil {
		return f.client.Close()
	}
	return nil
}
