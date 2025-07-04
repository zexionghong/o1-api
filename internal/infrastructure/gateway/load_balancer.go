package gateway

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"
)

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	StrategyRoundRobin     LoadBalanceStrategy = "round_robin"
	StrategyWeighted       LoadBalanceStrategy = "weighted"
	StrategyLeastConnections LoadBalanceStrategy = "least_connections"
	StrategyRandom         LoadBalanceStrategy = "random"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// SelectProvider 选择提供商
	SelectProvider(ctx context.Context, providers []*entities.Provider) (*entities.Provider, error)
	
	// RecordResponse 记录响应结果
	RecordResponse(ctx context.Context, providerID int64, success bool, duration time.Duration) error
	
	// GetProviderStats 获取提供商统计信息
	GetProviderStats(ctx context.Context, providerID int64) (*ProviderStats, error)
	
	// SetStrategy 设置负载均衡策略
	SetStrategy(strategy LoadBalanceStrategy)
}

// ProviderStats 提供商统计信息
type ProviderStats struct {
	ProviderID         int64         `json:"provider_id"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	SuccessRate        float64       `json:"success_rate"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	CurrentConnections int           `json:"current_connections"`
	Weight             int           `json:"weight"`
	IsHealthy          bool          `json:"is_healthy"`
	LastUsed           time.Time     `json:"last_used"`
}

// loadBalancerImpl 负载均衡器实现
type loadBalancerImpl struct {
	strategy    LoadBalanceStrategy
	stats       map[int64]*ProviderStats
	roundRobin  map[string]int // 轮询计数器
	mu          sync.RWMutex
	logger      logger.Logger
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(strategy LoadBalanceStrategy, logger logger.Logger) LoadBalancer {
	return &loadBalancerImpl{
		strategy:   strategy,
		stats:      make(map[int64]*ProviderStats),
		roundRobin: make(map[string]int),
		logger:     logger,
	}
}

// SelectProvider 选择提供商
func (lb *loadBalancerImpl) SelectProvider(ctx context.Context, providers []*entities.Provider) (*entities.Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}
	
	// 过滤健康的提供商
	healthyProviders := make([]*entities.Provider, 0, len(providers))
	for _, provider := range providers {
		if provider.IsAvailable() {
			healthyProviders = append(healthyProviders, provider)
		}
	}
	
	if len(healthyProviders) == 0 {
		return nil, fmt.Errorf("no healthy providers available")
	}
	
	// 根据策略选择提供商
	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.selectRoundRobin(healthyProviders), nil
	case StrategyWeighted:
		return lb.selectWeighted(healthyProviders), nil
	case StrategyLeastConnections:
		return lb.selectLeastConnections(healthyProviders), nil
	case StrategyRandom:
		return lb.selectRandom(healthyProviders), nil
	default:
		return lb.selectRoundRobin(healthyProviders), nil
	}
}

// selectRoundRobin 轮询选择
func (lb *loadBalancerImpl) selectRoundRobin(providers []*entities.Provider) *entities.Provider {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	key := "default"
	count := lb.roundRobin[key]
	provider := providers[count%len(providers)]
	lb.roundRobin[key] = count + 1
	
	return provider
}

// selectWeighted 加权选择
func (lb *loadBalancerImpl) selectWeighted(providers []*entities.Provider) *entities.Provider {
	// 计算总权重
	totalWeight := 0
	for _, provider := range providers {
		totalWeight += provider.Priority
	}
	
	if totalWeight == 0 {
		return lb.selectRoundRobin(providers)
	}
	
	// 随机选择
	random := rand.Intn(totalWeight)
	currentWeight := 0
	
	for _, provider := range providers {
		currentWeight += provider.Priority
		if random < currentWeight {
			return provider
		}
	}
	
	return providers[0]
}

// selectLeastConnections 最少连接选择
func (lb *loadBalancerImpl) selectLeastConnections(providers []*entities.Provider) *entities.Provider {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	var selectedProvider *entities.Provider
	minConnections := int(^uint(0) >> 1) // 最大int值
	
	for _, provider := range providers {
		stats := lb.getProviderStats(provider.ID)
		if stats.CurrentConnections < minConnections {
			minConnections = stats.CurrentConnections
			selectedProvider = provider
		}
	}
	
	if selectedProvider == nil {
		return providers[0]
	}
	
	return selectedProvider
}

// selectRandom 随机选择
func (lb *loadBalancerImpl) selectRandom(providers []*entities.Provider) *entities.Provider {
	index := rand.Intn(len(providers))
	return providers[index]
}

// RecordResponse 记录响应结果
func (lb *loadBalancerImpl) RecordResponse(ctx context.Context, providerID int64, success bool, duration time.Duration) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	stats := lb.getProviderStats(providerID)
	stats.TotalRequests++
	stats.LastUsed = time.Now()
	
	if success {
		stats.SuccessfulRequests++
	} else {
		stats.FailedRequests++
	}
	
	// 更新成功率
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests)
	}
	
	// 更新平均响应时间（简单移动平均）
	if stats.TotalRequests == 1 {
		stats.AvgResponseTime = duration
	} else {
		// 使用指数移动平均
		alpha := 0.1
		stats.AvgResponseTime = time.Duration(float64(stats.AvgResponseTime)*(1-alpha) + float64(duration)*alpha)
	}
	
	lb.logger.WithFields(map[string]interface{}{
		"provider_id":    providerID,
		"success":        success,
		"duration":       duration,
		"total_requests": stats.TotalRequests,
		"success_rate":   stats.SuccessRate,
	}).Debug("Recorded provider response")
	
	return nil
}

// GetProviderStats 获取提供商统计信息
func (lb *loadBalancerImpl) GetProviderStats(ctx context.Context, providerID int64) (*ProviderStats, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	stats := lb.getProviderStats(providerID)
	
	// 返回副本
	return &ProviderStats{
		ProviderID:         stats.ProviderID,
		TotalRequests:      stats.TotalRequests,
		SuccessfulRequests: stats.SuccessfulRequests,
		FailedRequests:     stats.FailedRequests,
		SuccessRate:        stats.SuccessRate,
		AvgResponseTime:    stats.AvgResponseTime,
		CurrentConnections: stats.CurrentConnections,
		Weight:             stats.Weight,
		IsHealthy:          stats.IsHealthy,
		LastUsed:           stats.LastUsed,
	}, nil
}

// SetStrategy 设置负载均衡策略
func (lb *loadBalancerImpl) SetStrategy(strategy LoadBalanceStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.strategy = strategy
	lb.logger.WithField("strategy", strategy).Info("Load balancer strategy updated")
}

// getProviderStats 获取提供商统计信息（内部方法，需要持有锁）
func (lb *loadBalancerImpl) getProviderStats(providerID int64) *ProviderStats {
	stats, exists := lb.stats[providerID]
	if !exists {
		stats = &ProviderStats{
			ProviderID:         providerID,
			TotalRequests:      0,
			SuccessfulRequests: 0,
			FailedRequests:     0,
			SuccessRate:        0.0,
			AvgResponseTime:    0,
			CurrentConnections: 0,
			Weight:             1,
			IsHealthy:          true,
			LastUsed:           time.Now(),
		}
		lb.stats[providerID] = stats
	}
	return stats
}

// IncrementConnections 增加连接数
func (lb *loadBalancerImpl) IncrementConnections(providerID int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	stats := lb.getProviderStats(providerID)
	stats.CurrentConnections++
}

// DecrementConnections 减少连接数
func (lb *loadBalancerImpl) DecrementConnections(providerID int64) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	stats := lb.getProviderStats(providerID)
	if stats.CurrentConnections > 0 {
		stats.CurrentConnections--
	}
}

// UpdateHealthStatus 更新健康状态
func (lb *loadBalancerImpl) UpdateHealthStatus(providerID int64, isHealthy bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	stats := lb.getProviderStats(providerID)
	stats.IsHealthy = isHealthy
}

// GetAllStats 获取所有提供商统计信息
func (lb *loadBalancerImpl) GetAllStats() map[int64]*ProviderStats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	result := make(map[int64]*ProviderStats)
	for id, stats := range lb.stats {
		result[id] = &ProviderStats{
			ProviderID:         stats.ProviderID,
			TotalRequests:      stats.TotalRequests,
			SuccessfulRequests: stats.SuccessfulRequests,
			FailedRequests:     stats.FailedRequests,
			SuccessRate:        stats.SuccessRate,
			AvgResponseTime:    stats.AvgResponseTime,
			CurrentConnections: stats.CurrentConnections,
			Weight:             stats.Weight,
			IsHealthy:          stats.IsHealthy,
			LastUsed:           stats.LastUsed,
		}
	}
	return result
}
