package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"
)

// CacheService 缓存服务
type CacheService struct {
	client  *RedisClient
	logger  logger.Logger
	enabled bool
}

// NewCacheService 创建缓存服务
func NewCacheService(client *RedisClient, logger logger.Logger) *CacheService {
	return &CacheService{
		client:  client,
		logger:  logger,
		enabled: viper.GetBool("cache.enabled"),
	}
}

// IsEnabled 检查缓存是否启用
func (c *CacheService) IsEnabled() bool {
	return c.enabled
}

// Set 设置缓存
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.WithFields(map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to marshal cache value")
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"key":   key,
			"ttl":   ttl,
			"error": err.Error(),
		}).Error("Failed to set cache")
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"key": key,
		"ttl": ttl,
	}).Debug("Cache set successfully")

	return nil
}

// Get 获取缓存
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	if !c.enabled {
		return redis.Nil
	}

	data, err := c.client.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			c.logger.WithFields(map[string]interface{}{
				"key": key,
			}).Debug("Cache miss")
		} else {
			c.logger.WithFields(map[string]interface{}{
				"key":   key,
				"error": err.Error(),
			}).Error("Failed to get cache")
		}
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"key":   key,
			"error": err.Error(),
		}).Error("Failed to unmarshal cache value")
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"key": key,
	}).Debug("Cache hit")

	return nil
}

// Delete 删除缓存
func (c *CacheService) Delete(ctx context.Context, keys ...string) error {
	if !c.enabled {
		return nil
	}

	if err := c.client.Del(ctx, keys...); err != nil {
		c.logger.WithFields(map[string]interface{}{
			"keys":  keys,
			"error": err.Error(),
		}).Error("Failed to delete cache")
		return err
	}

	c.logger.WithFields(map[string]interface{}{
		"keys": keys,
	}).Debug("Cache deleted successfully")

	return nil
}

// Exists 检查缓存是否存在
func (c *CacheService) Exists(ctx context.Context, keys ...string) (int64, error) {
	if !c.enabled {
		return 0, nil
	}

	return c.client.Exists(ctx, keys...)
}

// TTL 获取缓存剩余时间
func (c *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if !c.enabled {
		return 0, nil
	}

	return c.client.TTL(ctx, key)
}

// Expire 设置缓存过期时间
func (c *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	return c.client.Expire(ctx, key, ttl)
}

// 用户缓存相关方法

// SetUser 缓存用户信息
func (c *CacheService) SetUser(ctx context.Context, user *entities.User) error {
	key := GetUserCacheKey(user.ID)
	ttl := viper.GetDuration("cache.user_ttl")
	return c.Set(ctx, key, user, ttl)
}

// GetUser 获取用户缓存
func (c *CacheService) GetUser(ctx context.Context, userID int64) (*entities.User, error) {
	key := GetUserCacheKey(userID)
	var user entities.User
	if err := c.Get(ctx, key, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser 删除用户缓存
func (c *CacheService) DeleteUser(ctx context.Context, userID int64) error {
	key := GetUserCacheKey(userID)
	return c.Delete(ctx, key)
}

// API密钥缓存相关方法

// SetAPIKey 缓存API密钥信息
func (c *CacheService) SetAPIKey(ctx context.Context, apiKey *entities.APIKey) error {
	key := GetAPIKeyCacheKey(apiKey.Key)
	ttl := viper.GetDuration("cache.api_key_ttl")
	return c.Set(ctx, key, apiKey, ttl)
}

// GetAPIKey 获取API密钥缓存
func (c *CacheService) GetAPIKey(ctx context.Context, apiKeyStr string) (*entities.APIKey, error) {
	key := GetAPIKeyCacheKey(apiKeyStr)
	var apiKey entities.APIKey
	if err := c.Get(ctx, key, &apiKey); err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// DeleteAPIKey 删除API密钥缓存
func (c *CacheService) DeleteAPIKey(ctx context.Context, apiKeyStr string) error {
	key := GetAPIKeyCacheKey(apiKeyStr)
	return c.Delete(ctx, key)
}

// 模型缓存相关方法

// SetModel 缓存模型信息
func (c *CacheService) SetModel(ctx context.Context, model *entities.Model) error {
	key := GetModelCacheKey(model.ID)
	ttl := viper.GetDuration("cache.model_ttl")
	return c.Set(ctx, key, model, ttl)
}

// GetModel 获取模型缓存
func (c *CacheService) GetModel(ctx context.Context, modelID int64) (*entities.Model, error) {
	key := GetModelCacheKey(modelID)
	var model entities.Model
	if err := c.Get(ctx, key, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

// DeleteModel 删除模型缓存
func (c *CacheService) DeleteModel(ctx context.Context, modelID int64) error {
	key := GetModelCacheKey(modelID)
	return c.Delete(ctx, key)
}

// 提供商缓存相关方法

// SetProvider 缓存提供商信息
func (c *CacheService) SetProvider(ctx context.Context, provider *entities.Provider) error {
	key := GetProviderCacheKey(provider.ID)
	ttl := viper.GetDuration("cache.provider_ttl")
	return c.Set(ctx, key, provider, ttl)
}

// GetProvider 获取提供商缓存
func (c *CacheService) GetProvider(ctx context.Context, providerID int64) (*entities.Provider, error) {
	key := GetProviderCacheKey(providerID)
	var provider entities.Provider
	if err := c.Get(ctx, key, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

// DeleteProvider 删除提供商缓存
func (c *CacheService) DeleteProvider(ctx context.Context, providerID int64) error {
	key := GetProviderCacheKey(providerID)
	return c.Delete(ctx, key)
}

// 配额缓存相关方法

// SetQuota 缓存配额信息
func (c *CacheService) SetQuota(ctx context.Context, quota *entities.Quota) error {
	key := GetQuotaCacheKey(quota.ID)
	ttl := viper.GetDuration("cache.quota_ttl")
	return c.Set(ctx, key, quota, ttl)
}

// GetQuota 获取配额缓存
func (c *CacheService) GetQuota(ctx context.Context, quotaID int64) (*entities.Quota, error) {
	key := GetQuotaCacheKey(quotaID)
	var quota entities.Quota
	if err := c.Get(ctx, key, &quota); err != nil {
		return nil, err
	}
	return &quota, nil
}

// DeleteQuota 删除配额缓存
func (c *CacheService) DeleteQuota(ctx context.Context, quotaID int64) error {
	key := GetQuotaCacheKey(quotaID)
	return c.Delete(ctx, key)
}

// SetUserQuotas 缓存用户配额列表
func (c *CacheService) SetUserQuotas(ctx context.Context, userID int64, quotas []*entities.Quota) error {
	key := GetUserQuotasCacheKey(userID)
	ttl := viper.GetDuration("cache.quota_ttl")
	return c.Set(ctx, key, quotas, ttl)
}

// GetUserQuotas 获取用户配额列表缓存
func (c *CacheService) GetUserQuotas(ctx context.Context, userID int64) ([]*entities.Quota, error) {
	key := GetUserQuotasCacheKey(userID)
	var quotas []*entities.Quota
	if err := c.Get(ctx, key, &quotas); err != nil {
		return nil, err
	}
	return quotas, nil
}

// DeleteUserQuotas 删除用户配额列表缓存
func (c *CacheService) DeleteUserQuotas(ctx context.Context, userID int64) error {
	key := GetUserQuotasCacheKey(userID)
	return c.Delete(ctx, key)
}

// 缓存键生成函数

// GetUserCacheKey 生成用户缓存键
func GetUserCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

// GetAPIKeyCacheKey 生成API密钥缓存键
func GetAPIKeyCacheKey(apiKeyStr string) string {
	return fmt.Sprintf("apikey:%s", apiKeyStr)
}

// GetModelCacheKey 生成模型缓存键
func GetModelCacheKey(modelID int64) string {
	return fmt.Sprintf("model:%d", modelID)
}

// GetProviderCacheKey 生成提供商缓存键
func GetProviderCacheKey(providerID int64) string {
	return fmt.Sprintf("provider:%d", providerID)
}

// GetQuotaCacheKey 生成配额缓存键
func GetQuotaCacheKey(quotaID int64) string {
	return fmt.Sprintf("quota:%d", quotaID)
}

// GetUserQuotasCacheKey 生成用户配额列表缓存键
func GetUserQuotasCacheKey(userID int64) string {
	return fmt.Sprintf("user_quotas:%d", userID)
}

// GetQuotaUsageCacheKey 生成配额使用缓存键
func GetQuotaUsageCacheKey(userID int64, quotaType, period string) string {
	return fmt.Sprintf("quota_usage:%d:%s:%s", userID, quotaType, period)
}
