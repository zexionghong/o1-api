package entities

import (
	"time"
)

// UsageLog 使用日志实体
type UsageLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	APIKeyID     int64     `json:"api_key_id" db:"api_key_id"`
	ProviderID   int64     `json:"provider_id" db:"provider_id"`
	ModelID      int64     `json:"model_id" db:"model_id"`
	RequestID    string    `json:"request_id" db:"request_id"`
	Method       string    `json:"method" db:"method"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	InputTokens  int       `json:"input_tokens" db:"input_tokens"`
	OutputTokens int       `json:"output_tokens" db:"output_tokens"`
	TotalTokens  int       `json:"total_tokens" db:"total_tokens"`
	RequestSize  int       `json:"request_size" db:"request_size"`
	ResponseSize int       `json:"response_size" db:"response_size"`
	DurationMs   int       `json:"duration_ms" db:"duration_ms"`
	StatusCode   int       `json:"status_code" db:"status_code"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	Cost         float64   `json:"cost" db:"cost"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
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
	ID          int64         `json:"id" db:"id"`
	UserID      int64         `json:"user_id" db:"user_id"`
	UsageLogID  int64         `json:"usage_log_id" db:"usage_log_id"`
	Amount      float64       `json:"amount" db:"amount"`
	Currency    string        `json:"currency" db:"currency"`
	BillingType BillingType   `json:"billing_type" db:"billing_type"`
	Description *string       `json:"description,omitempty" db:"description"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty" db:"processed_at"`
	Status      BillingStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
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
