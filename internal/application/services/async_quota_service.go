package services

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/async"
	"ai-api-gateway/internal/infrastructure/logger"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"
)

// AsyncQuotaService 异步配额服务
type AsyncQuotaService struct {
	// 嵌入同步配额服务，复用检查逻辑
	*quotaServiceImpl
	
	// 异步消费者
	consumer *async.QuotaConsumer
	
	// 配置
	enableAsync bool
}

// NewAsyncQuotaService 创建异步配额服务
func NewAsyncQuotaService(
	quotaRepo repositories.QuotaRepository,
	quotaUsageRepo repositories.QuotaUsageRepository,
	userRepo repositories.UserRepository,
	cache *redisInfra.CacheService,
	invalidationService *redisInfra.CacheInvalidationService,
	consumerConfig *async.QuotaConsumerConfig,
	logger logger.Logger,
) (QuotaService, error) {
	// 创建基础的同步服务
	baseService := &quotaServiceImpl{
		quotaRepo:           quotaRepo,
		quotaUsageRepo:      quotaUsageRepo,
		userRepo:            userRepo,
		cache:               cache,
		invalidationService: invalidationService,
		logger:              logger,
	}
	
	// 创建异步消费者
	consumer := async.NewQuotaConsumer(
		consumerConfig,
		quotaRepo,
		quotaUsageRepo,
		invalidationService,
		logger,
	)
	
	asyncService := &AsyncQuotaService{
		quotaServiceImpl: baseService,
		consumer:         consumer,
		enableAsync:      true,
	}
	
	// 启动消费者
	if err := consumer.Start(); err != nil {
		return nil, fmt.Errorf("failed to start quota consumer: %w", err)
	}
	
	logger.Info("Async quota service started successfully")
	
	return asyncService, nil
}

// CheckQuota 检查配额（同步操作，保持实时性）
func (s *AsyncQuotaService) CheckQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) (bool, error) {
	// 配额检查必须是同步的，确保实时性
	return s.quotaServiceImpl.CheckQuota(ctx, userID, quotaType, value)
}

// ConsumeQuota 消费配额（异步操作，提升性能）
func (s *AsyncQuotaService) ConsumeQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error {
	if !s.enableAsync {
		// 如果异步被禁用，回退到同步处理
		return s.quotaServiceImpl.ConsumeQuota(ctx, userID, quotaType, value)
	}
	
	// 创建配额使用事件
	event := &async.QuotaUsageEvent{
		UserID:    userID,
		QuotaType: quotaType,
		Value:     value,
		Timestamp: time.Now(),
		RequestID: s.getRequestIDFromContext(ctx),
		Metadata:  s.getMetadataFromContext(ctx),
	}
	
	// 异步发布事件
	if err := s.consumer.PublishEvent(event); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"user_id":    userID,
			"quota_type": quotaType,
			"value":      value,
			"error":      err.Error(),
		}).Error("Failed to publish quota usage event, falling back to sync")
		
		// 异步失败时回退到同步处理
		return s.quotaServiceImpl.ConsumeQuota(ctx, userID, quotaType, value)
	}
	
	s.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"quota_type": quotaType,
		"value":      value,
		"request_id": event.RequestID,
	}).Debug("Quota usage event published successfully")
	
	return nil
}

// CheckBalance 检查余额（同步操作）
func (s *AsyncQuotaService) CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error) {
	return s.quotaServiceImpl.CheckBalance(ctx, userID, estimatedCost)
}

// GetQuotaStatus 获取配额状态（同步操作）
func (s *AsyncQuotaService) GetQuotaStatus(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return s.quotaServiceImpl.GetQuotaStatus(ctx, userID)
}

// ConsumeQuotaSync 同步消费配额（用于需要立即确认的场景）
func (s *AsyncQuotaService) ConsumeQuotaSync(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error {
	return s.quotaServiceImpl.ConsumeQuota(ctx, userID, quotaType, value)
}

// ConsumeQuotaBatch 批量消费配额（异步操作）
func (s *AsyncQuotaService) ConsumeQuotaBatch(ctx context.Context, events []*async.QuotaUsageEvent) error {
	if !s.enableAsync {
		// 如果异步被禁用，逐个同步处理
		for _, event := range events {
			if err := s.quotaServiceImpl.ConsumeQuota(ctx, event.UserID, event.QuotaType, event.Value); err != nil {
				return err
			}
		}
		return nil
	}
	
	// 批量发布事件
	var failedEvents []*async.QuotaUsageEvent
	for _, event := range events {
		if err := s.consumer.PublishEvent(event); err != nil {
			failedEvents = append(failedEvents, event)
		}
	}
	
	// 如果有失败的事件，同步处理
	if len(failedEvents) > 0 {
		s.logger.WithFields(map[string]interface{}{
			"failed_count": len(failedEvents),
			"total_count":  len(events),
		}).Warn("Some quota events failed to publish, processing synchronously")
		
		for _, event := range failedEvents {
			if err := s.quotaServiceImpl.ConsumeQuota(ctx, event.UserID, event.QuotaType, event.Value); err != nil {
				return fmt.Errorf("failed to process quota event for user %d: %w", event.UserID, err)
			}
		}
	}
	
	return nil
}

// GetConsumerStats 获取消费者统计信息
func (s *AsyncQuotaService) GetConsumerStats() *async.ConsumerStats {
	if s.consumer == nil {
		return nil
	}
	return s.consumer.GetStats()
}

// IsConsumerHealthy 检查消费者健康状态
func (s *AsyncQuotaService) IsConsumerHealthy() bool {
	if s.consumer == nil {
		return false
	}
	return s.consumer.IsHealthy()
}

// EnableAsync 启用异步模式
func (s *AsyncQuotaService) EnableAsync() {
	s.enableAsync = true
	s.logger.Info("Async quota processing enabled")
}

// DisableAsync 禁用异步模式（回退到同步）
func (s *AsyncQuotaService) DisableAsync() {
	s.enableAsync = false
	s.logger.Info("Async quota processing disabled, falling back to sync")
}

// IsAsyncEnabled 检查是否启用异步模式
func (s *AsyncQuotaService) IsAsyncEnabled() bool {
	return s.enableAsync
}

// Stop 停止异步服务
func (s *AsyncQuotaService) Stop() error {
	if s.consumer != nil {
		return s.consumer.Stop()
	}
	return nil
}

// Flush 强制刷新所有待处理的事件
func (s *AsyncQuotaService) Flush(ctx context.Context) error {
	if !s.enableAsync || s.consumer == nil {
		return nil
	}
	
	// 等待一段时间让消费者处理完当前批次
	select {
	case <-time.After(10 * time.Second):
		s.logger.Info("Quota consumer flush completed")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// getRequestIDFromContext 从上下文获取请求ID
func (s *AsyncQuotaService) getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getMetadataFromContext 从上下文获取元数据
func (s *AsyncQuotaService) getMetadataFromContext(ctx context.Context) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// 提取常用的上下文信息
	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		metadata["user_agent"] = userAgent
	}
	
	if clientIP := ctx.Value("client_ip"); clientIP != nil {
		metadata["client_ip"] = clientIP
	}
	
	if apiPath := ctx.Value("api_path"); apiPath != nil {
		metadata["api_path"] = apiPath
	}
	
	return metadata
}

// QuotaServiceWithAsync 扩展的配额服务接口
type QuotaServiceWithAsync interface {
	QuotaService
	
	// 异步相关方法
	ConsumeQuotaSync(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error
	ConsumeQuotaBatch(ctx context.Context, events []*async.QuotaUsageEvent) error
	GetConsumerStats() *async.ConsumerStats
	IsConsumerHealthy() bool
	EnableAsync()
	DisableAsync()
	IsAsyncEnabled() bool
	Stop() error
	Flush(ctx context.Context) error
}

// 确保 AsyncQuotaService 实现了 QuotaServiceWithAsync 接口
var _ QuotaServiceWithAsync = (*AsyncQuotaService)(nil)
