package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ai-api-gateway/internal/config"
	"ai-api-gateway/internal/infrastructure/cache"
)

func main() {
	fmt.Println("🔧 Redis缓存功能测试")
	fmt.Println("==================================================")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 配置加载失败: %v", err)
	}

	// 初始化Redis缓存
	fmt.Println("🔌 连接Redis...")
	cacheManager, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		log.Fatalf("❌ Redis连接失败: %v", err)
	}
	defer cacheManager.Close()

	fmt.Println("✅ Redis连接成功!")
	fmt.Printf("📍 Redis地址: %s:%d (DB: %d)\n", cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.DB)
	fmt.Println()

	ctx := context.Background()

	// 测试1: 基本缓存操作
	fmt.Println("🧪 测试1: 基本缓存操作")
	testBasicCache(ctx, cacheManager)
	fmt.Println()

	// 测试2: 分布式锁
	fmt.Println("🔒 测试2: 分布式锁功能")
	testDistributedLock(ctx, cacheManager)
	fmt.Println()

	// 测试3: 缓存过期
	fmt.Println("⏰ 测试3: 缓存过期功能")
	testCacheExpiration(ctx, cacheManager)
	fmt.Println()

	// 测试4: 检查现有缓存
	fmt.Println("📊 测试4: 检查现有缓存数据")
	testExistingCache(ctx, cacheManager)
	fmt.Println()

	fmt.Println("🎉 所有Redis缓存测试完成!")
}

func testBasicCache(ctx context.Context, cache cache.Cache) {
	testKey := "test:cache:basic"
	testValue := "Hello Redis Cache!"

	// 设置缓存
	fmt.Printf("   📝 设置缓存: %s = %s\n", testKey, testValue)
	err := cache.Set(ctx, testKey, testValue, 5*time.Minute)
	if err != nil {
		fmt.Printf("   ❌ 设置缓存失败: %v\n", err)
		return
	}

	// 获取缓存
	fmt.Printf("   📖 读取缓存: %s\n", testKey)
	result, err := cache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("   ❌ 读取缓存失败: %v\n", err)
		return
	}

	if result == testValue {
		fmt.Printf("   ✅ 缓存读写成功! 值: %s\n", result)
	} else {
		fmt.Printf("   ❌ 缓存值不匹配! 期望: %s, 实际: %s\n", testValue, result)
	}

	// 删除缓存
	fmt.Printf("   🗑️  删除缓存: %s\n", testKey)
	err = cache.Delete(ctx, testKey)
	if err != nil {
		fmt.Printf("   ❌ 删除缓存失败: %v\n", err)
		return
	}

	// 验证删除
	_, err = cache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("   ✅ 缓存删除成功!\n")
	} else {
		fmt.Printf("   ❌ 缓存删除失败，仍能读取到值\n")
	}
}

func testDistributedLock(ctx context.Context, cache cache.Cache) {
	lockKey := "test:lock:distributed"
	lockValue := "test-lock-value"
	lockTTL := 10 * time.Second

	// 获取锁
	fmt.Printf("   🔐 获取分布式锁: %s\n", lockKey)
	acquired, err := cache.AcquireLock(ctx, lockKey, lockValue, lockTTL)
	if err != nil {
		fmt.Printf("   ❌ 获取锁失败: %v\n", err)
		return
	}

	if acquired {
		fmt.Printf("   ✅ 成功获取分布式锁!\n")

		// 尝试再次获取同一个锁（应该失败）
		fmt.Printf("   🔄 尝试重复获取锁...\n")
		acquired2, err := cache.AcquireLock(ctx, lockKey, "another-value", lockTTL)
		if err != nil {
			fmt.Printf("   ❌ 重复获取锁时出错: %v\n", err)
		} else if !acquired2 {
			fmt.Printf("   ✅ 正确拒绝了重复锁请求!\n")
		} else {
			fmt.Printf("   ❌ 错误地允许了重复锁请求!\n")
		}

		// 释放锁
		fmt.Printf("   🔓 释放分布式锁: %s\n", lockKey)
		released, err := cache.ReleaseLock(ctx, lockKey, lockValue)
		if err != nil {
			fmt.Printf("   ❌ 释放锁失败: %v\n", err)
		} else if released {
			fmt.Printf("   ✅ 成功释放分布式锁!\n")
		} else {
			fmt.Printf("   ❌ 锁释放失败（可能已过期）\n")
		}
	} else {
		fmt.Printf("   ❌ 获取分布式锁失败!\n")
	}
}

func testCacheExpiration(ctx context.Context, cache cache.Cache) {
	testKey := "test:cache:expiration"
	testValue := "This will expire soon"

	// 设置短期缓存
	fmt.Printf("   ⏱️  设置2秒过期的缓存: %s\n", testKey)
	err := cache.Set(ctx, testKey, testValue, 2*time.Second)
	if err != nil {
		fmt.Printf("   ❌ 设置缓存失败: %v\n", err)
		return
	}

	// 立即读取
	result, err := cache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("   ❌ 立即读取失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 立即读取成功: %s\n", result)

	// 等待过期
	fmt.Printf("   ⏳ 等待3秒让缓存过期...\n")
	time.Sleep(3 * time.Second)

	// 尝试读取过期缓存
	_, err = cache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("   ✅ 缓存正确过期，无法读取!\n")
	} else {
		fmt.Printf("   ❌ 缓存未正确过期，仍能读取!\n")
	}
}

func testExistingCache(ctx context.Context, cache cache.Cache) {
	// 检查一些具体的缓存键
	specificKeys := []string{
		"gateway:stats",
		"gateway:health", 
		"gateway:config",
		"user:1",
		"user:2",
		"api_key:1",
		"api_key:2",
		"quota:user:1",
		"quota:user:2",
	}

	fmt.Printf("   🔍 检查现有缓存数据...\n")
	foundCount := 0
	
	for _, key := range specificKeys {
		value, err := cache.Get(ctx, key)
		if err == nil {
			fmt.Printf("   ✅ 发现缓存: %s = %s\n", key, value)
			foundCount++
		} else {
			fmt.Printf("   ℹ️  缓存不存在: %s\n", key)
		}
	}

	if foundCount > 0 {
		fmt.Printf("   📊 总共发现 %d 个缓存项\n", foundCount)
	} else {
		fmt.Printf("   📊 未发现任何现有缓存项（这是正常的）\n")
	}
}
