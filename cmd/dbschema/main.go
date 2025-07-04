package main

import (
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

	// 查询所有表
	tables, err := getTables(db)
	if err != nil {
		log.Fatalf("Failed to get tables: %v", err)
	}

	fmt.Println("Database Tables:")
	for _, table := range tables {
		fmt.Printf("\n=== Table: %s ===\n", table)
		
		// 获取表结构
		schema, err := getTableSchema(db, table)
		if err != nil {
			log.Printf("Failed to get schema for table %s: %v", table, err)
			continue
		}
		
		fmt.Println(schema)
	}
}

func getTables(db *sql.DB) ([]string, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func getTableSchema(db *sql.DB, tableName string) (string, error) {
	query := "SELECT sql FROM sqlite_master WHERE type='table' AND name=?"
	var schema string
	err := db.QueryRow(query, tableName).Scan(&schema)
	if err != nil {
		return "", err
	}
	return schema, nil
}
