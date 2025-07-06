package utils

import (
	"ai-api-gateway/internal/domain/entities"
	"github.com/gin-gonic/gin"
)

// ContextHelper 上下文助手
type ContextHelper struct{}

// NewContextHelper 创建上下文助手
func NewContextHelper() *ContextHelper {
	return &ContextHelper{}
}

// AuthInfo 认证信息
type AuthInfo struct {
	User      *entities.User   `json:"user,omitempty"`
	APIKey    *entities.APIKey `json:"api_key,omitempty"`
	UserID    int64            `json:"user_id,omitempty"`
	APIKeyID  int64            `json:"api_key_id,omitempty"`
	RequestID string           `json:"request_id,omitempty"`
}

// GetAuthInfo 获取完整的认证信息
func (h *ContextHelper) GetAuthInfo(c *gin.Context) *AuthInfo {
	info := &AuthInfo{}
	
	// 获取用户信息
	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*entities.User); ok {
			info.User = user
		}
	}
	
	// 获取API密钥信息
	if apiKeyInterface, exists := c.Get("api_key"); exists {
		if apiKey, ok := apiKeyInterface.(*entities.APIKey); ok {
			info.APIKey = apiKey
		}
	}
	
	// 获取用户ID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if userID, ok := userIDInterface.(int64); ok {
			info.UserID = userID
		}
	}
	
	// 获取API密钥ID
	if apiKeyIDInterface, exists := c.Get("api_key_id"); exists {
		if apiKeyID, ok := apiKeyIDInterface.(int64); ok {
			info.APIKeyID = apiKeyID
		}
	}
	
	// 获取请求ID
	if requestIDInterface, exists := c.Get("request_id"); exists {
		if requestID, ok := requestIDInterface.(string); ok {
			info.RequestID = requestID
		}
	}
	
	return info
}

// GetUserFromContext 从上下文中获取用户信息
func (h *ContextHelper) GetUserFromContext(c *gin.Context) (*entities.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInterface.(*entities.User)
	return user, ok
}

// GetAPIKeyFromContext 从上下文中获取API密钥信息
func (h *ContextHelper) GetAPIKeyFromContext(c *gin.Context) (*entities.APIKey, bool) {
	apiKeyInterface, exists := c.Get("api_key")
	if !exists {
		return nil, false
	}

	apiKey, ok := apiKeyInterface.(*entities.APIKey)
	return apiKey, ok
}

// GetUserIDFromContext 从上下文中获取用户ID
func (h *ContextHelper) GetUserIDFromContext(c *gin.Context) (int64, bool) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := userIDInterface.(int64)
	return userID, ok
}

// GetAPIKeyIDFromContext 从上下文中获取API密钥ID
func (h *ContextHelper) GetAPIKeyIDFromContext(c *gin.Context) (int64, bool) {
	apiKeyIDInterface, exists := c.Get("api_key_id")
	if !exists {
		return 0, false
	}

	apiKeyID, ok := apiKeyIDInterface.(int64)
	return apiKeyID, ok
}

// GetRequestIDFromContext 从上下文中获取请求ID
func (h *ContextHelper) GetRequestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RequireAuth 要求认证，返回用户ID，如果未认证则返回false
func (h *ContextHelper) RequireAuth(c *gin.Context) (int64, bool) {
	userID, exists := h.GetUserIDFromContext(c)
	return userID, exists && userID > 0
}

// RequireUser 要求用户信息，如果未认证则返回false
func (h *ContextHelper) RequireUser(c *gin.Context) (*entities.User, bool) {
	user, exists := h.GetUserFromContext(c)
	return user, exists && user != nil
}

// GetIdentifier 获取用户标识符（用于限流等）
func (h *ContextHelper) GetIdentifier(c *gin.Context) string {
	if apiKeyID, exists := h.GetAPIKeyIDFromContext(c); exists {
		return "api_key_" + string(rune(apiKeyID))
	}
	if userID, exists := h.GetUserIDFromContext(c); exists {
		return "user_" + string(rune(userID))
	}
	return "ip_" + c.ClientIP()
}

// SetAuthInfo 设置认证信息到上下文
func (h *ContextHelper) SetAuthInfo(c *gin.Context, user *entities.User, apiKey *entities.APIKey) {
	if user != nil {
		c.Set("user", user)
		c.Set("user_id", user.ID)
	}
	if apiKey != nil {
		c.Set("api_key", apiKey)
		c.Set("api_key_id", apiKey.ID)
	}
}

// 全局实例
var DefaultContextHelper = NewContextHelper()

// 便捷函数
func GetAuthInfo(c *gin.Context) *AuthInfo {
	return DefaultContextHelper.GetAuthInfo(c)
}

func GetUserFromContext(c *gin.Context) (*entities.User, bool) {
	return DefaultContextHelper.GetUserFromContext(c)
}

func GetAPIKeyFromContext(c *gin.Context) (*entities.APIKey, bool) {
	return DefaultContextHelper.GetAPIKeyFromContext(c)
}

func GetUserIDFromContext(c *gin.Context) (int64, bool) {
	return DefaultContextHelper.GetUserIDFromContext(c)
}

func GetAPIKeyIDFromContext(c *gin.Context) (int64, bool) {
	return DefaultContextHelper.GetAPIKeyIDFromContext(c)
}

func GetRequestIDFromContext(c *gin.Context) string {
	return DefaultContextHelper.GetRequestIDFromContext(c)
}

func RequireAuth(c *gin.Context) (int64, bool) {
	return DefaultContextHelper.RequireAuth(c)
}

func RequireUser(c *gin.Context) (*entities.User, bool) {
	return DefaultContextHelper.RequireUser(c)
}

func GetIdentifier(c *gin.Context) string {
	return DefaultContextHelper.GetIdentifier(c)
}

func SetAuthInfo(c *gin.Context, user *entities.User, apiKey *entities.APIKey) {
	DefaultContextHelper.SetAuthInfo(c, user, apiKey)
}
