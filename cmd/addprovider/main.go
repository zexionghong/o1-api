package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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

	// 添加兼容提供商
	query := `
		INSERT INTO providers (name, slug, base_url, status, health_status, priority, timeout_seconds, retry_attempts, health_check_interval, created_at, updated_at)
		VALUES ('OpenAI Compatible', 'openai-compatible', 'https://api.compatible-provider.com/v1', 'active', 'healthy', 3, 30, 3, 60, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Fatalf("Failed to insert provider: %v", err)
	}

	providerID, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("Failed to get provider ID: %v", err)
	}

	fmt.Printf("✅ Added provider: OpenAI Compatible (ID: %d)\n", providerID)

	// 添加对所有OpenAI模型的支持
	openaiModels := []string{
		"gpt-4", "gpt-4-32k", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
		"gpt-3.5-turbo", "gpt-3.5-turbo-16k",
		"text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002",
	}

	supportQuery := `
		INSERT INTO provider_model_support (provider_id, model_slug, upstream_model_name, priority, enabled, created_at, updated_at)
		VALUES (?, ?, ?, 2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	for _, model := range openaiModels {
		_, err := db.ExecContext(ctx, supportQuery, providerID, model, model)
		if err != nil {
			log.Printf("Warning: Failed to add support for model %s: %v", model, err)
		} else {
			fmt.Printf("   ✅ Added support for: %s\n", model)
		}
	}

	fmt.Println("\n🎉 Successfully added OpenAI Compatible provider with support for all OpenAI models!")
}
