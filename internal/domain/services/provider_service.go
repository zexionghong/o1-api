package services

import (
	"context"
	"ai-api-gateway/internal/domain/entities"
)

// ProviderService 提供商服务接口
type ProviderService interface {
	// GetAvailableProviders 获取可用的提供商列表
	GetAvailableProviders(ctx context.Context) ([]*entities.Provider, error)
	
	// SelectProvider 选择提供商（负载均衡）
	SelectProvider(ctx context.Context, modelSlug string) (*entities.Provider, error)
	
	// HealthCheck 健康检查
	HealthCheck(ctx context.Context, providerID int64) error
	
	// PerformHealthChecks 执行健康检查
	PerformHealthChecks(ctx context.Context) error
	
	// UpdateProviderHealth 更新提供商健康状态
	UpdateProviderHealth(ctx context.Context, providerID int64, status entities.HealthStatus) error
	
	// GetProviderBySlug 根据slug获取提供商
	GetProviderBySlug(ctx context.Context, slug string) (*entities.Provider, error)
	
	// GetProviderModels 获取提供商的模型列表
	GetProviderModels(ctx context.Context, providerID int64) ([]*entities.Model, error)
	
	// IsProviderAvailable 检查提供商是否可用
	IsProviderAvailable(ctx context.Context, providerID int64) (bool, error)
}

// LoadBalancerService 负载均衡服务接口
type LoadBalancerService interface {
	// SelectProvider 选择提供商
	SelectProvider(ctx context.Context, providers []*entities.Provider, request *LoadBalanceRequest) (*entities.Provider, error)
	
	// RecordResponse 记录响应（用于统计）
	RecordResponse(ctx context.Context, providerID int64, success bool, duration int64) error
	
	// GetProviderStats 获取提供商统计信息
	GetProviderStats(ctx context.Context, providerID int64) (*ProviderStats, error)
	
	// UpdateStrategy 更新负载均衡策略
	UpdateStrategy(ctx context.Context, strategy LoadBalanceStrategy) error
}

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	LoadBalanceStrategyRoundRobin     LoadBalanceStrategy = "round_robin"
	LoadBalanceStrategyWeighted       LoadBalanceStrategy = "weighted"
	LoadBalanceStrategyLeastConnections LoadBalanceStrategy = "least_connections"
	LoadBalanceStrategyRandom         LoadBalanceStrategy = "random"
)

// LoadBalanceRequest 负载均衡请求
type LoadBalanceRequest struct {
	ModelSlug    string            `json:"model_slug"`
	UserID       int64             `json:"user_id"`
	APIKeyID     int64             `json:"api_key_id"`
	RequestSize  int               `json:"request_size"`
	Headers      map[string]string `json:"headers"`
	Priority     int               `json:"priority"`
}

// ProviderStats 提供商统计信息
type ProviderStats struct {
	ProviderID        int64   `json:"provider_id"`
	TotalRequests     int64   `json:"total_requests"`
	SuccessfulRequests int64  `json:"successful_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	SuccessRate       float64 `json:"success_rate"`
	AvgResponseTime   float64 `json:"avg_response_time_ms"`
	CurrentConnections int    `json:"current_connections"`
	Weight            int     `json:"weight"`
	IsHealthy         bool    `json:"is_healthy"`
	LastHealthCheck   string  `json:"last_health_check"`
}
