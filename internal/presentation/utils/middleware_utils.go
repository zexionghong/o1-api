package utils

import (
	"net/http"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// MiddlewareHelper 中间件助手
type MiddlewareHelper struct {
	logger        logger.Logger
	contextHelper *ContextHelper
	errorHandler  *ErrorHandler
}

// NewMiddlewareHelper 创建中间件助手
func NewMiddlewareHelper(logger logger.Logger) *MiddlewareHelper {
	return &MiddlewareHelper{
		logger:        logger,
		contextHelper: NewContextHelper(),
		errorHandler:  NewErrorHandler(logger),
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid    bool
	User     *entities.User
	APIKey   *entities.APIKey
	UserID   int64
	APIKeyID int64
	Error    error
}

// RequireAuthentication 要求认证
func (h *MiddlewareHelper) RequireAuthentication(c *gin.Context) *ValidationResult {
	result := &ValidationResult{}

	// 获取用户ID
	userID, exists := h.contextHelper.GetUserIDFromContext(c)
	if !exists || userID <= 0 {
		result.Valid = false
		h.errorHandler.RespondWithAuthError(c, "Authentication required")
		return result
	}

	result.Valid = true
	result.UserID = userID

	// 尝试获取用户信息
	if user, exists := h.contextHelper.GetUserFromContext(c); exists {
		result.User = user
	}

	// 尝试获取API密钥信息
	if apiKey, exists := h.contextHelper.GetAPIKeyFromContext(c); exists {
		result.APIKey = apiKey
		if apiKeyID, exists := h.contextHelper.GetAPIKeyIDFromContext(c); exists {
			result.APIKeyID = apiKeyID
		}
	}

	return result
}

// RequireUser 要求用户信息
func (h *MiddlewareHelper) RequireUser(c *gin.Context) *ValidationResult {
	result := &ValidationResult{}

	user, exists := h.contextHelper.GetUserFromContext(c)
	if !exists || user == nil {
		result.Valid = false
		h.errorHandler.RespondWithAuthError(c, "User information required")
		return result
	}

	result.Valid = true
	result.User = user
	result.UserID = user.ID

	return result
}

// CheckBalance 检查用户余额
func (h *MiddlewareHelper) CheckBalance(c *gin.Context, estimatedCost float64) *ValidationResult {
	result := h.RequireUser(c)
	if !result.Valid {
		return result
	}

	user := result.User

	// 检查余额是否足够
	if !user.CanMakeRequest() {
		result.Valid = false
		h.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"balance": user.Balance,
		}).Warn("Insufficient balance")

		h.errorHandler.RespondWithInsufficientBalance(c, user.Balance, estimatedCost)
		return result
	}

	// 检查余额是否足够支付预估成本
	if estimatedCost > 0 && user.Balance < estimatedCost {
		result.Valid = false
		h.logger.WithFields(map[string]interface{}{
			"user_id":        user.ID,
			"balance":        user.Balance,
			"estimated_cost": estimatedCost,
		}).Warn("Insufficient balance for estimated cost")

		h.errorHandler.RespondWithInsufficientBalance(c, user.Balance, estimatedCost)
		return result
	}

	return result
}

// LogRequest 记录请求日志
func (h *MiddlewareHelper) LogRequest(c *gin.Context, level string, message string, extraFields map[string]interface{}) {
	fields := map[string]interface{}{
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}

	// 添加请求ID
	if requestID := h.contextHelper.GetRequestIDFromContext(c); requestID != "" {
		fields["request_id"] = requestID
	}

	// 添加用户信息
	if userID, exists := h.contextHelper.GetUserIDFromContext(c); exists {
		fields["user_id"] = userID
	}

	// 添加API密钥信息
	if apiKeyID, exists := h.contextHelper.GetAPIKeyIDFromContext(c); exists {
		fields["api_key_id"] = apiKeyID
	}

	// 添加额外字段
	for k, v := range extraFields {
		fields[k] = v
	}

	logEntry := h.logger.WithFields(fields)

	switch level {
	case "debug":
		logEntry.Debug(message)
	case "info":
		logEntry.Info(message)
	case "warn":
		logEntry.Warn(message)
	case "error":
		logEntry.Error(message)
	default:
		logEntry.Info(message)
	}
}

// CreateAuthMiddleware 创建认证中间件
func (h *MiddlewareHelper) CreateAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		result := h.RequireAuthentication(c)
		if !result.Valid {
			c.Abort()
			return
		}
		c.Next()
	}
}

// CreateBalanceCheckMiddleware 创建余额检查中间件
func (h *MiddlewareHelper) CreateBalanceCheckMiddleware(estimatedCost float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := h.CheckBalance(c, estimatedCost)
		if !result.Valid {
			c.Abort()
			return
		}
		c.Next()
	}
}

// CreateQuotaCheckMiddleware 创建配额检查中间件
func (h *MiddlewareHelper) CreateQuotaCheckMiddleware(quotaType entities.QuotaType, amount int) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := h.RequireAuthentication(c)
		if !result.Valid {
			c.Abort()
			return
		}

		// 这里可以添加配额检查逻辑
		// 为了简化，暂时跳过具体实现

		c.Next()
	}
}

// HandlePanic 处理panic
func (h *MiddlewareHelper) HandlePanic(c *gin.Context, recovered interface{}) {
	h.logger.WithFields(map[string]interface{}{
		"panic":      recovered,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}).Error("Panic recovered")

	c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
		ErrCodeInternalError,
		"Internal server error",
		nil,
	))
}

// ValidatePagination 验证分页参数
func (h *MiddlewareHelper) ValidatePagination(c *gin.Context) (*dto.PaginationRequest, bool) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid pagination parameters")
		h.errorHandler.RespondWithValidationError(c, "Invalid pagination parameters", map[string]interface{}{
			"details": err.Error(),
		})
		return nil, false
	}

	pagination.SetDefaults()
	return &pagination, true
}

// ValidateContentType 验证内容类型
func (h *MiddlewareHelper) ValidateContentType(c *gin.Context, allowedTypes []string) bool {
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			h.errorHandler.RespondWithValidationError(c, "Content-Type header is required", nil)
			return false
		}

		if len(allowedTypes) > 0 {
			allowed := false
			for _, allowedType := range allowedTypes {
				if contentType == allowedType {
					allowed = true
					break
				}
			}
			if !allowed {
				h.errorHandler.RespondWithValidationError(c, "Unsupported Content-Type", map[string]interface{}{
					"content_type":  contentType,
					"allowed_types": allowedTypes,
				})
				return false
			}
		}
	}

	return true
}

// ValidateRequestSize 验证请求大小
func (h *MiddlewareHelper) ValidateRequestSize(c *gin.Context, maxSize int64) bool {
	if c.Request.ContentLength > maxSize {
		h.errorHandler.RespondWithError(c, ErrCodeRequestTooLarge, "Request entity too large", http.StatusRequestEntityTooLarge, map[string]interface{}{
			"content_length": c.Request.ContentLength,
			"max_size":       maxSize,
		})
		return false
	}

	return true
}

// 全局中间件助手实例
var DefaultMiddlewareHelper *MiddlewareHelper

// InitDefaultMiddlewareHelper 初始化默认中间件助手
func InitDefaultMiddlewareHelper(logger logger.Logger) {
	DefaultMiddlewareHelper = NewMiddlewareHelper(logger)
}

// 便捷函数
func RequireAuthentication(c *gin.Context) *ValidationResult {
	if DefaultMiddlewareHelper != nil {
		return DefaultMiddlewareHelper.RequireAuthentication(c)
	}
	return &ValidationResult{Valid: false}
}

func RequireUserValidation(c *gin.Context) *ValidationResult {
	if DefaultMiddlewareHelper != nil {
		return DefaultMiddlewareHelper.RequireUser(c)
	}
	return &ValidationResult{Valid: false}
}

func CheckBalance(c *gin.Context, estimatedCost float64) *ValidationResult {
	if DefaultMiddlewareHelper != nil {
		return DefaultMiddlewareHelper.CheckBalance(c, estimatedCost)
	}
	return &ValidationResult{Valid: false}
}
