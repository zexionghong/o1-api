package entities

import (
	"time"
)

// UsageLog 使用日志实体
type UsageLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID       int64     `json:"user_id" gorm:"not null;index"`
	APIKeyID     int64     `json:"api_key_id" gorm:"not null;index"`
	ProviderID   int64     `json:"provider_id" gorm:"not null;index"`
	ModelID      int64     `json:"model_id" gorm:"not null;index"`
	RequestID    string    `json:"request_id" gorm:"uniqueIndex;not null;size:100"`
	Method       string    `json:"method" gorm:"not null;size:10"`
	Endpoint     string    `json:"endpoint" gorm:"not null;size:200"`
	InputTokens  int       `json:"input_tokens" gorm:"default:0"`
	OutputTokens int       `json:"output_tokens" gorm:"default:0"`
	TotalTokens  int       `json:"total_tokens" gorm:"default:0"`
	RequestSize  int       `json:"request_size" gorm:"default:0"`
	ResponseSize int       `json:"response_size" gorm:"default:0"`
	DurationMs   int       `json:"duration_ms" gorm:"not null"`
	StatusCode   int       `json:"status_code" gorm:"not null;index"`
	ErrorMessage *string   `json:"error_message,omitempty" gorm:"type:text"`
	Cost         float64   `json:"cost" gorm:"type:numeric(15,8);default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;autoCreateTime;index"`
}

// TableName 指定表名
func (UsageLog) TableName() string {
	return "usage_logs"
}

// IsSuccessful 检查请求是否成功
func (ul *UsageLog) IsSuccessful() bool {
	return ul.StatusCode >= 200 && ul.StatusCode < 300
}

// HasError 检查是否有错误
func (ul *UsageLog) HasError() bool {
	return ul.ErrorMessage != nil && *ul.ErrorMessage != ""
}

// GetDuration 获取请求持续时间
func (ul *UsageLog) GetDuration() time.Duration {
	return time.Duration(ul.DurationMs) * time.Millisecond
}

// CalculateTotalTokens 计算总token数
func (ul *UsageLog) CalculateTotalTokens() {
	ul.TotalTokens = ul.InputTokens + ul.OutputTokens
}

// GetTokensUsed 获取使用的token数（用于配额计算）
func (ul *UsageLog) GetTokensUsed() int {
	if ul.TotalTokens > 0 {
		return ul.TotalTokens
	}
	return ul.InputTokens + ul.OutputTokens
}

// BillingType 计费类型枚举
type BillingType string

const (
	BillingTypeUsage      BillingType = "usage"
	BillingTypeAdjustment BillingType = "adjustment"
	BillingTypeRefund     BillingType = "refund"
)

// BillingStatus 计费状态枚举
type BillingStatus string

const (
	BillingStatusPending   BillingStatus = "pending"
	BillingStatusProcessed BillingStatus = "processed"
	BillingStatusFailed    BillingStatus = "failed"
)

// BillingRecord 计费记录实体
type BillingRecord struct {
	ID          int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      int64         `json:"user_id" gorm:"not null;index"`
	UsageLogID  int64         `json:"usage_log_id" gorm:"not null;index"`
	Amount      float64       `json:"amount" gorm:"type:numeric(15,8);not null"`
	Currency    string        `json:"currency" gorm:"not null;default:USD;size:3"`
	BillingType BillingType   `json:"billing_type" gorm:"not null;size:20;index"`
	Description *string       `json:"description,omitempty" gorm:"type:text"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty"`
	Status      BillingStatus `json:"status" gorm:"not null;default:pending;size:20;index"`
	CreatedAt   time.Time     `json:"created_at" gorm:"not null;autoCreateTime;index"`
}

// TableName 指定表名
func (BillingRecord) TableName() string {
	return "billing_records"
}

// IsPending 检查计费记录是否待处理
func (br *BillingRecord) IsPending() bool {
	return br.Status == BillingStatusPending
}

// IsProcessed 检查计费记录是否已处理
func (br *BillingRecord) IsProcessed() bool {
	return br.Status == BillingStatusProcessed
}

// IsFailed 检查计费记录是否处理失败
func (br *BillingRecord) IsFailed() bool {
	return br.Status == BillingStatusFailed
}

// MarkAsProcessed 标记为已处理
func (br *BillingRecord) MarkAsProcessed() {
	br.Status = BillingStatusProcessed
	now := time.Now()
	br.ProcessedAt = &now
}

// MarkAsFailed 标记为处理失败
func (br *BillingRecord) MarkAsFailed() {
	br.Status = BillingStatusFailed
}

// IsDebit 检查是否为借记（扣费）
func (br *BillingRecord) IsDebit() bool {
	return br.Amount > 0 && (br.BillingType == BillingTypeUsage || br.BillingType == BillingTypeAdjustment)
}

// IsCredit 检查是否为贷记（退费）
func (br *BillingRecord) IsCredit() bool {
	return br.Amount < 0 || br.BillingType == BillingTypeRefund
}

// 统计数据结构
type UsageStats struct {
	TotalRequests     int64   `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AvgDurationMs     float64 `json:"avg_duration_ms"`
}

type ModelUsageStats struct {
	ModelID           int64   `json:"model_id"`
	RequestCount      int64   `json:"request_count"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
}

type ProviderStats struct {
	ProviderID        int64   `json:"provider_id"`
	TotalRequests     int64   `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AvgDurationMs     float64 `json:"avg_duration_ms"`
	SuccessCount      int64   `json:"success_count"`
	ErrorCount        int64   `json:"error_count"`
}

type ModelStats struct {
	ModelID           int64   `json:"model_id"`
	TotalRequests     int64   `json:"total_requests"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalTokens       int64   `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AvgDurationMs     float64 `json:"avg_duration_ms"`
	SuccessCount      int64   `json:"success_count"`
	ErrorCount        int64   `json:"error_count"`
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
