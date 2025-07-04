package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath     = flag.String("db", "./data/gateway.db", "Database file path")
		gatewayURL = flag.String("gateway", "http://localhost:8080", "Gateway URL")
		action     = flag.String("action", "setup", "Action: setup, test, cleanup")
		apiKey     = flag.String("apikey", "", "API key for testing (auto-generated if empty)")
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
		key, err := setupTestEnvironment(ctx, db)
		if err != nil {
			log.Fatalf("Failed to setup test environment: %v", err)
		}
		fmt.Printf("🎉 Test environment setup completed!\n")
		fmt.Printf("📋 API Key: %s\n", key)
		fmt.Printf("🚀 Ready for testing!\n")
	case "test":
		if *apiKey == "" {
			log.Fatalf("API key is required for testing. Use -apikey flag or run setup first.")
		}
		if err := runE2ETest(ctx, *gatewayURL, *apiKey); err != nil {
			log.Fatalf("E2E test failed: %v", err)
		}
		fmt.Println("✅ E2E test completed successfully!")
	case "cleanup":
		if err := cleanupTestEnvironment(ctx, db); err != nil {
			log.Fatalf("Failed to cleanup test environment: %v", err)
		}
		fmt.Println("🧹 Test environment cleaned up!")
	default:
		fmt.Printf("Invalid action: %s\n", *action)
		fmt.Println("Available actions: setup, test, cleanup")
	}
}

func setupTestEnvironment(ctx context.Context, db *sql.DB) (string, error) {
	fmt.Println("🔧 Setting up test environment...")

	// 创建仓储
	repoFactory := repositories.NewRepositoryFactory(db)
	userRepo := repoFactory.UserRepository()
	apiKeyRepo := repoFactory.APIKeyRepository()

	// 1. 创建测试用户
	fmt.Println("👤 Creating test user...")
	user := &entities.User{
		Username: "e2e-test-user",
		Email:    "e2e-test@example.com",
		FullName: stringPtr("E2E Test User"),
		Status:   entities.UserStatusActive,
		Balance:  100.0, // 给用户100美元余额
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return "", fmt.Errorf("failed to create test user: %w", err)
	}
	fmt.Printf("   ✅ Created user: ID=%d, Username=%s, Balance=%.2f\n", 
		user.ID, user.Username, user.Balance)

	// 2. 创建API密钥
	fmt.Println("🔑 Creating API key...")
	apiKeyStr := generateAPIKey()
	apiKey := &entities.APIKey{
		UserID:    user.ID,
		Key:       apiKeyStr,
		KeyPrefix: "sk-e2e",
		Name:      stringPtr("E2E Test Key"),
		Status:    entities.APIKeyStatusActive,
	}

	if err := apiKeyRepo.Create(ctx, apiKey); err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}
	fmt.Printf("   ✅ Created API key: ID=%d, Key=%s\n", apiKey.ID, apiKey.Key)

	// 3. 验证提供商和模型配置
	fmt.Println("🔍 Verifying providers and models...")
	if err := verifyConfiguration(ctx, db); err != nil {
		return "", fmt.Errorf("failed to verify configuration: %w", err)
	}

	return apiKeyStr, nil
}

func runE2ETest(ctx context.Context, gatewayURL, apiKey string) error {
	fmt.Println("🧪 Running E2E test...")

	// 测试用例
	testCases := []struct {
		name        string
		model       string
		messages    []map[string]string
		expectError bool
	}{
		{
			name:  "GPT-4 Chat Completion",
			model: "gpt-4",
			messages: []map[string]string{
				{"role": "user", "content": "Hello! Please respond with exactly 'Test successful' and nothing else."},
			},
			expectError: false,
		},
		{
			name:  "GPT-3.5 Turbo Chat Completion",
			model: "gpt-3.5-turbo",
			messages: []map[string]string{
				{"role": "user", "content": "Say 'Hello from GPT-3.5' and nothing else."},
			},
			expectError: false,
		},
		{
			name:  "Claude 3 Haiku Chat Completion",
			model: "claude-3-haiku",
			messages: []map[string]string{
				{"role": "user", "content": "Respond with 'Claude test successful' only."},
			},
			expectError: false,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\n📝 Test %d: %s\n", i+1, tc.name)
		
		if err := testChatCompletion(gatewayURL, apiKey, tc.model, tc.messages); err != nil {
			if tc.expectError {
				fmt.Printf("   ✅ Expected error occurred: %v\n", err)
			} else {
				fmt.Printf("   ❌ Unexpected error: %v\n", err)
				return fmt.Errorf("test case '%s' failed: %w", tc.name, err)
			}
		} else {
			if tc.expectError {
				fmt.Printf("   ❌ Expected error but test passed\n")
				return fmt.Errorf("test case '%s' should have failed", tc.name)
			} else {
				fmt.Printf("   ✅ Test passed\n")
			}
		}
	}

	return nil
}

func testChatCompletion(gatewayURL, apiKey, model string, messages []map[string]string) error {
	// 构建请求
	requestBody := map[string]interface{}{
		"model":      model,
		"messages":   messages,
		"max_tokens": 50,
		"temperature": 0.1,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	url := fmt.Sprintf("%s/v1/chat/completions", gatewayURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("   📊 Status: %d\n", resp.StatusCode)
	fmt.Printf("   📄 Response: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 验证响应结构
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					fmt.Printf("   💬 AI Response: %s\n", content)
				}
			}
		}
	}

	return nil
}

func verifyConfiguration(ctx context.Context, db *sql.DB) error {
	// 检查提供商
	var providerCount int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM providers WHERE status = 'active'").Scan(&providerCount)
	if err != nil {
		return fmt.Errorf("failed to count providers: %w", err)
	}
	fmt.Printf("   📊 Active providers: %d\n", providerCount)

	// 检查模型
	var modelCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM models WHERE status = 'active'").Scan(&modelCount)
	if err != nil {
		return fmt.Errorf("failed to count models: %w", err)
	}
	fmt.Printf("   📊 Active models: %d\n", modelCount)

	// 检查模型支持
	var supportCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM provider_model_support WHERE enabled = true").Scan(&supportCount)
	if err != nil {
		return fmt.Errorf("failed to count model support: %w", err)
	}
	fmt.Printf("   📊 Model support mappings: %d\n", supportCount)

	// 检查定价
	var pricingCount int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM model_pricing").Scan(&pricingCount)
	if err != nil {
		return fmt.Errorf("failed to count pricing: %w", err)
	}
	fmt.Printf("   📊 Pricing records: %d\n", pricingCount)

	if providerCount == 0 || modelCount == 0 || supportCount == 0 || pricingCount == 0 {
		return fmt.Errorf("incomplete configuration: providers=%d, models=%d, support=%d, pricing=%d", 
			providerCount, modelCount, supportCount, pricingCount)
	}

	return nil
}

func cleanupTestEnvironment(ctx context.Context, db *sql.DB) error {
	fmt.Println("🧹 Cleaning up test environment...")

	// 删除测试数据
	queries := []string{
		"DELETE FROM billing_records WHERE user_id IN (SELECT id FROM users WHERE username = 'e2e-test-user')",
		"DELETE FROM usage_logs WHERE user_id IN (SELECT id FROM users WHERE username = 'e2e-test-user')",
		"DELETE FROM api_keys WHERE user_id IN (SELECT id FROM users WHERE username = 'e2e-test-user')",
		"DELETE FROM users WHERE username = 'e2e-test-user'",
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			log.Printf("Warning: Failed to execute cleanup query: %s, error: %v", query, err)
		}
	}

	fmt.Println("   ✅ Test data cleaned up")
	return nil
}

func generateAPIKey() string {
	// 生成32字节随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "sk-e2e" + hex.EncodeToString(bytes)
}

func stringPtr(s string) *string {
	return &s
}
