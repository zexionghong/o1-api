package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
	"time"
)

// QuotaRepository 配额仓储接口
type QuotaRepository interface {
	// Create 创建配额
	Create(ctx context.Context, quota *entities.Quota) error

	// GetByID 根据ID获取配额
	GetByID(ctx context.Context, id int64) (*entities.Quota, error)

	// GetByAPIKeyID 根据API Key ID获取配额列表
	GetByAPIKeyID(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error)

	// GetByAPIKeyAndType 根据API Key ID和配额类型获取配额
	GetByAPIKeyAndType(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) (*entities.Quota, error)

	// Update 更新配额
	Update(ctx context.Context, quota *entities.Quota) error

	// Delete 删除配额
	Delete(ctx context.Context, id int64) error

	// List 获取配额列表
	List(ctx context.Context, offset, limit int) ([]*entities.Quota, error)

	// Count 获取配额总数
	Count(ctx context.Context) (int64, error)

	// GetActiveQuotas 获取活跃的配额列表
	GetActiveQuotas(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error)
}

// QuotaUsageRepository 配额使用情况仓储接口
type QuotaUsageRepository interface {
	// Create 创建配额使用记录
	Create(ctx context.Context, usage *entities.QuotaUsage) error

	// GetByID 根据ID获取配额使用记录
	GetByID(ctx context.Context, id int64) (*entities.QuotaUsage, error)

	// GetByQuotaAndPeriod 根据配额ID和周期获取使用记录
	GetByQuotaAndPeriod(ctx context.Context, apiKeyID, quotaID int64, periodStart, periodEnd *time.Time) (*entities.QuotaUsage, error)

	// GetCurrentUsage 获取当前周期的使用情况
	GetCurrentUsage(ctx context.Context, apiKeyID int64, quotaID int64, at time.Time) (*entities.QuotaUsage, error)

	// Update 更新配额使用记录
	Update(ctx context.Context, usage *entities.QuotaUsage) error

	// IncrementUsage 增加使用量
	IncrementUsage(ctx context.Context, apiKeyID, quotaID int64, value float64, periodStart, periodEnd *time.Time) error

	// Delete 删除配额使用记录
	Delete(ctx context.Context, id int64) error

	// List 获取配额使用记录列表
	List(ctx context.Context, offset, limit int) ([]*entities.QuotaUsage, error)

	// Count 获取配额使用记录总数
	Count(ctx context.Context) (int64, error)

	// GetUsageByAPIKey 根据API Key ID获取使用记录列表
	GetUsageByAPIKey(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.QuotaUsage, error)

	// GetUsageByPeriod 根据时间周期获取使用记录列表
	GetUsageByPeriod(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.QuotaUsage, error)

	// CleanupExpiredUsage 清理过期的使用记录
	CleanupExpiredUsage(ctx context.Context, before time.Time) error
}
