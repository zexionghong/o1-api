package routes

import (
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/gateway"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/presentation/handlers"
	"ai-api-gateway/internal/presentation/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ai-api-gateway/docs" // 导入swagger文档
)

// Router 路由器
type Router struct {
	engine         *gin.Engine
	config         *config.Config
	logger         logger.Logger
	serviceFactory *services.ServiceFactory
	gatewayService gateway.GatewayService
}

// NewRouter 创建路由器
func NewRouter(
	config *config.Config,
	logger logger.Logger,
	serviceFactory *services.ServiceFactory,
	gatewayService gateway.GatewayService,
) *Router {
	// 设置Gin模式
	if config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	return &Router{
		engine:         engine,
		config:         config,
		logger:         logger,
		serviceFactory: serviceFactory,
		gatewayService: gatewayService,
	}
}

// SetupRoutes 设置路由
func (r *Router) SetupRoutes() {
	// 创建中间件
	authMiddleware := middleware.NewAuthMiddleware(
		r.serviceFactory.APIKeyService(),
		r.serviceFactory.JWTService(),
		r.serviceFactory.UserService(),
		r.serviceFactory.UserRepository(),
		r.logger,
	)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(&r.config.RateLimit, r.logger)
	quotaMiddleware := middleware.NewQuotaMiddleware(r.serviceFactory.QuotaService(), r.logger)

	// 全局中间件
	r.engine.Use(middleware.RecoveryMiddleware(r.logger))
	r.engine.Use(middleware.LoggingMiddleware(r.logger))
	r.engine.Use(middleware.CORSMiddleware())
	r.engine.Use(middleware.SecurityMiddleware())
	r.engine.Use(middleware.RequestIDMiddleware())
	r.engine.Use(middleware.TimeoutMiddleware(30 * time.Second))

	// 创建处理器
	aiHandler := handlers.NewAIHandler(r.gatewayService, r.logger)
	userHandler := handlers.NewUserHandler(r.serviceFactory.UserService(), r.logger)
	apiKeyHandler := handlers.NewAPIKeyHandler(
		r.serviceFactory.APIKeyService(),
		r.serviceFactory.UsageLogRepository(),
		r.serviceFactory.BillingRecordRepository(),
		r.serviceFactory.ModelRepository(),
		r.logger,
	)
	healthHandler := handlers.NewHealthHandler(r.gatewayService, r.logger)
	authHandler := handlers.NewAuthHandler(r.serviceFactory.AuthService(), r.logger)
	toolHandler := handlers.NewToolHandler(r.serviceFactory.ToolService(), r.logger)
	quotaHandler := handlers.NewQuotaHandler(r.serviceFactory.QuotaService(), r.logger)

	// 健康检查路由（无需认证）
	health := r.engine.Group("/health")
	{
		health.GET("/", healthHandler.HealthCheck)
		health.GET("/ready", healthHandler.ReadinessCheck)
		health.GET("/live", healthHandler.LivenessCheck)
		health.GET("/stats", healthHandler.GetStats)
		health.GET("/version", healthHandler.GetVersion)
	}

	// 认证路由（无需认证）
	auth := r.engine.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)

		// 需要认证的认证路由
		authProtected := auth.Group("/")
		authProtected.Use(authMiddleware.Authenticate())
		{
			authProtected.GET("/profile", authHandler.GetProfile)
			authProtected.POST("/change-password", authHandler.ChangePassword)
			authProtected.POST("/recharge", authHandler.Recharge)
		}
	}

	// 监控指标路由（无需认证）
	r.engine.GET("/metrics", healthHandler.GetMetrics)

	// Swagger文档路由（无需认证）
	swaggerGroup := r.engine.Group("/swagger")
	swaggerGroup.Use(func(c *gin.Context) {
		// 设置 CSP 头部以允许 Swagger UI 正常工作
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'")
		c.Next()
	})
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// OpenAI兼容的API路由
	v1 := r.engine.Group("/v1")
	v1.Use(rateLimitMiddleware.IPRateLimit(100)) // IP级别限流
	{
		// AI请求路由（需要认证和配额检查）
		aiRoutes := v1.Group("/")
		aiRoutes.Use(authMiddleware.Authenticate())
		aiRoutes.Use(rateLimitMiddleware.RateLimit())
		aiRoutes.Use(quotaMiddleware.CheckQuota())
		aiRoutes.Use(quotaMiddleware.ConsumeQuota()) // 在请求完成后消费配额
		{
			aiRoutes.POST("/chat/completions", aiHandler.ChatCompletions)
			aiRoutes.POST("/completions", aiHandler.Completions)
			aiRoutes.GET("/models", aiHandler.Models)
		}

		// 使用情况路由（需要认证）
		usageRoutes := v1.Group("/")
		usageRoutes.Use(authMiddleware.Authenticate())
		{
			usageRoutes.GET("/usage", aiHandler.Usage)
		}
	}

	// 管理API路由
	admin := r.engine.Group("/admin")
	admin.Use(rateLimitMiddleware.CustomRateLimit(200)) // 管理API更高的限流
	admin.Use(authMiddleware.Authenticate())            // 需要JWT认证
	{
		// 用户管理路由
		users := admin.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/:id/balance", userHandler.UpdateBalance)
			// 用户的API密钥路由
			users.GET("/:id/api-keys", apiKeyHandler.GetUserAPIKeys)
		}

		// API密钥管理路由
		apiKeys := admin.Group("/api-keys")
		{
			apiKeys.POST("/", apiKeyHandler.CreateAPIKey)
			apiKeys.GET("/", apiKeyHandler.ListAPIKeys)
			apiKeys.GET("/:id/usage-logs", apiKeyHandler.GetAPIKeyUsageLogs)
			apiKeys.GET("/:id/billing-records", apiKeyHandler.GetAPIKeyBillingRecords)
			apiKeys.POST("/:id/revoke", apiKeyHandler.RevokeAPIKey)
			apiKeys.GET("/:id", apiKeyHandler.GetAPIKey)
			apiKeys.PUT("/:id", apiKeyHandler.UpdateAPIKey)
			apiKeys.DELETE("/:id", apiKeyHandler.DeleteAPIKey)

			// API密钥配额管理路由
			apiKeys.GET("/:id/quotas", quotaHandler.GetAPIKeyQuotas)
			apiKeys.POST("/:id/quotas", quotaHandler.CreateAPIKeyQuota)
			apiKeys.GET("/:id/quota-status", quotaHandler.GetQuotaStatus)
		}

		// 配额管理路由
		quotas := admin.Group("/quotas")
		{
			quotas.PUT("/:quota_id", quotaHandler.UpdateQuota)
			quotas.DELETE("/:quota_id", quotaHandler.DeleteQuota)
		}

		// 工具管理路由
		tools := admin.Group("/tools")
		{
			// 用户工具实例路由
			tools.GET("/", toolHandler.GetUserToolInstances)
			tools.POST("/", toolHandler.CreateUserToolInstance)
			tools.GET("/:id", toolHandler.GetUserToolInstance)
			tools.PUT("/:id", toolHandler.UpdateUserToolInstance)
			tools.DELETE("/:id", toolHandler.DeleteUserToolInstance)
			tools.POST("/:id/usage", toolHandler.IncrementUsage)

			// 工具相关资源路由
			tools.GET("/api-keys", toolHandler.GetUserAPIKeys)
		}

	}

	// 公开工具路由（无需认证）
	publicTools := r.engine.Group("/tools")
	{
		publicTools.GET("/types", toolHandler.GetTools)
		publicTools.GET("/models", toolHandler.GetModels)
		publicTools.GET("/public", toolHandler.GetPublicTools)
		publicTools.GET("/share/:token", toolHandler.GetSharedToolInstance)
	}

	// 404处理
	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Endpoint not found",
			},
			"timestamp": time.Now(),
		})
	})

	// 405处理
	r.engine.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "Method not allowed",
			},
			"timestamp": time.Now(),
		})
	})
}

// GetEngine 获取Gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Start 启动服务器
func (r *Router) Start() error {
	address := r.config.Server.GetAddress()
	r.logger.WithField("address", address).Info("Starting HTTP server")

	return r.engine.Run(address)
}
