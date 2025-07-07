package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	var (
		direction = flag.String("direction", "up", "Migration direction: up or down")
		steps     = flag.Int("steps", 0, "Number of migration steps (0 for all)")
		dsn       = flag.String("dsn", "host=localhost port=5432 user=gateway password=gateway_password dbname=gateway sslmode=disable", "PostgreSQL DSN")
	)
	flag.Parse()

	// 打开数据库连接
	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// 创建migrate驱动
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migrate driver: %v", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://migrations-postgres"
	if _, err := os.Stat("migrations-postgres"); os.IsNotExist(err) {
		// 如果在当前目录找不到migrations-postgres，尝试从项目根目录查找
		if _, err := os.Stat("../../migrations-postgres"); err == nil {
			migrationsPath = "file://../../migrations-postgres"
		} else {
			log.Fatalf("PostgreSQL migrations directory not found")
		}
	}

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
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
		fmt.Printf("PostgreSQL migration %s completed successfully\n", *direction)
	}
}
