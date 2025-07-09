package clients

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// StreamChunk 流式响应数据块
type StreamChunk struct {
	ID           string   `json:"id"`
	Object       string   `json:"object"`
	Created      int64    `json:"created"`
	Model        string   `json:"model"`
	Content      string   `json:"content"`
	FinishReason *string  `json:"finish_reason"`
	Usage        *AIUsage `json:"usage,omitempty"`
	Cost         *AICost  `json:"cost,omitempty"`
}

// AIRequest AI请求 (通用结构)
type AIRequest struct {
	Model       string                 `json:"model"`
	Messages    []AIMessage            `json:"messages,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Tools       []Tool                 `json:"tools,omitempty"`       // Function call tools
	ToolChoice  interface{}            `json:"tool_choice,omitempty"` // Tool choice strategy
	WebSearch   bool                   `json:"web_search,omitempty"`  // 是否启用联网搜索
	Extra       map[string]interface{} `json:"-"`                     // 额外参数
}

// ChatCompletionRequest 聊天补全请求
type ChatCompletionRequest struct {
	Model       string      `json:"model" binding:"required" example:"gpt-3.5-turbo"`
	Messages    []AIMessage `json:"messages" binding:"required,min=1"`
	MaxTokens   int         `json:"max_tokens,omitempty" example:"150"`
	Temperature float64     `json:"temperature,omitempty" example:"0.7"`
	Stream      bool        `json:"stream,omitempty" example:"false"`
	Tools       []Tool      `json:"tools,omitempty"`                      // Function call tools
	ToolChoice  interface{} `json:"tool_choice,omitempty"`                // Tool choice strategy
	WebSearch   bool        `json:"web_search,omitempty" example:"false"` // 是否启用联网搜索
}

// CompletionRequest 文本补全请求
type CompletionRequest struct {
	Model       string  `json:"model" binding:"required" example:"gpt-3.5-turbo"`
	Prompt      string  `json:"prompt" binding:"required" example:"Once upon a time"`
	MaxTokens   int     `json:"max_tokens,omitempty" example:"150"`
	Temperature float64 `json:"temperature,omitempty" example:"0.7"`
	Stream      bool    `json:"stream,omitempty" example:"false"`
	WebSearch   bool    `json:"web_search,omitempty" example:"false"` // 是否启用联网搜索
}

// AIMessage AI消息
type AIMessage struct {
	Role       string     `json:"role" binding:"required" example:"user" enums:"system,user,assistant,tool"`
	Content    string     `json:"content" example:"Hello, how are you?"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // Function calls made by assistant
	ToolCallID string     `json:"tool_call_id,omitempty"` // ID of the tool call this message is responding to
	Name       string     `json:"name,omitempty"`         // Name of the function for tool messages
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
	Index        int        `json:"index"`
	Message      AIMessage  `json:"message,omitempty"`
	Text         string     `json:"text,omitempty"`
	FinishReason string     `json:"finish_reason"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"` // Function calls in streaming mode
}

// AIUsage AI使用情况
type AIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AICost AI成本信息
type AICost struct {
	PromptCost     float64 `json:"prompt_cost"`
	CompletionCost float64 `json:"completion_cost"`
	TotalCost      float64 `json:"total_cost"`
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

// Tool Function call tool definition
type Tool struct {
	Type     string   `json:"type" example:"function"`
	Function Function `json:"function"`
}

// Function Function definition for tool calls
type Function struct {
	Name        string      `json:"name" example:"search"`
	Description string      `json:"description" example:"Search for information"`
	Parameters  interface{} `json:"parameters"` // JSON Schema for function parameters
}

// ToolCall Function call made by the assistant
type ToolCall struct {
	ID       string       `json:"id" example:"call_123"`
	Type     string       `json:"type" example:"function"`
	Function FunctionCall `json:"function"`
}

// FunctionCall Function call details
type FunctionCall struct {
	Name      string `json:"name" example:"search"`
	Arguments string `json:"arguments"` // JSON string of function arguments
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

// SendStreamRequest 发送流式请求到AI提供商
func (c *aiProviderClientImpl) SendStreamRequest(ctx context.Context, provider *entities.Provider, request *AIRequest, streamChan chan<- *StreamChunk) error {
	// 确保请求是流式的
	request.Stream = true

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

	// 打印流式请求信息
	fmt.Println("=== SENDING STREAM REQUEST ===")
	fmt.Println("URL:", url)
	fmt.Println("Request struct:", request)

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
	fmt.Println("===============================")

	// 发送流式请求
	return c.sendStreamRequestToProvider(ctx, url, request, headers, streamChan)
}

// 工厂方法
func NewOpenAIClient(httpClient HTTPClient) AIProviderClient {
	return NewAIProviderClient(httpClient)
}

func NewAnthropicClient(httpClient HTTPClient) AIProviderClient {
	return NewAIProviderClient(httpClient)
}

// sendStreamRequestToProvider 发送流式请求到提供商的具体实现
func (c *aiProviderClientImpl) sendStreamRequestToProvider(ctx context.Context, url string, request *AIRequest, headers map[string]string, streamChan chan<- *StreamChunk) error {
	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	fmt.Println("Request JSON:", string(requestBody))

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	for key, value := range headers {
		httpReq.Header.Set(key, value)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 发送请求
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 处理流式响应
	return c.processStreamResponse(ctx, resp.Body, streamChan, request.Model)
}

// processStreamResponse 处理流式响应
func (c *aiProviderClientImpl) processStreamResponse(ctx context.Context, body io.Reader, streamChan chan<- *StreamChunk, model string) error {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line := scanner.Text()

			// 跳过空行和注释行
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			// 移除 "data: " 前缀
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				fmt.Println("Stream completed with [DONE] marker")
				return nil
			}

			// 解析JSON数据
			var sseData map[string]interface{}
			if err := json.Unmarshal([]byte(data), &sseData); err != nil {
				fmt.Printf("Failed to parse SSE data: %s, error: %v\n", data, err)
				continue
			}
			fmt.Printf("Received SSE data: %s\n", sseData)

			// 提取内容
			content := ""
			if choices, ok := sseData["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if deltaContent, ok := delta["content"].(string); ok {
							content = deltaContent
						}
					}
				}
			}

			// 构造流式数据块
			chunk := &StreamChunk{
				ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
				Object:  "chat.completion.chunk",
				Created: time.Now().Unix(),
				Model:   model,
				Content: content,
			}

			// 发送数据块
			select {
			case streamChan <- chunk:
				fmt.Printf("Received and forwarded chunk: %s\n", content)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}
