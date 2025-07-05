package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/async"
)

// 这是一个异步配额处理的使用示例
func main() {
	// 模拟API请求处理
	simulateAPIRequests()
}

// simulateAPIRequests 模拟API请求处理
func simulateAPIRequests() {
	fmt.Println("🚀 异步配额处理示例")
	fmt.Println("==================")

	// 1. 创建异步配额服务（在实际项目中通过ServiceFactory创建）
	quotaService := createMockAsyncQuotaService()

	// 2. 模拟高并发API请求
	userID := int64(123)
	quotaType := entities.QuotaTypeRequests

	fmt.Printf("📊 开始处理用户 %d 的API请求...\n", userID)

	// 3. 处理100个并发请求
	start := time.Now()
	
	for i := 0; i < 100; i++ {
		go func(requestID int) {
			processAPIRequest(quotaService, userID, quotaType, requestID)
		}(i)
	}

	// 4. 等待一段时间让异步处理完成
	time.Sleep(2 * time.Second)

	elapsed := time.Since(start)
	fmt.Printf("⏱️  处理100个请求耗时: %v\n", elapsed)

	// 5. 获取消费者统计信息
	if asyncService, ok := quotaService.(services.QuotaServiceWithAsync); ok {
		stats := asyncService.GetConsumerStats()
		if stats != nil {
			fmt.Println("\n📈 异步消费者统计:")
			fmt.Printf("   总事件数: %d\n", stats.TotalEvents)
			fmt.Printf("   已处理事件数: %d\n", stats.ProcessedEvents)
			fmt.Printf("   失败事件数: %d\n", stats.FailedEvents)
			fmt.Printf("   丢弃事件数: %d\n", stats.DroppedEvents)
			fmt.Printf("   批次数量: %d\n", stats.BatchCount)
		}

		// 6. 检查消费者健康状态
		if asyncService.IsConsumerHealthy() {
			fmt.Println("✅ 异步消费者状态: 健康")
		} else {
			fmt.Println("❌ 异步消费者状态: 异常")
		}
	}

	fmt.Println("\n🎉 示例完成!")
}

// processAPIRequest 处理单个API请求
func processAPIRequest(quotaService services.QuotaService, userID int64, quotaType entities.QuotaType, requestID int) {
	ctx := context.Background()

	// 1. 配额检查（同步操作，确保实时性）
	allowed, err := quotaService.CheckQuota(ctx, userID, quotaType, 1)
	if err != nil {
		fmt.Printf("❌ 请求 %d: 配额检查失败 - %v\n", requestID, err)
		return
	}

	if !allowed {
		fmt.Printf("🚫 请求 %d: 配额不足\n", requestID)
		return
	}

	// 2. 模拟业务处理
	processBusinessLogic(requestID)

	// 3. 配额消费（异步操作，提升性能）
	err = quotaService.ConsumeQuota(ctx, userID, quotaType, 1)
	if err != nil {
		fmt.Printf("⚠️  请求 %d: 配额消费失败 - %v\n", requestID, err)
		return
	}

	fmt.Printf("✅ 请求 %d: 处理成功\n", requestID)
}

// processBusinessLogic 模拟业务逻辑处理
func processBusinessLogic(requestID int) {
	// 模拟一些业务处理时间
	time.Sleep(10 * time.Millisecond)
}

// createMockAsyncQuotaService 创建模拟的异步配额服务
func createMockAsyncQuotaService() services.QuotaService {
	// 在实际项目中，这会通过ServiceFactory创建
	// 这里只是为了示例，创建一个模拟的服务

	fmt.Println("🔧 创建异步配额服务...")

	// 配置异步消费者
	config := &async.QuotaConsumerConfig{
		WorkerCount:   2,                // 2个工作协程
		ChannelSize:   100,              // 100个事件缓冲
		BatchSize:     5,                // 每批处理5个事件
		FlushInterval: 1 * time.Second,  // 1秒强制刷新
		RetryAttempts: 2,                // 重试2次
		RetryDelay:    50 * time.Millisecond, // 50ms重试延迟
	}

	// 在实际项目中，这里会传入真实的Repository和Cache
	// asyncService, err := services.NewAsyncQuotaService(
	//     quotaRepo,
	//     quotaUsageRepo,
	//     userRepo,
	//     cache,
	//     invalidationService,
	//     config,
	//     logger,
	// )

	// 为了示例，返回一个模拟的服务
	return &MockQuotaService{}
}

// MockQuotaService 模拟配额服务（仅用于示例）
type MockQuotaService struct{}

func (m *MockQuotaService) CheckQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) (bool, error) {
	// 模拟配额检查，总是返回允许
	return true, nil
}

func (m *MockQuotaService) ConsumeQuota(ctx context.Context, userID int64, quotaType entities.QuotaType, value float64) error {
	// 模拟异步配额消费
	return nil
}

func (m *MockQuotaService) CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error) {
	return true, nil
}

func (m *MockQuotaService) GetQuotaStatus(ctx context.Context, userID int64) (map[string]interface{}, error) {
	return map[string]interface{}{
		"user_id": userID,
		"quotas":  []interface{}{},
	}, nil
}

// 性能对比示例
func performanceComparison() {
	fmt.Println("\n🏁 性能对比测试")
	fmt.Println("================")

	userID := int64(123)
	quotaType := entities.QuotaTypeRequests
	requestCount := 1000

	// 1. 同步处理性能测试
	fmt.Printf("🐌 同步处理 %d 个请求...\n", requestCount)
	syncStart := time.Now()
	
	for i := 0; i < requestCount; i++ {
		// 模拟同步配额处理（包含数据库写入延迟）
		time.Sleep(5 * time.Millisecond) // 模拟数据库写入延迟
	}
	
	syncDuration := time.Since(syncStart)
	fmt.Printf("   同步处理耗时: %v\n", syncDuration)
	fmt.Printf("   平均每请求: %v\n", syncDuration/time.Duration(requestCount))

	// 2. 异步处理性能测试
	fmt.Printf("\n🚀 异步处理 %d 个请求...\n", requestCount)
	asyncStart := time.Now()
	
	// 创建channel模拟异步处理
	eventChan := make(chan int, 100)
	
	// 启动消费者goroutine
	go func() {
		batch := make([]int, 0, 10)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					// 处理剩余批次
					if len(batch) > 0 {
						processBatch(batch)
					}
					return
				}
				
				batch = append(batch, event)
				if len(batch) >= 10 {
					processBatch(batch)
					batch = batch[:0]
				}
				
			case <-ticker.C:
				if len(batch) > 0 {
					processBatch(batch)
					batch = batch[:0]
				}
			}
		}
	}()
	
	// 发送事件（异步）
	for i := 0; i < requestCount; i++ {
		eventChan <- i
	}
	
	asyncDuration := time.Since(asyncStart)
	fmt.Printf("   异步处理耗时: %v\n", asyncDuration)
	fmt.Printf("   平均每请求: %v\n", asyncDuration/time.Duration(requestCount))
	
	// 关闭channel并等待处理完成
	close(eventChan)
	time.Sleep(500 * time.Millisecond)

	// 3. 性能提升计算
	improvement := float64(syncDuration-asyncDuration) / float64(syncDuration) * 100
	fmt.Printf("\n📈 性能提升: %.1f%%\n", improvement)
	fmt.Printf("🚀 吞吐量提升: %.1fx\n", float64(syncDuration)/float64(asyncDuration))
}

// processBatch 模拟批量处理
func processBatch(batch []int) {
	// 模拟批量数据库操作
	time.Sleep(20 * time.Millisecond) // 批量操作比单个操作更高效
	log.Printf("📦 处理批次: %d 个事件", len(batch))
}

// 运行完整示例
func init() {
	// 可以在这里添加性能对比测试
	// performanceComparison()
}
