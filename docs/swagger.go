// Package docs AI API Gateway API Documentation
//
//	@title			AI API Gateway
//	@version		1.0.0
//	@description	AI API Gateway是一个高性能的AI API网关，提供统一的API接口来访问多个AI提供商。
//	@description
//	@description	## 主要功能
//	@description	- 🤖 多AI提供商支持（OpenAI、Anthropic等）
//	@description	- ⚖️ 智能负载均衡和故障转移
//	@description	- 📊 精确的配额管理和计费
//	@description	- 🔐 完整的认证和授权
//	@description	- 📈 实时监控和统计
//	@description
//	@description	## 认证方式
//	@description	所有API请求都需要在请求头中包含有效的API密钥：
//	@description	```
//	@description	Authorization: Bearer YOUR_API_KEY
//	@description	```
//	@description
//	@description	## 快速开始
//	@description	1. 使用管理API创建用户账户
//	@description	2. 为用户生成API密钥
//	@description	3. 使用API密钥调用AI接口
//	@description
//	@description	## 支持的模型
//	@description	- GPT-3.5-turbo
//	@description	- GPT-4
//	@description	- Claude-3-haiku
//	@description	- Claude-3-sonnet
//	@description	- Claude-3-opus
//
//	@contact.name	AI API Gateway Team
//	@contact.email	support@example.com
//	@contact.url	https://example.com/support
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/
//	@schemes	http https
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				API密钥认证，格式：Bearer YOUR_API_KEY
package docs

import (
	_ "ai-api-gateway/internal/application/dto"
	_ "ai-api-gateway/internal/domain/entities"
)

// 响应模型定义

// 通用错误响应
// swagger:response errorResponse
type ErrorResponseWrapper struct {
	// 错误响应
	// in: body
	Body struct {
		Success   bool   `json:"success" example:"false"`
		Error     Error  `json:"error"`
		Timestamp string `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	}
}

// Error 错误详情
type Error struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"请求参数无效"`
	Details any    `json:"details,omitempty"`
}

// 健康检查响应
// swagger:response healthResponse
type HealthResponseWrapper struct {
	// 健康检查响应
	// in: body
	Body struct {
		Success bool   `json:"success" example:"true"`
		Status  string `json:"status" example:"healthy"`
		Message string `json:"message" example:"Service is healthy"`
	}
}

// 系统统计响应
// swagger:response statsResponse
type StatsResponseWrapper struct {
	// 系统统计响应
	// in: body
	Body struct {
		Success bool  `json:"success" example:"true"`
		Data    Stats `json:"data"`
	}
}

// 系统统计数据
type Stats struct {
	Uptime           string `json:"uptime" example:"1h30m45s"`
	TotalRequests    int64  `json:"total_requests" example:"12345"`
	ActiveUsers      int64  `json:"active_users" example:"123"`
	HealthyProviders int    `json:"healthy_providers" example:"2"`
}

// 聊天补全响应
// swagger:response chatCompletionResponse
type ChatCompletionResponseWrapper struct {
	// 聊天补全响应
	// in: body
	Body struct {
		ID      string   `json:"id" example:"chatcmpl-123"`
		Object  string   `json:"object" example:"chat.completion"`
		Created int64    `json:"created" example:"1640995200"`
		Model   string   `json:"model" example:"gpt-3.5-turbo"`
		Choices []Choice `json:"choices"`
		Usage   Usage    `json:"usage"`
	}
}

// 选择项
type Choice struct {
	Index        int     `json:"index" example:"0"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason" example:"stop"`
}

// 消息
type Message struct {
	Role    string `json:"role" example:"assistant"`
	Content string `json:"content" example:"Hello! How can I help you today?"`
}

// 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens" example:"10"`
	CompletionTokens int `json:"completion_tokens" example:"20"`
	TotalTokens      int `json:"total_tokens" example:"30"`
}

// 文本补全响应
// swagger:response completionResponse
type CompletionResponseWrapper struct {
	// 文本补全响应
	// in: body
	Body struct {
		ID      string             `json:"id" example:"cmpl-123"`
		Object  string             `json:"object" example:"text_completion"`
		Created int64              `json:"created" example:"1640995200"`
		Model   string             `json:"model" example:"gpt-3.5-turbo"`
		Choices []CompletionChoice `json:"choices"`
		Usage   Usage              `json:"usage"`
	}
}

// 补全选择项
type CompletionChoice struct {
	Text         string `json:"text" example:"This is a completion."`
	Index        int    `json:"index" example:"0"`
	FinishReason string `json:"finish_reason" example:"stop"`
}

// 模型列表响应
// swagger:response modelsResponse
type ModelsResponseWrapper struct {
	// 模型列表响应
	// in: body
	Body struct {
		Object string  `json:"object" example:"list"`
		Data   []Model `json:"data"`
	}
}

// 模型信息
type Model struct {
	ID      string `json:"id" example:"gpt-3.5-turbo"`
	Object  string `json:"object" example:"model"`
	Created int64  `json:"created" example:"1640995200"`
	OwnedBy string `json:"owned_by" example:"openai"`
}

// 使用情况响应
// swagger:response usageResponse
type UsageResponseWrapper struct {
	// 使用情况响应
	// in: body
	Body struct {
		Success bool      `json:"success" example:"true"`
		Data    UsageData `json:"data"`
	}
}

// 使用数据
type UsageData struct {
	TotalRequests int     `json:"total_requests" example:"100"`
	TotalTokens   int     `json:"total_tokens" example:"5000"`
	TotalCost     float64 `json:"total_cost" example:"0.05"`
}

// 用户响应
// swagger:response userResponse
type UserResponseWrapper struct {
	// 用户响应
	// in: body
	Body struct {
		Success bool     `json:"success" example:"true"`
		Data    UserData `json:"data"`
	}
}

// 用户数据
type UserData struct {
	ID        int64   `json:"id" example:"1"`
	Username  string  `json:"username" example:"john_doe"`
	Email     string  `json:"email" example:"john@example.com"`
	Balance   float64 `json:"balance" example:"100.50"`
	Status    string  `json:"status" example:"active"`
	CreatedAt string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// 用户列表响应
// swagger:response usersListResponse
type UsersListResponseWrapper struct {
	// 用户列表响应
	// in: body
	Body struct {
		Success bool       `json:"success" example:"true"`
		Data    []UserData `json:"data"`
		Total   int        `json:"total" example:"100"`
		Page    int        `json:"page" example:"1"`
		Limit   int        `json:"limit" example:"10"`
	}
}

// API密钥响应
// swagger:response apiKeyResponse
type APIKeyResponseWrapper struct {
	// API密钥响应
	// in: body
	Body struct {
		Success bool       `json:"success" example:"true"`
		Data    APIKeyData `json:"data"`
	}
}

// API密钥数据
type APIKeyData struct {
	ID         int64  `json:"id" example:"1"`
	UserID     int64  `json:"user_id" example:"1"`
	KeyPrefix  string `json:"key_prefix" example:"ak_abc123"`
	Name       string `json:"name" example:"My API Key"`
	Status     string `json:"status" example:"active"`
	CreatedAt  string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt  string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	LastUsedAt string `json:"last_used_at,omitempty" example:"2024-01-01T00:00:00Z"`
	ExpiresAt  string `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

// API密钥创建响应
// swagger:response apiKeyCreateResponse
type APIKeyCreateResponseWrapper struct {
	// API密钥创建响应
	// in: body
	Body struct {
		Success bool             `json:"success" example:"true"`
		Data    APIKeyCreateData `json:"data"`
		Message string           `json:"message" example:"API密钥创建成功"`
	}
}

// API密钥创建数据
type APIKeyCreateData struct {
	ID        int64  `json:"id" example:"1"`
	UserID    int64  `json:"user_id" example:"1"`
	Key       string `json:"key" example:"ak_1234567890abcdef1234567890abcdef12345678"`
	KeyPrefix string `json:"key_prefix" example:"ak_abc123"`
	Name      string `json:"name" example:"My API Key"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	ExpiresAt string `json:"expires_at,omitempty" example:"2024-12-31T23:59:59Z"`
}

// API密钥列表响应
// swagger:response apiKeysListResponse
type APIKeysListResponseWrapper struct {
	// API密钥列表响应
	// in: body
	Body struct {
		Success bool         `json:"success" example:"true"`
		Data    []APIKeyData `json:"data"`
		Total   int          `json:"total" example:"10"`
		Page    int          `json:"page" example:"1"`
		Limit   int          `json:"limit" example:"10"`
	}
}
