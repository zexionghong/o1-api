package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"ai-api-gateway/internal/domain/values"

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

	// 生成符合格式的API密钥
	keyGen := values.NewAPIKeyGenerator()
	newKey, _, prefix, err := keyGen.Generate()
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	fmt.Printf("🔑 Generated new API key: %s\n", newKey)
	fmt.Printf("📋 Prefix: %s\n", prefix)

	// 验证格式
	if keyGen.ValidateFormat(newKey) {
		fmt.Println("✅ Key format is valid")
	} else {
		fmt.Println("❌ Key format is invalid")
	}

	// 更新数据库中的API密钥
	query := `UPDATE api_keys SET key = ?, key_prefix = ? WHERE user_id = 2`
	_, err = db.ExecContext(ctx, query, newKey, prefix)
	if err != nil {
		log.Fatalf("Failed to update API key: %v", err)
	}

	fmt.Printf("✅ Updated API key in database\n")
	fmt.Printf("🎯 Use this API key for testing: %s\n", newKey)

	// 验证更新
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM api_keys WHERE key = ?", newKey).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}

	if count > 0 {
		fmt.Println("✅ API key successfully updated in database")
	} else {
		fmt.Println("❌ API key not found in database")
	}
}
