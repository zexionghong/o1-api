package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// quotaUsageRepositoryImpl 配额使用仓储实现
type quotaUsageRepositoryImpl struct {
	db *sql.DB
}

// NewQuotaUsageRepository 创建配额使用仓储
func NewQuotaUsageRepository(db *sql.DB) repositories.QuotaUsageRepository {
	return &quotaUsageRepositoryImpl{
		db: db,
	}
}

// Create 创建配额使用记录
func (r *quotaUsageRepositoryImpl) Create(ctx context.Context, usage *entities.QuotaUsage) error {
	query := `
		INSERT INTO quota_usage (
			api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	usage.CreatedAt = now
	usage.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		usage.APIKeyID,
		usage.QuotaID,
		usage.PeriodStart,
		usage.PeriodEnd,
		usage.UsedValue,
		usage.CreatedAt,
		usage.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create quota usage: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	usage.ID = id
	return nil
}

// GetByID 根据ID获取配额使用记录
func (r *quotaUsageRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.QuotaUsage, error) {
	query := `
		SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage WHERE id = ?
	`

	usage := &entities.QuotaUsage{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&usage.ID,
		&usage.APIKeyID,
		&usage.QuotaID,
		&usage.PeriodStart,
		&usage.PeriodEnd,
		&usage.UsedValue,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get quota usage by id: %w", err)
	}

	return usage, nil
}

// GetByQuotaIDAndPeriod 根据配额ID和周期获取使用记录
func (r *quotaUsageRepositoryImpl) GetByQuotaIDAndPeriod(ctx context.Context, quotaID int64, periodStart, periodEnd time.Time, offset, limit int) ([]*entities.QuotaUsage, error) {
	query := `
		SELECT id, user_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage 
		WHERE quota_id = ? AND period_start = ? AND period_end = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, quotaID, periodStart, periodEnd, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota usage by quota id and period: %w", err)
	}
	defer rows.Close()

	var usages []*entities.QuotaUsage
	for rows.Next() {
		usage := &entities.QuotaUsage{}
		err := rows.Scan(
			&usage.ID,
			&usage.APIKeyID,
			&usage.QuotaID,
			&usage.PeriodStart,
			&usage.PeriodEnd,
			&usage.UsedValue,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota usage: %w", err)
		}
		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quota usages: %w", err)
	}

	return usages, nil
}

// GetByQuotaAndPeriod 根据配额ID和周期获取使用记录（单个记录）
func (r *quotaUsageRepositoryImpl) GetByQuotaAndPeriod(ctx context.Context, apiKeyID, quotaID int64, periodStart, periodEnd *time.Time) (*entities.QuotaUsage, error) {
	var query string
	var args []interface{}

	if periodStart == nil || periodEnd == nil {
		// 总限额查询
		query = `
			SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
			FROM quota_usage
			WHERE api_key_id = ? AND quota_id = ? AND period_start IS NULL AND period_end IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{apiKeyID, quotaID}
	} else {
		// 周期限额查询
		query = `
			SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
			FROM quota_usage
			WHERE api_key_id = ? AND quota_id = ? AND period_start = ? AND period_end = ?
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{apiKeyID, quotaID, *periodStart, *periodEnd}
	}

	usage := &entities.QuotaUsage{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&usage.ID,
		&usage.APIKeyID,
		&usage.QuotaID,
		&usage.PeriodStart,
		&usage.PeriodEnd,
		&usage.UsedValue,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get quota usage by quota and period: %w", err)
	}

	return usage, nil
}

// GetCurrentUsage 获取当前周期的使用情况
func (r *quotaUsageRepositoryImpl) GetCurrentUsage(ctx context.Context, apiKeyID int64, quotaID int64, at time.Time) (*entities.QuotaUsage, error) {
	query := `
		SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage
		WHERE api_key_id = ? AND quota_id = ? AND period_start <= ? AND period_end > ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	usage := &entities.QuotaUsage{}
	err := r.db.QueryRowContext(ctx, query, apiKeyID, quotaID, at, at).Scan(
		&usage.ID,
		&usage.APIKeyID,
		&usage.QuotaID,
		&usage.PeriodStart,
		&usage.PeriodEnd,
		&usage.UsedValue,
		&usage.CreatedAt,
		&usage.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}

	return usage, nil
}

// IncrementUsage 增加使用量
func (r *quotaUsageRepositoryImpl) IncrementUsage(ctx context.Context, apiKeyID, quotaID int64, value float64, periodStart, periodEnd *time.Time) error {
	var updateQuery string
	var args []interface{}

	if periodStart == nil || periodEnd == nil {
		// 总限额更新
		updateQuery = `
			UPDATE quota_usage
			SET used_value = used_value + ?, updated_at = ?
			WHERE api_key_id = ? AND quota_id = ? AND period_start IS NULL AND period_end IS NULL
		`
		args = []interface{}{value, time.Now(), apiKeyID, quotaID}
	} else {
		// 周期限额更新
		updateQuery = `
			UPDATE quota_usage
			SET used_value = used_value + ?, updated_at = ?
			WHERE api_key_id = ? AND quota_id = ? AND period_start = ? AND period_end = ?
		`
		args = []interface{}{value, time.Now(), apiKeyID, quotaID, *periodStart, *periodEnd}
	}

	result, err := r.db.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// 如果没有更新任何记录，创建新记录
	if rowsAffected == 0 {
		usage := &entities.QuotaUsage{
			APIKeyID:    apiKeyID,
			QuotaID:     quotaID,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			UsedValue:   value,
		}
		return r.Create(ctx, usage)
	}

	return nil
}

// GetUsageByAPIKey 根据API Key ID获取使用记录列表
func (r *quotaUsageRepositoryImpl) GetUsageByAPIKey(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.QuotaUsage, error) {
	query := `
		SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage
		WHERE api_key_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by api key: %w", err)
	}
	defer rows.Close()

	var usages []*entities.QuotaUsage
	for rows.Next() {
		usage := &entities.QuotaUsage{}
		err := rows.Scan(
			&usage.ID,
			&usage.APIKeyID,
			&usage.QuotaID,
			&usage.PeriodStart,
			&usage.PeriodEnd,
			&usage.UsedValue,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota usage: %w", err)
		}
		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quota usages: %w", err)
	}

	return usages, nil
}

// GetUsageByPeriod 根据时间周期获取使用记录列表
func (r *quotaUsageRepositoryImpl) GetUsageByPeriod(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.QuotaUsage, error) {
	query := `
		SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage
		WHERE period_start >= ? AND period_end <= ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage by period: %w", err)
	}
	defer rows.Close()

	var usages []*entities.QuotaUsage
	for rows.Next() {
		usage := &entities.QuotaUsage{}
		err := rows.Scan(
			&usage.ID,
			&usage.APIKeyID,
			&usage.QuotaID,
			&usage.PeriodStart,
			&usage.PeriodEnd,
			&usage.UsedValue,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota usage: %w", err)
		}
		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quota usages: %w", err)
	}

	return usages, nil
}

// Update 更新配额使用记录
func (r *quotaUsageRepositoryImpl) Update(ctx context.Context, usage *entities.QuotaUsage) error {
	query := `
		UPDATE quota_usage 
		SET used_value = ?, updated_at = ?
		WHERE id = ?
	`

	usage.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		usage.UsedValue,
		usage.UpdatedAt,
		usage.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update quota usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// Delete 删除配额使用记录
func (r *quotaUsageRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM quota_usage WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quota usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrUserNotFound
	}

	return nil
}

// List 获取配额使用记录列表
func (r *quotaUsageRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.QuotaUsage, error) {
	query := `
		SELECT id, api_key_id, quota_id, period_start, period_end, used_value, created_at, updated_at
		FROM quota_usage
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list quota usages: %w", err)
	}
	defer rows.Close()

	var usages []*entities.QuotaUsage
	for rows.Next() {
		usage := &entities.QuotaUsage{}
		err := rows.Scan(
			&usage.ID,
			&usage.APIKeyID,
			&usage.QuotaID,
			&usage.PeriodStart,
			&usage.PeriodEnd,
			&usage.UsedValue,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota usage: %w", err)
		}
		usages = append(usages, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quota usages: %w", err)
	}

	return usages, nil
}

// Count 获取配额使用记录总数
func (r *quotaUsageRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM quota_usage`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count quota usages: %w", err)
	}

	return count, nil
}

// CleanupExpiredUsage 清理过期的使用记录
func (r *quotaUsageRepositoryImpl) CleanupExpiredUsage(ctx context.Context, before time.Time) error {
	query := `DELETE FROM quota_usage WHERE period_end < ?`

	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired quota usage: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	fmt.Printf("Cleaned up %d expired quota usage records\n", rowsAffected)
	return nil
}
