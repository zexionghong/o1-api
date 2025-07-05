package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"
	redisInfra "ai-api-gateway/internal/infrastructure/redis"
)

// QuotaUsageEvent 配额使用事件
type QuotaUsageEvent struct {
	UserID      int64                  `json:"user_id"`
	QuotaType   entities.QuotaType     `json:"quota_type"`
	Value       float64                `json:"value"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QuotaConsumerConfig 配额消费者配置
type QuotaConsumerConfig struct {
	WorkerCount    int           `yaml:"worker_count"`    // 工作协程数量
	ChannelSize    int           `yaml:"channel_size"`    // 通道缓冲区大小
	BatchSize      int           `yaml:"batch_size"`      // 批量处理大小
	FlushInterval  time.Duration `yaml:"flush_interval"`  // 强制刷新间隔
	RetryAttempts  int           `yaml:"retry_attempts"`  // 重试次数
	RetryDelay     time.Duration `yaml:"retry_delay"`     // 重试延迟
}

// DefaultQuotaConsumerConfig 默认配置
func DefaultQuotaConsumerConfig() *QuotaConsumerConfig {
	return &QuotaConsumerConfig{
		WorkerCount:   3,                // 3个工作协程
		ChannelSize:   1000,             // 1000个事件缓冲
		BatchSize:     10,               // 每批处理10个事件
		FlushInterval: 5 * time.Second,  // 5秒强制刷新
		RetryAttempts: 3,                // 重试3次
		RetryDelay:    100 * time.Millisecond, // 100ms重试延迟
	}
}

// QuotaConsumer 异步配额消费者
type QuotaConsumer struct {
	config              *QuotaConsumerConfig
	eventChannel        chan *QuotaUsageEvent
	quotaRepo           repositories.QuotaRepository
	quotaUsageRepo      repositories.QuotaUsageRepository
	invalidationService *redisInfra.CacheInvalidationService
	logger              logger.Logger
	
	// 控制相关
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	started    bool
	mu         sync.RWMutex
	
	// 统计信息
	stats *ConsumerStats
}

// ConsumerStats 消费者统计信息
type ConsumerStats struct {
	TotalEvents     int64 `json:"total_events"`
	ProcessedEvents int64 `json:"processed_events"`
	FailedEvents    int64 `json:"failed_events"`
	DroppedEvents   int64 `json:"dropped_events"`
	BatchCount      int64 `json:"batch_count"`
	mu              sync.RWMutex
}

// NewQuotaConsumer 创建配额消费者
func NewQuotaConsumer(
	config *QuotaConsumerConfig,
	quotaRepo repositories.QuotaRepository,
	quotaUsageRepo repositories.QuotaUsageRepository,
	invalidationService *redisInfra.CacheInvalidationService,
	logger logger.Logger,
) *QuotaConsumer {
	if config == nil {
		config = DefaultQuotaConsumerConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &QuotaConsumer{
		config:              config,
		eventChannel:        make(chan *QuotaUsageEvent, config.ChannelSize),
		quotaRepo:           quotaRepo,
		quotaUsageRepo:      quotaUsageRepo,
		invalidationService: invalidationService,
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
		stats:               &ConsumerStats{},
	}
}

// Start 启动消费者
func (c *QuotaConsumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.started {
		return fmt.Errorf("quota consumer already started")
	}
	
	c.logger.WithFields(map[string]interface{}{
		"worker_count":   c.config.WorkerCount,
		"channel_size":   c.config.ChannelSize,
		"batch_size":     c.config.BatchSize,
		"flush_interval": c.config.FlushInterval,
	}).Info("Starting quota consumer")
	
	// 启动工作协程
	for i := 0; i < c.config.WorkerCount; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}
	
	c.started = true
	return nil
}

// Stop 停止消费者
func (c *QuotaConsumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.started {
		return fmt.Errorf("quota consumer not started")
	}
	
	c.logger.Info("Stopping quota consumer")
	
	// 关闭事件通道
	close(c.eventChannel)
	
	// 取消上下文
	c.cancel()
	
	// 等待所有工作协程结束
	c.wg.Wait()
	
	c.started = false
	c.logger.Info("Quota consumer stopped")
	
	return nil
}

// PublishEvent 发布配额使用事件
func (c *QuotaConsumer) PublishEvent(event *QuotaUsageEvent) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if !c.started {
		return fmt.Errorf("quota consumer not started")
	}
	
	// 增加统计
	c.stats.mu.Lock()
	c.stats.TotalEvents++
	c.stats.mu.Unlock()
	
	select {
	case c.eventChannel <- event:
		return nil
	default:
		// 通道满了，丢弃事件
		c.stats.mu.Lock()
		c.stats.DroppedEvents++
		c.stats.mu.Unlock()
		
		c.logger.WithFields(map[string]interface{}{
			"user_id":    event.UserID,
			"quota_type": event.QuotaType,
			"value":      event.Value,
		}).Warn("Quota usage event dropped due to full channel")
		
		return fmt.Errorf("event channel is full")
	}
}

// worker 工作协程
func (c *QuotaConsumer) worker(workerID int) {
	defer c.wg.Done()
	
	c.logger.WithFields(map[string]interface{}{
		"worker_id": workerID,
	}).Info("Quota consumer worker started")
	
	batch := make([]*QuotaUsageEvent, 0, c.config.BatchSize)
	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case event, ok := <-c.eventChannel:
			if !ok {
				// 通道关闭，处理剩余批次
				if len(batch) > 0 {
					c.processBatch(workerID, batch)
				}
				c.logger.WithFields(map[string]interface{}{
					"worker_id": workerID,
				}).Info("Quota consumer worker stopped")
				return
			}
			
			batch = append(batch, event)
			
			// 批次满了，处理批次
			if len(batch) >= c.config.BatchSize {
				c.processBatch(workerID, batch)
				batch = batch[:0] // 重置批次
			}
			
		case <-ticker.C:
			// 定时刷新，处理未满的批次
			if len(batch) > 0 {
				c.processBatch(workerID, batch)
				batch = batch[:0] // 重置批次
			}
			
		case <-c.ctx.Done():
			// 上下文取消，处理剩余批次
			if len(batch) > 0 {
				c.processBatch(workerID, batch)
			}
			c.logger.WithFields(map[string]interface{}{
				"worker_id": workerID,
			}).Info("Quota consumer worker stopped by context")
			return
		}
	}
}

// processBatch 处理批次
func (c *QuotaConsumer) processBatch(workerID int, batch []*QuotaUsageEvent) {
	if len(batch) == 0 {
		return
	}
	
	c.logger.WithFields(map[string]interface{}{
		"worker_id":   workerID,
		"batch_size":  len(batch),
	}).Debug("Processing quota usage batch")
	
	// 增加批次统计
	c.stats.mu.Lock()
	c.stats.BatchCount++
	c.stats.mu.Unlock()
	
	// 按用户分组处理
	userGroups := c.groupEventsByUser(batch)
	
	for userID, events := range userGroups {
		if err := c.processUserEvents(userID, events); err != nil {
			c.logger.WithFields(map[string]interface{}{
				"worker_id":    workerID,
				"user_id":      userID,
				"events_count": len(events),
				"error":        err.Error(),
			}).Error("Failed to process user quota events")
			
			// 增加失败统计
			c.stats.mu.Lock()
			c.stats.FailedEvents += int64(len(events))
			c.stats.mu.Unlock()
		} else {
			// 增加成功统计
			c.stats.mu.Lock()
			c.stats.ProcessedEvents += int64(len(events))
			c.stats.mu.Unlock()
		}
	}
}

// groupEventsByUser 按用户分组事件
func (c *QuotaConsumer) groupEventsByUser(batch []*QuotaUsageEvent) map[int64][]*QuotaUsageEvent {
	groups := make(map[int64][]*QuotaUsageEvent)
	
	for _, event := range batch {
		groups[event.UserID] = append(groups[event.UserID], event)
	}
	
	return groups
}

// processUserEvents 处理用户事件
func (c *QuotaConsumer) processUserEvents(userID int64, events []*QuotaUsageEvent) error {
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()
	
	// 按配额类型分组
	quotaGroups := make(map[entities.QuotaType]float64)
	for _, event := range events {
		quotaGroups[event.QuotaType] += event.Value
	}
	
	// 获取用户配额设置
	quotas, err := c.quotaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user quotas: %w", err)
	}
	
	// 处理每种配额类型
	for quotaType, totalValue := range quotaGroups {
		if err := c.processQuotaType(ctx, userID, quotaType, totalValue, quotas); err != nil {
			c.logger.WithFields(map[string]interface{}{
				"user_id":    userID,
				"quota_type": quotaType,
				"value":      totalValue,
				"error":      err.Error(),
			}).Error("Failed to process quota type")
			// 继续处理其他配额类型
		}
	}
	
	// 失效用户配额缓存
	if c.invalidationService != nil {
		if err := c.invalidationService.InvalidateQuotaCache(ctx, 0, userID, "", ""); err != nil {
			c.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Warn("Failed to invalidate quota cache")
		}
	}
	
	return nil
}

// processQuotaType 处理特定配额类型
func (c *QuotaConsumer) processQuotaType(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64, quotas []*entities.Quota) error {
	now := time.Now()
	
	for _, quota := range quotas {
		if quota.QuotaType != quotaType || !quota.IsActive() {
			continue
		}
		
		// 计算周期边界
		periodStart := quota.GetPeriodStart(now)
		periodEnd := quota.GetPeriodEnd(now)
		
		// 更新配额使用量（带重试）
		if err := c.incrementUsageWithRetry(ctx, userID, quota.ID, value, periodStart, periodEnd); err != nil {
			return fmt.Errorf("failed to increment usage for quota %d: %w", quota.ID, err)
		}
	}
	
	return nil
}

// incrementUsageWithRetry 带重试的使用量增加
func (c *QuotaConsumer) incrementUsageWithRetry(ctx context.Context, userID, quotaID int64, value float64, periodStart, periodEnd time.Time) error {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// 重试延迟
			select {
			case <-time.After(c.config.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		err := c.quotaUsageRepo.IncrementUsage(ctx, userID, quotaID, value, periodStart, periodEnd)
		if err == nil {
			return nil
		}
		
		lastErr = err
		c.logger.WithFields(map[string]interface{}{
			"user_id":  userID,
			"quota_id": quotaID,
			"attempt":  attempt + 1,
			"error":    err.Error(),
		}).Warn("Failed to increment quota usage, retrying")
	}
	
	return fmt.Errorf("failed after %d attempts: %w", c.config.RetryAttempts+1, lastErr)
}

// GetStats 获取统计信息
func (c *QuotaConsumer) GetStats() *ConsumerStats {
	c.stats.mu.RLock()
	defer c.stats.mu.RUnlock()
	
	return &ConsumerStats{
		TotalEvents:     c.stats.TotalEvents,
		ProcessedEvents: c.stats.ProcessedEvents,
		FailedEvents:    c.stats.FailedEvents,
		DroppedEvents:   c.stats.DroppedEvents,
		BatchCount:      c.stats.BatchCount,
	}
}

// IsHealthy 检查消费者健康状态
func (c *QuotaConsumer) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return c.started && c.ctx.Err() == nil
}
