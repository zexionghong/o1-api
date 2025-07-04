package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/infrastructure/redis"
)

func main() {
	fmt.Println("🧪 Testing Redis and Distributed Lock functionality...")

	// 初始化配置
	if err := config.InitConfig("configs/config.yaml"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化日志
	log := logger.NewLogger()

	// 创建Redis工厂
	redisFactory, err := redis.NewRedisFactory(log)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Fatal("Failed to create Redis factory")
	}
	defer redisFactory.Close()

	fmt.Println("✅ Redis connection established")

	// 获取服务
	cache := redisFactory.GetCacheService()
	lockService := redisFactory.GetLockService()

	ctx := context.Background()

	// 测试1: 基本缓存功能
	fmt.Println("\n🔍 Test 1: Basic cache functionality")
	testBasicCache(ctx, cache)

	// 测试2: 分布式锁功能
	fmt.Println("\n🔍 Test 2: Distributed lock functionality")
	testDistributedLock(ctx, lockService)

	// 测试3: 并发锁测试
	fmt.Println("\n🔍 Test 3: Concurrent lock test")
	testConcurrentLock(ctx, lockService)

	fmt.Println("\n🎉 All Redis tests completed!")
}

func testBasicCache(ctx context.Context, cache *redis.CacheService) {
	// 测试设置和获取
	testKey := "test:user:123"
	testValue := map[string]interface{}{
		"id":      123,
		"name":    "Test User",
		"balance": 100.50,
	}

	// 设置缓存
	if err := cache.Set(ctx, testKey, testValue, 5*time.Minute); err != nil {
		fmt.Printf("❌ Failed to set cache: %v\n", err)
		return
	}
	fmt.Println("✅ Cache set successfully")

	// 获取缓存
	var retrieved map[string]interface{}
	if err := cache.Get(ctx, testKey, &retrieved); err != nil {
		fmt.Printf("❌ Failed to get cache: %v\n", err)
		return
	}
	fmt.Printf("✅ Cache retrieved: %+v\n", retrieved)

	// 检查TTL
	ttl, err := cache.TTL(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Failed to get TTL: %v\n", err)
		return
	}
	fmt.Printf("✅ Cache TTL: %v\n", ttl)

	// 删除缓存
	if err := cache.Delete(ctx, testKey); err != nil {
		fmt.Printf("❌ Failed to delete cache: %v\n", err)
		return
	}
	fmt.Println("✅ Cache deleted successfully")
}

func testDistributedLock(ctx context.Context, lockService *redis.DistributedLockService) {
	lockKey := "test:lock:user:456"

	// 创建锁
	lock := lockService.NewLock(lockKey, nil)

	// 获取锁
	if err := lock.Lock(ctx); err != nil {
		fmt.Printf("❌ Failed to acquire lock: %v\n", err)
		return
	}
	fmt.Println("✅ Lock acquired successfully")

	// 检查锁是否被持有
	held, err := lock.IsHeld(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to check lock status: %v\n", err)
		return
	}
	fmt.Printf("✅ Lock held status: %v\n", held)

	// 延长锁
	if err := lock.Extend(ctx, 1*time.Minute); err != nil {
		fmt.Printf("❌ Failed to extend lock: %v\n", err)
		return
	}
	fmt.Println("✅ Lock extended successfully")

	// 释放锁
	if err := lock.Unlock(ctx); err != nil {
		fmt.Printf("❌ Failed to release lock: %v\n", err)
		return
	}
	fmt.Println("✅ Lock released successfully")
}

func testConcurrentLock(ctx context.Context, lockService *redis.DistributedLockService) {
	lockKey := "test:concurrent:lock"
	numGoroutines := 5
	var wg sync.WaitGroup
	var successCount int32
	var mu sync.Mutex

	fmt.Printf("Starting %d concurrent goroutines to compete for lock...\n", numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 使用WithLock方法
			err := lockService.WithLock(ctx, lockKey, nil, func() error {
				fmt.Printf("🔒 Goroutine %d acquired lock\n", id)
				
				// 模拟一些工作
				time.Sleep(100 * time.Millisecond)
				
				mu.Lock()
				successCount++
				mu.Unlock()
				
				fmt.Printf("🔓 Goroutine %d releasing lock\n", id)
				return nil
			})

			if err != nil {
				fmt.Printf("❌ Goroutine %d failed to execute with lock: %v\n", id, err)
			}
		}(i)
	}

	wg.Wait()

	mu.Lock()
	finalCount := successCount
	mu.Unlock()

	fmt.Printf("✅ Concurrent test completed. %d/%d goroutines successfully executed with lock\n", 
		finalCount, numGoroutines)

	if finalCount == int32(numGoroutines) {
		fmt.Println("✅ All goroutines executed successfully - lock is working correctly!")
	} else {
		fmt.Printf("⚠️  Only %d out of %d goroutines succeeded\n", finalCount, numGoroutines)
	}
}
