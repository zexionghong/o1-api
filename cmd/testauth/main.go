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
	testAPIKey := "sk-e2e2a3c9e06e2a99a9b826c0d075a98850937828db4bb2fa61cdeb7ac89bbfc0012"
	
	fmt.Printf("🔍 Testing API key validation for: %s\n", testAPIKey)

	apiKey, user, err := apiKeyService.ValidateAPIKey(ctx, testAPIKey)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		
		// 尝试直接查询数据库
		fmt.Println("\n🔍 Direct database query:")
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM api_keys WHERE key = ?", testAPIKey).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Database query failed: %v\n", err)
		} else {
			fmt.Printf("📊 Found %d matching records\n", count)
		}
		
		// 查看所有API密钥
		fmt.Println("\n📋 All API keys in database:")
		rows, err := db.QueryContext(ctx, "SELECT id, user_id, key, status FROM api_keys")
		if err != nil {
			fmt.Printf("❌ Failed to query all keys: %v\n", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var id, userID int64
				var key, status string
				if err := rows.Scan(&id, &userID, &key, &status); err != nil {
					fmt.Printf("❌ Failed to scan row: %v\n", err)
					continue
				}
				fmt.Printf("   ID: %d, UserID: %d, Key: %s, Status: %s\n", id, userID, key, status)
			}
		}
		
		return
	}

	fmt.Printf("✅ Validation successful!\n")
	fmt.Printf("   API Key ID: %d\n", apiKey.ID)
	fmt.Printf("   User ID: %d\n", user.ID)
	fmt.Printf("   User: %s (%s)\n", user.Username, user.Email)
	fmt.Printf("   Status: %s\n", apiKey.Status)
	fmt.Printf("   Balance: %.6f USD\n", user.Balance)
}
