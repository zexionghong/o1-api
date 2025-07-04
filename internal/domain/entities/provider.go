package entities

import (
	"time"
)

// ProviderStatus 提供商状态枚举
type ProviderStatus string

const (
	ProviderStatusActive      ProviderStatus = "active"
	ProviderStatusInactive    ProviderStatus = "inactive"
	ProviderStatusMaintenance ProviderStatus = "maintenance"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Provider 服务提供商实体
type Provider struct {
	ID                    int64          `json:"id" db:"id"`
	Name                  string         `json:"name" db:"name"`
	Slug                  string         `json:"slug" db:"slug"`
	BaseURL               string         `json:"base_url" db:"base_url"`
	APIKeyEncrypted       *string        `json:"-" db:"api_key_encrypted"` // 不在JSON中暴露
	Status                ProviderStatus `json:"status" db:"status"`
	Priority              int            `json:"priority" db:"priority"`
	TimeoutSeconds        int            `json:"timeout_seconds" db:"timeout_seconds"`
	RetryAttempts         int            `json:"retry_attempts" db:"retry_attempts"`
	HealthCheckURL        *string        `json:"health_check_url,omitempty" db:"health_check_url"`
	HealthCheckInterval   int            `json:"health_check_interval" db:"health_check_interval"`
	LastHealthCheck       *time.Time     `json:"last_health_check,omitempty" db:"last_health_check"`
	HealthStatus          HealthStatus   `json:"health_status" db:"health_status"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at" db:"updated_at"`
}

// IsAvailable 检查提供商是否可用
func (p *Provider) IsAvailable() bool {
	return p.Status == ProviderStatusActive && p.HealthStatus == HealthStatusHealthy
}

// IsActive 检查提供商是否处于活跃状态
func (p *Provider) IsActive() bool {
	return p.Status == ProviderStatusActive
}

// IsHealthy 检查提供商是否健康
func (p *Provider) IsHealthy() bool {
	return p.HealthStatus == HealthStatusHealthy
}

// NeedsHealthCheck 检查是否需要进行健康检查
func (p *Provider) NeedsHealthCheck() bool {
	if p.LastHealthCheck == nil {
		return true
	}
	
	interval := time.Duration(p.HealthCheckInterval) * time.Second
	return time.Since(*p.LastHealthCheck) >= interval
}

// UpdateHealthStatus 更新健康状态
func (p *Provider) UpdateHealthStatus(status HealthStatus) {
	p.HealthStatus = status
	now := time.Now()
	p.LastHealthCheck = &now
	p.UpdatedAt = now
}

// GetTimeout 获取超时时间
func (p *Provider) GetTimeout() time.Duration {
	return time.Duration(p.TimeoutSeconds) * time.Second
}

// ShouldRetry 检查是否应该重试
func (p *Provider) ShouldRetry(attemptCount int) bool {
	return attemptCount < p.RetryAttempts
}
