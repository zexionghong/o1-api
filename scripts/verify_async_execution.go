package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/async"
	"ai-api-gateway/internal/infrastructure/logger"
)

// 验证异步处理是否真的在执行
func main() {
	fmt.Println("🔍 深度验证异步配额处理执行状态")
	fmt.Println("=====================================")

	// 创建真实的异步配额服务
	asyncService, err := createRealAsyncQuotaService()
	if err != nil {
		log.Fatalf("❌ 创建异步配额服务失败: %v", err)
	}
	defer asyncService.Stop()

	// 验证1: 检查服务类型
	verifyServiceType(asyncService)

	// 验证2: 检查异步模式状态
	verifyAsyncMode(asyncService)

	// 验证3: 检查消费者启动状态
	verifyConsumerStartup(asyncService)

	// 验证4: 验证事件发布和处理
	verifyEventProcessing(asyncService)

	// 验证5: 验证批量处理
	verifyBatchProcessing(asyncService)

	// 验证6: 验证统计信息更新
	verifyStatsUpdating(asyncService)

	// 验证7: 验证异步vs同步性能差异
	verifyPerformanceDifference(asyncService)

	fmt.Println("\n🎉 异步处理验证完成!")
}

// createRealAsyncQuotaService 创建真实的异步配额服务
func createRealAsyncQuotaService() (services.QuotaServiceWithAsync, error) {
	// 创建真实的logger
	realLogger := &RealLogger{}

	// 创建异步消费者配置
	config := &async.QuotaConsumerConfig{
		WorkerCount:   2,                // 2个工作协程
		ChannelSize:   50,               // 50个事件缓冲
		BatchSize:     3,                // 每批处理3个事件
		FlushInterval: 500 * time.Millisecond, // 500ms强制刷新
		RetryAttempts: 2,                // 重试2次
		RetryDelay:    100 * time.Millisecond, // 100ms重试延迟
	}

	// 创建模拟的Repository（用于测试）
	quotaRepo := &MockQuotaRepository{}
	quotaUsageRepo := &MockQuotaUsageRepository{}
	userRepo := &MockUserRepository{}

	// 创建真实的异步配额服务
	return services.NewAsyncQuotaService(
		quotaRepo,
		quotaUsageRepo,
		userRepo,
		nil, // cache
		nil, // invalidationService
		config,
		realLogger,
	)
}

// verifyServiceType 验证服务类型
func verifyServiceType(service services.QuotaService) {
	fmt.Println("\n📋 验证1: 服务类型检查")
	
	if asyncService, ok := service.(services.QuotaServiceWithAsync); ok {
		fmt.Println("✅ 服务实现了QuotaServiceWithAsync接口")
		
		// 检查异步特有的方法
		stats := asyncService.GetConsumerStats()
		if stats != nil {
			fmt.Println("✅ GetConsumerStats()方法可用")
		} else {
			fmt.Println("❌ GetConsumerStats()返回nil")
		}
		
		healthy := asyncService.IsConsumerHealthy()
		fmt.Printf("✅ IsConsumerHealthy(): %v\n", healthy)
		
		enabled := asyncService.IsAsyncEnabled()
		fmt.Printf("✅ IsAsyncEnabled(): %v\n", enabled)
	} else {
		fmt.Println("❌ 服务没有实现QuotaServiceWithAsync接口")
		os.Exit(1)
	}
}

// verifyAsyncMode 验证异步模式状态
func verifyAsyncMode(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证2: 异步模式状态")
	
	if service.IsAsyncEnabled() {
		fmt.Println("✅ 异步模式已启用")
	} else {
		fmt.Println("❌ 异步模式未启用")
		
		// 尝试启用异步模式
		service.EnableAsync()
		if service.IsAsyncEnabled() {
			fmt.Println("✅ 成功启用异步模式")
		} else {
			fmt.Println("❌ 无法启用异步模式")
		}
	}
}

// verifyConsumerStartup 验证消费者启动状态
func verifyConsumerStartup(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证3: 消费者启动状态")
	
	if service.IsConsumerHealthy() {
		fmt.Println("✅ 消费者已启动且健康")
	} else {
		fmt.Println("❌ 消费者未启动或不健康")
	}
	
	// 检查初始统计信息
	stats := service.GetConsumerStats()
	if stats != nil {
		fmt.Printf("📊 初始统计: 总事件=%d, 已处理=%d, 失败=%d\n", 
			stats.TotalEvents, stats.ProcessedEvents, stats.FailedEvents)
	}
}

// verifyEventProcessing 验证事件发布和处理
func verifyEventProcessing(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证4: 事件发布和处理")
	
	ctx := context.Background()
	userID := int64(123)
	quotaType := entities.QuotaTypeRequests
	
	// 获取处理前的统计
	statsBefore := service.GetConsumerStats()
	eventsBefore := int64(0)
	if statsBefore != nil {
		eventsBefore = statsBefore.TotalEvents
	}
	
	// 发布单个事件
	fmt.Println("📤 发布单个配额消费事件...")
	err := service.ConsumeQuota(ctx, userID, quotaType, 1)
	if err != nil {
		fmt.Printf("❌ 事件发布失败: %v\n", err)
		return
	}
	
	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
	
	// 检查统计是否更新
	statsAfter := service.GetConsumerStats()
	if statsAfter != nil && statsAfter.TotalEvents > eventsBefore {
		fmt.Printf("✅ 事件发布成功: 总事件数从 %d 增加到 %d\n", 
			eventsBefore, statsAfter.TotalEvents)
	} else {
		fmt.Println("❌ 事件统计未更新，可能异步处理未工作")
	}
}

// verifyBatchProcessing 验证批量处理
func verifyBatchProcessing(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证5: 批量处理")
	
	ctx := context.Background()
	userID := int64(456)
	quotaType := entities.QuotaTypeTokens
	
	// 获取处理前的统计
	statsBefore := service.GetConsumerStats()
	
	// 快速发布多个事件
	fmt.Println("📤 快速发布5个配额消费事件...")
	for i := 0; i < 5; i++ {
		err := service.ConsumeQuota(ctx, userID, quotaType, float64(i+1))
		if err != nil {
			fmt.Printf("❌ 事件 %d 发布失败: %v\n", i+1, err)
		}
	}
	
	// 等待批量处理完成
	fmt.Println("⏳ 等待批量处理完成...")
	time.Sleep(1 * time.Second)
	
	// 检查统计
	statsAfter := service.GetConsumerStats()
	if statsAfter != nil && statsBefore != nil {
		eventIncrease := statsAfter.TotalEvents - statsBefore.TotalEvents
		processedIncrease := statsAfter.ProcessedEvents - statsBefore.ProcessedEvents
		
		fmt.Printf("📊 批量处理结果:\n")
		fmt.Printf("   新增事件: %d\n", eventIncrease)
		fmt.Printf("   已处理事件: %d\n", processedIncrease)
		fmt.Printf("   批次数量: %d\n", statsAfter.BatchCount)
		
		if eventIncrease >= 5 {
			fmt.Println("✅ 批量事件发布成功")
		} else {
			fmt.Println("❌ 批量事件发布可能有问题")
		}
	}
}

// verifyStatsUpdating 验证统计信息更新
func verifyStatsUpdating(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证6: 统计信息更新")
	
	stats := service.GetConsumerStats()
	if stats == nil {
		fmt.Println("❌ 无法获取统计信息")
		return
	}
	
	fmt.Printf("📊 当前统计信息:\n")
	fmt.Printf("   总事件数: %d\n", stats.TotalEvents)
	fmt.Printf("   已处理事件数: %d\n", stats.ProcessedEvents)
	fmt.Printf("   失败事件数: %d\n", stats.FailedEvents)
	fmt.Printf("   丢弃事件数: %d\n", stats.DroppedEvents)
	fmt.Printf("   批次数量: %d\n", stats.BatchCount)
	
	if stats.TotalEvents > 0 {
		successRate := float64(stats.ProcessedEvents) / float64(stats.TotalEvents) * 100
		fmt.Printf("   处理成功率: %.1f%%\n", successRate)
		
		if successRate > 80 {
			fmt.Println("✅ 异步处理工作正常")
		} else {
			fmt.Println("⚠️  异步处理成功率较低")
		}
	} else {
		fmt.Println("❌ 没有处理任何事件")
	}
}

// verifyPerformanceDifference 验证异步vs同步性能差异
func verifyPerformanceDifference(service services.QuotaServiceWithAsync) {
	fmt.Println("\n📋 验证7: 性能差异测试")
	
	ctx := context.Background()
	userID := int64(789)
	quotaType := entities.QuotaTypeCost
	requestCount := 50
	
	// 异步处理性能测试
	fmt.Printf("🚀 异步处理 %d 个请求...\n", requestCount)
	asyncStart := time.Now()
	
	for i := 0; i < requestCount; i++ {
		service.ConsumeQuota(ctx, userID, quotaType, 0.1)
	}
	
	asyncDuration := time.Since(asyncStart)
	fmt.Printf("   异步处理耗时: %v\n", asyncDuration)
	fmt.Printf("   平均每请求: %v\n", asyncDuration/time.Duration(requestCount))
	
	// 同步处理性能测试
	fmt.Printf("\n🐌 同步处理 %d 个请求...\n", requestCount)
	syncStart := time.Now()
	
	for i := 0; i < requestCount; i++ {
		service.ConsumeQuotaSync(ctx, userID, quotaType, 0.1)
	}
	
	syncDuration := time.Since(syncStart)
	fmt.Printf("   同步处理耗时: %v\n", syncDuration)
	fmt.Printf("   平均每请求: %v\n", syncDuration/time.Duration(requestCount))
	
	// 性能对比
	if asyncDuration < syncDuration {
		improvement := float64(syncDuration-asyncDuration) / float64(syncDuration) * 100
		speedup := float64(syncDuration) / float64(asyncDuration)
		fmt.Printf("\n📈 性能提升: %.1f%%\n", improvement)
		fmt.Printf("🚀 速度提升: %.1fx\n", speedup)
		fmt.Println("✅ 异步处理确实提升了性能")
	} else {
		fmt.Println("\n⚠️  异步处理没有显示出性能优势")
		fmt.Println("   这可能是因为测试环境或事件处理延迟导致的")
	}
	
	// 等待异步处理完成
	time.Sleep(2 * time.Second)
	
	// 最终统计
	finalStats := service.GetConsumerStats()
	if finalStats != nil {
		fmt.Printf("\n📊 最终统计:\n")
		fmt.Printf("   总事件数: %d\n", finalStats.TotalEvents)
		fmt.Printf("   已处理事件数: %d\n", finalStats.ProcessedEvents)
		
		if finalStats.TotalEvents > 0 {
			fmt.Printf("   处理成功率: %.1f%%\n", 
				float64(finalStats.ProcessedEvents)/float64(finalStats.TotalEvents)*100)
		}
	}
}

// RealLogger 真实的logger实现
type RealLogger struct{}

func (l *RealLogger) Debug(msg string) { log.Printf("[DEBUG] %s", msg) }
func (l *RealLogger) Info(msg string)  { log.Printf("[INFO] %s", msg) }
func (l *RealLogger) Warn(msg string)  { log.Printf("[WARN] %s", msg) }
func (l *RealLogger) Error(msg string) { log.Printf("[ERROR] %s", msg) }
func (l *RealLogger) WithFields(fields map[string]interface{}) logger.Logger { 
	return l 
}

// Mock repositories for testing
type MockQuotaRepository struct{}
func (m *MockQuotaRepository) Create(ctx context.Context, quota *entities.Quota) error { return nil }
func (m *MockQuotaRepository) GetByID(ctx context.Context, id int64) (*entities.Quota, error) { return nil, nil }
func (m *MockQuotaRepository) GetByUserID(ctx context.Context, userID int64) ([]*entities.Quota, error) { return []*entities.Quota{}, nil }
func (m *MockQuotaRepository) GetByUserAndType(ctx context.Context, userID int64, quotaType entities.QuotaType, period entities.QuotaPeriod) (*entities.Quota, error) { return nil, nil }
func (m *MockQuotaRepository) Update(ctx context.Context, quota *entities.Quota) error { return nil }
func (m *MockQuotaRepository) Delete(ctx context.Context, id int64) error { return nil }
func (m *MockQuotaRepository) List(ctx context.Context, offset, limit int) ([]*entities.Quota, error) { return []*entities.Quota{}, nil }
func (m *MockQuotaRepository) Count(ctx context.Context) (int64, error) { return 0, nil }

type MockQuotaUsageRepository struct{}
func (m *MockQuotaUsageRepository) Create(ctx context.Context, usage *entities.QuotaUsage) error { return nil }
func (m *MockQuotaUsageRepository) GetByID(ctx context.Context, id int64) (*entities.QuotaUsage, error) { return nil, nil }
func (m *MockQuotaUsageRepository) GetByQuotaAndPeriod(ctx context.Context, userID, quotaID int64, periodStart, periodEnd time.Time) (*entities.QuotaUsage, error) { return nil, nil }
func (m *MockQuotaUsageRepository) GetCurrentUsage(ctx context.Context, userID int64, quotaID int64, at time.Time) (*entities.QuotaUsage, error) { return nil, nil }
func (m *MockQuotaUsageRepository) Update(ctx context.Context, usage *entities.QuotaUsage) error { return nil }
func (m *MockQuotaUsageRepository) IncrementUsage(ctx context.Context, userID, quotaID int64, value float64, periodStart, periodEnd time.Time) error { 
	// 模拟一些处理延迟
	time.Sleep(1 * time.Millisecond)
	return nil 
}
func (m *MockQuotaUsageRepository) Delete(ctx context.Context, id int64) error { return nil }
func (m *MockQuotaUsageRepository) List(ctx context.Context, offset, limit int) ([]*entities.QuotaUsage, error) { return []*entities.QuotaUsage{}, nil }
func (m *MockQuotaUsageRepository) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *MockQuotaUsageRepository) GetUsageByUser(ctx context.Context, userID int64, offset, limit int) ([]*entities.QuotaUsage, error) { return []*entities.QuotaUsage{}, nil }
func (m *MockQuotaUsageRepository) GetUsageByPeriod(ctx context.Context, start, end time.Time, offset, limit int) ([]*entities.QuotaUsage, error) { return []*entities.QuotaUsage{}, nil }
func (m *MockQuotaUsageRepository) CleanupExpiredUsage(ctx context.Context, before time.Time) error { return nil }

type MockUserRepository struct{}
func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) { return &entities.User{ID: id}, nil }
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) { return nil, nil }
func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error { return nil }
func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*entities.User, error) { return []*entities.User{}, nil }
func (m *MockUserRepository) Count(ctx context.Context) (int64, error) { return 0, nil }
