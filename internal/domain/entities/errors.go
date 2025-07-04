package entities

import (
	"errors"
	"fmt"
)

// 领域层错误定义

// 用户相关错误
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserInactive        = errors.New("user is inactive")
	ErrUserSuspended       = errors.New("user is suspended")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// API密钥相关错误
var (
	ErrAPIKeyNotFound   = errors.New("api key not found")
	ErrAPIKeyInactive   = errors.New("api key is inactive")
	ErrAPIKeyExpired    = errors.New("api key has expired")
	ErrAPIKeyRevoked    = errors.New("api key has been revoked")
	ErrAPIKeyInvalid    = errors.New("invalid api key")
	ErrPermissionDenied = errors.New("permission denied")
)

// 提供商相关错误
var (
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderUnavailable = errors.New("provider is unavailable")
	ErrProviderUnhealthy   = errors.New("provider is unhealthy")
	ErrNoAvailableProvider = errors.New("no available provider")
)

// 模型相关错误
var (
	ErrModelNotFound     = errors.New("model not found")
	ErrModelUnavailable  = errors.New("model is unavailable")
	ErrModelNotSupported = errors.New("model is not supported")
	ErrInvalidModelType  = errors.New("invalid model type")
)

// 提供商模型支持相关错误
var (
	ErrProviderModelSupportNotFound = errors.New("provider model support not found")
)

// 配额相关错误
var (
	ErrQuotaExceeded     = errors.New("quota exceeded")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrQuotaNotFound     = errors.New("quota not found")
	ErrInvalidQuotaType  = errors.New("invalid quota type")
)

// 请求相关错误
var (
	ErrInvalidRequest     = errors.New("invalid request")
	ErrRequestTooLarge    = errors.New("request too large")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrMissingParameter   = errors.New("missing required parameter")
)

// 业务逻辑错误
var (
	ErrOperationFailed  = errors.New("operation failed")
	ErrConcurrentUpdate = errors.New("concurrent update detected")
	ErrResourceLocked   = errors.New("resource is locked")
	ErrInvalidOperation = errors.New("invalid operation")
)

// DomainError 领域错误接口
type DomainError interface {
	error
	Code() string
	Message() string
	Details() map[string]interface{}
}

// domainError 领域错误实现
type domainError struct {
	code    string
	message string
	details map[string]interface{}
	cause   error
}

// Error 实现error接口
func (e *domainError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code 返回错误代码
func (e *domainError) Code() string {
	return e.code
}

// Message 返回错误消息
func (e *domainError) Message() string {
	return e.message
}

// Details 返回错误详情
func (e *domainError) Details() map[string]interface{} {
	return e.details
}

// Unwrap 返回原始错误
func (e *domainError) Unwrap() error {
	return e.cause
}

// NewDomainError 创建新的领域错误
func NewDomainError(code, message string, details map[string]interface{}) DomainError {
	return &domainError{
		code:    code,
		message: message,
		details: details,
	}
}

// WrapDomainError 包装错误为领域错误
func WrapDomainError(code, message string, cause error, details map[string]interface{}) DomainError {
	return &domainError{
		code:    code,
		message: message,
		details: details,
		cause:   cause,
	}
}

// 预定义的领域错误代码
const (
	ErrorCodeUserNotFound        = "USER_NOT_FOUND"
	ErrorCodeUserInactive        = "USER_INACTIVE"
	ErrorCodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	ErrorCodeAPIKeyInvalid       = "API_KEY_INVALID"
	ErrorCodeAPIKeyExpired       = "API_KEY_EXPIRED"
	ErrorCodePermissionDenied    = "PERMISSION_DENIED"
	ErrorCodeQuotaExceeded       = "QUOTA_EXCEEDED"
	ErrorCodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
	ErrorCodeProviderUnavailable = "PROVIDER_UNAVAILABLE"
	ErrorCodeModelNotSupported   = "MODEL_NOT_SUPPORTED"
	ErrorCodeInvalidRequest      = "INVALID_REQUEST"
	ErrorCodeOperationFailed     = "OPERATION_FAILED"
)
