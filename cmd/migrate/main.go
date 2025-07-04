package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func main() {
	var (
		direction = flag.String("direction", "up", "Migration direction: up or down")
		steps     = flag.Int("steps", 0, "Number of migration steps (0 for all)")
		dbPath    = flag.String("db", "./data/gateway.db", "Database file path")
	)
	flag.Parse()

	// 确保数据目录存在
	dataDir := filepath.Dir(*dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 创建migrate驱动
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("Failed to create migrate driver: %v", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://migrations"
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		// 如果在当前目录找不到migrations，尝试从项目根目录查找
		if _, err := os.Stat("../../migrations"); err == nil {
			migrationsPath = "file://../../migrations"
		} else {
			log.Fatalf("Migrations directory not found")
		}
	}

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "sqlite3", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// 执行迁移
	switch *direction {
	case "up":
		if *steps == 0 {
			err = m.Up()
		} else {
			err = m.Steps(*steps)
		}
	case "down":
		if *steps == 0 {
			err = m.Down()
		} else {
			err = m.Steps(-*steps)
		}
	default:
		log.Fatalf("Invalid direction: %s (use 'up' or 'down')", *direction)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to apply")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		fmt.Printf("Migration %s completed successfully\n", *direction)
	}
}
