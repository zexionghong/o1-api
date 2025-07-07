package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// quotaRepositoryImpl 配额仓储实现
type quotaRepositoryImpl struct {
	db *sql.DB
}

// NewQuotaRepository 创建配额仓储
func NewQuotaRepository(db *sql.DB) repositories.QuotaRepository {
	return &quotaRepositoryImpl{
		db: db,
	}
}

// Create 创建配额
func (r *quotaRepositoryImpl) Create(ctx context.Context, quota *entities.Quota) error {
	query := `
		INSERT INTO quotas (
			api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	quota.CreatedAt = now
	quota.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		quota.APIKeyID,
		quota.QuotaType,
		quota.Period,
		quota.LimitValue,
		quota.ResetTime,
		quota.Status,
		quota.CreatedAt,
		quota.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create quota: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	quota.ID = id
	return nil
}

// GetByID 根据ID获取配额
func (r *quotaRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.Quota, error) {
	query := `
		SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
		FROM quotas WHERE id = ?
	`

	quota := &entities.Quota{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&quota.ID,
		&quota.APIKeyID,
		&quota.QuotaType,
		&quota.Period,
		&quota.LimitValue,
		&quota.ResetTime,
		&quota.Status,
		&quota.CreatedAt,
		&quota.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound // 使用现有的错误类型
		}
		return nil, fmt.Errorf("failed to get quota by id: %w", err)
	}

	return quota, nil
}

// GetByAPIKeyID 根据API Key ID获取配额列表
func (r *quotaRepositoryImpl) GetByAPIKeyID(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	query := `
		SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
		FROM quotas
		WHERE api_key_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotas by api key id: %w", err)
	}
	defer rows.Close()

	var quotas []*entities.Quota
	for rows.Next() {
		quota := &entities.Quota{}
		err := rows.Scan(
			&quota.ID,
			&quota.APIKeyID,
			&quota.QuotaType,
			&quota.Period,
			&quota.LimitValue,
			&quota.ResetTime,
			&quota.Status,
			&quota.CreatedAt,
			&quota.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota: %w", err)
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quotas: %w", err)
	}

	return quotas, nil
}

// GetByAPIKeyAndType 根据API Key ID和配额类型获取配额
func (r *quotaRepositoryImpl) GetByAPIKeyAndType(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) (*entities.Quota, error) {
	var query string
	var args []interface{}

	if period == nil {
		// 查询总限额
		query = `
			SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
			FROM quotas
			WHERE api_key_id = ? AND quota_type = ? AND period IS NULL AND status = 'active'
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{apiKeyID, quotaType}
	} else {
		// 查询周期限额
		query = `
			SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
			FROM quotas
			WHERE api_key_id = ? AND quota_type = ? AND period = ? AND status = 'active'
			ORDER BY created_at DESC
			LIMIT 1
		`
		args = []interface{}{apiKeyID, quotaType, *period}
	}

	quota := &entities.Quota{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&quota.ID,
		&quota.APIKeyID,
		&quota.QuotaType,
		&quota.Period,
		&quota.LimitValue,
		&quota.ResetTime,
		&quota.Status,
		&quota.CreatedAt,
		&quota.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get quota by api key and type: %w", err)
	}

	return quota, nil
}

// Update 更新配额
func (r *quotaRepositoryImpl) Update(ctx context.Context, quota *entities.Quota) error {
	query := `
		UPDATE quotas 
		SET quota_type = ?, period = ?, limit_value = ?, reset_time = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	quota.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		quota.QuotaType,
		quota.Period,
		quota.LimitValue,
		quota.ResetTime,
		quota.Status,
		quota.UpdatedAt,
		quota.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
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

// Delete 删除配额
func (r *quotaRepositoryImpl) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM quotas WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quota: %w", err)
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

// List 获取配额列表
func (r *quotaRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.Quota, error) {
	query := `
		SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
		FROM quotas
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotas: %w", err)
	}
	defer rows.Close()

	var quotas []*entities.Quota
	for rows.Next() {
		quota := &entities.Quota{}
		err := rows.Scan(
			&quota.ID,
			&quota.APIKeyID,
			&quota.QuotaType,
			&quota.Period,
			&quota.LimitValue,
			&quota.ResetTime,
			&quota.Status,
			&quota.CreatedAt,
			&quota.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota: %w", err)
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quotas: %w", err)
	}

	return quotas, nil
}

// Count 获取配额总数
func (r *quotaRepositoryImpl) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM quotas`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count quotas: %w", err)
	}

	return count, nil
}

// GetActiveQuotas 获取活跃的配额
func (r *quotaRepositoryImpl) GetActiveQuotas(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	query := `
		SELECT id, api_key_id, quota_type, period, limit_value, reset_time, status, created_at, updated_at
		FROM quotas
		WHERE api_key_id = ? AND status = 'active'
		ORDER BY quota_type, period
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active quotas: %w", err)
	}
	defer rows.Close()

	var quotas []*entities.Quota
	for rows.Next() {
		quota := &entities.Quota{}
		err := rows.Scan(
			&quota.ID,
			&quota.APIKeyID,
			&quota.QuotaType,
			&quota.Period,
			&quota.LimitValue,
			&quota.ResetTime,
			&quota.Status,
			&quota.CreatedAt,
			&quota.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota: %w", err)
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate quotas: %w", err)
	}

	return quotas, nil
}
