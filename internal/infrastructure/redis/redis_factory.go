package redis

import (
	"ai-api-gateway/internal/infrastructure/logger"
)

// RedisFactory Redis工厂
type RedisFactory struct {
	client      *RedisClient
	cache       *CacheService
	lockService *DistributedLockService
	logger      logger.Logger
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

	return &RedisFactory{
		client:      client,
		cache:       cache,
		lockService: lockService,
		logger:      log,
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

// Close 关闭Redis连接
func (f *RedisFactory) Close() error {
	if f.client != nil {
		return f.client.Close()
	}
	return nil
}
