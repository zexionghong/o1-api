package services

import (
	"context"
	"time"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
)

// BillingService 计费服务接口
type BillingService interface {
	// CalculateCost 计算请求成本
	CalculateCost(ctx context.Context, modelID int64, inputTokens, outputTokens int) (*CostCalculation, error)
	
	// CreateBillingRecord 创建计费记录
	CreateBillingRecord(ctx context.Context, usageLogID int64, amount float64, billingType entities.BillingType, description string) (*entities.BillingRecord, error)
	
	// ProcessBilling 处理计费
	ProcessBilling(ctx context.Context, usageLog *entities.UsageLog) error
	
	// ProcessPendingBilling 处理待处理的计费记录
	ProcessPendingBilling(ctx context.Context, limit int) error
	
	// GetUserBilling 获取用户计费信息
	GetUserBilling(ctx context.Context, userID int64, start, end time.Time) (*UserBillingInfo, error)
	
	// GetBillingStats 获取计费统计
	GetBillingStats(ctx context.Context, userID int64, start, end time.Time) (*repositories.BillingStats, error)
	
	// RefundUsage 退款
	RefundUsage(ctx context.Context, usageLogID int64, reason string) error
	
	// AdjustBalance 调整余额
	AdjustBalance(ctx context.Context, userID int64, amount float64, reason string) error
	
	// DeductBalance 扣减余额
	DeductBalance(ctx context.Context, userID int64, amount float64) error
	
	// AddBalance 增加余额
	AddBalance(ctx context.Context, userID int64, amount float64) error
}

// PricingService 定价服务接口
type PricingService interface {
	// GetModelPricing 获取模型定价
	GetModelPricing(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error)
	
	// GetCurrentPricing 获取当前有效定价
	GetCurrentPricing(ctx context.Context, modelID int64, pricingType entities.PricingType) (*entities.ModelPricing, error)
	
	// UpdatePricing 更新定价
	UpdatePricing(ctx context.Context, modelID int64, pricingType entities.PricingType, pricePerUnit float64, effectiveFrom time.Time) error
	
	// CalculateTokenCost 计算token成本
	CalculateTokenCost(ctx context.Context, modelID int64, inputTokens, outputTokens int) (float64, error)
	
	// CalculateRequestCost 计算请求成本
	CalculateRequestCost(ctx context.Context, modelID int64) (float64, error)
	
	// GetPricingHistory 获取定价历史
	GetPricingHistory(ctx context.Context, modelID int64) ([]*entities.ModelPricing, error)
}

// CostCalculation 成本计算结果
type CostCalculation struct {
	ModelID      int64   `json:"model_id"`
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
	Currency     string  `json:"currency"`
	Breakdown    []*CostBreakdown `json:"breakdown"`
}

// CostBreakdown 成本明细
type CostBreakdown struct {
	Type         entities.PricingType `json:"type"`
	Units        int                  `json:"units"`
	PricePerUnit float64              `json:"price_per_unit"`
	Cost         float64              `json:"cost"`
	Currency     string               `json:"currency"`
}

// UserBillingInfo 用户计费信息
type UserBillingInfo struct {
	UserID          int64                    `json:"user_id"`
	Period          DateRange                `json:"period"`
	TotalAmount     float64                  `json:"total_amount"`
	ProcessedAmount float64                  `json:"processed_amount"`
	PendingAmount   float64                  `json:"pending_amount"`
	Records         []*entities.BillingRecord `json:"records"`
	Stats           *repositories.BillingStats `json:"stats"`
}

// DateRange 日期范围
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
