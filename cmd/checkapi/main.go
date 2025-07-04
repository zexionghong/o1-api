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

	// 检查API密钥表结构
	fmt.Println("=== API Keys Table Structure ===")
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(api_keys)")
	if err != nil {
		log.Fatalf("Failed to get table info: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue *string

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatalf("Failed to scan table info: %v", err)
		}

		fmt.Printf("Column: %s, Type: %s, NotNull: %d, PK: %d\n", name, dataType, notNull, pk)
	}

	// 检查现有的API密钥
	fmt.Println("\n=== Existing API Keys ===")
	apiRows, err := db.QueryContext(ctx, "SELECT id, user_id, key, key_prefix, name, status FROM api_keys")
	if err != nil {
		log.Fatalf("Failed to query api keys: %v", err)
	}
	defer apiRows.Close()

	for apiRows.Next() {
		var id, userID int64
		var key, keyPrefix string
		var name, status *string

		err := apiRows.Scan(&id, &userID, &key, &keyPrefix, &name, &status)
		if err != nil {
			log.Fatalf("Failed to scan api key: %v", err)
		}

		nameStr := "NULL"
		if name != nil {
			nameStr = *name
		}
		statusStr := "NULL"
		if status != nil {
			statusStr = *status
		}

		fmt.Printf("ID: %d, UserID: %d, Key: %s, Prefix: %s, Name: %s, Status: %s\n", 
			id, userID, key, keyPrefix, nameStr, statusStr)
	}
}
