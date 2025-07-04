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

	// 检查迁移状态
	fmt.Println("=== Migration Status ===")
	rows, err := db.QueryContext(ctx, "SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Printf("Failed to query migrations: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var version, appliedAt string
			if err := rows.Scan(&version, &appliedAt); err != nil {
				log.Printf("Failed to scan migration: %v", err)
				continue
			}
			fmt.Printf("Version: %s, Applied: %s\n", version, appliedAt)
		}
	}

	// 检查表是否存在
	fmt.Println("\n=== Table Status ===")
	tables := []string{"users", "api_keys", "providers", "models", "model_pricing", "provider_model_support", "table_comments"}
	for _, table := range tables {
		var exists bool
		err := db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)", table).Scan(&exists)
		if err != nil {
			fmt.Printf("❌ %s: Error checking - %v\n", table, err)
		} else if exists {
			fmt.Printf("✅ %s: Exists\n", table)
		} else {
			fmt.Printf("❌ %s: Not found\n", table)
		}
	}
}
