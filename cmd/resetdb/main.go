package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath = flag.String("db", "./data/gateway.db", "Database file path")
		force  = flag.Bool("force", false, "Force reset database")
	)
	flag.Parse()

	if !*force {
		fmt.Println("This will reset the database. Use -force to confirm.")
		return
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 删除所有表
	tables := []string{
		"billing_records",
		"usage_logs", 
		"quota_usage",
		"quotas",
		"model_pricing",
		"models",
		"providers",
		"api_keys",
		"users",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		if err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			fmt.Printf("Dropped table: %s\n", table)
		}
	}

	fmt.Println("Database reset completed!")
}
