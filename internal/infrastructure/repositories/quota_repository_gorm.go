package repositories

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"gorm.io/gorm"
)

// quotaRepositoryGorm GORM配额仓储实现
type quotaRepositoryGorm struct {
	db *gorm.DB
}

// NewQuotaRepositoryGorm 创建GORM配额仓储
func NewQuotaRepositoryGorm(db *gorm.DB) repositories.QuotaRepository {
	return &quotaRepositoryGorm{
		db: db,
	}
}

// Create 创建配额
func (r *quotaRepositoryGorm) Create(ctx context.Context, quota *entities.Quota) error {
	if err := r.db.WithContext(ctx).Create(quota).Error; err != nil {
		return fmt.Errorf("failed to create quota: %w", err)
	}
	return nil
}

// GetByID 根据ID获取配额
func (r *quotaRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.Quota, error) {
	var quota entities.Quota
	if err := r.db.WithContext(ctx).First(&quota, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrQuotaNotFound
		}
		return nil, fmt.Errorf("failed to get quota by id: %w", err)
	}
	return &quota, nil
}

// GetByAPIKeyID 根据API Key ID获取配额列表
func (r *quotaRepositoryGorm) GetByAPIKeyID(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	var quotas []*entities.Quota
	if err := r.db.WithContext(ctx).
		Where("api_key_id = ?", apiKeyID).
		Order("created_at DESC").
		Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("failed to get quotas by api key id: %w", err)
	}
	return quotas, nil
}

// GetByAPIKeyAndType 根据API Key ID和配额类型获取配额
func (r *quotaRepositoryGorm) GetByAPIKeyAndType(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) (*entities.Quota, error) {
	var quota entities.Quota
	query := r.db.WithContext(ctx).Where("api_key_id = ? AND quota_type = ?", apiKeyID, quotaType)
	
	if period == nil {
		query = query.Where("period IS NULL")
	} else {
		query = query.Where("period = ?", *period)
	}
	
	if err := query.First(&quota).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrQuotaNotFound
		}
		return nil, fmt.Errorf("failed to get quota by api key and type: %w", err)
	}
	return &quota, nil
}

// Update 更新配额
func (r *quotaRepositoryGorm) Update(ctx context.Context, quota *entities.Quota) error {
	quota.UpdatedAt = time.Now()
	
	result := r.db.WithContext(ctx).Save(quota)
	if result.Error != nil {
		return fmt.Errorf("failed to update quota: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrQuotaNotFound
	}
	
	return nil
}

// Delete 删除配额
func (r *quotaRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.Quota{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete quota: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrQuotaNotFound
	}
	
	return nil
}

// List 获取配额列表
func (r *quotaRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.Quota, error) {
	var quotas []*entities.Quota
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}
	return quotas, nil
}

// Count 获取配额总数
func (r *quotaRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.Quota{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count quotas: %w", err)
	}
	return count, nil
}

// GetActiveQuotas 获取活跃的配额列表
func (r *quotaRepositoryGorm) GetActiveQuotas(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	var quotas []*entities.Quota
	if err := r.db.WithContext(ctx).
		Where("api_key_id = ? AND status = ?", apiKeyID, entities.QuotaStatusActive).
		Order("quota_type ASC, period ASC").
		Find(&quotas).Error; err != nil {
		return nil, fmt.Errorf("failed to get active quotas: %w", err)
	}
	return quotas, nil
}

// quotaUsageRepositoryGorm GORM配额使用仓储实现
type quotaUsageRepositoryGorm struct {
	db *gorm.DB
}

// NewQuotaUsageRepositoryGorm 创建GORM配额使用仓储
func NewQuotaUsageRepositoryGorm(db *gorm.DB) repositories.QuotaUsageRepository {
	return &quotaUsageRepositoryGorm{
		db: db,
	}
}

// Create 创建配额使用记录
func (r *quotaUsageRepositoryGorm) Create(ctx context.Context, usage *entities.QuotaUsage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		return fmt.Errorf("failed to create quota usage: %w", err)
	}
	return nil
}

// GetByID 根据ID获取配额使用记录
func (r *quotaUsageRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.QuotaUsage, error) {
	var usage entities.QuotaUsage
	if err := r.db.WithContext(ctx).First(&usage, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrQuotaUsageNotFound
		}
		return nil, fmt.Errorf("failed to get quota usage by id: %w", err)
	}
	return &usage, nil
}

// GetByQuotaAndPeriod 根据配额ID和周期获取使用记录
func (r *quotaUsageRepositoryGorm) GetByQuotaAndPeriod(ctx context.Context, apiKeyID, quotaID int64, periodStart, periodEnd *time.Time) (*entities.QuotaUsage, error) {
	var usage entities.QuotaUsage
	query := r.db.WithContext(ctx).Where("api_key_id = ? AND quota_id = ?", apiKeyID, quotaID)
	
	if periodStart == nil && periodEnd == nil {
		// 总限额查询
		query = query.Where("period_start IS NULL AND period_end IS NULL")
	} else {
		// 周期限额查询
		query = query.Where("period_start = ? AND period_end = ?", periodStart, periodEnd)
	}
	
	if err := query.First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrQuotaUsageNotFound
		}
		return nil, fmt.Errorf("failed to get quota usage by quota and period: %w", err)
	}
	return &usage, nil
}

// GetCurrentUsage 获取当前周期的使用情况
func (r *quotaUsageRepositoryGorm) GetCurrentUsage(ctx context.Context, apiKeyID int64, quotaID int64, at time.Time) (*entities.QuotaUsage, error) {
	var usage entities.QuotaUsage
	
	// 对于周期配额，查找包含指定时间的周期记录
	if err := r.db.WithContext(ctx).
		Where("api_key_id = ? AND quota_id = ? AND ((period_start IS NULL AND period_end IS NULL) OR (period_start <= ? AND period_end > ?))", 
			apiKeyID, quotaID, at, at).
		First(&usage).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrQuotaUsageNotFound
		}
		return nil, fmt.Errorf("failed to get current quota usage: %w", err)
	}
	return &usage, nil
}

// Update 更新配额使用记录
func (r *quotaUsageRepositoryGorm) Update(ctx context.Context, usage *entities.QuotaUsage) error {
	usage.UpdatedAt = time.Now()
	
	result := r.db.WithContext(ctx).Save(usage)
	if result.Error != nil {
		return fmt.Errorf("failed to update quota usage: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrQuotaUsageNotFound
	}
	
	return nil
}

// IncrementUsage 增加使用量
func (r *quotaUsageRepositoryGorm) IncrementUsage(ctx context.Context, apiKeyID, quotaID int64, value float64, periodStart, periodEnd *time.Time) error {
	// 使用事务确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usage entities.QuotaUsage
		query := tx.Where("api_key_id = ? AND quota_id = ?", apiKeyID, quotaID)
		
		if periodStart == nil && periodEnd == nil {
			query = query.Where("period_start IS NULL AND period_end IS NULL")
		} else {
			query = query.Where("period_start = ? AND period_end = ?", periodStart, periodEnd)
		}
		
		err := query.First(&usage).Error
		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			usage = entities.QuotaUsage{
				APIKeyID:    apiKeyID,
				QuotaID:     quotaID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				UsedValue:   value,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			return tx.Create(&usage).Error
		} else if err != nil {
			return fmt.Errorf("failed to get quota usage: %w", err)
		}
		
		// 更新现有记录
		usage.UsedValue += value
		usage.UpdatedAt = time.Now()
		return tx.Save(&usage).Error
	})
}

// Delete 删除配额使用记录
func (r *quotaUsageRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.QuotaUsage{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete quota usage: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return entities.ErrQuotaUsageNotFound
	}
	
	return nil
}

// List 获取配额使用记录列表
func (r *quotaUsageRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.QuotaUsage, error) {
	var usages []*entities.QuotaUsage
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to list quota usage: %w", err)
	}
	return usages, nil
}

// Count 获取配额使用记录总数
func (r *quotaUsageRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.QuotaUsage{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count quota usage: %w", err)
	}
	return count, nil
}

// GetUsageByAPIKey 根据API Key ID获取使用记录列表
func (r *quotaUsageRepositoryGorm) GetUsageByAPIKey(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.QuotaUsage, error) {
	var usages []*entities.QuotaUsage
	if err := r.db.WithContext(ctx).
		Where("api_key_id = ?", apiKeyID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to get quota usage by api key: %w", err)
	}
	return usages, nil
}

// GetUsageByPeriod 根据时间周期获取使用记录列表
func (r *quotaUsageRepositoryGorm) GetUsageByPeriod(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.QuotaUsage, error) {
	var usages []*entities.QuotaUsage
	if err := r.db.WithContext(ctx).
		Where("(period_start >= ? AND period_start < ?) OR (period_start IS NULL)", start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to get quota usage by period: %w", err)
	}
	return usages, nil
}

// CleanupExpiredUsage 清理过期的使用记录
func (r *quotaUsageRepositoryGorm) CleanupExpiredUsage(ctx context.Context, before time.Time) error {
	result := r.db.WithContext(ctx).
		Where("period_end IS NOT NULL AND period_end < ?", before).
		Delete(&entities.QuotaUsage{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired quota usage: %w", result.Error)
	}
	
	return nil
}
