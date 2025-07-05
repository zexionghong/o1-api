package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-api-gateway/internal/infrastructure/logger"
)

// CacheManager 缓存管理器
type CacheManager struct {
	cache               *CacheService
	client              *RedisClient
	invalidationService *CacheInvalidationService
	logger              logger.Logger
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(
	cache *CacheService,
	client *RedisClient,
	invalidationService *CacheInvalidationService,
	logger logger.Logger,
) *CacheManager {
	return &CacheManager{
		cache:               cache,
		client:              client,
		invalidationService: invalidationService,
		logger:              logger,
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	TotalKeys     int64                    `json:"total_keys"`
	KeysByPattern map[string]int64         `json:"keys_by_pattern"`
	MemoryUsage   string                   `json:"memory_usage"`
	HitRate       float64                  `json:"hit_rate"`
	Timestamp     time.Time                `json:"timestamp"`
	TTLStats      map[string]time.Duration `json:"ttl_stats"`
}

// GetCacheStats 获取缓存统计信息
func (m *CacheManager) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	if !m.cache.IsEnabled() {
		return &CacheStats{
			Timestamp: time.Now(),
		}, nil
	}

	stats := &CacheStats{
		KeysByPattern: make(map[string]int64),
		TTLStats:      make(map[string]time.Duration),
		Timestamp:     time.Now(),
	}

	// 获取所有键
	allKeys, err := m.client.Keys(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}

	stats.TotalKeys = int64(len(allKeys))

	// 按模式统计键数量
	patterns := []string{
		"user:*",
		"apikey:*",
		"model:*",
		"provider:*",
		"quota:*",
		"models:*",
		"providers:*",
		"usage_logs:*",
		"quota_usage:*",
		"stats:*",
	}

	for _, pattern := range patterns {
		count := int64(0)
		for _, key := range allKeys {
			if matched, _ := m.matchPattern(key, pattern); matched {
				count++
			}
		}
		if count > 0 {
			stats.KeysByPattern[pattern] = count
		}
	}

	// 获取内存使用情况（如果Redis支持）
	if memInfo, err := m.client.GetClient().Info(ctx, "memory").Result(); err == nil {
		lines := strings.Split(memInfo, "\r\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory_human:") {
				stats.MemoryUsage = strings.TrimPrefix(line, "used_memory_human:")
				break
			}
		}
	}

	// 采样一些键的TTL信息
	sampleKeys := allKeys
	if len(sampleKeys) > 10 {
		sampleKeys = sampleKeys[:10] // 只采样前10个键
	}

	for _, key := range sampleKeys {
		if ttl, err := m.client.TTL(ctx, key); err == nil && ttl > 0 {
			pattern := m.getKeyPattern(key)
			if pattern != "" {
				stats.TTLStats[pattern] = ttl
			}
		}
	}

	return stats, nil
}

// ClearCache 清除缓存
func (m *CacheManager) ClearCache(ctx context.Context, patterns ...string) error {
	if !m.cache.IsEnabled() {
		return nil
	}

	if len(patterns) == 0 {
		// 清除所有缓存
		patterns = []string{"*"}
	}

	var operations []InvalidationOperation
	for _, pattern := range patterns {
		operations = append(operations, NewPatternInvalidation(pattern))
	}

	return m.invalidationService.BatchInvalidate(ctx, operations)
}

// WarmupCache 预热缓存
func (m *CacheManager) WarmupCache(ctx context.Context) error {
	if !m.cache.IsEnabled() {
		return nil
	}

	m.logger.Info("Starting cache warmup")

	// 这里可以添加预热逻辑，例如：
	// 1. 预加载常用的模型列表
	// 2. 预加载活跃的提供商列表
	// 3. 预加载系统配置等

	// 示例：预热活跃模型列表
	// 注意：这需要依赖具体的Repository实现
	// 这里只是展示概念，实际实现需要注入相应的服务

	m.logger.Info("Cache warmup completed")
	return nil
}

// RefreshCache 刷新特定类型的缓存
func (m *CacheManager) RefreshCache(ctx context.Context, cacheType string) error {
	if !m.cache.IsEnabled() {
		return nil
	}

	switch cacheType {
	case "models":
		return m.invalidationService.InvalidateAllModelCache(ctx)
	case "providers":
		return m.invalidationService.InvalidateAllProviderCache(ctx)
	case "users":
		return m.invalidationService.InvalidateAllUserCache(ctx)
	case "all":
		return m.ClearCache(ctx)
	default:
		return fmt.Errorf("unknown cache type: %s", cacheType)
	}
}

// GetCacheHealth 获取缓存健康状态
func (m *CacheManager) GetCacheHealth(ctx context.Context) (*CacheHealth, error) {
	health := &CacheHealth{
		Enabled:   m.cache.IsEnabled(),
		Timestamp: time.Now(),
	}

	if !health.Enabled {
		health.Status = "disabled"
		return health, nil
	}

	// 测试Redis连接
	if err := m.client.GetClient().Ping(ctx).Err(); err != nil {
		health.Status = "unhealthy"
		health.Error = err.Error()
		return health, nil
	}

	health.Status = "healthy"

	// 获取连接池信息
	poolStats := m.client.GetClient().PoolStats()
	health.ConnectionPool = &ConnectionPoolStats{
		TotalConns: poolStats.TotalConns,
		IdleConns:  poolStats.IdleConns,
		StaleConns: poolStats.StaleConns,
	}

	return health, nil
}

// CacheHealth 缓存健康状态
type CacheHealth struct {
	Enabled        bool                 `json:"enabled"`
	Status         string               `json:"status"`
	Error          string               `json:"error,omitempty"`
	ConnectionPool *ConnectionPoolStats `json:"connection_pool,omitempty"`
	Timestamp      time.Time            `json:"timestamp"`
}

// ConnectionPoolStats 连接池统计
type ConnectionPoolStats struct {
	TotalConns uint32 `json:"total_conns"`
	IdleConns  uint32 `json:"idle_conns"`
	StaleConns uint32 `json:"stale_conns"`
}

// matchPattern 简单的模式匹配
func (m *CacheManager) matchPattern(key, pattern string) (bool, error) {
	if pattern == "*" {
		return true, nil
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix), nil
	}

	return key == pattern, nil
}

// getKeyPattern 根据键获取对应的模式
func (m *CacheManager) getKeyPattern(key string) string {
	patterns := map[string]string{
		"user:":        "user:*",
		"apikey:":      "apikey:*",
		"model:":       "model:*",
		"provider:":    "provider:*",
		"quota:":       "quota:*",
		"models:":      "models:*",
		"providers:":   "providers:*",
		"usage_logs:":  "usage_logs:*",
		"quota_usage:": "quota_usage:*",
		"stats:":       "stats:*",
	}

	for prefix, pattern := range patterns {
		if strings.HasPrefix(key, prefix) {
			return pattern
		}
	}

	return "other"
}

// MonitorCache 监控缓存性能
func (m *CacheManager) MonitorCache(ctx context.Context, interval time.Duration) {
	if !m.cache.IsEnabled() {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats, err := m.GetCacheStats(ctx)
			if err != nil {
				m.logger.WithFields(map[string]interface{}{
					"error": err.Error(),
				}).Error("Failed to get cache stats")
				continue
			}

			m.logger.WithFields(map[string]interface{}{
				"total_keys":     stats.TotalKeys,
				"memory_usage":   stats.MemoryUsage,
				"keys_by_pattern": stats.KeysByPattern,
			}).Info("Cache monitoring stats")
		}
	}
}
