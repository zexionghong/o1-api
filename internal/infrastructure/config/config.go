package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"ai-api-gateway/internal/infrastructure/database"
)

// Config 应用配置
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    database.Config   `mapstructure:"database"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limiting"`
	Providers   ProvidersConfig   `mapstructure:"providers"`
	LoadBalance LoadBalanceConfig `mapstructure:"load_balancer"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
	Billing     BillingConfig     `mapstructure:"billing"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	DefaultRequestsPerMinute int `mapstructure:"default_requests_per_minute"`
	DefaultRequestsPerHour   int `mapstructure:"default_requests_per_hour"`
	DefaultRequestsPerDay    int `mapstructure:"default_requests_per_day"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Name                 string        `mapstructure:"name"`
	BaseURL              string        `mapstructure:"base_url"`
	Enabled              bool          `mapstructure:"enabled"`
	Priority             int           `mapstructure:"priority"`
	Timeout              time.Duration `mapstructure:"timeout"`
	RetryAttempts        int           `mapstructure:"retry_attempts"`
	HealthCheckInterval  time.Duration `mapstructure:"health_check_interval"`
}

// ProvidersConfig 提供商配置集合
type ProvidersConfig struct {
	OpenAI    ProviderConfig `mapstructure:"openai"`
	Anthropic ProviderConfig `mapstructure:"anthropic"`
}

// LoadBalanceConfig 负载均衡配置
type LoadBalanceConfig struct {
	Strategy            string `mapstructure:"strategy"`
	HealthCheckEnabled  bool   `mapstructure:"health_check_enabled"`
	FailoverEnabled     bool   `mapstructure:"failover_enabled"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	MetricsEnabled    bool   `mapstructure:"metrics_enabled"`
	MetricsPort       int    `mapstructure:"metrics_port"`
	HealthCheckPath   string `mapstructure:"health_check_path"`
}

// BillingConfig 计费配置
type BillingConfig struct {
	Currency  string `mapstructure:"currency"`
	Precision int    `mapstructure:"precision"`
	BatchSize int    `mapstructure:"batch_size"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}
	
	// 设置环境变量
	viper.AutomaticEnv()
	
	// 设置默认值
	setDefaults()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return &config, nil
}

// setDefaults 设置默认值
func setDefaults() {
	// 服务器默认值
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	
	// 数据库默认值
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dsn", "./data/gateway.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "300s")
	
	// 日志默认值
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	
	// 速率限制默认值
	viper.SetDefault("rate_limiting.default_requests_per_minute", 60)
	viper.SetDefault("rate_limiting.default_requests_per_hour", 1000)
	viper.SetDefault("rate_limiting.default_requests_per_day", 10000)
	
	// 负载均衡默认值
	viper.SetDefault("load_balancer.strategy", "round_robin")
	viper.SetDefault("load_balancer.health_check_enabled", true)
	viper.SetDefault("load_balancer.failover_enabled", true)
	
	// 监控默认值
	viper.SetDefault("monitoring.metrics_enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
	viper.SetDefault("monitoring.health_check_path", "/health")
	
	// 计费默认值
	viper.SetDefault("billing.currency", "USD")
	viper.SetDefault("billing.precision", 6)
	viper.SetDefault("billing.batch_size", 100)
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	
	// 验证数据库配置
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	
	if config.Database.DSN == "" {
		return fmt.Errorf("database dsn is required")
	}
	
	// 验证日志配置
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[config.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", config.Logging.Level)
	}
	
	// 验证负载均衡策略
	validStrategies := map[string]bool{
		"round_robin": true, "weighted": true, "least_connections": true, "random": true,
	}
	if !validStrategies[config.LoadBalance.Strategy] {
		return fmt.Errorf("invalid load balance strategy: %s", config.LoadBalance.Strategy)
	}
	
	return nil
}

// GetAddress 获取服务器地址
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetMetricsAddress 获取监控地址
func (c *MonitoringConfig) GetMetricsAddress() string {
	return fmt.Sprintf(":%d", c.MetricsPort)
}
