package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver
)

// Config 数据库配置
type Config struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// Connection 数据库连接管理器
type Connection struct {
	db     *sql.DB
	dbx    *sqlx.DB
	config *Config
}

// NewConnection 创建数据库连接
func NewConnection(config *Config) (*Connection, error) {
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建sqlx连接
	dbx := sqlx.NewDb(db, config.Driver)

	return &Connection{
		db:     db,
		dbx:    dbx,
		config: config,
	}, nil
}

// DB 获取数据库连接
func (c *Connection) DB() *sql.DB {
	return c.db
}

// DBX 获取sqlx数据库连接
func (c *Connection) DBX() *sqlx.DB {
	return c.dbx
}

// Close 关闭数据库连接
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping 测试数据库连接
func (c *Connection) Ping() error {
	return c.db.Ping()
}

// Stats 获取连接池统计信息
func (c *Connection) Stats() sql.DBStats {
	return c.db.Stats()
}

// BeginTx 开始事务
func (c *Connection) BeginTx() (*sql.Tx, error) {
	return c.db.Begin()
}

// DefaultConfig 默认数据库配置
func DefaultConfig() *Config {
	return &Config{
		Driver:          "sqlite",
		DSN:             "./data/gateway.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300 * time.Second,
	}
}
