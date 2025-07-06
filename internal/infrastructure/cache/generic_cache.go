package cache

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/infrastructure/redis"
)

// GenericCache 通用缓存接口
type GenericCache[T any] interface {
	Set(ctx context.Context, key string, value *T, ttl time.Duration) error
	Get(ctx context.Context, key string) (*T, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetWithKeyFunc(ctx context.Context, keyFunc func(*T) string, value *T, ttl time.Duration) error
	GetByKeyFunc(ctx context.Context, keyFunc func() string) (*T, error)
}

// genericCacheImpl 通用缓存实现
type genericCacheImpl[T any] struct {
	cache  redis.CacheService
	logger logger.Logger
	prefix string
}

// NewGenericCache 创建通用缓存
func NewGenericCache[T any](cache redis.CacheService, logger logger.Logger, prefix string) GenericCache[T] {
	return &genericCacheImpl[T]{
		cache:  cache,
		logger: logger,
		prefix: prefix,
	}
}

// Set 设置缓存
func (c *genericCacheImpl[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}

	fullKey := c.getFullKey(key)
	return c.cache.Set(ctx, fullKey, value, ttl)
}

// Get 获取缓存
func (c *genericCacheImpl[T]) Get(ctx context.Context, key string) (*T, error) {
	fullKey := c.getFullKey(key)
	var value T
	err := c.cache.Get(ctx, fullKey, &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// Delete 删除缓存
func (c *genericCacheImpl[T]) Delete(ctx context.Context, key string) error {
	fullKey := c.getFullKey(key)
	return c.cache.Delete(ctx, fullKey)
}

// Exists 检查缓存是否存在
func (c *genericCacheImpl[T]) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.getFullKey(key)
	return c.cache.Exists(ctx, fullKey)
}

// SetWithKeyFunc 使用键函数设置缓存
func (c *genericCacheImpl[T]) SetWithKeyFunc(ctx context.Context, keyFunc func(*T) string, value *T, ttl time.Duration) error {
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}
	key := keyFunc(value)
	return c.Set(ctx, key, value, ttl)
}

// GetByKeyFunc 使用键函数获取缓存
func (c *genericCacheImpl[T]) GetByKeyFunc(ctx context.Context, keyFunc func() string) (*T, error) {
	key := keyFunc()
	return c.Get(ctx, key)
}

// getFullKey 获取完整的缓存键
func (c *genericCacheImpl[T]) getFullKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// CacheManager 缓存管理器
type CacheManager struct {
	cache  redis.CacheService
	logger logger.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cache redis.CacheService, logger logger.Logger) *CacheManager {
	return &CacheManager{
		cache:  cache,
		logger: logger,
	}
}

// GetUserCache 获取用户缓存
func (m *CacheManager) GetUserCache() GenericCache[entities.User] {
	return NewGenericCache[entities.User](m.cache, m.logger, "user")
}

// GetAPIKeyCache 获取API密钥缓存
func (m *CacheManager) GetAPIKeyCache() GenericCache[entities.APIKey] {
	return NewGenericCache[entities.APIKey](m.cache, m.logger, "api_key")
}

// GetModelCache 获取模型缓存
func (m *CacheManager) GetModelCache() GenericCache[entities.Model] {
	return NewGenericCache[entities.Model](m.cache, m.logger, "model")
}

// GetProviderCache 获取提供商缓存
func (m *CacheManager) GetProviderCache() GenericCache[entities.Provider] {
	return NewGenericCache[entities.Provider](m.cache, m.logger, "provider")
}

// GetQuotaCache 获取配额缓存
func (m *CacheManager) GetQuotaCache() GenericCache[entities.Quota] {
	return NewGenericCache[entities.Quota](m.cache, m.logger, "quota")
}

// CacheHelper 缓存助手
type CacheHelper struct {
	manager *CacheManager
}

// NewCacheHelper 创建缓存助手
func NewCacheHelper(manager *CacheManager) *CacheHelper {
	return &CacheHelper{
		manager: manager,
	}
}

// CacheUser 缓存用户
func (h *CacheHelper) CacheUser(ctx context.Context, user *entities.User, ttl time.Duration) error {
	cache := h.manager.GetUserCache()
	return cache.SetWithKeyFunc(ctx, func(u *entities.User) string {
		return fmt.Sprintf("%d", u.ID)
	}, user, ttl)
}

// GetCachedUser 获取缓存的用户
func (h *CacheHelper) GetCachedUser(ctx context.Context, userID int64) (*entities.User, error) {
	cache := h.manager.GetUserCache()
	return cache.Get(ctx, fmt.Sprintf("%d", userID))
}

// CacheAPIKey 缓存API密钥
func (h *CacheHelper) CacheAPIKey(ctx context.Context, apiKey *entities.APIKey, ttl time.Duration) error {
	cache := h.manager.GetAPIKeyCache()
	return cache.SetWithKeyFunc(ctx, func(ak *entities.APIKey) string {
		return ak.Key
	}, apiKey, ttl)
}

// GetCachedAPIKey 获取缓存的API密钥
func (h *CacheHelper) GetCachedAPIKey(ctx context.Context, key string) (*entities.APIKey, error) {
	cache := h.manager.GetAPIKeyCache()
	return cache.Get(ctx, key)
}

// CacheModel 缓存模型
func (h *CacheHelper) CacheModel(ctx context.Context, model *entities.Model, ttl time.Duration) error {
	cache := h.manager.GetModelCache()
	return cache.SetWithKeyFunc(ctx, func(m *entities.Model) string {
		return fmt.Sprintf("%d", m.ID)
	}, model, ttl)
}

// GetCachedModel 获取缓存的模型
func (h *CacheHelper) GetCachedModel(ctx context.Context, modelID int64) (*entities.Model, error) {
	cache := h.manager.GetModelCache()
	return cache.Get(ctx, fmt.Sprintf("%d", modelID))
}

// InvalidateUser 使用户缓存失效
func (h *CacheHelper) InvalidateUser(ctx context.Context, userID int64) error {
	cache := h.manager.GetUserCache()
	return cache.Delete(ctx, fmt.Sprintf("%d", userID))
}

// InvalidateAPIKey 使API密钥缓存失效
func (h *CacheHelper) InvalidateAPIKey(ctx context.Context, key string) error {
	cache := h.manager.GetAPIKeyCache()
	return cache.Delete(ctx, key)
}

// InvalidateModel 使模型缓存失效
func (h *CacheHelper) InvalidateModel(ctx context.Context, modelID int64) error {
	cache := h.manager.GetModelCache()
	return cache.Delete(ctx, fmt.Sprintf("%d", modelID))
}
