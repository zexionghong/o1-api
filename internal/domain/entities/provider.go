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
	ID                  int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string         `json:"name" gorm:"uniqueIndex;not null;size:100"`
	Slug                string         `json:"slug" gorm:"uniqueIndex;not null;size:50"`
	BaseURL             string         `json:"base_url" gorm:"column:base_url;not null;size:500"`
	APIKeyEncrypted     *string        `json:"api_key_encrypted" gorm:"column:api_key_encrypted;type:text"` // 不在JSON中暴露
	Status              ProviderStatus `json:"status" gorm:"not null;default:active;size:20;index"`
	Priority            int            `json:"priority" gorm:"not null;default:1"`
	TimeoutSeconds      int            `json:"timeout_seconds" gorm:"column:timeout_seconds;not null;default:30"`
	RetryAttempts       int            `json:"retry_attempts" gorm:"column:retry_attempts;not null;default:3"`
	HealthCheckURL      *string        `json:"health_check_url,omitempty" gorm:"column:health_check_url;size:500"`
	HealthCheckInterval int            `json:"health_check_interval" gorm:"column:health_check_interval;not null;default:60"`
	LastHealthCheck     *time.Time     `json:"last_health_check,omitempty" gorm:"column:last_health_check"`
	HealthStatus        HealthStatus   `json:"health_status" gorm:"column:health_status;default:unknown;size:20;index"`
	CreatedAt           time.Time      `json:"created_at" gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt           time.Time      `json:"updated_at" gorm:"column:updated_at;not null;autoUpdateTime"`
}

// TableName 指定表名
func (Provider) TableName() string {
	return "providers"
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
	if p.TimeoutSeconds <= 0 {
		return 30 * time.Second // 默认30秒
	}
	return time.Duration(p.TimeoutSeconds) * time.Second
}

// ShouldRetry 检查是否应该重试
func (p *Provider) ShouldRetry(attemptCount int) bool {
	return attemptCount < p.RetryAttempts
}
