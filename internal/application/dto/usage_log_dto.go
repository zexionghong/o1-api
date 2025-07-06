package dto

import (
	"ai-api-gateway/internal/domain/entities"
	"time"
)

// UsageLogResponse 使用日志响应
type UsageLogResponse struct {
	ID          int64     `json:"id"`
	APIKeyID    int64     `json:"api_key_id"`
	UserID      int64     `json:"user_id"`
	Model       string    `json:"model"`
	TokensUsed  int       `json:"tokens_used"`
	Cost        float64   `json:"cost"`
	RequestType string    `json:"request_type"`
	Status      string    `json:"status"`
	RequestID   string    `json:"request_id"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Timestamp   time.Time `json:"timestamp"`
}

// BillingRecordResponse 扣费记录响应
type BillingRecordResponse struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	BalanceBefore   float64   `json:"balance_before"`
	BalanceAfter    float64   `json:"balance_after"`
	Timestamp       time.Time `json:"timestamp"`
}

// UsageLogListRequest 使用日志列表请求
type UsageLogListRequest struct {
	APIKeyID  int64      `form:"api_key_id"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02T15:04:05Z07:00"`
	Page      int        `form:"page,default=1" validate:"min=1"`
	PageSize  int        `form:"page_size,default=10" validate:"min=1,max=100"`
}

// BillingRecordListRequest 扣费记录列表请求
type BillingRecordListRequest struct {
	APIKeyID  int64      `form:"api_key_id"`
	StartDate *time.Time `form:"start_date" time_format:"2006-01-02T15:04:05Z07:00"`
	EndDate   *time.Time `form:"end_date" time_format:"2006-01-02T15:04:05Z07:00"`
	Page      int        `form:"page,default=1" validate:"min=1"`
	PageSize  int        `form:"page_size,default=10" validate:"min=1,max=100"`
}

// FromUsageLogEntity 从使用日志实体转换
func (r *UsageLogResponse) FromEntity(log *entities.UsageLog) *UsageLogResponse {
	// 根据现有实体结构映射
	status := "success"
	if !log.IsSuccessful() {
		status = "failed"
	}

	// 简化的模型映射
	model := "unknown"
	requestType := log.Method

	return &UsageLogResponse{
		ID:          log.ID,
		APIKeyID:    log.APIKeyID,
		UserID:      log.UserID,
		Model:       model,
		TokensUsed:  log.GetTokensUsed(),
		Cost:        log.Cost,
		RequestType: requestType,
		Status:      status,
		RequestID:   log.RequestID,
		IPAddress:   "", // 如果需要可以添加
		UserAgent:   "", // 如果需要可以添加
		Timestamp:   log.CreatedAt,
	}
}

// FromBillingRecordEntity 从扣费记录实体转换
func (r *BillingRecordResponse) FromEntity(record *entities.BillingRecord) *BillingRecordResponse {
	description := "API Usage"
	if record.Description != nil {
		description = *record.Description
	}

	return &BillingRecordResponse{
		ID:              record.ID,
		UserID:          record.UserID,
		Amount:          -record.Amount, // 负数表示扣费
		Description:     description,
		TransactionType: string(record.BillingType),
		BalanceBefore:   0, // 需要从账户余额计算
		BalanceAfter:    0, // 需要从账户余额计算
		Timestamp:       record.CreatedAt,
	}
}

// FromUsageLogEntities 从使用日志实体列表转换
func FromUsageLogEntities(logs []*entities.UsageLog) []*UsageLogResponse {
	responses := make([]*UsageLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = (&UsageLogResponse{}).FromEntity(log)
	}
	return responses
}

// FromBillingRecordEntities 从扣费记录实体列表转换
func FromBillingRecordEntities(records []*entities.BillingRecord) []*BillingRecordResponse {
	responses := make([]*BillingRecordResponse, len(records))
	for i, record := range records {
		responses[i] = (&BillingRecordResponse{}).FromEntity(record)
	}
	return responses
}
