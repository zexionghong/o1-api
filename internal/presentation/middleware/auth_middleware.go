package middleware

import (
	"net/http"
	"strings"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	apiKeyService services.APIKeyService
	jwtService    services.JWTService
	userService   services.UserService
	userRepo      repositories.UserRepository
	logger        logger.Logger
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(apiKeyService services.APIKeyService, jwtService services.JWTService, userService services.UserService, userRepo repositories.UserRepository, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		apiKeyService: apiKeyService,
		jwtService:    jwtService,
		userService:   userService,
		userRepo:      userRepo,
		logger:        logger,
	}
}

// Authenticate 认证中间件函数
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试JWT认证
		if m.tryJWTAuth(c) {
			c.Next()
			return
		}

		// 尝试API Key认证
		if m.tryAPIKeyAuth(c) {
			c.Next()
			return
		}

		// 两种认证都失败
		m.logger.Warn("Authentication failed: no valid credentials provided")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"Valid authentication credentials are required",
			nil,
		))
		c.Abort()
	}
}

// extractAPIKey 从请求中提取API密钥
func (m *AuthMiddleware) extractAPIKey(c *gin.Context) string {
	// 1. 从Authorization头中提取 (Bearer token)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// 2. 从X-API-Key头中提取
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// 3. 从查询参数中提取
	apiKey = c.Query("api_key")
	if apiKey != "" {
		return apiKey
	}

	return ""
}

// maskAPIKey 掩码API密钥用于日志记录
func (m *AuthMiddleware) maskAPIKey(apiKey string) string {
	if len(apiKey) < 8 {
		return "***"
	}
	return apiKey[:8] + "..."
}

// RequirePermission 权限检查中间件
func (m *AuthMiddleware) RequirePermission(providerSlug, modelSlug string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取API密钥
		apiKeyInterface, exists := c.Get("api_key")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
				"AUTHENTICATION_REQUIRED",
				"Authentication required",
				nil,
			))
			c.Abort()
			return
		}

		apiKey, ok := apiKeyInterface.(*entities.APIKey)
		if !ok {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
				"INTERNAL_ERROR",
				"Internal authentication error",
				nil,
			))
			c.Abort()
			return
		}

		// 检查提供商权限
		if providerSlug != "" && !apiKey.HasPermissionForProvider(providerSlug) {
			m.logger.WithFields(map[string]interface{}{
				"api_key_id":    apiKey.ID,
				"provider_slug": providerSlug,
			}).Warn("Provider permission denied")

			c.JSON(http.StatusForbidden, dto.ErrorResponse(
				"PROVIDER_PERMISSION_DENIED",
				"Access to this provider is not allowed",
				map[string]interface{}{
					"provider": providerSlug,
				},
			))
			c.Abort()
			return
		}

		// 检查模型权限
		if modelSlug != "" && !apiKey.HasPermissionForModel(modelSlug) {
			m.logger.WithFields(map[string]interface{}{
				"api_key_id": apiKey.ID,
				"model_slug": modelSlug,
			}).Warn("Model permission denied")

			c.JSON(http.StatusForbidden, dto.ErrorResponse(
				"MODEL_PERMISSION_DENIED",
				"Access to this model is not allowed",
				map[string]interface{}{
					"model": modelSlug,
				},
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// tryJWTAuth 尝试JWT认证
func (m *AuthMiddleware) tryJWTAuth(c *gin.Context) bool {
	// 从Authorization头获取JWT token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return false
	}

	// 提取token
	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return false
	}

	// 验证JWT token
	claims, err := m.jwtService.ValidateAccessToken(c.Request.Context(), token)
	if err != nil {
		m.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Debug("JWT token validation failed")
		return false
	}

	// 获取用户信息并检查余额
	user, err := m.userRepo.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		m.logger.WithFields(map[string]interface{}{
			"user_id": claims.UserID,
			"error":   err.Error(),
		}).Debug("Failed to get user for JWT authentication")
		return false
	}

	// 检查用户是否可以发起请求（包括余额检查）
	if !user.CanMakeRequest() {
		m.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"balance": user.Balance,
			"status":  user.Status,
		}).Debug("User cannot make request - insufficient balance or inactive status")
		return false
	}

	// 将认证信息存储到上下文中
	c.Set("auth_type", "jwt")
	c.Set("user", user)
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("jwt_claims", claims)

	m.logger.WithFields(map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"balance":  user.Balance,
	}).Debug("JWT authentication successful")

	return true
}

// tryAPIKeyAuth 尝试API Key认证
func (m *AuthMiddleware) tryAPIKeyAuth(c *gin.Context) bool {
	// 从请求头中获取API密钥
	apiKey := m.extractAPIKey(c)
	if apiKey == "" {
		return false
	}

	// 验证API密钥
	apiKeyEntity, user, err := m.apiKeyService.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil {
		m.logger.WithFields(map[string]interface{}{
			"api_key_prefix": m.maskAPIKey(apiKey),
			"error":          err.Error(),
		}).Debug("API key validation failed")
		return false
	}

	// 检查用户是否可以发起请求（包括余额检查）
	if !user.CanMakeRequest() {
		m.logger.WithFields(map[string]interface{}{
			"user_id":        user.ID,
			"balance":        user.Balance,
			"status":         user.Status,
			"api_key_prefix": apiKeyEntity.KeyPrefix,
		}).Debug("User cannot make request - insufficient balance or inactive status")
		return false
	}

	// 将认证信息存储到上下文中
	c.Set("auth_type", "api_key")
	c.Set("api_key", apiKeyEntity)
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("api_key_id", apiKeyEntity.ID)

	m.logger.WithFields(map[string]interface{}{
		"user_id":        user.ID,
		"api_key_id":     apiKeyEntity.ID,
		"api_key_prefix": apiKeyEntity.KeyPrefix,
		"balance":        user.Balance,
	}).Debug("API key authentication successful")

	return true
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(c *gin.Context) (*entities.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*entities.User)
	return user, ok
}

// GetAPIKeyFromContext 从上下文中获取API密钥信息
func GetAPIKeyFromContext(c *gin.Context) (*entities.APIKey, bool) {
	apiKeyInterface, exists := c.Get("api_key")
	if !exists {
		return nil, false
	}

	apiKey, ok := apiKeyInterface.(*entities.APIKey)
	return apiKey, ok
}

// GetUserIDFromContext 从上下文中获取用户ID
func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := userIDInterface.(int64)
	return userID, ok
}

// GetAPIKeyIDFromContext 从上下文中获取API密钥ID
func GetAPIKeyIDFromContext(c *gin.Context) (int64, bool) {
	apiKeyIDInterface, exists := c.Get("api_key_id")
	if !exists {
		return 0, false
	}

	apiKeyID, ok := apiKeyIDInterface.(int64)
	return apiKeyID, ok
}
