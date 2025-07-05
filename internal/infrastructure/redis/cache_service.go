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

// 扩展的缓存方法

// 用户查询缓存方法

// SetUserByUsername 缓存按用户名查询的用户信息
func (c *CacheService) SetUserByUsername(ctx context.Context, username string, user *entities.User) error {
	key := GetUserByUsernameCacheKey(username)
	ttl := viper.GetDuration("cache.query.user_lookup_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.user_ttl") // 向后兼容
	}
	return c.Set(ctx, key, user, ttl)
}

// GetUserByUsername 获取按用户名查询的用户缓存
func (c *CacheService) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	key := GetUserByUsernameCacheKey(username)
	var user entities.User
	if err := c.Get(ctx, key, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SetUserByEmail 缓存按邮箱查询的用户信息
func (c *CacheService) SetUserByEmail(ctx context.Context, email string, user *entities.User) error {
	key := GetUserByEmailCacheKey(email)
	ttl := viper.GetDuration("cache.query.user_lookup_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.user_ttl") // 向后兼容
	}
	return c.Set(ctx, key, user, ttl)
}

// GetUserByEmail 获取按邮箱查询的用户缓存
func (c *CacheService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	key := GetUserByEmailCacheKey(email)
	var user entities.User
	if err := c.Get(ctx, key, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// 模型列表缓存方法

// SetActiveModels 缓存活跃模型列表
func (c *CacheService) SetActiveModels(ctx context.Context, models []*entities.Model) error {
	key := GetActiveModelsCacheKey()
	ttl := viper.GetDuration("cache.query.model_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.model_ttl") // 向后兼容
	}
	return c.Set(ctx, key, models, ttl)
}

// GetActiveModels 获取活跃模型列表缓存
func (c *CacheService) GetActiveModels(ctx context.Context) ([]*entities.Model, error) {
	key := GetActiveModelsCacheKey()
	var models []*entities.Model
	if err := c.Get(ctx, key, &models); err != nil {
		return nil, err
	}
	return models, nil
}

// SetModelsByType 缓存按类型查询的模型列表
func (c *CacheService) SetModelsByType(ctx context.Context, modelType string, models []*entities.Model) error {
	key := GetModelsByTypeCacheKey(modelType)
	ttl := viper.GetDuration("cache.query.model_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.model_ttl") // 向后兼容
	}
	return c.Set(ctx, key, models, ttl)
}

// GetModelsByType 获取按类型查询的模型列表缓存
func (c *CacheService) GetModelsByType(ctx context.Context, modelType string) ([]*entities.Model, error) {
	key := GetModelsByTypeCacheKey(modelType)
	var models []*entities.Model
	if err := c.Get(ctx, key, &models); err != nil {
		return nil, err
	}
	return models, nil
}

// SetAvailableModels 缓存可用模型列表
func (c *CacheService) SetAvailableModels(ctx context.Context, models []*entities.Model) error {
	key := GetAvailableModelsCacheKey()
	ttl := viper.GetDuration("cache.query.model_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.model_ttl") // 向后兼容
	}
	return c.Set(ctx, key, models, ttl)
}

// GetAvailableModels 获取可用模型列表缓存
func (c *CacheService) GetAvailableModels(ctx context.Context) ([]*entities.Model, error) {
	key := GetAvailableModelsCacheKey()
	var models []*entities.Model
	if err := c.Get(ctx, key, &models); err != nil {
		return nil, err
	}
	return models, nil
}

// 提供商列表缓存方法

// SetAvailableProviders 缓存可用提供商列表
func (c *CacheService) SetAvailableProviders(ctx context.Context, providers []*entities.Provider) error {
	key := GetAvailableProvidersCacheKey()
	ttl := viper.GetDuration("cache.query.provider_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.provider_ttl") // 向后兼容
	}
	return c.Set(ctx, key, providers, ttl)
}

// GetAvailableProviders 获取可用提供商列表缓存
func (c *CacheService) GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error) {
	key := GetAvailableProvidersCacheKey()
	var providers []*entities.Provider
	if err := c.Get(ctx, key, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

// SetActiveProviders 缓存活跃提供商列表
func (c *CacheService) SetActiveProviders(ctx context.Context, providers []*entities.Provider) error {
	key := GetActiveProvidersCacheKey()
	ttl := viper.GetDuration("cache.query.provider_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.provider_ttl") // 向后兼容
	}
	return c.Set(ctx, key, providers, ttl)
}

// GetActiveProviders 获取活跃提供商列表缓存
func (c *CacheService) GetActiveProviders(ctx context.Context) ([]*entities.Provider, error) {
	key := GetActiveProvidersCacheKey()
	var providers []*entities.Provider
	if err := c.Get(ctx, key, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

// SetProvidersNeedingHealthCheck 缓存需要健康检查的提供商列表
func (c *CacheService) SetProvidersNeedingHealthCheck(ctx context.Context, providers []*entities.Provider) error {
	key := GetProvidersNeedingHealthCheckCacheKey()
	ttl := viper.GetDuration("cache.query.provider_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.provider_ttl") // 向后兼容
	}
	return c.Set(ctx, key, providers, ttl)
}

// GetProvidersNeedingHealthCheck 获取需要健康检查的提供商列表缓存
func (c *CacheService) GetProvidersNeedingHealthCheck(ctx context.Context) ([]*entities.Provider, error) {
	key := GetProvidersNeedingHealthCheckCacheKey()
	var providers []*entities.Provider
	if err := c.Get(ctx, key, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

// API密钥列表缓存方法

// SetUserAPIKeys 缓存用户API密钥列表
func (c *CacheService) SetUserAPIKeys(ctx context.Context, userID int64, apiKeys []*entities.APIKey) error {
	key := GetUserAPIKeysCacheKey(userID)
	ttl := viper.GetDuration("cache.query.api_key_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.api_key_ttl") // 向后兼容
	}
	return c.Set(ctx, key, apiKeys, ttl)
}

// GetUserAPIKeys 获取用户API密钥列表缓存
func (c *CacheService) GetUserAPIKeys(ctx context.Context, userID int64) ([]*entities.APIKey, error) {
	key := GetUserAPIKeysCacheKey(userID)
	var apiKeys []*entities.APIKey
	if err := c.Get(ctx, key, &apiKeys); err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// SetActiveUserAPIKeys 缓存用户活跃API密钥列表
func (c *CacheService) SetActiveUserAPIKeys(ctx context.Context, userID int64, apiKeys []*entities.APIKey) error {
	key := GetActiveUserAPIKeysCacheKey(userID)
	ttl := viper.GetDuration("cache.query.api_key_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.api_key_ttl") // 向后兼容
	}
	return c.Set(ctx, key, apiKeys, ttl)
}

// GetActiveUserAPIKeys 获取用户活跃API密钥列表缓存
func (c *CacheService) GetActiveUserAPIKeys(ctx context.Context, userID int64) ([]*entities.APIKey, error) {
	key := GetActiveUserAPIKeysCacheKey(userID)
	var apiKeys []*entities.APIKey
	if err := c.Get(ctx, key, &apiKeys); err != nil {
		return nil, err
	}
	return apiKeys, nil
}

// 配额使用情况缓存方法

// SetQuotaUsage 缓存配额使用情况
func (c *CacheService) SetQuotaUsage(ctx context.Context, userID int64, quotaType, period string, usage *entities.QuotaUsage) error {
	key := GetQuotaUsageCacheKey(userID, quotaType, period)
	ttl := viper.GetDuration("cache.query.quota_usage_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.quota_ttl") // 向后兼容
	}
	return c.Set(ctx, key, usage, ttl)
}

// GetQuotaUsage 获取配额使用情况缓存
func (c *CacheService) GetQuotaUsage(ctx context.Context, userID int64, quotaType, period string) (*entities.QuotaUsage, error) {
	key := GetQuotaUsageCacheKey(userID, quotaType, period)
	var usage entities.QuotaUsage
	if err := c.Get(ctx, key, &usage); err != nil {
		return nil, err
	}
	return &usage, nil
}

// DeleteQuotaUsage 删除配额使用情况缓存
func (c *CacheService) DeleteQuotaUsage(ctx context.Context, userID int64, quotaType, period string) error {
	key := GetQuotaUsageCacheKey(userID, quotaType, period)
	return c.Delete(ctx, key)
}

// SetActiveQuotas 缓存活跃配额列表
func (c *CacheService) SetActiveQuotas(ctx context.Context, userID int64, quotas []*entities.Quota) error {
	key := GetActiveQuotasCacheKey(userID)
	ttl := viper.GetDuration("cache.query.user_quota_list_ttl")
	if ttl == 0 {
		ttl = viper.GetDuration("cache.quota_ttl") // 向后兼容
	}
	return c.Set(ctx, key, quotas, ttl)
}

// GetActiveQuotas 获取活跃配额列表缓存
func (c *CacheService) GetActiveQuotas(ctx context.Context, userID int64) ([]*entities.Quota, error) {
	key := GetActiveQuotasCacheKey(userID)
	var quotas []*entities.Quota
	if err := c.Get(ctx, key, &quotas); err != nil {
		return nil, err
	}
	return quotas, nil
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

// 扩展的缓存键生成函数

// GetUserByUsernameCacheKey 生成按用户名查询用户的缓存键
func GetUserByUsernameCacheKey(username string) string {
	return fmt.Sprintf("user:username:%s", username)
}

// GetUserByEmailCacheKey 生成按邮箱查询用户的缓存键
func GetUserByEmailCacheKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// GetActiveModelsCacheKey 生成活跃模型列表缓存键
func GetActiveModelsCacheKey() string {
	return "models:active"
}

// GetModelsByTypeCacheKey 生成按类型查询模型的缓存键
func GetModelsByTypeCacheKey(modelType string) string {
	return fmt.Sprintf("models:type:%s", modelType)
}

// GetAvailableModelsCacheKey 生成可用模型列表缓存键
func GetAvailableModelsCacheKey() string {
	return "models:available"
}

// GetAvailableProvidersCacheKey 生成可用提供商列表缓存键
func GetAvailableProvidersCacheKey() string {
	return "providers:available"
}

// GetActiveProvidersCacheKey 生成活跃提供商列表缓存键
func GetActiveProvidersCacheKey() string {
	return "providers:active"
}

// GetProvidersNeedingHealthCheckCacheKey 生成需要健康检查的提供商列表缓存键
func GetProvidersNeedingHealthCheckCacheKey() string {
	return "providers:health_check_needed"
}

// GetUserAPIKeysCacheKey 生成用户API密钥列表缓存键
func GetUserAPIKeysCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d:apikeys", userID)
}

// GetActiveUserAPIKeysCacheKey 生成用户活跃API密钥列表缓存键
func GetActiveUserAPIKeysCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d:apikeys:active", userID)
}

// GetQuotaByUserAndTypeCacheKey 生成按用户和类型查询配额的缓存键
func GetQuotaByUserAndTypeCacheKey(userID int64, quotaType, period string) string {
	return fmt.Sprintf("quota:user:%d:type:%s:period:%s", userID, quotaType, period)
}

// GetActiveQuotasCacheKey 生成活跃配额列表缓存键
func GetActiveQuotasCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d:quotas:active", userID)
}

// GetUsageLogsByUserCacheKey 生成用户使用日志列表缓存键
func GetUsageLogsByUserCacheKey(userID int64, offset, limit int) string {
	return fmt.Sprintf("usage_logs:user:%d:page:%d:%d", userID, offset, limit)
}

// GetQuotaUsageByUserCacheKey 生成用户配额使用列表缓存键
func GetQuotaUsageByUserCacheKey(userID int64, offset, limit int) string {
	return fmt.Sprintf("quota_usage:user:%d:page:%d:%d", userID, offset, limit)
}

// 统计缓存键

// GetUserCountCacheKey 生成用户总数缓存键
func GetUserCountCacheKey() string {
	return "stats:users:count"
}

// GetQuotaCountCacheKey 生成配额总数缓存键
func GetQuotaCountCacheKey() string {
	return "stats:quotas:count"
}

// GetActiveUsersCacheKey 生成活跃用户分页列表缓存键
func GetActiveUsersCacheKey(offset, limit int) string {
	return fmt.Sprintf("users:active:page:%d:%d", offset, limit)
}
