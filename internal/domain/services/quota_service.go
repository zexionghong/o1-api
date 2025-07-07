package services

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
	"time"
)

// QuotaService 配额服务接口
type QuotaService interface {
	// CheckQuota 检查配额是否足够
	CheckQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, value float64) (*QuotaCheckResult, error)

	// ConsumeQuota 消费配额
	ConsumeQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, value float64) error

	// GetAPIKeyQuotas 获取API Key配额列表
	GetAPIKeyQuotas(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error)

	// GetQuotaUsage 获取配额使用情况
	GetQuotaUsage(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) (*QuotaUsageInfo, error)

	// CreateQuota 创建配额
	CreateQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod, limit float64) (*entities.Quota, error)

	// UpdateQuota 更新配额
	UpdateQuota(ctx context.Context, quotaID int64, limit float64) error

	// DeleteQuota 删除配额
	DeleteQuota(ctx context.Context, quotaID int64) error

	// ResetQuota 重置配额使用情况
	ResetQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) error

	// GetQuotaStatus 获取配额状态
	GetQuotaStatus(ctx context.Context, apiKeyID int64) (*QuotaStatus, error)

	// CleanupExpiredUsage 清理过期的使用记录
	CleanupExpiredUsage(ctx context.Context) error
}

// RateLimitService 速率限制服务接口
type RateLimitService interface {
	// CheckRateLimit 检查速率限制
	CheckRateLimit(ctx context.Context, userID int64, apiKeyID int64) (*RateLimitResult, error)

	// RecordRequest 记录请求
	RecordRequest(ctx context.Context, userID int64, apiKeyID int64) error

	// GetRateLimitStatus 获取速率限制状态
	GetRateLimitStatus(ctx context.Context, userID int64, apiKeyID int64) (*RateLimitStatus, error)

	// ResetRateLimit 重置速率限制
	ResetRateLimit(ctx context.Context, userID int64, apiKeyID int64) error
}

// QuotaCheckResult 配额检查结果
type QuotaCheckResult struct {
	Allowed   bool      `json:"allowed"`
	Remaining float64   `json:"remaining"`
	Limit     float64   `json:"limit"`
	Used      float64   `json:"used"`
	ResetTime time.Time `json:"reset_time"`
	Reason    string    `json:"reason,omitempty"`
}

// QuotaUsageInfo 配额使用信息
type QuotaUsageInfo struct {
	QuotaID     int64                 `json:"quota_id"`
	QuotaType   entities.QuotaType    `json:"quota_type"`
	Period      *entities.QuotaPeriod `json:"period,omitempty"` // NULL表示总限额
	Limit       float64               `json:"limit"`
	Used        float64               `json:"used"`
	Remaining   float64               `json:"remaining"`
	PeriodStart *time.Time            `json:"period_start,omitempty"` // 总限额时为NULL
	PeriodEnd   *time.Time            `json:"period_end,omitempty"`   // 总限额时为NULL
	Percentage  float64               `json:"percentage"`
}

// QuotaStatus 配额状态
type QuotaStatus struct {
	APIKeyID  int64             `json:"api_key_id"`
	Quotas    []*QuotaUsageInfo `json:"quotas"`
	IsBlocked bool              `json:"is_blocked"`
	Reason    string            `json:"reason,omitempty"`
}

// RateLimitResult 速率限制结果
type RateLimitResult struct {
	Allowed    bool      `json:"allowed"`
	Remaining  int       `json:"remaining"`
	Limit      int       `json:"limit"`
	ResetTime  time.Time `json:"reset_time"`
	RetryAfter int       `json:"retry_after_seconds,omitempty"`
	Reason     string    `json:"reason,omitempty"`
}

// RateLimitStatus 速率限制状态
type RateLimitStatus struct {
	UserID   int64          `json:"user_id"`
	APIKeyID int64          `json:"api_key_id"`
	Minute   *RateLimitInfo `json:"minute,omitempty"`
	Hour     *RateLimitInfo `json:"hour,omitempty"`
	Day      *RateLimitInfo `json:"day,omitempty"`
}

// RateLimitInfo 速率限制信息
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Used      int       `json:"used"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
}
