package redis

import (
	"context"
	"fmt"

	"ai-api-gateway/internal/infrastructure/logger"
)

// CacheInvalidationService 缓存失效服务
type CacheInvalidationService struct {
	cache  *CacheService
	client *RedisClient
	logger logger.Logger
}

// NewCacheInvalidationService 创建缓存失效服务
func NewCacheInvalidationService(cache *CacheService, client *RedisClient, logger logger.Logger) *CacheInvalidationService {
	return &CacheInvalidationService{
		cache:  cache,
		client: client,
		logger: logger,
	}
}

// InvalidateUserCache 失效用户相关缓存
func (s *CacheInvalidationService) InvalidateUserCache(ctx context.Context, userID int64, username, email string) error {
	keys := []string{
		GetUserCacheKey(userID),
	}

	// 如果提供了用户名，添加用户名缓存键
	if username != "" {
		keys = append(keys, GetUserByUsernameCacheKey(username))
	}

	// 如果提供了邮箱，添加邮箱缓存键
	if email != "" {
		keys = append(keys, GetUserByEmailCacheKey(email))
	}

	// 添加用户相关的列表缓存
	keys = append(keys,
		GetUserQuotasCacheKey(userID),
		GetUserAPIKeysCacheKey(userID),
		GetActiveUserAPIKeysCacheKey(userID),
		GetActiveQuotasCacheKey(userID),
	)

	// 删除分页缓存（使用模式匹配）
	if err := s.deleteByPattern(ctx, "users:active:page:*"); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Warn("Failed to delete user pagination cache")
	}

	// 删除统计缓存
	keys = append(keys, GetUserCountCacheKey())

	return s.cache.Delete(ctx, keys...)
}

// InvalidateAPIKeyCache 失效API密钥相关缓存
func (s *CacheInvalidationService) InvalidateAPIKeyCache(ctx context.Context, userID int64, apiKeyStr string) error {
	keys := []string{
		GetAPIKeyCacheKey(apiKeyStr),
		GetUserAPIKeysCacheKey(userID),
		GetActiveUserAPIKeysCacheKey(userID),
	}

	return s.cache.Delete(ctx, keys...)
}

// InvalidateModelCache 失效模型相关缓存
func (s *CacheInvalidationService) InvalidateModelCache(ctx context.Context, modelID int64, modelType string) error {
	keys := []string{
		GetModelCacheKey(modelID),
		GetActiveModelsCacheKey(),
		GetAvailableModelsCacheKey(),
	}

	// 如果提供了模型类型，删除对应的类型缓存
	if modelType != "" {
		keys = append(keys, GetModelsByTypeCacheKey(modelType))
	}

	return s.cache.Delete(ctx, keys...)
}

// InvalidateProviderCache 失效提供商相关缓存
func (s *CacheInvalidationService) InvalidateProviderCache(ctx context.Context, providerID int64) error {
	keys := []string{
		GetProviderCacheKey(providerID),
		GetAvailableProvidersCacheKey(),
		GetActiveProvidersCacheKey(),
		GetProvidersNeedingHealthCheckCacheKey(),
	}

	return s.cache.Delete(ctx, keys...)
}

// InvalidateQuotaCache 失效配额相关缓存
func (s *CacheInvalidationService) InvalidateQuotaCache(ctx context.Context, quotaID, userID int64, quotaType, period string) error {
	keys := []string{
		GetQuotaCacheKey(quotaID),
		GetUserQuotasCacheKey(userID),
		GetActiveQuotasCacheKey(userID),
	}

	// 如果提供了配额类型和周期，删除对应的复合查询缓存
	if quotaType != "" && period != "" {
		keys = append(keys, GetQuotaByUserAndTypeCacheKey(userID, quotaType, period))
	}

	// 删除配额使用相关缓存
	if err := s.deleteByPattern(ctx, fmt.Sprintf("quota_usage:user:%d:*", userID)); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id":  userID,
			"quota_id": quotaID,
			"error":    err.Error(),
		}).Warn("Failed to delete quota usage cache")
	}

	// 删除统计缓存
	keys = append(keys, GetQuotaCountCacheKey())

	return s.cache.Delete(ctx, keys...)
}

// InvalidateUsageLogCache 失效使用日志相关缓存
func (s *CacheInvalidationService) InvalidateUsageLogCache(ctx context.Context, userID int64) error {
	// 删除用户使用日志分页缓存
	return s.deleteByPattern(ctx, fmt.Sprintf("usage_logs:user:%d:page:*", userID))
}

// InvalidateQuotaUsageCache 失效配额使用相关缓存
func (s *CacheInvalidationService) InvalidateQuotaUsageCache(ctx context.Context, userID int64) error {
	// 删除用户配额使用分页缓存
	return s.deleteByPattern(ctx, fmt.Sprintf("quota_usage:user:%d:page:*", userID))
}

// InvalidateAllModelCache 失效所有模型相关缓存
func (s *CacheInvalidationService) InvalidateAllModelCache(ctx context.Context) error {
	// 删除所有模型相关缓存
	return s.deleteByPattern(ctx, "models:*")
}

// InvalidateAllProviderCache 失效所有提供商相关缓存
func (s *CacheInvalidationService) InvalidateAllProviderCache(ctx context.Context) error {
	// 删除所有提供商相关缓存
	return s.deleteByPattern(ctx, "providers:*")
}

// InvalidateAllUserCache 失效所有用户相关缓存
func (s *CacheInvalidationService) InvalidateAllUserCache(ctx context.Context) error {
	// 删除所有用户相关缓存
	patterns := []string{
		"user:*",
		"users:*",
		"stats:users:*",
	}

	for _, pattern := range patterns {
		if err := s.deleteByPattern(ctx, pattern); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"pattern": pattern,
				"error":   err.Error(),
			}).Error("Failed to delete cache by pattern")
		}
	}

	return nil
}

// deleteByPattern 根据模式删除缓存键
func (s *CacheInvalidationService) deleteByPattern(ctx context.Context, pattern string) error {
	if !s.cache.IsEnabled() {
		return nil
	}

	// 获取匹配的键
	keys, err := s.client.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	// 删除匹配的键
	if err := s.cache.Delete(ctx, keys...); err != nil {
		return fmt.Errorf("failed to delete keys by pattern %s: %w", pattern, err)
	}

	s.logger.WithFields(map[string]interface{}{
		"pattern":    pattern,
		"keys_count": len(keys),
	}).Debug("Cache keys deleted by pattern")

	return nil
}

// BatchInvalidate 批量失效缓存
func (s *CacheInvalidationService) BatchInvalidate(ctx context.Context, operations []InvalidationOperation) error {
	var allKeys []string
	var patterns []string

	for _, op := range operations {
		switch op.Type {
		case InvalidationTypeKeys:
			allKeys = append(allKeys, op.Keys...)
		case InvalidationTypePattern:
			patterns = append(patterns, op.Pattern)
		}
	}

	// 删除直接指定的键
	if len(allKeys) > 0 {
		if err := s.cache.Delete(ctx, allKeys...); err != nil {
			return fmt.Errorf("failed to delete keys in batch: %w", err)
		}
	}

	// 删除模式匹配的键
	for _, pattern := range patterns {
		if err := s.deleteByPattern(ctx, pattern); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"pattern": pattern,
				"error":   err.Error(),
			}).Error("Failed to delete cache by pattern in batch")
		}
	}

	return nil
}

// InvalidationOperation 失效操作
type InvalidationOperation struct {
	Type    InvalidationType
	Keys    []string
	Pattern string
}

// InvalidationType 失效类型
type InvalidationType string

const (
	InvalidationTypeKeys    InvalidationType = "keys"
	InvalidationTypePattern InvalidationType = "pattern"
)

// NewKeysInvalidation 创建键失效操作
func NewKeysInvalidation(keys ...string) InvalidationOperation {
	return InvalidationOperation{
		Type: InvalidationTypeKeys,
		Keys: keys,
	}
}

// NewPatternInvalidation 创建模式失效操作
func NewPatternInvalidation(pattern string) InvalidationOperation {
	return InvalidationOperation{
		Type:    InvalidationTypePattern,
		Pattern: pattern,
	}
}
