package database

import (
	"fmt"
	"log"
	"time"

	"ai-api-gateway/internal/domain/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormConfig GORM数据库配置
type GormConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// NewGormDB 创建GORM数据库连接
func NewGormDB(config GormConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode, config.TimeZone)

	// 配置GORM日志
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // 慢SQL阈值
				LogLevel:                  logger.Info, // 日志级别
				IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound错误
				Colorful:                  false,       // 禁用彩色打印
			},
		),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间

	return db, nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	// 定义所有需要迁移的模型
	models := []interface{}{
		&entities.User{},
		&entities.APIKey{},
		&entities.Provider{},
		&entities.Model{},
		&entities.ModelPricing{},
		&entities.ProviderModelSupport{},
		&entities.Quota{},
		&entities.QuotaUsage{},
		&entities.UsageLog{},
		&entities.BillingRecord{},
		&entities.Tool{},
		&entities.UserToolInstance{},
		&entities.ToolUsageLog{},
	}

	// 执行自动迁移
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	return nil
}

// CreateIndexes 创建额外的索引
func CreateIndexes(db *gorm.DB) error {
	// 用户表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)").Error; err != nil {
		return fmt.Errorf("failed to create users username index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return fmt.Errorf("failed to create users email index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)").Error; err != nil {
		return fmt.Errorf("failed to create users status index: %w", err)
	}

	// API密钥表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create api_keys user_id index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix)").Error; err != nil {
		return fmt.Errorf("failed to create api_keys key_prefix index: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_api_keys_status ON api_keys(status)").Error; err != nil {
		return fmt.Errorf("failed to create api_keys status index: %w", err)
	}

	return nil
}

// InitializeDatabase 初始化数据库（迁移+索引）
func InitializeDatabase(db *gorm.DB) error {
	// 执行自动迁移
	if err := AutoMigrate(db); err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	// 创建额外索引
	if err := CreateIndexes(db); err != nil {
		return fmt.Errorf("create indexes failed: %w", err)
	}

	return nil
}

// HealthCheck 数据库健康检查
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetDBStats 获取数据库连接池统计信息
func GetDBStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
