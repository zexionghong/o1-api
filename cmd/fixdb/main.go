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

	// 删除脏状态的迁移记录
	_, err = db.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = '005'")
	if err != nil {
		log.Printf("Warning: Failed to delete migration record: %v", err)
	}

	// 删除可能已创建的table_comments表
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS table_comments")
	if err != nil {
		log.Printf("Warning: Failed to drop table_comments: %v", err)
	}

	fmt.Println("✅ Database state fixed!")
}
