package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath = flag.String("db", "./data/gateway.db", "Database file path")
		action = flag.String("action", "setup", "Action: setup, test, cleanup")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	switch *action {
	case "setup":
		if err := setupTestData(ctx, db); err != nil {
			log.Fatalf("Failed to setup test data: %v", err)
		}
		fmt.Println("Test data setup completed successfully!")
	case "test":
		if err := testBilling(ctx, db); err != nil {
			log.Fatalf("Failed to test billing: %v", err)
		}
		fmt.Println("Billing test completed successfully!")
	case "cleanup":
		if err := cleanupTestData(ctx, db); err != nil {
			log.Fatalf("Failed to cleanup test data: %v", err)
		}
		fmt.Println("Test data cleanup completed successfully!")
	default:
		fmt.Printf("Invalid action: %s\n", *action)
		fmt.Println("Available actions: setup, test, cleanup")
	}
}

func setupTestData(ctx context.Context, db *sql.DB) error {
	fmt.Println("Setting up test data...")

	// 创建仓储工厂
	repoFactory := repositories.NewRepositoryFactory(db)
	userRepo := repoFactory.UserRepository()
	apiKeyRepo := repoFactory.APIKeyRepository()

	// 创建测试用户
	user := &entities.User{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: stringPtr("Test User"),
		Status:   entities.UserStatusActive,
		Balance:  10.0, // 给用户10美元余额
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}
	fmt.Printf("Created test user: ID=%d, Username=%s, Balance=%.2f\n", user.ID, user.Username, user.Balance)

	// 创建测试API密钥
	apiKey := &entities.APIKey{
		UserID:    user.ID,
		Key:       "test-api-key-12345",
		KeyPrefix: "test-",
		Name:      stringPtr("Test API Key"),
		Status:    entities.APIKeyStatusActive,
	}

	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		return fmt.Errorf("failed to create test API key: %w", err)
	}
	fmt.Printf("Created test API key: ID=%d, Key=%s\n", apiKey.ID, apiKey.Key)

	return nil
}

func testBilling(ctx context.Context, db *sql.DB) error {
	fmt.Println("Testing billing functionality...")

	// 创建服务
	repoFactory := repositories.NewRepositoryFactory(db)
	serviceFactory := services.NewServiceFactory(repoFactory)
	billingService := serviceFactory.BillingService()
	userRepo := repoFactory.UserRepository()
	usageLogRepo := repoFactory.UsageLogRepository()

	// 查找测试用户
	user, err := userRepo.GetByUsername(ctx, "testuser")
	if err != nil {
		return fmt.Errorf("failed to find test user: %w", err)
	}
	fmt.Printf("Found test user: ID=%d, Balance=%.6f\n", user.ID, user.Balance)

	// 测试成本计算
	fmt.Println("\n=== Testing Cost Calculation ===")
	modelID := int64(1) // GPT-4
	inputTokens := 1000
	outputTokens := 500

	cost, err := billingService.CalculateCost(ctx, modelID, inputTokens, outputTokens)
	if err != nil {
		return fmt.Errorf("failed to calculate cost: %w", err)
	}
	fmt.Printf("Cost calculation: Model=%d, Input=%d tokens, Output=%d tokens, Cost=%.8f USD\n", 
		modelID, inputTokens, outputTokens, cost)

	// 创建使用日志
	fmt.Println("\n=== Testing Usage Log and Billing ===")
	usageLog := &entities.UsageLog{
		UserID:       user.ID,
		APIKeyID:     1, // 假设API密钥ID为1
		ProviderID:   1, // OpenAI
		ModelID:      modelID,
		RequestID:    fmt.Sprintf("test-req-%d", time.Now().Unix()),
		Method:       "POST",
		Endpoint:     "/v1/chat/completions",
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		RequestSize:  1024,
		ResponseSize: 2048,
		DurationMs:   1500,
		StatusCode:   200,
		Cost:         cost,
		CreatedAt:    time.Now(),
	}

	if err := usageLogRepo.Create(ctx, usageLog); err != nil {
		return fmt.Errorf("failed to create usage log: %w", err)
	}
	fmt.Printf("Created usage log: ID=%d, Cost=%.8f\n", usageLog.ID, usageLog.Cost)

	// 获取用户当前余额
	userBefore, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user before billing: %w", err)
	}
	fmt.Printf("User balance before billing: %.6f USD\n", userBefore.Balance)

	// 处理计费
	if err := billingService.ProcessBilling(ctx, usageLog); err != nil {
		return fmt.Errorf("failed to process billing: %w", err)
	}
	fmt.Println("Billing processed successfully")

	// 获取用户计费后余额
	userAfter, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user after billing: %w", err)
	}
	fmt.Printf("User balance after billing: %.6f USD\n", userAfter.Balance)
	fmt.Printf("Amount deducted: %.6f USD\n", userBefore.Balance-userAfter.Balance)

	// 测试余额不足的情况
	fmt.Println("\n=== Testing Insufficient Balance ===")
	// 创建一个高成本的使用日志
	expensiveUsageLog := &entities.UsageLog{
		UserID:       user.ID,
		APIKeyID:     1,
		ProviderID:   1,
		ModelID:      modelID,
		RequestID:    fmt.Sprintf("expensive-req-%d", time.Now().Unix()),
		Method:       "POST",
		Endpoint:     "/v1/chat/completions",
		InputTokens:  100000, // 大量token
		OutputTokens: 50000,
		TotalTokens:  150000,
		RequestSize:  10240,
		ResponseSize: 20480,
		DurationMs:   15000,
		StatusCode:   200,
		Cost:         userAfter.Balance + 1.0, // 超过用户余额
		CreatedAt:    time.Now(),
	}

	if err := usageLogRepo.Create(ctx, expensiveUsageLog); err != nil {
		return fmt.Errorf("failed to create expensive usage log: %w", err)
	}

	fmt.Printf("Created expensive usage log: Cost=%.6f USD (exceeds balance)\n", expensiveUsageLog.Cost)

	// 尝试处理计费（应该失败）
	err = billingService.ProcessBilling(ctx, expensiveUsageLog)
	if err != nil {
		fmt.Printf("Expected billing failure: %v\n", err)
	} else {
		fmt.Println("WARNING: Billing should have failed due to insufficient balance!")
	}

	// 验证余额没有变化
	userFinal, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get user final balance: %w", err)
	}
	fmt.Printf("User balance after failed billing: %.6f USD (should be unchanged)\n", userFinal.Balance)

	return nil
}

func cleanupTestData(ctx context.Context, db *sql.DB) error {
	fmt.Println("Cleaning up test data...")

	// 删除测试数据
	queries := []string{
		"DELETE FROM billing_records WHERE user_id IN (SELECT id FROM users WHERE username = 'testuser')",
		"DELETE FROM usage_logs WHERE user_id IN (SELECT id FROM users WHERE username = 'testuser')",
		"DELETE FROM api_keys WHERE user_id IN (SELECT id FROM users WHERE username = 'testuser')",
		"DELETE FROM users WHERE username = 'testuser'",
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			log.Printf("Warning: Failed to execute cleanup query: %s, error: %v", query, err)
		}
	}

	fmt.Println("Test data cleanup completed")
	return nil
}

func stringPtr(s string) *string {
	return &s
}
