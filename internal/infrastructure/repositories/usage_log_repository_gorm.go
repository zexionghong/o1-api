package repositories

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"

	"gorm.io/gorm"
)

// usageLogRepositoryGorm GORM使用日志仓储实现
type usageLogRepositoryGorm struct {
	db *gorm.DB
}

// NewUsageLogRepositoryGorm 创建GORM使用日志仓储
func NewUsageLogRepositoryGorm(db *gorm.DB) repositories.UsageLogRepository {
	return &usageLogRepositoryGorm{
		db: db,
	}
}

// Create 创建使用日志
func (r *usageLogRepositoryGorm) Create(ctx context.Context, log *entities.UsageLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create usage log: %w", err)
	}
	return nil
}

// GetByID 根据ID获取使用日志
func (r *usageLogRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.UsageLog, error) {
	var log entities.UsageLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrUsageLogNotFound
		}
		return nil, fmt.Errorf("failed to get usage log by id: %w", err)
	}
	return &log, nil
}

// GetByRequestID 根据请求ID获取使用日志
func (r *usageLogRepositoryGorm) GetByRequestID(ctx context.Context, requestID string) (*entities.UsageLog, error) {
	var log entities.UsageLog
	if err := r.db.WithContext(ctx).Where("request_id = ?", requestID).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrUsageLogNotFound
		}
		return nil, fmt.Errorf("failed to get usage log by request id: %w", err)
	}
	return &log, nil
}

// GetByUserID 根据用户ID获取使用日志列表
func (r *usageLogRepositoryGorm) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage logs by user id: %w", err)
	}
	return logs, nil
}

// GetByAPIKeyID 根据API Key ID获取使用日志列表
func (r *usageLogRepositoryGorm) GetByAPIKeyID(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("api_key_id = ?", apiKeyID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage logs by api key id: %w", err)
	}
	return logs, nil
}

// GetByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志列表
func (r *usageLogRepositoryGorm) GetByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	query := r.db.WithContext(ctx).Where("api_key_id = ?", apiKeyID)

	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage logs by api key id and date range: %w", err)
	}
	return logs, nil
}

// CountByAPIKeyID 根据API密钥ID获取使用日志总数
func (r *usageLogRepositoryGorm) CountByAPIKeyID(ctx context.Context, apiKeyID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.UsageLog{}).
		Where("api_key_id = ?", apiKeyID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count usage logs by api key id: %w", err)
	}
	return count, nil
}

// CountByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志总数
func (r *usageLogRepositoryGorm) CountByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entities.UsageLog{}).
		Where("api_key_id = ?", apiKeyID)

	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count usage logs by api key id and date range: %w", err)
	}
	return count, nil
}

// GetByTimeRange 根据时间范围获取使用日志列表
func (r *usageLogRepositoryGorm) GetByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage logs by time range: %w", err)
	}
	return logs, nil
}

// Update 更新使用日志
func (r *usageLogRepositoryGorm) Update(ctx context.Context, log *entities.UsageLog) error {
	result := r.db.WithContext(ctx).Save(log)
	if result.Error != nil {
		return fmt.Errorf("failed to update usage log: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUsageLogNotFound
	}

	return nil
}

// Delete 删除使用日志
func (r *usageLogRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.UsageLog{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete usage log: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrUsageLogNotFound
	}

	return nil
}

// List 获取使用日志列表
func (r *usageLogRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to list usage logs: %w", err)
	}
	return logs, nil
}

// Count 获取使用日志总数
func (r *usageLogRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.UsageLog{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count usage logs: %w", err)
	}
	return count, nil
}

// GetUsageStats 获取使用统计
func (r *usageLogRepositoryGorm) GetUsageStats(ctx context.Context, userID int64, start, end time.Time) (*repositories.UsageStats, error) {
	var stats repositories.UsageStats

	// 使用原生SQL查询统计数据
	query := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as successful_requests,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) as failed_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(duration_ms), 0) as avg_duration
		FROM usage_logs
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
	`

	row := r.db.WithContext(ctx).Raw(query, userID, start, end).Row()
	err := row.Scan(
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.FailedRequests,
		&stats.TotalTokens,
		&stats.InputTokens,
		&stats.OutputTokens,
		&stats.TotalCost,
		&stats.AvgDuration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage stats: %w", err)
	}

	return &stats, nil
}

// GetModelUsageStats 获取模型使用统计
func (r *usageLogRepositoryGorm) GetModelUsageStats(ctx context.Context, userID int64, start, end time.Time) ([]*repositories.ModelStats, error) {
	var stats []*repositories.ModelStats

	query := `
		SELECT
			model_id,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(total_tokens), 0) / COUNT(*) ELSE 0 END as avg_tokens_per_request
		FROM usage_logs
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
		GROUP BY model_id
		ORDER BY total_cost DESC
	`

	rows, err := r.db.WithContext(ctx).Raw(query, userID, start, end).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get model usage stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat repositories.ModelStats
		err := rows.Scan(
			&stat.ModelID,
			&stat.TotalRequests,
			&stat.TotalTokens,
			&stat.InputTokens,
			&stat.OutputTokens,
			&stat.TotalCost,
			&stat.AvgTokensPerRequest,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan model usage stats: %w", err)
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

// GetByDateRange 根据日期范围获取使用日志
func (r *usageLogRepositoryGorm) GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get usage logs by date range: %w", err)
	}
	return logs, nil
}

// GetSuccessfulLogs 获取成功的使用日志
func (r *usageLogRepositoryGorm) GetSuccessfulLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status_code >= 200 AND status_code < 300 AND created_at >= ? AND created_at < ?",
			userID, start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get successful usage logs: %w", err)
	}
	return logs, nil
}

// GetErrorLogs 获取错误的使用日志
func (r *usageLogRepositoryGorm) GetErrorLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var logs []*entities.UsageLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status_code >= 400 AND created_at >= ? AND created_at < ?",
			userID, start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get error usage logs: %w", err)
	}
	return logs, nil
}

// GetProviderStats 获取提供商使用统计
func (r *usageLogRepositoryGorm) GetProviderStats(ctx context.Context, providerID int64, start, end time.Time) (*repositories.ProviderStats, error) {
	var stats repositories.ProviderStats

	query := `
		SELECT
			? as provider_id,
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) as successful_requests,
			COUNT(CASE WHEN status_code >= 400 THEN 1 END) as failed_requests,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(cost), 0) as total_cost
		FROM usage_logs
		WHERE provider_id = ? AND created_at >= ? AND created_at < ?
	`

	row := r.db.WithContext(ctx).Raw(query, providerID, providerID, start, end).Row()
	err := row.Scan(
		&stats.ProviderID,
		&stats.TotalRequests,
		&stats.SuccessfulRequests,
		&stats.FailedRequests,
		&stats.AvgDuration,
		&stats.TotalCost,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider stats: %w", err)
	}

	return &stats, nil
}

// GetModelStats 获取模型使用统计
func (r *usageLogRepositoryGorm) GetModelStats(ctx context.Context, modelID int64, start, end time.Time) (*repositories.ModelStats, error) {
	var stats repositories.ModelStats

	query := `
		SELECT
			? as model_id,
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(input_tokens), 0) as input_tokens,
			COALESCE(SUM(output_tokens), 0) as output_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(total_tokens), 0) / COUNT(*) ELSE 0 END as avg_tokens_per_request
		FROM usage_logs
		WHERE model_id = ? AND created_at >= ? AND created_at < ?
	`

	row := r.db.WithContext(ctx).Raw(query, modelID, modelID, start, end).Row()
	err := row.Scan(
		&stats.ModelID,
		&stats.TotalRequests,
		&stats.TotalTokens,
		&stats.InputTokens,
		&stats.OutputTokens,
		&stats.TotalCost,
		&stats.AvgTokensPerRequest,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get model stats: %w", err)
	}

	return &stats, nil
}

// CleanupOldLogs 清理旧日志
func (r *usageLogRepositoryGorm) CleanupOldLogs(ctx context.Context, before time.Time) error {
	result := r.db.WithContext(ctx).
		Where("created_at < ?", before).
		Delete(&entities.UsageLog{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old usage logs: %w", result.Error)
	}

	return nil
}

// billingRecordRepositoryGorm GORM计费记录仓储实现
type billingRecordRepositoryGorm struct {
	db *gorm.DB
}

// NewBillingRecordRepositoryGorm 创建GORM计费记录仓储
func NewBillingRecordRepositoryGorm(db *gorm.DB) repositories.BillingRecordRepository {
	return &billingRecordRepositoryGorm{
		db: db,
	}
}

// Create 创建计费记录
func (r *billingRecordRepositoryGorm) Create(ctx context.Context, record *entities.BillingRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create billing record: %w", err)
	}
	return nil
}

// GetByID 根据ID获取计费记录
func (r *billingRecordRepositoryGorm) GetByID(ctx context.Context, id int64) (*entities.BillingRecord, error) {
	var record entities.BillingRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrBillingRecordNotFound
		}
		return nil, fmt.Errorf("failed to get billing record by id: %w", err)
	}
	return &record, nil
}

// GetByUserID 根据用户ID获取计费记录列表
func (r *billingRecordRepositoryGorm) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get billing records by user id: %w", err)
	}
	return records, nil
}

// Update 更新计费记录
func (r *billingRecordRepositoryGorm) Update(ctx context.Context, record *entities.BillingRecord) error {
	result := r.db.WithContext(ctx).Save(record)
	if result.Error != nil {
		return fmt.Errorf("failed to update billing record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrBillingRecordNotFound
	}

	return nil
}

// UpdateStatus 更新计费状态
func (r *billingRecordRepositoryGorm) UpdateStatus(ctx context.Context, id int64, status entities.BillingStatus) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}

	if status == entities.BillingStatusProcessed {
		updates["processed_at"] = &now
	}

	result := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update billing record status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrBillingRecordNotFound
	}

	return nil
}

// Delete 删除计费记录
func (r *billingRecordRepositoryGorm) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&entities.BillingRecord{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete billing record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return entities.ErrBillingRecordNotFound
	}

	return nil
}

// List 获取计费记录列表
func (r *billingRecordRepositoryGorm) List(ctx context.Context, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to list billing records: %w", err)
	}
	return records, nil
}

// Count 获取计费记录总数
func (r *billingRecordRepositoryGorm) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count billing records: %w", err)
	}
	return count, nil
}

// GetPendingRecords 获取待处理的计费记录
func (r *billingRecordRepositoryGorm) GetPendingRecords(ctx context.Context, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Where("status = ?", entities.BillingStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending billing records: %w", err)
	}
	return records, nil
}

// GetByTimeRange 根据时间范围获取计费记录
func (r *billingRecordRepositoryGorm) GetByTimeRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get billing records by time range: %w", err)
	}
	return records, nil
}

// GetByAPIKeyID 根据API密钥ID获取计费记录列表
func (r *billingRecordRepositoryGorm) GetByAPIKeyID(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Where("usage_log_id IN (SELECT id FROM usage_logs WHERE api_key_id = ?)", apiKeyID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get billing records by api key id: %w", err)
	}
	return records, nil
}

// GetByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取计费记录列表
func (r *billingRecordRepositoryGorm) GetByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	query := r.db.WithContext(ctx).
		Where("usage_log_id IN (SELECT id FROM usage_logs WHERE api_key_id = ?)", apiKeyID)

	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}

	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get billing records by api key id and date range: %w", err)
	}
	return records, nil
}

// CountByAPIKeyID 根据API密钥ID获取计费记录总数
func (r *billingRecordRepositoryGorm) CountByAPIKeyID(ctx context.Context, apiKeyID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).
		Where("usage_log_id IN (SELECT id FROM usage_logs WHERE api_key_id = ?)", apiKeyID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count billing records by api key id: %w", err)
	}
	return count, nil
}

// CountByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取计费记录总数
func (r *billingRecordRepositoryGorm) CountByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).
		Where("usage_log_id IN (SELECT id FROM usage_logs WHERE api_key_id = ?)", apiKeyID)

	if start != nil && end != nil {
		query = query.Where("created_at >= ? AND created_at < ?", *start, *end)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count billing records by api key id and date range: %w", err)
	}
	return count, nil
}

// GetByUsageLogID 根据使用日志ID获取计费记录
func (r *billingRecordRepositoryGorm) GetByUsageLogID(ctx context.Context, usageLogID int64) (*entities.BillingRecord, error) {
	var record entities.BillingRecord
	if err := r.db.WithContext(ctx).Where("usage_log_id = ?", usageLogID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrBillingRecordNotFound
		}
		return nil, fmt.Errorf("failed to get billing record by usage log id: %w", err)
	}
	return &record, nil
}

// GetByDateRange 根据日期范围获取计费记录
func (r *billingRecordRepositoryGorm) GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.BillingRecord, error) {
	var records []*entities.BillingRecord
	if err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at < ?", start, end).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get billing records by date range: %w", err)
	}
	return records, nil
}

// GetBillingStats 获取计费统计
func (r *billingRecordRepositoryGorm) GetBillingStats(ctx context.Context, userID int64, start, end time.Time) (*repositories.BillingStats, error) {
	var stats repositories.BillingStats

	query := `
		SELECT
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 'processed' THEN amount ELSE 0 END), 0) as processed_amount,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as pending_amount,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN amount ELSE 0 END), 0) as failed_amount,
			COUNT(*) as total_records,
			COUNT(CASE WHEN status = 'processed' THEN 1 END) as processed_records,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_records,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_records
		FROM billing_records
		WHERE user_id = ? AND created_at >= ? AND created_at < ?
	`

	row := r.db.WithContext(ctx).Raw(query, userID, start, end).Row()
	err := row.Scan(
		&stats.TotalAmount,
		&stats.ProcessedAmount,
		&stats.PendingAmount,
		&stats.FailedAmount,
		&stats.TotalRecords,
		&stats.ProcessedRecords,
		&stats.PendingRecords,
		&stats.FailedRecords,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing stats: %w", err)
	}

	return &stats, nil
}

// BatchUpdateStatus 批量更新状态
func (r *billingRecordRepositoryGorm) BatchUpdateStatus(ctx context.Context, ids []int64, status entities.BillingStatus) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}

	if status == entities.BillingStatusProcessed {
		updates["processed_at"] = &now
	}

	result := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).
		Where("id IN ?", ids).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to batch update billing record status: %w", result.Error)
	}

	return nil
}

// GetTotalAmountByUser 获取用户总计费金额
func (r *billingRecordRepositoryGorm) GetTotalAmountByUser(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	var total float64

	query := r.db.WithContext(ctx).Model(&entities.BillingRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND status = ?", userID, entities.BillingStatusProcessed)

	if !start.IsZero() && !end.IsZero() {
		query = query.Where("created_at >= ? AND created_at < ?", start, end)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to get total amount by user: %w", err)
	}

	return total, nil
}
