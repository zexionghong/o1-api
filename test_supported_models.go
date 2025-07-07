package main

import (
	"context"
	"fmt"
	"log"

	"ai-api-gateway/internal/infrastructure/database"
	"ai-api-gateway/internal/infrastructure/repositories"
)

func main() {
	// 连接数据库 (使用PostgreSQL)
	gormConfig := database.GormConfig{
		Host:     "47.76.73.118",
		Port:     5432,
		User:     "proxy",
		Password: "pPhnbrlIKfYA",
		DBName:   "ai",
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	db, err := database.NewGormDB(gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建仓储
	toolRepo := repositories.NewToolRepositoryGorm(db)

	ctx := context.Background()

	// 测试 GetTools 方法
	fmt.Println("=== Testing GetTools ===")
	tools, err := toolRepo.GetTools(ctx)
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	for _, tool := range tools {
		fmt.Printf("Tool: %s (%s)\n", tool.Name, tool.ID)
		fmt.Printf("  Category: %s\n", tool.Category)
		fmt.Printf("  Supported Models (%d):\n", len(tool.SupportedModels))
		for _, model := range tool.SupportedModels {
			fmt.Printf("    - %s (ID: %d, Type: %s)\n", model.Name, model.ID, model.ModelType)
		}
		fmt.Println()
	}

	// 测试 GetToolByID 方法
	if len(tools) > 0 {
		fmt.Println("=== Testing GetToolByID ===")
		firstToolID := tools[0].ID
		tool, err := toolRepo.GetToolByID(ctx, firstToolID)
		if err != nil {
			log.Fatalf("Failed to get tool by ID: %v", err)
		}

		fmt.Printf("Tool: %s (%s)\n", tool.Name, tool.ID)
		fmt.Printf("  Category: %s\n", tool.Category)
		fmt.Printf("  Supported Models (%d):\n", len(tool.SupportedModels))
		for _, model := range tool.SupportedModels {
			fmt.Printf("    - %s (ID: %d, Type: %s)\n", model.Name, model.ID, model.ModelType)
		}
	}

	fmt.Println("Test completed successfully!")
}
