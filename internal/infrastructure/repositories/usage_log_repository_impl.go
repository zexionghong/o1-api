package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// usageLogRepositoryImpl 使用日志仓储实现
type usageLogRepositoryImpl struct {
	db *sql.DB
}

// NewUsageLogRepository 创建使用日志仓储
func NewUsageLogRepository(db *sql.DB) repositories.UsageLogRepository {
	return &usageLogRepositoryImpl{
		db: db,
	}
}

// Create 创建使用日志
func (r *usageLogRepositoryImpl) Create(ctx context.Context, log *entities.UsageLog) error {
	query := `
		INSERT INTO usage_logs (
			user_id, api_key_id, provider_id, model_id, request_id,
			method, endpoint, input_tokens, output_tokens, total_tokens,
			request_size, response_size, duration_ms, status_code,
			error_message, cost, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	log.CreatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		log.UserID,
		log.APIKeyID,
		log.ProviderID,
		log.ModelID,
		log.RequestID,
		log.Method,
		log.Endpoint,
		log.InputTokens,
		log.OutputTokens,
		log.TotalTokens,
		log.RequestSize,
		log.ResponseSize,
		log.DurationMs,
		log.StatusCode,
		log.ErrorMessage,
		log.Cost,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create usage log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.ID = id
	return nil
}

// GetByID 根据ID获取使用日志
func (r *usageLogRepositoryImpl) GetByID(ctx context.Context, id int64) (*entities.UsageLog, error) {
	query := `
		SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
			   method, endpoint, input_tokens, output_tokens, total_tokens,
			   request_size, response_size, duration_ms, status_code,
			   error_message, cost, created_at
		FROM usage_logs WHERE id = ?
	`

	log := &entities.UsageLog{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.APIKeyID,
		&log.ProviderID,
		&log.ModelID,
		&log.RequestID,
		&log.Method,
		&log.Endpoint,
		&log.InputTokens,
		&log.OutputTokens,
		&log.TotalTokens,
		&log.RequestSize,
		&log.ResponseSize,
		&log.DurationMs,
		&log.StatusCode,
		&log.ErrorMessage,
		&log.Cost,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound // 使用现有的错误类型
		}
		return nil, fmt.Errorf("failed to get usage log by id: %w", err)
	}

	return log, nil
}

// GetByRequestID 根据请求ID获取使用日志
func (r *usageLogRepositoryImpl) GetByRequestID(ctx context.Context, requestID string) (*entities.UsageLog, error) {
	query := `
		SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
			   method, endpoint, input_tokens, output_tokens, total_tokens,
			   request_size, response_size, duration_ms, status_code,
			   error_message, cost, created_at
		FROM usage_logs WHERE request_id = ?
	`

	log := &entities.UsageLog{}
	err := r.db.QueryRowContext(ctx, query, requestID).Scan(
		&log.ID,
		&log.UserID,
		&log.APIKeyID,
		&log.ProviderID,
		&log.ModelID,
		&log.RequestID,
		&log.Method,
		&log.Endpoint,
		&log.InputTokens,
		&log.OutputTokens,
		&log.TotalTokens,
		&log.RequestSize,
		&log.ResponseSize,
		&log.DurationMs,
		&log.StatusCode,
		&log.ErrorMessage,
		&log.Cost,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entities.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get usage log by request id: %w", err)
	}

	return log, nil
}

// GetByUserID 根据用户ID获取使用日志列表
func (r *usageLogRepositoryImpl) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.UsageLog, error) {
	query := `
		SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
			   method, endpoint, input_tokens, output_tokens, total_tokens,
			   request_size, response_size, duration_ms, status_code,
			   error_message, cost, created_at
		FROM usage_logs 
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage logs by user id: %w", err)
	}
	defer rows.Close()

	var logs []*entities.UsageLog
	for rows.Next() {
		log := &entities.UsageLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.APIKeyID,
			&log.ProviderID,
			&log.ModelID,
			&log.RequestID,
			&log.Method,
			&log.Endpoint,
			&log.InputTokens,
			&log.OutputTokens,
			&log.TotalTokens,
			&log.RequestSize,
			&log.ResponseSize,
			&log.DurationMs,
			&log.StatusCode,
			&log.ErrorMessage,
			&log.Cost,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage logs: %w", err)
	}

	return logs, nil
}

// GetByAPIKeyID 根据API密钥ID获取使用日志列表
func (r *usageLogRepositoryImpl) GetByAPIKeyID(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.UsageLog, error) {
	query := `
		SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
			   method, endpoint, input_tokens, output_tokens, total_tokens,
			   request_size, response_size, duration_ms, status_code,
			   error_message, cost, created_at
		FROM usage_logs
		WHERE api_key_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage logs by api key id: %w", err)
	}
	defer rows.Close()

	var logs []*entities.UsageLog
	for rows.Next() {
		log := &entities.UsageLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.APIKeyID,
			&log.ProviderID,
			&log.ModelID,
			&log.RequestID,
			&log.Method,
			&log.Endpoint,
			&log.InputTokens,
			&log.OutputTokens,
			&log.TotalTokens,
			&log.RequestSize,
			&log.ResponseSize,
			&log.DurationMs,
			&log.StatusCode,
			&log.ErrorMessage,
			&log.Cost,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage logs: %w", err)
	}

	return logs, nil
}

// GetByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志列表
func (r *usageLogRepositoryImpl) GetByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	var query string
	var args []interface{}

	if start != nil && end != nil {
		query = `
			SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
				   method, endpoint, input_tokens, output_tokens, total_tokens,
				   request_size, response_size, duration_ms, status_code,
				   error_message, cost, created_at
			FROM usage_logs
			WHERE api_key_id = ? AND created_at >= ? AND created_at <= ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{apiKeyID, *start, *end, limit, offset}
	} else if start != nil {
		query = `
			SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
				   method, endpoint, input_tokens, output_tokens, total_tokens,
				   request_size, response_size, duration_ms, status_code,
				   error_message, cost, created_at
			FROM usage_logs
			WHERE api_key_id = ? AND created_at >= ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{apiKeyID, *start, limit, offset}
	} else if end != nil {
		query = `
			SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
				   method, endpoint, input_tokens, output_tokens, total_tokens,
				   request_size, response_size, duration_ms, status_code,
				   error_message, cost, created_at
			FROM usage_logs
			WHERE api_key_id = ? AND created_at <= ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{apiKeyID, *end, limit, offset}
	} else {
		// 没有日期过滤，直接调用GetByAPIKeyID
		return r.GetByAPIKeyID(ctx, apiKeyID, offset, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage logs by api key id and date range: %w", err)
	}
	defer rows.Close()

	var logs []*entities.UsageLog
	for rows.Next() {
		log := &entities.UsageLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.APIKeyID,
			&log.ProviderID,
			&log.ModelID,
			&log.RequestID,
			&log.Method,
			&log.Endpoint,
			&log.InputTokens,
			&log.OutputTokens,
			&log.TotalTokens,
			&log.RequestSize,
			&log.ResponseSize,
			&log.DurationMs,
			&log.StatusCode,
			&log.ErrorMessage,
			&log.Cost,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage logs: %w", err)
	}

	return logs, nil
}

// CountByAPIKeyID 根据API密钥ID获取使用日志总数
func (r *usageLogRepositoryImpl) CountByAPIKeyID(ctx context.Context, apiKeyID int64) (int64, error) {
	query := `SELECT COUNT(*) FROM usage_logs WHERE api_key_id = ?`

	var count int64
	err := r.db.QueryRowContext(ctx, query, apiKeyID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count usage logs by api key id: %w", err)
	}

	return count, nil
}

// CountByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志总数
func (r *usageLogRepositoryImpl) CountByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time) (int64, error) {
	var query string
	var args []interface{}

	if start != nil && end != nil {
		query = `SELECT COUNT(*) FROM usage_logs WHERE api_key_id = ? AND created_at >= ? AND created_at <= ?`
		args = []interface{}{apiKeyID, *start, *end}
	} else if start != nil {
		query = `SELECT COUNT(*) FROM usage_logs WHERE api_key_id = ? AND created_at >= ?`
		args = []interface{}{apiKeyID, *start}
	} else if end != nil {
		query = `SELECT COUNT(*) FROM usage_logs WHERE api_key_id = ? AND created_at <= ?`
		args = []interface{}{apiKeyID, *end}
	} else {
		// 没有日期过滤，直接调用CountByAPIKeyID
		return r.CountByAPIKeyID(ctx, apiKeyID)
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count usage logs by api key id and date range: %w", err)
	}

	return count, nil
}

// Update 更新使用日志
func (r *usageLogRepositoryImpl) Update(ctx context.Context, log *entities.UsageLog) error {
	// TODO: 实现更新使用日志
	return nil
}

// Delete 删除使用日志
func (r *usageLogRepositoryImpl) Delete(ctx context.Context, id int64) error {
	// TODO: 实现删除使用日志
	return nil
}

// List 获取使用日志列表
func (r *usageLogRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.UsageLog, error) {
	// TODO: 实现获取使用日志列表
	return nil, nil
}

// Count 获取使用日志总数
func (r *usageLogRepositoryImpl) Count(ctx context.Context) (int64, error) {
	// TODO: 实现获取使用日志总数
	return 0, nil
}

// GetByDateRange 根据日期范围获取使用日志
func (r *usageLogRepositoryImpl) GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	query := `
		SELECT id, user_id, api_key_id, provider_id, model_id, request_id,
			   method, endpoint, input_tokens, output_tokens, total_tokens,
			   request_size, response_size, duration_ms, status_code,
			   error_message, cost, created_at
		FROM usage_logs
		WHERE created_at >= ? AND created_at <= ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage logs by date range: %w", err)
	}
	defer rows.Close()

	var logs []*entities.UsageLog
	for rows.Next() {
		log := &entities.UsageLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.APIKeyID,
			&log.ProviderID,
			&log.ModelID,
			&log.RequestID,
			&log.Method,
			&log.Endpoint,
			&log.InputTokens,
			&log.OutputTokens,
			&log.TotalTokens,
			&log.RequestSize,
			&log.ResponseSize,
			&log.DurationMs,
			&log.StatusCode,
			&log.ErrorMessage,
			&log.Cost,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage logs: %w", err)
	}

	return logs, nil
}

// GetSuccessfulLogs 获取成功的使用日志
func (r *usageLogRepositoryImpl) GetSuccessfulLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	// TODO: 实现获取成功的使用日志
	return nil, nil
}

// GetErrorLogs 获取错误的使用日志
func (r *usageLogRepositoryImpl) GetErrorLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error) {
	// TODO: 实现获取错误的使用日志
	return nil, nil
}

// GetUsageStats 获取使用统计
func (r *usageLogRepositoryImpl) GetUsageStats(ctx context.Context, userID int64, start, end time.Time) (*repositories.UsageStats, error) {
	// TODO: 实现获取使用统计
	return nil, nil
}

// GetProviderStats 获取提供商使用统计
func (r *usageLogRepositoryImpl) GetProviderStats(ctx context.Context, providerID int64, start, end time.Time) (*repositories.ProviderStats, error) {
	// TODO: 实现获取提供商使用统计
	return nil, nil
}

// GetModelStats 获取模型使用统计
func (r *usageLogRepositoryImpl) GetModelStats(ctx context.Context, modelID int64, start, end time.Time) (*repositories.ModelStats, error) {
	// TODO: 实现获取模型使用统计
	return nil, nil
}

// CleanupOldLogs 清理旧的日志记录
func (r *usageLogRepositoryImpl) CleanupOldLogs(ctx context.Context, before time.Time) error {
	// TODO: 实现清理旧的日志记录
	return nil
}
