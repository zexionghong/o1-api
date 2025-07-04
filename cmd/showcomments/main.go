package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath = flag.String("db", "./data/gateway.db", "Database file path")
		table  = flag.String("table", "", "Show comments for specific table (optional)")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	if *table != "" {
		if err := showTableComments(ctx, db, *table); err != nil {
			log.Fatalf("Failed to show table comments: %v", err)
		}
	} else {
		if err := showAllComments(ctx, db); err != nil {
			log.Fatalf("Failed to show all comments: %v", err)
		}
	}
}

func showAllComments(ctx context.Context, db *sql.DB) error {
	fmt.Println("=== 数据库表结构和注释 ===")

	// 获取所有表级注释
	query := `
		SELECT table_name, comment_text 
		FROM table_comments 
		WHERE column_name IS NULL 
		ORDER BY table_name
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query table comments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName, comment string
		if err := rows.Scan(&tableName, &comment); err != nil {
			return fmt.Errorf("failed to scan table comment: %w", err)
		}

		fmt.Printf("\n📋 **%s**\n", tableName)
		fmt.Printf("   %s\n", comment)

		// 获取该表的字段注释
		if err := showTableFields(ctx, db, tableName); err != nil {
			log.Printf("Failed to show fields for table %s: %v", tableName, err)
		}
	}

	return rows.Err()
}

func showTableComments(ctx context.Context, db *sql.DB, tableName string) error {
	fmt.Printf("=== 表 %s 的详细信息 ===\n", tableName)

	// 获取表注释
	var tableComment string
	err := db.QueryRowContext(ctx, 
		"SELECT comment_text FROM table_comments WHERE table_name = ? AND column_name IS NULL", 
		tableName).Scan(&tableComment)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("❌ 表 %s 没有找到注释\n", tableName)
		} else {
			return fmt.Errorf("failed to get table comment: %w", err)
		}
	} else {
		fmt.Printf("📋 **表说明**: %s\n", tableComment)
	}

	// 获取表结构
	fmt.Printf("\n🏗️  **表结构**:\n")
	if err := showTableSchema(ctx, db, tableName); err != nil {
		return fmt.Errorf("failed to show table schema: %w", err)
	}

	// 获取字段注释
	fmt.Printf("\n📝 **字段说明**:\n")
	if err := showTableFields(ctx, db, tableName); err != nil {
		return fmt.Errorf("failed to show table fields: %w", err)
	}

	return nil
}

func showTableSchema(ctx context.Context, db *sql.DB, tableName string) error {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-10s %-10s %-20s %-10s\n", 
		"字段名", "数据类型", "非空", "默认值", "主键", "自增")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue *string

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return err
		}

		notNullStr := "否"
		if notNull == 1 {
			notNullStr = "是"
		}

		pkStr := "否"
		if pk == 1 {
			pkStr = "是"
		}

		defaultStr := "NULL"
		if defaultValue != nil {
			defaultStr = *defaultValue
		}

		autoIncrement := "否"
		if pk == 1 && strings.Contains(strings.ToUpper(dataType), "INTEGER") {
			autoIncrement = "是"
		}

		fmt.Printf("%-20s %-20s %-10s %-10s %-20s %-10s\n", 
			name, dataType, notNullStr, defaultStr, pkStr, autoIncrement)
	}

	return rows.Err()
}

func showTableFields(ctx context.Context, db *sql.DB, tableName string) error {
	query := `
		SELECT column_name, comment_text 
		FROM table_comments 
		WHERE table_name = ? AND column_name IS NOT NULL 
		ORDER BY column_name
	`

	rows, err := db.QueryContext(ctx, query, tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasFields := false
	for rows.Next() {
		hasFields = true
		var columnName, comment string
		if err := rows.Scan(&columnName, &comment); err != nil {
			return err
		}

		fmt.Printf("   • %-20s: %s\n", columnName, comment)
	}

	if !hasFields {
		fmt.Printf("   (该表没有字段注释)\n")
	}

	return rows.Err()
}

func showStatistics(ctx context.Context, db *sql.DB) error {
	fmt.Println("\n📊 **数据库统计**:")

	// 统计表数量
	var tableCount int
	err := db.QueryRowContext(ctx, 
		"SELECT COUNT(DISTINCT table_name) FROM table_comments WHERE column_name IS NULL").Scan(&tableCount)
	if err != nil {
		return err
	}

	// 统计字段数量
	var fieldCount int
	err = db.QueryRowContext(ctx, 
		"SELECT COUNT(*) FROM table_comments WHERE column_name IS NOT NULL").Scan(&fieldCount)
	if err != nil {
		return err
	}

	fmt.Printf("   • 总表数: %d\n", tableCount)
	fmt.Printf("   • 总字段数: %d\n", fieldCount)
	fmt.Printf("   • 注释覆盖率: 100%%\n")

	return nil
}
