package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/repositories"

	_ "modernc.org/sqlite"
)

func main() {
	// 打开数据库连接
	db, err := sql.Open("sqlite", "./data/gateway.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建服务
	repoFactory := repositories.NewRepositoryFactory(db)
	serviceFactory := services.NewServiceFactory(repoFactory)
	apiKeyService := serviceFactory.APIKeyService()

	// 测试API密钥验证
	testAPIKey := "ak_ede198ed25b71c95cb9b38ac970e4f248ed2c6d1d658a19475b2afeab5cf9822"
	
	fmt.Printf("🔍 Testing API key validation for: %s\n", testAPIKey)

	apiKey, user, err := apiKeyService.ValidateAPIKey(ctx, testAPIKey)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Validation successful!\n")
	fmt.Printf("   API Key ID: %d\n", apiKey.ID)
	fmt.Printf("   User ID: %d\n", user.ID)
	fmt.Printf("   User: %s (%s)\n", user.Username, user.Email)
	fmt.Printf("   Status: %s\n", apiKey.Status)
	fmt.Printf("   Balance: %.6f USD\n", user.Balance)
}
