package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// billingRecordRepositoryImpl 计费记录仓储实现
type billingRecordRepositoryImpl struct {
	db *sql.DB
}

// NewBillingRecordRepository 创建计费记录仓储
func NewBillingRecordRepository(db *sql.DB) repositories.BillingRecordRepository {
	return &billingRecordRepositoryImpl{
		db: db,
	}
}

// Create 创建计费记录
func (r *billingRecordRepositoryImpl) Create(ctx context.Context, record *entities.BillingRecord) error {
	query := `
		INSERT INTO billing_records (
			user_id, usage_log_id, amount, currency, billing_type,
			description, status, processed_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	record.CreatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		record.UserID,
		record.UsageLogID,
		record.Amount,
		record.Currency,
		record.BillingType,
		record.Description,
		record.Status,
		record.ProcessedAt,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create billing record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	record.ID = id
	return nil
}

// GetByID 根据ID获取计费记录
func (r *billingRecordRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.BillingRecord, error) {
	query := `
		SELECT id, user_id, usage_log_id, amount, currency, billing_type,
			   description, status, processed_at, created_at
		FROM billing_records WHERE id = ?
	`

	record := &entities.BillingRecord{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.UsageLogID,
		&record.Amount,
		&record.Currency,
		&record.BillingType,
		&record.Description,
		&record.Status,
		&record.ProcessedAt,
		&record.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound // 使用现有的错误类型
		}
		return nil, fmt.Errorf("failed to get billing record by id: %w", err)
	}

	return record, nil
}

// GetByUserID 根据用户ID获取计费记录列表
func (r *billingRecordRepositoryImpl) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.BillingRecord, error) {
	query := `
		SELECT id, user_id, usage_log_id, amount, currency, billing_type,
			   description, status, processed_at, created_at
		FROM billing_records 
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing records by user id: %w", err)
	}
	defer rows.Close()

	var records []*entities.BillingRecord
	for rows.Next() {
		record := &entities.BillingRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.UsageLogID,
			&record.Amount,
			&record.Currency,
			&record.BillingType,
			&record.Description,
			&record.Status,
			&record.ProcessedAt,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan billing record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate billing records: %w", err)
	}

	return records, nil
}

// GetByUsageLogID 根据使用日志ID获取计费记录
func (r *billingRecordRepositoryImpl) GetByUsageLogID(ctx context.Context, usageLogID int64) (*entities.BillingRecord, error) {
	query := `
		SELECT id, user_id, usage_log_id, amount, currency, billing_type,
			   description, status, processed_at, created_at
		FROM billing_records WHERE usage_log_id = ?
	`

	record := &entities.BillingRecord{}
	err := r.db.QueryRowContext(ctx, query, usageLogID).Scan(
		&record.ID,
		&record.UserID,
		&record.UsageLogID,
		&record.Amount,
		&record.Currency,
		&record.BillingType,
		&record.Description,
		&record.Status,
		&record.ProcessedAt,
		&record.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get billing record by usage log id: %w", err)
	}

	return record, nil
}

// Update 更新计费记录
func (r *billingRecordRepositoryImpl) Update(ctx context.Context, record *entities.BillingRecord) error {
	query := `
		UPDATE billing_records 
		SET amount = ?, currency = ?, billing_type = ?, description = ?, 
			status = ?, processed_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		record.Amount,
		record.Currency,
		record.BillingType,
		record.Description,
		record.Status,
		record.ProcessedAt,
		record.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update billing record: %w", err)
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

// UpdateStatus 更新计费状态
func (r *billingRecordRepositoryImpl) UpdateStatus(ctx context.Context, id int64, status entities.BillingStatus) error {
	query := `
		UPDATE billing_records 
		SET status = ?, processed_at = ?
		WHERE id = ?
	`

	processedAt := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, processedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update billing record status: %w", err)
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

// Delete 删除计费记录
func (r *billingRecordRepositoryImpl) Delete(ctx context.Context, id int64) error {
	// TODO: 实现删除计费记录
	return nil
}

// List 获取计费记录列表
func (r *billingRecordRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.BillingRecord, error) {
	// TODO: 实现获取计费记录列表
	return nil, nil
}

// Count 获取计费记录总数
func (r *billingRecordRepositoryImpl) Count(ctx context.Context) (int64, error) {
	// TODO: 实现获取计费记录总数
	return 0, nil
}

// GetPendingRecords 获取待处理的计费记录
func (r *billingRecordRepositoryImpl) GetPendingRecords(ctx context.Context, limit int) ([]*entities.BillingRecord, error) {
	// TODO: 实现获取待处理的计费记录
	return nil, nil
}

// GetByDateRange 根据日期范围获取计费记录
func (r *billingRecordRepositoryImpl) GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.BillingRecord, error) {
	// TODO: 实现根据日期范围获取计费记录
	return nil, nil
}

// GetBillingStats 获取计费统计
func (r *billingRecordRepositoryImpl) GetBillingStats(ctx context.Context, userID int64, start, end time.Time) (*repositories.BillingStats, error) {
	// TODO: 实现获取计费统计
	return nil, nil
}

// BatchUpdateStatus 批量更新状态
func (r *billingRecordRepositoryImpl) BatchUpdateStatus(ctx context.Context, ids []int64, status entities.BillingStatus) error {
	// TODO: 实现批量更新状态
	return nil
}
