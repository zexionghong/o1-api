package services

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/domain/services"
	"ai-api-gateway/internal/infrastructure/logger"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"
)

// quotaServiceImpl 配额服务实现
type quotaServiceImpl struct {
	quotaRepo           repositories.QuotaRepository
	quotaUsageRepo      repositories.QuotaUsageRepository
	userRepo            repositories.UserRepository
	cache               *redisInfra.CacheService
	invalidationService *redisInfra.CacheInvalidationService
	logger              logger.Logger
}

// NewQuotaService 创建配额服务
func NewQuotaService(
	quotaRepo repositories.QuotaRepository,
	quotaUsageRepo repositories.QuotaUsageRepository,
	userRepo repositories.UserRepository,
	logger logger.Logger,
) services.QuotaService {
	return &quotaServiceImpl{
		quotaRepo:      quotaRepo,
		quotaUsageRepo: quotaUsageRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// NewQuotaServiceWithCache 创建带缓存的配额服务
func NewQuotaServiceWithCache(
	quotaRepo repositories.QuotaRepository,
	quotaUsageRepo repositories.QuotaUsageRepository,
	userRepo repositories.UserRepository,
	cache *redisInfra.CacheService,
	invalidationService *redisInfra.CacheInvalidationService,
	logger logger.Logger,
) services.QuotaService {
	return &quotaServiceImpl{
		quotaRepo:           quotaRepo,
		quotaUsageRepo:      quotaUsageRepo,
		userRepo:            userRepo,
		cache:               cache,
		invalidationService: invalidationService,
		logger:              logger,
	}
}

// CheckQuota 检查配额是否足够
func (s *quotaServiceImpl) CheckQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, value float64) (*services.QuotaCheckResult, error) {
	// 获取API Key的配额设置（带缓存）
	quotas, err := s.getAPIKeyQuotasWithCache(ctx, apiKeyID)
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"api_key_id": apiKeyID,
			"quota_type": quotaType,
			"error":      err.Error(),
		}).Error("Failed to get API key quotas")
		return &services.QuotaCheckResult{
			Allowed: false,
			Reason:  "Failed to get quota settings",
		}, fmt.Errorf("failed to get API key quotas: %w", err)
	}

	// 如果没有配额设置，默认允许（可以根据需要修改为拒绝）
	if len(quotas) == 0 {
		s.logger.WithFields(map[string]interface{}{
			"api_key_id": apiKeyID,
			"quota_type": quotaType,
		}).Debug("No quota settings found, allowing request")
		return &services.QuotaCheckResult{
			Allowed:   true,
			Remaining: -1, // 无限制
			Limit:     -1,
			Used:      0,
		}, nil
	}

	now := time.Now()

	// 检查所有相关的配额
	for _, quota := range quotas {
		if quota.QuotaType != quotaType || !quota.IsActive() {
			continue
		}

		// 获取当前周期的使用情况（带缓存）
		usage, err := s.getQuotaUsageWithCache(ctx, apiKeyID, quota.ID, quota.Period, now)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKeyID,
				"quota_id":   quota.ID,
				"error":      err.Error(),
			}).Error("Failed to get quota usage")
			continue
		}

		// 检查是否会超出配额
		if usage.UsedValue+value > quota.LimitValue {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id":    apiKeyID,
				"quota_type":    quotaType,
				"quota_id":      quota.ID,
				"used_value":    usage.UsedValue,
				"limit_value":   quota.LimitValue,
				"request_value": value,
			}).Warn("Quota would be exceeded")

			return &services.QuotaCheckResult{
				Allowed:   false,
				Remaining: quota.LimitValue - usage.UsedValue,
				Limit:     quota.LimitValue,
				Used:      usage.UsedValue,
				Reason:    "Quota exceeded",
			}, nil
		}
	}

	return &services.QuotaCheckResult{
		Allowed: true,
	}, nil
}

// ConsumeQuota 消费配额
func (s *quotaServiceImpl) ConsumeQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, value float64) error {
	// 获取API Key的配额设置（带缓存）
	quotas, err := s.getAPIKeyQuotasWithCache(ctx, apiKeyID)
	if err != nil {
		return fmt.Errorf("failed to get API key quotas: %w", err)
	}

	now := time.Now()

	// 更新所有相关配额的使用情况
	for _, quota := range quotas {
		if quota.QuotaType != quotaType || !quota.IsActive() {
			continue
		}

		// 计算当前周期的开始和结束时间
		var periodStart, periodEnd *time.Time
		if quota.IsPeriodQuota() {
			start, end := s.calculatePeriodBounds(*quota.Period, now)
			periodStart = &start
			periodEnd = &end
		}

		// 使用IncrementUsage方法直接增加使用量
		if err := s.quotaUsageRepo.IncrementUsage(ctx, apiKeyID, quota.ID, value, periodStart, periodEnd); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKeyID,
				"quota_id":   quota.ID,
				"value":      value,
				"error":      err.Error(),
			}).Error("Failed to increment quota usage")
			continue
		}

		s.logger.WithFields(map[string]interface{}{
			"api_key_id":  apiKeyID,
			"quota_type":  quotaType,
			"quota_id":    quota.ID,
			"limit_value": quota.LimitValue,
			"consumed":    value,
		}).Debug("Quota consumed successfully")
	}

	// 失效API Key配额相关缓存
	s.invalidateAPIKeyQuotaCache(ctx, apiKeyID)

	return nil
}

// getCurrentPeriodUsage 获取当前周期的使用情况
func (s *quotaServiceImpl) getCurrentPeriodUsage(ctx context.Context, apiKeyID, quotaID int64, period entities.QuotaPeriod, now time.Time) (*entities.QuotaUsage, error) {
	// 计算当前周期的开始和结束时间
	periodStart, periodEnd := s.calculatePeriodBounds(period, now)

	// 尝试获取现有的使用记录
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, apiKeyID, quotaID, &periodStart, &periodEnd)
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
		APIKeyID:    apiKeyID,
		QuotaID:     quotaID,
		PeriodStart: &periodStart,
		PeriodEnd:   &periodEnd,
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

// 缓存相关的辅助方法

// getAPIKeyQuotasWithCache 获取API Key配额列表（暂时不使用缓存）
func (s *quotaServiceImpl) getAPIKeyQuotasWithCache(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	// 暂时直接从数据库查询，等缓存服务支持API Key后再启用缓存
	quotas, err := s.quotaRepo.GetByAPIKeyID(ctx, apiKeyID)
	if err != nil {
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"api_key_id": apiKeyID,
		"count":      len(quotas),
		"source":     "database",
	}).Debug("Retrieved API key quotas from database")

	return quotas, nil
}

// getQuotaUsageWithCache 获取配额使用情况（暂时不使用缓存）
func (s *quotaServiceImpl) getQuotaUsageWithCache(ctx context.Context, apiKeyID, quotaID int64, period *entities.QuotaPeriod, now time.Time) (*entities.QuotaUsage, error) {
	// 计算当前周期的开始和结束时间
	var periodStart, periodEnd *time.Time
	if period != nil {
		start, end := s.calculatePeriodBounds(*period, now)
		periodStart = &start
		periodEnd = &end
	}

	// 暂时直接从数据库查询，等缓存服务支持API Key后再启用缓存
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, apiKeyID, quotaID, periodStart, periodEnd)
	if err != nil {
		// 如果没有找到记录，创建新记录
		if err == entities.ErrUserNotFound {
			usage = &entities.QuotaUsage{
				APIKeyID:    apiKeyID,
				QuotaID:     quotaID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				UsedValue:   0,
			}
		} else {
			return nil, fmt.Errorf("failed to get quota usage: %w", err)
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"api_key_id": apiKeyID,
		"quota_id":   quotaID,
		"period":     period,
		"source":     "database",
	}).Debug("Retrieved quota usage from database")

	return usage, nil
}

// invalidateAPIKeyQuotaCache 失效API Key配额相关缓存
func (s *quotaServiceImpl) invalidateAPIKeyQuotaCache(ctx context.Context, apiKeyID int64) {
	if s.invalidationService != nil {
		// 失效API Key配额设置缓存
		if err := s.invalidationService.InvalidateQuotaCache(ctx, apiKeyID, 0, "", ""); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKeyID,
				"error":      err.Error(),
			}).Warn("Failed to invalidate user quota cache")
		}

		// 失效API Key配额使用情况缓存（使用模式匹配）
		if err := s.invalidationService.BatchInvalidate(ctx, []redisInfra.InvalidationOperation{
			redisInfra.NewPatternInvalidation(fmt.Sprintf("quota_usage:api_key:%d:*", apiKeyID)),
		}); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKeyID,
				"error":      err.Error(),
			}).Warn("Failed to invalidate API key quota usage cache")
		}
	}
}

// GetAPIKeyQuotas 获取API Key配额列表
func (s *quotaServiceImpl) GetAPIKeyQuotas(ctx context.Context, apiKeyID int64) ([]*entities.Quota, error) {
	return s.getAPIKeyQuotasWithCache(ctx, apiKeyID)
}

// GetQuotaUsage 获取配额使用情况
func (s *quotaServiceImpl) GetQuotaUsage(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) (*services.QuotaUsageInfo, error) {
	// 获取配额设置
	quota, err := s.quotaRepo.GetByAPIKeyAndType(ctx, apiKeyID, quotaType, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	now := time.Now()
	var periodStart, periodEnd *time.Time

	if quota.IsPeriodQuota() {
		start := quota.GetPeriodStart(now)
		end := quota.GetPeriodEnd(now)
		periodStart = &start
		periodEnd = &end
	}

	// 获取使用情况
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, apiKeyID, quota.ID, periodStart, periodEnd)
	if err != nil && err != entities.ErrUserNotFound {
		return nil, fmt.Errorf("failed to get quota usage: %w", err)
	}

	usedValue := float64(0)
	if usage != nil {
		usedValue = usage.UsedValue
	}

	remaining := quota.LimitValue - usedValue
	percentage := (usedValue / quota.LimitValue) * 100

	return &services.QuotaUsageInfo{
		QuotaID:     quota.ID,
		QuotaType:   quota.QuotaType,
		Period:      quota.Period,
		Limit:       quota.LimitValue,
		Used:        usedValue,
		Remaining:   remaining,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Percentage:  percentage,
	}, nil
}

// CreateQuota 创建配额
func (s *quotaServiceImpl) CreateQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod, limit float64) (*entities.Quota, error) {
	// 检查是否已存在相同类型和周期的配额
	existingQuota, err := s.quotaRepo.GetByAPIKeyAndType(ctx, apiKeyID, quotaType, period)
	if err == nil && existingQuota != nil {
		return nil, fmt.Errorf("quota already exists for API key %d with type %s and period %v", apiKeyID, quotaType, period)
	}
	// 如果错误不是"未找到"，则返回错误
	if err != nil && err != entities.ErrQuotaNotFound {
		return nil, fmt.Errorf("failed to check existing quota: %w", err)
	}

	quota := &entities.Quota{
		APIKeyID:   apiKeyID,
		QuotaType:  quotaType,
		Period:     period,
		LimitValue: limit,
		Status:     entities.QuotaStatusActive,
	}

	if err := s.quotaRepo.Create(ctx, quota); err != nil {
		return nil, fmt.Errorf("failed to create quota: %w", err)
	}

	// 失效缓存
	s.invalidateAPIKeyQuotaCache(ctx, apiKeyID)

	return quota, nil
}

// UpdateQuota 更新配额
func (s *quotaServiceImpl) UpdateQuota(ctx context.Context, quotaID int64, limit float64) error {
	quota, err := s.quotaRepo.GetByID(ctx, quotaID)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	quota.LimitValue = limit
	if err := s.quotaRepo.Update(ctx, quota); err != nil {
		return fmt.Errorf("failed to update quota: %w", err)
	}

	// 失效缓存
	s.invalidateAPIKeyQuotaCache(ctx, quota.APIKeyID)

	return nil
}

// DeleteQuota 删除配额
func (s *quotaServiceImpl) DeleteQuota(ctx context.Context, quotaID int64) error {
	quota, err := s.quotaRepo.GetByID(ctx, quotaID)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	if err := s.quotaRepo.Delete(ctx, quotaID); err != nil {
		return fmt.Errorf("failed to delete quota: %w", err)
	}

	// 失效缓存
	s.invalidateAPIKeyQuotaCache(ctx, quota.APIKeyID)

	return nil
}

// ResetQuota 重置配额使用情况
func (s *quotaServiceImpl) ResetQuota(ctx context.Context, apiKeyID int64, quotaType entities.QuotaType, period *entities.QuotaPeriod) error {
	// 获取配额设置
	quota, err := s.quotaRepo.GetByAPIKeyAndType(ctx, apiKeyID, quotaType, period)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	now := time.Now()
	var periodStart, periodEnd *time.Time

	if quota.IsPeriodQuota() {
		start := quota.GetPeriodStart(now)
		end := quota.GetPeriodEnd(now)
		periodStart = &start
		periodEnd = &end
	}

	// 删除使用记录
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, apiKeyID, quota.ID, periodStart, periodEnd)
	if err != nil && err != entities.ErrUserNotFound {
		return fmt.Errorf("failed to get quota usage: %w", err)
	}

	if usage != nil {
		if err := s.quotaUsageRepo.Delete(ctx, usage.ID); err != nil {
			return fmt.Errorf("failed to delete quota usage: %w", err)
		}
	}

	// 失效缓存
	s.invalidateAPIKeyQuotaCache(ctx, apiKeyID)

	return nil
}

// GetQuotaStatus 获取配额状态
func (s *quotaServiceImpl) GetQuotaStatus(ctx context.Context, apiKeyID int64) (*services.QuotaStatus, error) {
	quotas, err := s.GetAPIKeyQuotas(ctx, apiKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key quotas: %w", err)
	}

	var quotaInfos []*services.QuotaUsageInfo
	isBlocked := false
	var reason string

	for _, quota := range quotas {
		if !quota.IsActive() {
			continue
		}

		usageInfo, err := s.GetQuotaUsage(ctx, apiKeyID, quota.QuotaType, quota.Period)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKeyID,
				"quota_id":   quota.ID,
				"error":      err.Error(),
			}).Warn("Failed to get quota usage for status")
			continue
		}

		quotaInfos = append(quotaInfos, usageInfo)

		// 检查是否超出限制
		if usageInfo.Used >= usageInfo.Limit {
			isBlocked = true
			if reason == "" {
				reason = fmt.Sprintf("%s quota exceeded", usageInfo.QuotaType)
			}
		}
	}

	return &services.QuotaStatus{
		APIKeyID:  apiKeyID,
		Quotas:    quotaInfos,
		IsBlocked: isBlocked,
		Reason:    reason,
	}, nil
}

// CleanupExpiredUsage 清理过期的使用记录
func (s *quotaServiceImpl) CleanupExpiredUsage(ctx context.Context) error {
	// 这里可以实现清理逻辑，比如删除超过一定时间的使用记录
	// 暂时返回nil，表示不需要清理
	return nil
}

// generatePeriodKey 生成包含日期的周期键
func (s *quotaServiceImpl) generatePeriodKey(period entities.QuotaPeriod, periodStart time.Time) string {
	switch period {
	case entities.QuotaPeriodMinute:
		// 格式：minute:2024-06-06:14:30
		return fmt.Sprintf("minute:%s", periodStart.Format("2006-01-02:15:04"))
	case entities.QuotaPeriodHour:
		// 格式：hour:2024-06-06:14
		return fmt.Sprintf("hour:%s", periodStart.Format("2006-01-02:15"))
	case entities.QuotaPeriodDay:
		// 格式：day:2024-06-06
		return fmt.Sprintf("day:%s", periodStart.Format("2006-01-02"))
	case entities.QuotaPeriodMonth:
		// 格式：month:2024-06
		return fmt.Sprintf("month:%s", periodStart.Format("2006-01"))
	default:
		// 默认为小时
		return fmt.Sprintf("hour:%s", periodStart.Format("2006-01-02:15"))
	}
}
