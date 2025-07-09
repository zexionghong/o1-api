package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"ai-api-gateway/internal/domain/entities"
)

// AIProviderClient AI提供商客户端接口
type AIProviderClient interface {
	// SendRequest 发送请求到AI提供商
	SendRequest(ctx context.Context, provider *entities.Provider, request *AIRequest) (*AIResponse, error)

	// SendStreamRequest 发送流式请求到AI提供商
	SendStreamRequest(ctx context.Context, provider *entities.Provider, request *AIRequest, streamChan chan<- *StreamChunk) error

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context, provider *entities.Provider) error

	// GetModels 获取提供商支持的模型列表
	GetModels(ctx context.Context, provider *entities.Provider) ([]*AIModel, error)
}

// StreamChunk 流式响应块（从gateway包移动到这里避免循环依赖）
type StreamChunk struct {
	ID           string    `json:"id"`
	Object       string    `json:"object"`
	Created      int64     `json:"created"`
	Model        string    `json:"model"`
	Content      string    `json:"content"`
	FinishReason *string   `json:"finish_reason"`
	Usage        *AIUsage  `json:"usage,omitempty"`
	Cost         *CostInfo `json:"cost,omitempty"`
}

// CostInfo 成本信息（从gateway包移动到这里避免循环依赖）
type CostInfo struct {
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost  float64 `json:"total_cost"`
	Currency   string  `json:"currency"`
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
	fmt.Println("-------------------")
	fmt.Println("URL:", url)
	fmt.Println("Request struct:", request)

	// 序列化请求体查看实际JSON
	if jsonBytes, err := json.Marshal(request); err == nil {
		fmt.Println("Request JSON:", string(jsonBytes))
	}

	if provider.APIKeyEncrypted != nil {
		fmt.Printf("API Key: %s\n", *provider.APIKeyEncrypted)
	} else {
		fmt.Println("API Key: nil")
	}

	for k, v := range headers {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Println("-------------------")
	resp, err := c.httpClient.Post(ctx, url, request, headers)
	if err != nil {
		fmt.Println("------------")
		fmt.Println(resp)
		fmt.Println("------------")
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

// SendStreamRequest 发送流式请求到AI提供商
func (c *aiProviderClientImpl) SendStreamRequest(ctx context.Context, provider *entities.Provider, request *AIRequest, streamChan chan<- *StreamChunk) error {
	// 构造请求URL
	url := fmt.Sprintf("%s/chat/completions", provider.BaseURL)

	// 构造请求头
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Accept":        "text/event-stream",
		"Cache-Control": "no-cache",
	}

	// 根据提供商类型设置认证头
	switch provider.Slug {
	case "302":
		if provider.APIKeyEncrypted != nil {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "openai":
		if provider.APIKeyEncrypted != nil {
			headers["Authorization"] = fmt.Sprintf("Bearer %s", *provider.APIKeyEncrypted)
		}
	case "anthropic":
		if provider.APIKeyEncrypted != nil {
			headers["x-api-key"] = *provider.APIKeyEncrypted
			headers["anthropic-version"] = "2023-06-01"
		}
		// Anthropic使用不同的端点
		url = fmt.Sprintf("%s/messages", provider.BaseURL)
	}

	// 确保请求是流式的
	streamRequest := *request
	streamRequest.Stream = true

	// 发送流式请求
	resp, err := c.httpClient.PostStream(ctx, url, &streamRequest, headers)
	if err != nil {
		return fmt.Errorf("failed to send stream request to provider %s: %w", provider.Name, err)
	}
	defer resp.Close()

	// 检查响应状态
	if !resp.IsSuccess() {
		return fmt.Errorf("provider %s returned error status: %d", provider.Name, resp.StatusCode)
	}

	// 解析SSE流
	return c.parseSSEStream(ctx, resp, streamChan, request.Model)
}

// parseSSEStream 解析SSE流
func (c *aiProviderClientImpl) parseSSEStream(ctx context.Context, resp *StreamResponse, streamChan chan<- *StreamChunk, model string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := resp.ReadLine()
			if err != nil {
				if err == io.EOF {
					return nil // 流结束
				}
				return fmt.Errorf("failed to read stream line: %w", err)
			}

			// 跳过空行和注释行
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// 解析SSE数据
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// 检查是否是结束标记
				if data == "[DONE]" {
					return nil
				}

				// 解析JSON数据
				var sseData map[string]interface{}
				if err := json.Unmarshal([]byte(data), &sseData); err != nil {
					// 跳过无法解析的数据
					continue
				}

				// 转换为StreamChunk
				chunk := c.convertToStreamChunk(sseData, model)
				if chunk != nil {
					select {
					case streamChan <- chunk:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}
}

// convertToStreamChunk 转换SSE数据为StreamChunk
func (c *aiProviderClientImpl) convertToStreamChunk(data map[string]interface{}, model string) *StreamChunk {
	chunk := &StreamChunk{
		Object:  "chat.completion.chunk",
		Model:   model,
		Created: time.Now().Unix(),
	}

	// 提取ID
	if id, ok := data["id"].(string); ok {
		chunk.ID = id
	} else {
		chunk.ID = fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
	}

	// 提取choices
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			// 提取delta内容
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if content, ok := delta["content"].(string); ok {
					chunk.Content = content
				}
			}

			// 提取finish_reason
			if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
				chunk.FinishReason = &finishReason
			}
		}
	}

	// 提取usage信息
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		chunk.Usage = &AIUsage{}
		if promptTokens, ok := usage["prompt_tokens"].(float64); ok {
			chunk.Usage.PromptTokens = int(promptTokens)
		}
		if completionTokens, ok := usage["completion_tokens"].(float64); ok {
			chunk.Usage.CompletionTokens = int(completionTokens)
		}
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			chunk.Usage.TotalTokens = int(totalTokens)
		}
	}

	return chunk
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
