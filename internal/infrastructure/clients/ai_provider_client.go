package clients

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/domain/entities"
)

// AIProviderClient AI提供商客户端接口
type AIProviderClient interface {
	// SendRequest 发送请求到AI提供商
	SendRequest(ctx context.Context, provider *entities.Provider, request *AIRequest) (*AIResponse, error)

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context, provider *entities.Provider) error

	// GetModels 获取提供商支持的模型列表
	GetModels(ctx context.Context, provider *entities.Provider) ([]*AIModel, error)
}

// AIRequest AI请求 (通用结构)
type AIRequest struct {
	Model       string                 `json:"model"`
	Messages    []AIMessage            `json:"messages,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Extra       map[string]interface{} `json:"-"` // 额外参数
}

// ChatCompletionRequest 聊天补全请求
type ChatCompletionRequest struct {
	Model       string      `json:"model" binding:"required" example:"gpt-3.5-turbo"`
	Messages    []AIMessage `json:"messages" binding:"required,min=1"`
	MaxTokens   int         `json:"max_tokens,omitempty" example:"150"`
	Temperature float64     `json:"temperature,omitempty" example:"0.7"`
	Stream      bool        `json:"stream,omitempty" example:"false"`
}

// CompletionRequest 文本补全请求
type CompletionRequest struct {
	Model       string  `json:"model" binding:"required" example:"gpt-3.5-turbo"`
	Prompt      string  `json:"prompt" binding:"required" example:"Once upon a time"`
	MaxTokens   int     `json:"max_tokens,omitempty" example:"150"`
	Temperature float64 `json:"temperature,omitempty" example:"0.7"`
	Stream      bool    `json:"stream,omitempty" example:"false"`
}

// AIMessage AI消息
type AIMessage struct {
	Role    string `json:"role" binding:"required" example:"user" enums:"system,user,assistant"`
	Content string `json:"content" binding:"required" example:"Hello, how are you?"`
}

// AIResponse AI响应
type AIResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int64      `json:"created"`
	Model   string     `json:"model"`
	Choices []AIChoice `json:"choices"`
	Usage   AIUsage    `json:"usage"`
	Error   *AIError   `json:"error,omitempty"`
}

// AIChoice AI选择
type AIChoice struct {
	Index        int       `json:"index"`
	Message      AIMessage `json:"message,omitempty"`
	Text         string    `json:"text,omitempty"`
	FinishReason string    `json:"finish_reason"`
}

// AIUsage AI使用情况
type AIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIError AI错误
type AIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// AIModel AI模型信息
type AIModel struct {
	ID         string        `json:"id"`
	Object     string        `json:"object"`
	Created    int64         `json:"created"`
	OwnedBy    string        `json:"owned_by"`
	Permission []interface{} `json:"permission"`
}

// ModelsResponse 模型列表响应
type ModelsResponse struct {
	Object string    `json:"object"`
	Data   []AIModel `json:"data"`
}

// UsageResponse 使用情况响应
type UsageResponse struct {
	TotalRequests int     `json:"total_requests"`
	TotalTokens   int     `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
}

// aiProviderClientImpl AI提供商客户端实现
type aiProviderClientImpl struct {
	httpClient HTTPClient
}

// NewAIProviderClient 创建AI提供商客户端
func NewAIProviderClient(httpClient HTTPClient) AIProviderClient {
	return &aiProviderClientImpl{
		httpClient: httpClient,
	}
}

// SendRequest 发送请求到AI提供商
func (c *aiProviderClientImpl) SendRequest(ctx context.Context, provider *entities.Provider, request *AIRequest) (*AIResponse, error) {
	// 构造请求URL
	url := fmt.Sprintf("%s/chat/completions", provider.BaseURL)

	// 构造请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 根据提供商类型设置认证头
	switch provider.Slug {
	case "302":
		if provider.APIKeyEncrypted != nil {
			// TODO: 解密API密钥
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "openai":
		if provider.APIKeyEncrypted != nil {
			// TODO: 解密API密钥
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "anthropic":
		if provider.APIKeyEncrypted != nil {
			// TODO: 解密API密钥
			headers["x-api-key"] = *provider.APIKeyEncrypted
			headers["anthropic-version"] = "2023-06-01"
		}
		// Anthropic使用不同的端点
		url = fmt.Sprintf("%s/messages", provider.BaseURL)
	}

	// 发送请求
	resp, err := c.httpClient.Post(ctx, url, request, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to provider %s: %w", provider.Name, err)
	}

	// 解析响应
	var aiResp AIResponse
	if err := resp.UnmarshalJSON(&aiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from provider %s: %w", provider.Name, err)
	}

	// 检查是否有错误
	if aiResp.Error != nil {
		return &aiResp, fmt.Errorf("provider %s returned error: %s", provider.Name, aiResp.Error.Message)
	}

	return &aiResp, nil
}

// HealthCheck 健康检查
func (c *aiProviderClientImpl) HealthCheck(ctx context.Context, provider *entities.Provider) error {
	// 使用模型列表端点进行健康检查
	url := fmt.Sprintf("%s/models", provider.BaseURL)

	// 构造请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 根据提供商类型设置认证头
	switch provider.Slug {
	case "openai":
		if provider.APIKeyEncrypted != nil {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "anthropic":
		if provider.APIKeyEncrypted != nil {
			headers["x-api-key"] = *provider.APIKeyEncrypted
			headers["anthropic-version"] = "2023-06-01"
		}
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, time.Duration(provider.TimeoutSeconds)*time.Second)
	defer cancel()

	// 发送请求
	resp, err := c.httpClient.Get(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("health check failed for provider %s: %w", provider.Name, err)
	}

	// 检查响应状态
	if !resp.IsSuccess() {
		return fmt.Errorf("health check failed for provider %s: status %d", provider.Name, resp.StatusCode)
	}

	return nil
}

// GetModels 获取提供商支持的模型列表
func (c *aiProviderClientImpl) GetModels(ctx context.Context, provider *entities.Provider) ([]*AIModel, error) {
	// 构造请求URL
	url := fmt.Sprintf("%s/models", provider.BaseURL)

	// 构造请求头
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 根据提供商类型设置认证头
	switch provider.Slug {
	case "openai":
		if provider.APIKeyEncrypted != nil {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "anthropic":
		if provider.APIKeyEncrypted != nil {
			headers["x-api-key"] = *provider.APIKeyEncrypted
			headers["anthropic-version"] = "2023-06-01"
		}
	}

	// 发送请求
	resp, err := c.httpClient.Get(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get models from provider %s: %w", provider.Name, err)
	}

	// 解析响应
	var modelsResp struct {
		Data []AIModel `json:"data"`
	}
	if err := resp.UnmarshalJSON(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal models response from provider %s: %w", provider.Name, err)
	}

	// 转换为指针切片
	models := make([]*AIModel, len(modelsResp.Data))
	for i := range modelsResp.Data {
		models[i] = &modelsResp.Data[i]
	}

	return models, nil
}

// 工厂方法
func NewOpenAIClient(httpClient HTTPClient) AIProviderClient {
	return NewAIProviderClient(httpClient)
}

func NewAnthropicClient(httpClient HTTPClient) AIProviderClient {
	return NewAIProviderClient(httpClient)
}
