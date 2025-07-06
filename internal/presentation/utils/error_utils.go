package utils

import (
	"net/http"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 错误处理器
type ErrorHandler struct {
	logger logger.Logger
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler(logger logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// ErrorCode 错误代码常量
const (
	// 认证相关错误
	ErrCodeAuthRequired        = "AUTHENTICATION_REQUIRED"
	ErrCodeInvalidCredentials  = "INVALID_CREDENTIALS"
	ErrCodeInvalidAPIKey       = "INVALID_API_KEY"
	ErrCodeAPIKeyExpired       = "API_KEY_EXPIRED"
	ErrCodeInsufficientBalance = "INSUFFICIENT_BALANCE"

	// 授权相关错误
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// 请求相关错误
	ErrCodeInvalidRequest      = "INVALID_REQUEST"
	ErrCodeInvalidPagination   = "INVALID_PAGINATION"
	ErrCodeInvalidParameters   = "INVALID_PARAMETERS"
	ErrCodeRequestTooLarge     = "REQUEST_TOO_LARGE"

	// 资源相关错误
	ErrCodeResourceNotFound    = "RESOURCE_NOT_FOUND"
	ErrCodeResourceExists      = "RESOURCE_EXISTS"
	ErrCodeResourceConflict    = "RESOURCE_CONFLICT"

	// 配额相关错误
	ErrCodeQuotaExceeded       = "QUOTA_EXCEEDED"
	ErrCodeQuotaCheckError     = "QUOTA_CHECK_ERROR"
	ErrCodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"

	// 服务相关错误
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout             = "TIMEOUT"
	ErrCodeDatabaseError       = "DATABASE_ERROR"
)

// CommonError 通用错误结构
type CommonError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
}

// Error 实现error接口
func (e *CommonError) Error() string {
	return e.Message
}

// NewCommonError 创建通用错误
func NewCommonError(code, message string, statusCode int, details map[string]interface{}) *CommonError {
	return &CommonError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// 预定义的常用错误
var (
	ErrAuthRequired = NewCommonError(
		ErrCodeAuthRequired,
		"Authentication required",
		http.StatusUnauthorized,
		nil,
	)

	ErrInvalidCredentials = NewCommonError(
		ErrCodeInvalidCredentials,
		"Invalid credentials",
		http.StatusUnauthorized,
		nil,
	)

	ErrInvalidAPIKey = NewCommonError(
		ErrCodeInvalidAPIKey,
		"Invalid API key",
		http.StatusUnauthorized,
		nil,
	)

	ErrForbidden = NewCommonError(
		ErrCodeForbidden,
		"Access forbidden",
		http.StatusForbidden,
		nil,
	)

	ErrInvalidRequest = NewCommonError(
		ErrCodeInvalidRequest,
		"Invalid request",
		http.StatusBadRequest,
		nil,
	)

	ErrResourceNotFound = NewCommonError(
		ErrCodeResourceNotFound,
		"Resource not found",
		http.StatusNotFound,
		nil,
	)

	ErrInternalError = NewCommonError(
		ErrCodeInternalError,
		"Internal server error",
		http.StatusInternalServerError,
		nil,
	)
)

// HandleError 处理错误并返回JSON响应
func (h *ErrorHandler) HandleError(c *gin.Context, err error) {
	if commonErr, ok := err.(*CommonError); ok {
		h.handleCommonError(c, commonErr)
		return
	}

	// 处理其他类型的错误
	h.handleGenericError(c, err)
}

// handleCommonError 处理通用错误
func (h *ErrorHandler) handleCommonError(c *gin.Context, err *CommonError) {
	// 记录错误日志
	h.logError(c, err.Code, err.Message, err.Details)

	// 返回错误响应
	c.JSON(err.StatusCode, dto.ErrorResponse(err.Code, err.Message, err.Details))
}

// handleGenericError 处理通用错误
func (h *ErrorHandler) handleGenericError(c *gin.Context, err error) {
	// 记录错误日志
	h.logError(c, ErrCodeInternalError, err.Error(), nil)

	// 返回通用错误响应
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
		ErrCodeInternalError,
		"An internal error occurred",
		nil,
	))
}

// logError 记录错误日志
func (h *ErrorHandler) logError(c *gin.Context, code, message string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"error_code":    code,
		"error_message": message,
		"path":          c.Request.URL.Path,
		"method":        c.Request.Method,
		"client_ip":     c.ClientIP(),
		"user_agent":    c.Request.UserAgent(),
	}

	// 添加请求ID
	if requestID := GetRequestIDFromContext(c); requestID != "" {
		fields["request_id"] = requestID
	}

	// 添加用户信息
	if userID, exists := GetUserIDFromContext(c); exists {
		fields["user_id"] = userID
	}

	// 添加错误详情
	if details != nil {
		fields["error_details"] = details
	}

	h.logger.WithFields(fields).Error("Request error")
}

// RespondWithError 快速错误响应
func (h *ErrorHandler) RespondWithError(c *gin.Context, code, message string, statusCode int, details map[string]interface{}) {
	err := NewCommonError(code, message, statusCode, details)
	h.HandleError(c, err)
}

// RespondWithAuthError 认证错误响应
func (h *ErrorHandler) RespondWithAuthError(c *gin.Context, message string) {
	h.RespondWithError(c, ErrCodeAuthRequired, message, http.StatusUnauthorized, nil)
}

// RespondWithValidationError 验证错误响应
func (h *ErrorHandler) RespondWithValidationError(c *gin.Context, message string, details map[string]interface{}) {
	h.RespondWithError(c, ErrCodeInvalidRequest, message, http.StatusBadRequest, details)
}

// RespondWithQuotaError 配额错误响应
func (h *ErrorHandler) RespondWithQuotaError(c *gin.Context, quotaType string) {
	h.RespondWithError(c, ErrCodeQuotaExceeded, "Quota exceeded", http.StatusTooManyRequests, map[string]interface{}{
		"quota_type": quotaType,
	})
}

// RespondWithInsufficientBalance 余额不足错误响应
func (h *ErrorHandler) RespondWithInsufficientBalance(c *gin.Context, currentBalance, estimatedCost float64) {
	h.RespondWithError(c, ErrCodeInsufficientBalance, "Insufficient account balance", http.StatusPaymentRequired, map[string]interface{}{
		"current_balance": currentBalance,
		"estimated_cost":  estimatedCost,
		"shortfall":       estimatedCost - currentBalance,
	})
}

// 全局错误处理器实例
var DefaultErrorHandler *ErrorHandler

// InitDefaultErrorHandler 初始化默认错误处理器
func InitDefaultErrorHandler(logger logger.Logger) {
	DefaultErrorHandler = NewErrorHandler(logger)
}

// 便捷函数
func HandleError(c *gin.Context, err error) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.HandleError(c, err)
	}
}

func RespondWithError(c *gin.Context, code, message string, statusCode int, details map[string]interface{}) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.RespondWithError(c, code, message, statusCode, details)
	}
}

func RespondWithAuthError(c *gin.Context, message string) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.RespondWithAuthError(c, message)
	}
}

func RespondWithValidationError(c *gin.Context, message string, details map[string]interface{}) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.RespondWithValidationError(c, message, details)
	}
}

func RespondWithQuotaError(c *gin.Context, quotaType string) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.RespondWithQuotaError(c, quotaType)
	}
}

func RespondWithInsufficientBalance(c *gin.Context, currentBalance, estimatedCost float64) {
	if DefaultErrorHandler != nil {
		DefaultErrorHandler.RespondWithInsufficientBalance(c, currentBalance, estimatedCost)
	}
}
