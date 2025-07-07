package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/database"
	"ai-api-gateway/internal/infrastructure/gateway"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/infrastructure/redis"
	"ai-api-gateway/internal/infrastructure/repositories"
	"ai-api-gateway/internal/presentation/routes"

	"github.com/spf13/viper"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志记录器
	logger.InitGlobalLogger(&cfg.Logging)
	log := logger.GetLogger()

	log.Info("Starting AI API Gateway")
	log.WithField("config", configPath).Info("Configuration loaded")

	// 初始化GORM数据库连接
	gormConfig := database.GormConfig{
		Host:     "47.76.73.118",
		Port:     5432,
		User:     "proxy",
		Password: "pPhnbrlIKfYA",
		DBName:   "ai",
		SSLMode:  "disable",
		TimeZone: "UTC",
	}

	gormDB, err := database.NewGormDB(gormConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to connect to PostgreSQL with GORM")
	}

	// 跳过自动迁移，因为我们使用手动创建的PostgreSQL表结构
	// 只进行健康检查
	if err := database.HealthCheck(gormDB); err != nil {
		log.WithField("error", err.Error()).Fatal("Database health check failed")
	}

	log.Info("PostgreSQL connection established with GORM")

	// 创建仓储工厂（全部使用GORM）
	repoFactory := repositories.NewRepositoryFactory(gormDB)

	// 创建Redis工厂（可选）
	var redisFactory *redis.RedisFactory
	if viper.GetBool("cache.enabled") || viper.GetBool("distributed_lock.enabled") {
		var err error
		redisFactory, err = redis.NewRedisFactory(log)
		if err != nil {
			log.WithFields(map[string]interface{}{
				"error": err.Error(),
			}).Warn("Failed to initialize Redis, continuing without cache and distributed locks")
			redisFactory = nil
		}
	}

	// 创建服务工厂
	serviceFactory := services.NewServiceFactory(repoFactory, redisFactory, cfg, log)

	// 创建HTTP客户端
	httpClient := clients.NewHTTPClient(30 * time.Second)

	// 创建AI提供商客户端
	aiClient := clients.NewAIProviderClient(httpClient)

	// 创建负载均衡器
	loadBalancer := gateway.NewLoadBalancer(
		gateway.LoadBalanceStrategy(cfg.LoadBalance.Strategy),
		log,
	)

	// 创建请求路由器
	requestRouter := gateway.NewRequestRouter(
		serviceFactory.ProviderService(),
		serviceFactory.ModelService(),
		repoFactory.ProviderModelSupportRepository(),
		loadBalancer,
		aiClient,
		log,
	)

	// 创建网关服务
	gatewayService := gateway.NewGatewayService(
		requestRouter,
		serviceFactory.UserService(),
		serviceFactory.APIKeyService(),
		serviceFactory.QuotaService(),
		serviceFactory.BillingService(),
		serviceFactory.UsageLogService(),
		repoFactory.BillingRecordRepository(),
		log,
	)

	// 创建路由器
	router := routes.NewRouter(cfg, log, serviceFactory, gatewayService)
	router.SetupRoutes()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      router.GetEngine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		log.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithField("error", err.Error()).Fatal("Failed to start HTTP server")
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithField("error", err.Error()).Fatal("Server forced to shutdown")
	} else {
		log.Info("Server shutdown complete")
	}
}
