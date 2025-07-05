package services

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"

	"github.com/go-redis/redis/v8"
)

// QuotaService 配额服务接口
type QuotaService interface {
	// CheckQuota 检查配额是否足够
	CheckQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) (bool, error)

	// ConsumeQuota 消费配额
	ConsumeQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error

	// CheckBalance 检查用户余额是否足够
	CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error)

	// GetQuotaStatus 获取配额状态
	GetQuotaStatus(ctx context.Context, userID int64) (map[string]interface{}, error)
}

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
) QuotaService {
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
) QuotaService {
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
func (s *quotaServiceImpl) CheckQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) (bool, error) {
	// 获取用户的配额设置（带缓存）
	quotas, err := s.getUserQuotasWithCache(ctx, userID)
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

		// 获取当前周期的使用情况（带缓存）
		usage, err := s.getQuotaUsageWithCache(ctx, userID, quota.ID, quota.Period, now)
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
	// 获取用户的配额设置（带缓存）
	quotas, err := s.getUserQuotasWithCache(ctx, userID)
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

	// 失效用户配额相关缓存
	s.invalidateUserQuotaCache(ctx, userID)

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
	quotas, err := s.getUserQuotasWithCache(ctx, userID)
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

		usage, err := s.getQuotaUsageWithCache(ctx, userID, quota.ID, quota.Period, now)
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

// 缓存相关的辅助方法

// getUserQuotasWithCache 获取用户配额列表（带缓存）
func (s *quotaServiceImpl) getUserQuotasWithCache(ctx context.Context, userID int64) ([]*entities.Quota, error) {
	// 如果缓存可用，先尝试从缓存获取
	if s.cache != nil && s.cache.IsEnabled() {
		if quotas, err := s.cache.GetUserQuotas(ctx, userID); err == nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"source":  "cache",
			}).Debug("User quotas retrieved from cache")
			return quotas, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to get user quotas from cache")
		}
	}

	// 从数据库查询
	quotas, err := s.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	if s.cache != nil && s.cache.IsEnabled() {
		if err := s.cache.SetUserQuotas(ctx, userID, quotas); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to cache user quotas")
		} else {
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"count":   len(quotas),
			}).Debug("User quotas cached successfully")
		}
	}

	return quotas, nil
}

// getQuotaUsageWithCache 获取配额使用情况（带缓存）
func (s *quotaServiceImpl) getQuotaUsageWithCache(ctx context.Context, userID, quotaID int64, period entities.QuotaPeriod, now time.Time) (*entities.QuotaUsage, error) {
	// 计算当前周期的开始和结束时间
	periodStart, periodEnd := s.calculatePeriodBounds(period, now)

	// 生成包含日期的缓存键
	periodKey := s.generatePeriodKey(period, periodStart)
	cacheKey := fmt.Sprintf("quota_usage:user:%d:quota:%d:period:%s", userID, quotaID, periodKey)

	// 如果缓存可用，先尝试从缓存获取
	if s.cache != nil && s.cache.IsEnabled() {
		var usage entities.QuotaUsage
		if err := s.cache.Get(ctx, cacheKey, &usage); err == nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"quota_id": quotaID,
				"period":   period,
				"source":   "cache",
			}).Debug("Quota usage retrieved from cache")
			return &usage, nil
		} else if err != redis.Nil {
			// 记录缓存错误但继续查询数据库
			s.logger.WithFields(map[string]interface{}{
				"user_id":  userID,
				"quota_id": quotaID,
				"error":    err.Error(),
			}).Warn("Failed to get quota usage from cache")
		}
	}

	// 从数据库查询
	usage, err := s.quotaUsageRepo.GetByQuotaAndPeriod(ctx, userID, quotaID, periodStart, periodEnd)
	if err != nil {
		// 如果没有找到记录，创建新记录
		if err == entities.ErrUserNotFound {
			usage = &entities.QuotaUsage{
				UserID:      userID,
				QuotaID:     quotaID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
				UsedValue:   0,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
		} else {
			return nil, fmt.Errorf("failed to get quota usage: %w", err)
		}
	}

	// 缓存结果（短期缓存）
	if s.cache != nil && s.cache.IsEnabled() {
		ttl := time.Duration(2) * time.Minute // 配额使用情况缓存2分钟
		if err := s.cache.Set(ctx, cacheKey, usage, ttl); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id":   userID,
				"quota_id":  quotaID,
				"cache_key": cacheKey,
				"error":     err.Error(),
			}).Warn("Failed to cache quota usage")
		} else {
			s.logger.WithFields(map[string]interface{}{
				"user_id":   userID,
				"quota_id":  quotaID,
				"cache_key": cacheKey,
				"period":    period,
			}).Debug("Quota usage cached successfully")
		}
	}

	return usage, nil
}

// invalidateUserQuotaCache 失效用户配额相关缓存
func (s *quotaServiceImpl) invalidateUserQuotaCache(ctx context.Context, userID int64) {
	if s.invalidationService != nil {
		// 失效用户配额设置缓存
		if err := s.invalidationService.InvalidateQuotaCache(ctx, 0, userID, "", ""); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to invalidate user quota cache")
		}

		// 失效用户配额使用情况缓存（使用模式匹配）
		if err := s.invalidationService.BatchInvalidate(ctx, []redisInfra.InvalidationOperation{
			redisInfra.NewPatternInvalidation(fmt.Sprintf("quota_usage:user:%d:*", userID)),
		}); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to invalidate user quota usage cache")
		}
	}
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
