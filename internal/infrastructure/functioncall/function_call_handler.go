package functioncall

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/logger"
)

// FunctionCallHandler Function call 处理器接口
type FunctionCallHandler interface {
	HandleFunctionCalls(ctx context.Context, messages []clients.AIMessage, toolCalls []clients.ToolCall) ([]clients.AIMessage, error)
	GetAvailableTools() []clients.Tool
}

// functionCallHandlerImpl Function call 处理器实现
type functionCallHandlerImpl struct {
	searchService SearchService
	logger        logger.Logger
}

// NewFunctionCallHandler 创建 Function call 处理器
func NewFunctionCallHandler(searchService SearchService, logger logger.Logger) FunctionCallHandler {
	return &functionCallHandlerImpl{
		searchService: searchService,
		logger:        logger,
	}
}

// GetAvailableTools 获取可用的工具列表
func (h *functionCallHandlerImpl) GetAvailableTools() []clients.Tool {
	return []clients.Tool{
		{
			Type: "function",
			Function: clients.Function{
				Name:        "search",
				Description: "Search for information on the internet",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The search query to execute",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: clients.Function{
				Name:        "news",
				Description: "Search for news articles",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "The news search query to execute",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: clients.Function{
				Name:        "crawler",
				Description: "Get the content of a specified URL",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "The URL of the webpage to crawl",
						},
					},
					"required": []string{"url"},
				},
			},
		},
	}
}

// HandleFunctionCalls 处理函数调用
func (h *functionCallHandlerImpl) HandleFunctionCalls(ctx context.Context, messages []clients.AIMessage, toolCalls []clients.ToolCall) ([]clients.AIMessage, error) {
	h.logger.WithFields(map[string]interface{}{
		"tool_calls_count": len(toolCalls),
	}).Info("Processing function calls")

	var newMessages []clients.AIMessage

	for _, toolCall := range toolCalls {
		h.logger.WithFields(map[string]interface{}{
			"tool_call_id":   toolCall.ID,
			"function_name":  toolCall.Function.Name,
			"function_args":  toolCall.Function.Arguments,
		}).Info("Executing function call")

		result, err := h.executeFunction(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
		if err != nil {
			h.logger.WithFields(map[string]interface{}{
				"tool_call_id":  toolCall.ID,
				"function_name": toolCall.Function.Name,
				"error":         err.Error(),
			}).Error("Function call execution failed")

			// 创建错误响应消息
			result = fmt.Sprintf("Error executing function %s: %s", toolCall.Function.Name, err.Error())
		}

		// 创建工具响应消息
		toolMessage := clients.AIMessage{
			Role:       "tool",
			Content:    result,
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
		}

		newMessages = append(newMessages, toolMessage)

		h.logger.WithFields(map[string]interface{}{
			"tool_call_id":  toolCall.ID,
			"function_name": toolCall.Function.Name,
			"result_length": len(result),
		}).Info("Function call executed successfully")
	}

	return newMessages, nil
}

// executeFunction 执行具体的函数
func (h *functionCallHandlerImpl) executeFunction(ctx context.Context, functionName, arguments string) (string, error) {
	// 解析函数参数
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse function arguments: %w", err)
		}
	}

	// 根据函数名执行相应的功能
	switch functionName {
	case "search":
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("search function requires a 'query' parameter of type string")
		}
		return h.searchService.Search(ctx, query)

	case "news":
		query, ok := args["query"].(string)
		if !ok {
			return "", fmt.Errorf("news function requires a 'query' parameter of type string")
		}
		return h.searchService.SearchNews(ctx, query)

	case "crawler":
		url, ok := args["url"].(string)
		if !ok {
			return "", fmt.Errorf("crawler function requires a 'url' parameter of type string")
		}
		return h.searchService.CrawlURL(ctx, url)

	default:
		return "", fmt.Errorf("unknown function: %s", functionName)
	}
}

// ShouldUseFunctionCall 判断是否应该使用 function call
func ShouldUseFunctionCall(messages []clients.AIMessage) bool {
	if len(messages) == 0 {
		return false
	}

	// 获取最后一条用户消息
	var lastUserMessage string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMessage = messages[i].Content
			break
		}
	}

	if lastUserMessage == "" {
		return false
	}

	// 简单的关键词检测，判断是否需要搜索
	searchKeywords := []string{
		"搜索", "查找", "search", "find", "lookup",
		"新闻", "news", "最新", "latest", "recent",
		"网页", "网站", "url", "webpage", "website",
		"什么是", "what is", "how to", "怎么",
		"今天", "today", "现在", "now", "当前", "current",
	}

	content := lastUserMessage
	for _, keyword := range searchKeywords {
		if contains(content, keyword) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	// 简单的包含检查，可以根据需要改进
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsInMiddle(s, substr)))
}

// containsInMiddle 检查字符串中间是否包含子字符串
func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GenerateToolCallID 生成工具调用ID
func GenerateToolCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}
