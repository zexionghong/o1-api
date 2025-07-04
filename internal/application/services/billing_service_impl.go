package services

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"
)

// billingServiceImpl 计费服务实现
type billingServiceImpl struct {
	billingRepo      repositories.BillingRecordRepository
	usageLogRepo     repositories.UsageLogRepository
	modelPricingRepo repositories.ModelPricingRepository
	userRepo         repositories.UserRepository
	logger           logger.Logger
}

// NewBillingService 创建计费服务实例
func NewBillingService(
	billingRepo repositories.BillingRecordRepository,
	usageLogRepo repositories.UsageLogRepository,
	modelPricingRepo repositories.ModelPricingRepository,
	userRepo repositories.UserRepository,
) BillingService {
	return &billingServiceImpl{
		billingRepo:      billingRepo,
		usageLogRepo:     usageLogRepo,
		modelPricingRepo: modelPricingRepo,
		userRepo:         userRepo,
		logger:           logger.GetLogger(), // 使用全局logger
	}
}

// CalculateCost 计算成本
func (s *billingServiceImpl) CalculateCost(ctx context.Context, modelID int64, inputTokens, outputTokens int) (float64, error) {
	// 获取输入token定价
	inputPricing, err := s.modelPricingRepo.GetPricingByType(ctx, modelID, entities.PricingTypeInput)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"model_id":     modelID,
			"pricing_type": entities.PricingTypeInput,
			"error":        err.Error(),
		}).Warn("Failed to get input pricing, using default")

		// 如果找不到定价，使用默认值（避免服务中断）
		inputPricing = &entities.ModelPricing{
			PricePerUnit: 0.001, // 默认每1000个token $0.001
			Unit:         entities.PricingUnitToken,
			Currency:     "USD",
		}
	}

	// 获取输出token定价
	outputPricing, err := s.modelPricingRepo.GetPricingByType(ctx, modelID, entities.PricingTypeOutput)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"model_id":     modelID,
			"pricing_type": entities.PricingTypeOutput,
			"error":        err.Error(),
		}).Warn("Failed to get output pricing, using default")

		// 如果找不到定价，使用默认值
		outputPricing = &entities.ModelPricing{
			PricePerUnit: 0.002, // 默认每1000个token $0.002
			Unit:         entities.PricingUnitToken,
			Currency:     "USD",
		}
	}

	// 计算成本
	// 注意：价格通常是按1000个token计算的，所以需要除以1000
	inputCost := float64(inputTokens) * inputPricing.PricePerUnit / 1000.0
	outputCost := float64(outputTokens) * outputPricing.PricePerUnit / 1000.0
	totalCost := inputCost + outputCost

	s.logger.WithFields(map[string]interface{}{
		"model_id":      modelID,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"input_cost":    inputCost,
		"output_cost":   outputCost,
		"total_cost":    totalCost,
	}).Debug("Cost calculation completed")

	return totalCost, nil
}

// ProcessBilling 处理计费
func (s *billingServiceImpl) ProcessBilling(ctx context.Context, usageLog *entities.UsageLog) error {
	if usageLog.Cost <= 0 {
		s.logger.WithField("usage_log_id", usageLog.ID).Debug("No cost to process")
		return nil
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, usageLog.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户余额
	if user.Balance < usageLog.Cost {
		s.logger.WithFields(map[string]interface{}{
			"user_id":      user.ID,
			"balance":      user.Balance,
			"cost":         usageLog.Cost,
			"usage_log_id": usageLog.ID,
		}).Warn("Insufficient balance for billing")

		// 创建失败的计费记录
		description := fmt.Sprintf("Insufficient balance for usage log %d", usageLog.ID)
		billingRecord := &entities.BillingRecord{
			UserID:      usageLog.UserID,
			UsageLogID:  usageLog.ID,
			Amount:      usageLog.Cost,
			Currency:    "USD",
			BillingType: entities.BillingTypeUsage,
			Description: &description,
			Status:      entities.BillingStatusFailed,
			CreatedAt:   time.Now(),
		}

		if err := s.billingRepo.Create(ctx, billingRecord); err != nil {
			s.logger.WithField("error", err.Error()).Error("Failed to create failed billing record")
		}

		return fmt.Errorf("insufficient balance: required %.8f, available %.6f", usageLog.Cost, user.Balance)
	}

	// 扣减用户余额
	if err := user.DeductBalance(usageLog.Cost); err != nil {
		return fmt.Errorf("failed to deduct balance: %w", err)
	}

	// 更新用户余额
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// 创建计费记录
	description := fmt.Sprintf("API usage cost for request %s", usageLog.RequestID)
	processedAt := time.Now()
	billingRecord := &entities.BillingRecord{
		UserID:      usageLog.UserID,
		UsageLogID:  usageLog.ID,
		Amount:      usageLog.Cost,
		Currency:    "USD",
		BillingType: entities.BillingTypeUsage,
		Description: &description,
		ProcessedAt: &processedAt,
		Status:      entities.BillingStatusProcessed,
		CreatedAt:   time.Now(),
	}

	if err := s.billingRepo.Create(ctx, billingRecord); err != nil {
		// 如果创建计费记录失败，需要回滚用户余额
		if rollbackErr := user.AddBalance(usageLog.Cost); rollbackErr != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":        user.ID,
				"cost":           usageLog.Cost,
				"rollback_error": rollbackErr.Error(),
			}).Error("Failed to rollback user balance after billing record creation failure")
		} else {
			if updateErr := s.userRepo.Update(ctx, user); updateErr != nil {
				s.logger.WithFields(map[string]interface{}{
					"user_id":      user.ID,
					"update_error": updateErr.Error(),
				}).Error("Failed to update user balance during rollback")
			}
		}
		return fmt.Errorf("failed to create billing record: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":           user.ID,
		"usage_log_id":      usageLog.ID,
		"billing_record_id": billingRecord.ID,
		"amount":            usageLog.Cost,
		"new_balance":       user.Balance,
	}).Info("Billing processed successfully")

	return nil
}

// GetBillingHistory 获取计费历史
func (s *billingServiceImpl) GetBillingHistory(ctx context.Context, userID int64, offset, limit int) ([]*entities.BillingRecord, error) {
	return s.billingRepo.GetByUserID(ctx, userID, offset, limit)
}

// GetBillingStats 获取计费统计
func (s *billingServiceImpl) GetBillingStats(ctx context.Context, userID int64, startTime, endTime time.Time) (*BillingStats, error) {
	records, err := s.billingRepo.GetByDateRange(ctx, startTime, endTime, 0, 1000) // 获取前1000条记录
	if err != nil {
		return nil, fmt.Errorf("failed to get billing records: %w", err)
	}

	// 过滤用户记录
	var userRecords []*entities.BillingRecord
	for _, record := range records {
		if record.UserID == userID {
			userRecords = append(userRecords, record)
		}
	}

	stats := &BillingStats{
		UserID:    userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	for _, record := range userRecords {
		stats.TotalAmount += record.Amount
		stats.TotalRecords++

		switch record.Status {
		case entities.BillingStatusProcessed:
			stats.ProcessedAmount += record.Amount
			stats.ProcessedRecords++
		case entities.BillingStatusFailed:
			stats.FailedAmount += record.Amount
			stats.FailedRecords++
		case entities.BillingStatusPending:
			stats.PendingAmount += record.Amount
			stats.PendingRecords++
		}
	}

	return stats, nil
}

// BillingStats 计费统计
type BillingStats struct {
	UserID           int64     `json:"user_id"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	TotalAmount      float64   `json:"total_amount"`
	TotalRecords     int       `json:"total_records"`
	ProcessedAmount  float64   `json:"processed_amount"`
	ProcessedRecords int       `json:"processed_records"`
	FailedAmount     float64   `json:"failed_amount"`
	FailedRecords    int       `json:"failed_records"`
	PendingAmount    float64   `json:"pending_amount"`
	PendingRecords   int       `json:"pending_records"`
}
