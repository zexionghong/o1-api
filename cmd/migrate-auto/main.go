package main

import (
	"ai-api-gateway/internal/infrastructure/config"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"      // PostgreSQL driver
	_ "modernc.org/sqlite"     // SQLite driver
)

func main() {
	var (
		direction  = flag.String("direction", "up", "Migration direction: up or down")
		steps      = flag.Int("steps", 0, "Number of migration steps (0 for all)")
		configPath = flag.String("config", "", "Path to configuration file")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Using database driver: %s\n", cfg.Database.Driver)
	fmt.Printf("Database DSN: %s\n", maskDSN(cfg.Database.DSN))

	// 根据数据库类型执行迁移
	switch cfg.Database.Driver {
	case "postgres":
		err = migratePostgreSQL(cfg.Database.DSN, *direction, *steps)
	case "sqlite":
		err = migrateSQLite(cfg.Database.DSN, *direction, *steps)
	default:
		log.Fatalf("Unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

// migratePostgreSQL 执行PostgreSQL迁移
func migratePostgreSQL(dsn, direction string, steps int) error {
	// 打开数据库连接
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	// 创建migrate驱动
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL migrate driver: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://migrations-postgres"
	if _, err := os.Stat("migrations-postgres"); os.IsNotExist(err) {
		// 如果在当前目录找不到migrations-postgres，尝试从项目根目录查找
		if _, err := os.Stat("../../migrations-postgres"); err == nil {
			migrationsPath = "file://../../migrations-postgres"
		} else {
			return fmt.Errorf("PostgreSQL migrations directory not found")
		}
	}

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create PostgreSQL migrate instance: %w", err)
	}

	// 执行迁移
	err = executeMigration(m, direction, steps)
	if err != nil {
		return err
	}

	fmt.Printf("PostgreSQL migration %s completed successfully\n", direction)
	return nil
}

// migrateSQLite 执行SQLite迁移
func migrateSQLite(dsn, direction string, steps int) error {
	// 确保数据目录存在
	dataDir := filepath.Dir(dsn)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}
	defer db.Close()

	// 创建migrate驱动
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create SQLite migrate driver: %w", err)
	}

	// 获取迁移文件路径
	migrationsPath := "file://migrations"
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		// 如果在当前目录找不到migrations，尝试从项目根目录查找
		if _, err := os.Stat("../../migrations"); err == nil {
			migrationsPath = "file://../../migrations"
		} else {
			return fmt.Errorf("SQLite migrations directory not found")
		}
	}

	// 创建migrate实例
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create SQLite migrate instance: %w", err)
	}

	// 执行迁移
	err = executeMigration(m, direction, steps)
	if err != nil {
		return err
	}

	fmt.Printf("SQLite migration %s completed successfully\n", direction)
	return nil
}

// executeMigration 执行迁移操作
func executeMigration(m *migrate.Migrate, direction string, steps int) error {
	switch direction {
	case "up":
		if steps == 0 {
			return m.Up()
		} else {
			return m.Steps(steps)
		}
	case "down":
		if steps == 0 {
			return m.Down()
		} else {
			return m.Steps(-steps)
		}
	default:
		return fmt.Errorf("invalid direction: %s (use 'up' or 'down')", direction)
	}
}

// maskDSN 隐藏DSN中的敏感信息
func maskDSN(dsn string) string {
	// 简单的密码隐藏逻辑
	// 这里可以根据需要实现更复杂的隐藏逻辑
	if len(dsn) > 50 {
		return dsn[:20] + "***" + dsn[len(dsn)-10:]
	}
	return dsn
}
