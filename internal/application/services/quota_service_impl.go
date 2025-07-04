package services

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"
)

// quotaServiceImpl 配额服务实现
type quotaServiceImpl struct {
	quotaRepo      repositories.QuotaRepository
	quotaUsageRepo repositories.QuotaUsageRepository
	userRepo       repositories.UserRepository
	logger         logger.Logger
}

// NewQuotaService 创建配额服务
func NewQuotaService(
	quotaRepo repositories.QuotaRepository,
	quotaUsageRepo repositories.QuotaUsageRepository,
	userRepo repositories.UserRepository,
	logger logger.Logger,
) QuotaService {
	return &quotaServiceImpl{
		quotaRepo:      quotaRepo,
		quotaUsageRepo: quotaUsageRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// CheckQuota 检查配额是否足够
func (s *quotaServiceImpl) CheckQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) (bool, error) {
	// 获取用户的配额设置
	quotas, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id":    userID,
			"quota_type": quotaType,
			"error":      err.Error(),
		}).Error("Failed to get user quotas")
		return false, fmt.Errorf("failed to get user quotas: %w", err)
	}

	// 如果没有配额设置，默认允许（可以根据需要修改为拒绝）
	if len(quotas) == 0 {
		s.logger.WithFields(map[string]interface{}{
			"user_id":    userID,
			"quota_type": quotaType,
		}).Debug("No quota settings found, allowing request")
		return true, nil
	}

	now := time.Now()

	// 检查所有相关的配额
	for _, quota := range quotas {
		if quota.QuotaType != quotaType || !quota.IsActive() {
			continue
		}

		// 获取当前周期的使用情况
		usage, err := s.getCurrentPeriodUsage(ctx, userID, quota.ID, quota.Period, now)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"quota_id": quota.ID,
				"error":    err.Error(),
			}).Error("Failed to get quota usage")
			continue
		}

		// 检查是否会超出配额
		if usage.UsedValue+value > quota.LimitValue {
			s.logger.WithFields(map[string]interface{}{
				"user_id":       userID,
				"quota_type":    quotaType,
				"quota_id":      quota.ID,
				"used_value":    usage.UsedValue,
				"limit_value":   quota.LimitValue,
				"request_value": value,
			}).Warn("Quota would be exceeded")
			return false, nil
		}
	}

	return true, nil
}

// ConsumeQuota 消费配额
func (s *quotaServiceImpl) ConsumeQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error {
	// 获取用户的配额设置
	quotas, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user quotas: %w", err)
	}

	now := time.Now()

	// 更新所有相关配额的使用情况
	for _, quota := range quotas {
		if quota.QuotaType != quotaType || !quota.IsActive() {
			continue
		}

		// 计算当前周期的开始和结束时间
		periodStart, periodEnd := s.calculatePeriodBounds(quota.Period, now)

		// 使用IncrementUsage方法直接增加使用量
		if err := s.quotaUsageRepo.IncrementUsage(ctx, userID, quota.ID, value, periodStart, periodEnd); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"quota_id": quota.ID,
				"value":    value,
				"error":    err.Error(),
			}).Error("Failed to increment quota usage")
			continue
		}

		s.logger.WithFields(map[string]interface{}{
			"user_id":     userID,
			"quota_type":  quotaType,
			"quota_id":    quota.ID,
			"limit_value": quota.LimitValue,
			"consumed":    value,
		}).Debug("Quota consumed successfully")
	}

	return nil
}

// getCurrentPeriodUsage 获取当前周期的使用情况
func (s *quotaServiceImpl) getCurrentPeriodUsage(ctx context.Context, userID, quotaID int64, period entities.QuotaPeriod, now time.Time) (*entities.QuotaUsage, error) {
	// 计算当前周期的开始和结束时间
	periodStart, periodEnd := s.calculatePeriodBounds(period, now)

	// 尝试获取现有的使用记录
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, userID, quotaID, periodStart, periodEnd)
	if err != nil {
		// 如果没有找到记录，这是正常的，我们将创建新记录
		if err != entities.ErrUserNotFound {
			return nil, fmt.Errorf("failed to get quota usage: %w", err)
		}
	} else {
		// 如果找到记录，直接返回
		return usage, nil
	}

	// 如果不存在，创建新的使用记录
	usage = &entities.QuotaUsage{
		UserID:      userID,
		QuotaID:     quotaID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		UsedValue:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.quotaUsageRepo.Create(ctx, usage); err != nil {
		return nil, fmt.Errorf("failed to create quota usage: %w", err)
	}

	return usage, nil
}

// calculatePeriodBounds 计算周期边界
func (s *quotaServiceImpl) calculatePeriodBounds(period entities.QuotaPeriod, now time.Time) (time.Time, time.Time) {
	switch period {
	case entities.QuotaPeriodMinute:
		start := now.Truncate(time.Minute)
		return start, start.Add(time.Minute)

	case entities.QuotaPeriodHour:
		start := now.Truncate(time.Hour)
		return start, start.Add(time.Hour)

	case entities.QuotaPeriodDay:
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 0, 1)

	case entities.QuotaPeriodMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 1, 0)

	default:
		// 默认为小时
		start := now.Truncate(time.Hour)
		return start, start.Add(time.Hour)
	}
}

// CheckBalance 检查用户余额是否足够
func (s *quotaServiceImpl) CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if !user.CanMakeRequest() {
		s.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"status":  user.Status,
		}).Warn("User cannot make requests due to status")
		return false, nil
	}

	// 检查余额是否足够
	if user.Balance < estimatedCost {
		s.logger.WithFields(map[string]interface{}{
			"user_id":        userID,
			"balance":        user.Balance,
			"estimated_cost": estimatedCost,
			"shortfall":      estimatedCost - user.Balance,
		}).Warn("Insufficient balance")
		return false, nil
	}

	return true, nil
}

// GetQuotaStatus 获取用户配额状态
func (s *quotaServiceImpl) GetQuotaStatus(ctx context.Context, userID int64) (map[string]interface{}, error) {
	quotas, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user quotas: %w", err)
	}

	now := time.Now()
	status := make(map[string]interface{})
	quotaDetails := make([]map[string]interface{}, 0)

	for _, quota := range quotas {
		if !quota.IsActive() {
			continue
		}

		usage, err := s.getCurrentPeriodUsage(ctx, userID, quota.ID, quota.Period, now)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"quota_id": quota.ID,
				"error":    err.Error(),
			}).Error("Failed to get quota usage for status")
			continue
		}

		quotaDetail := map[string]interface{}{
			"quota_id":     quota.ID,
			"quota_type":   quota.QuotaType,
			"period":       quota.Period,
			"limit_value":  quota.LimitValue,
			"used_value":   usage.UsedValue,
			"remaining":    quota.LimitValue - usage.UsedValue,
			"percentage":   usage.GetUsagePercentage(quota.LimitValue),
			"period_start": usage.PeriodStart,
			"period_end":   usage.PeriodEnd,
			"is_exceeded":  usage.IsExceeded(quota.LimitValue),
		}

		quotaDetails = append(quotaDetails, quotaDetail)
	}

	status["user_id"] = userID
	status["quotas"] = quotaDetails
	status["timestamp"] = now

	return status, nil
}
