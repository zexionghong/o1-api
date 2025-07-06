package repositories

import (
	"ai-api-gateway/internal/domain/entities"
	"context"
	"time"
)

// UsageLogRepository 使用日志仓储接口
type UsageLogRepository interface {
	// Create 创建使用日志
	Create(ctx context.Context, log *entities.UsageLog) error

	// GetByID 根据ID获取使用日志
	GetByID(ctx context.Context, id int64) (*entities.UsageLog, error)

	// GetByRequestID 根据请求ID获取使用日志
	GetByRequestID(ctx context.Context, requestID string) (*entities.UsageLog, error)

	// GetByUserID 根据用户ID获取使用日志列表
	GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.UsageLog, error)

	// GetByAPIKeyID 根据API密钥ID获取使用日志列表
	GetByAPIKeyID(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.UsageLog, error)

	// GetByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志列表
	GetByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time, offset, limit int) ([]*entities.UsageLog, error)

	// CountByAPIKeyID 根据API密钥ID获取使用日志总数
	CountByAPIKeyID(ctx context.Context, apiKeyID int64) (int64, error)

	// CountByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取使用日志总数
	CountByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time) (int64, error)

	// Update 更新使用日志
	Update(ctx context.Context, log *entities.UsageLog) error

	// Delete 删除使用日志
	Delete(ctx context.Context, id int64) error

	// List 获取使用日志列表
	List(ctx context.Context, offset, limit int) ([]*entities.UsageLog, error)

	// Count 获取使用日志总数
	Count(ctx context.Context) (int64, error)

	// GetByDateRange 根据日期范围获取使用日志
	GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error)

	// GetSuccessfulLogs 获取成功的使用日志
	GetSuccessfulLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error)

	// GetErrorLogs 获取错误的使用日志
	GetErrorLogs(ctx context.Context, userID int64, start, end time.Time, offset, limit int) ([]*entities.UsageLog, error)

	// GetUsageStats 获取使用统计
	GetUsageStats(ctx context.Context, userID int64, start, end time.Time) (*UsageStats, error)

	// GetProviderStats 获取提供商使用统计
	GetProviderStats(ctx context.Context, providerID int64, start, end time.Time) (*ProviderStats, error)

	// GetModelStats 获取模型使用统计
	GetModelStats(ctx context.Context, modelID int64, start, end time.Time) (*ModelStats, error)

	// CleanupOldLogs 清理旧的日志记录
	CleanupOldLogs(ctx context.Context, before time.Time) error
}

// BillingRecordRepository 计费记录仓储接口
type BillingRecordRepository interface {
	// Create 创建计费记录
	Create(ctx context.Context, record *entities.BillingRecord) error

	// GetByID 根据ID获取计费记录
	GetByID(ctx context.Context, id int64) (*entities.BillingRecord, error)

	// GetByUserID 根据用户ID获取计费记录列表
	GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*entities.BillingRecord, error)

	// GetByAPIKeyID 根据API密钥ID获取计费记录列表
	GetByAPIKeyID(ctx context.Context, apiKeyID int64, offset, limit int) ([]*entities.BillingRecord, error)

	// GetByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取计费记录列表
	GetByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time, offset, limit int) ([]*entities.BillingRecord, error)

	// CountByAPIKeyID 根据API密钥ID获取计费记录总数
	CountByAPIKeyID(ctx context.Context, apiKeyID int64) (int64, error)

	// CountByAPIKeyIDAndDateRange 根据API密钥ID和日期范围获取计费记录总数
	CountByAPIKeyIDAndDateRange(ctx context.Context, apiKeyID int64, start, end *time.Time) (int64, error)

	// GetByUsageLogID 根据使用日志ID获取计费记录
	GetByUsageLogID(ctx context.Context, usageLogID int64) (*entities.BillingRecord, error)

	// Update 更新计费记录
	Update(ctx context.Context, record *entities.BillingRecord) error

	// UpdateStatus 更新计费状态
	UpdateStatus(ctx context.Context, id int64, status entities.BillingStatus) error

	// Delete 删除计费记录
	Delete(ctx context.Context, id int64) error

	// List 获取计费记录列表
	List(ctx context.Context, offset, limit int) ([]*entities.BillingRecord, error)

	// Count 获取计费记录总数
	Count(ctx context.Context) (int64, error)

	// GetPendingRecords 获取待处理的计费记录
	GetPendingRecords(ctx context.Context, limit int) ([]*entities.BillingRecord, error)

	// GetByDateRange 根据日期范围获取计费记录
	GetByDateRange(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.BillingRecord, error)

	// GetBillingStats 获取计费统计
	GetBillingStats(ctx context.Context, userID int64, start, end time.Time) (*BillingStats, error)

	// BatchUpdateStatus 批量更新状态
	BatchUpdateStatus(ctx context.Context, ids []int64, status entities.BillingStatus) error
}

// 统计数据结构
type UsageStats struct {
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	TotalTokens        int64   `json:"total_tokens"`
	InputTokens        int64   `json:"input_tokens"`
	OutputTokens       int64   `json:"output_tokens"`
	TotalCost          float64 `json:"total_cost"`
	AvgDuration        float64 `json:"avg_duration_ms"`
}

type ProviderStats struct {
	ProviderID         int64   `json:"provider_id"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	AvgDuration        float64 `json:"avg_duration_ms"`
	TotalCost          float64 `json:"total_cost"`
}

type ModelStats struct {
	ModelID             int64   `json:"model_id"`
	TotalRequests       int64   `json:"total_requests"`
	TotalTokens         int64   `json:"total_tokens"`
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	TotalCost           float64 `json:"total_cost"`
	AvgTokensPerRequest float64 `json:"avg_tokens_per_request"`
}

type BillingStats struct {
	TotalAmount      float64 `json:"total_amount"`
	ProcessedAmount  float64 `json:"processed_amount"`
	PendingAmount    float64 `json:"pending_amount"`
	FailedAmount     float64 `json:"failed_amount"`
	TotalRecords     int64   `json:"total_records"`
	ProcessedRecords int64   `json:"processed_records"`
	PendingRecords   int64   `json:"pending_records"`
	FailedRecords    int64   `json:"failed_records"`
}
