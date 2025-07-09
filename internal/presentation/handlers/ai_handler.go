package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/gateway"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/presentation/middleware"

	"github.com/gin-gonic/gin"
)

// AIHandler AI请求处理器
type AIHandler struct {
	gatewayService gateway.GatewayService
	logger         logger.Logger
}

// NewAIHandler 创建AI请求处理器
func NewAIHandler(gatewayService gateway.GatewayService, logger logger.Logger) *AIHandler {
	return &AIHandler{
		gatewayService: gatewayService,
		logger:         logger,
	}
}

// handleStreamingRequest 处理流式请求
func (h *AIHandler) handleStreamingRequest(c *gin.Context, gatewayRequest *gateway.GatewayRequest, requestID string, userID, apiKeyID int64) {
	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Request-ID", requestID)

	// 获取响应写入器
	w := c.Writer

	// 创建流式响应通道
	streamChan := make(chan *clients.StreamChunk, 100)
	errorChan := make(chan error, 1)

	// 启动流式处理
	go func() {
		defer func() {
			// 安全关闭channels
			select {
			case <-streamChan:
			default:
				close(streamChan)
			}

			select {
			case <-errorChan:
			default:
				close(errorChan)
			}
		}()

		err := h.gatewayService.ProcessStreamRequest(c.Request.Context(), gatewayRequest, streamChan)
		if err != nil {
			select {
			case errorChan <- err:
			case <-c.Request.Context().Done():
				// 如果上下文已取消，不发送错误
			}
		}
	}()

	// 发送流式数据
	var totalTokens int
	var totalCost float64

	for {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				// 流结束，发送结束标记
				_, err := w.Write([]byte("data: [DONE]\n\n"))
				if err != nil {
					h.logger.WithFields(map[string]interface{}{
						"request_id": requestID,
						"error":      err.Error(),
					}).Error("Failed to write stream end marker")
				}
				w.Flush()

				// 输出流式AI提供商响应结果
				h.logger.WithFields(map[string]interface{}{
					"request_id":   requestID,
					"user_id":      userID,
					"api_key_id":   apiKeyID,
					"total_tokens": totalTokens,
					"total_cost":   totalCost,
					"stream_type":  "completed",
				}).Info("AI provider streaming response completed successfully")

				// 设置使用量到上下文
				c.Set("tokens_used", totalTokens)
				c.Set("cost_used", totalCost)
				return
			}

			// 累计使用量
			if chunk.Usage != nil {
				totalTokens += chunk.Usage.TotalTokens
			}
			if chunk.Cost != nil {
				totalCost += chunk.Cost.TotalCost
			}

			// 构造SSE数据
			data := map[string]interface{}{
				"id":      chunk.ID,
				"object":  "chat.completion.chunk",
				"created": chunk.Created,
				"model":   chunk.Model,
				"choices": []map[string]interface{}{
					{
						"index": 0,
						"delta": map[string]interface{}{
							"content": chunk.Content,
						},
						"finish_reason": chunk.FinishReason,
					},
				},
			}

			// 序列化为JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				h.logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to marshal stream chunk")
				continue
			}

			// 发送SSE数据
			_, err = w.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
			if err != nil {
				h.logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to write stream chunk")
				return
			}

			// 立即刷新缓冲区
			w.Flush()

		case err := <-errorChan:
			if err != nil {
				h.logger.WithFields(map[string]interface{}{
					"request_id": requestID,
					"user_id":    userID,
					"api_key_id": apiKeyID,
					"error":      err.Error(),
				}).Error("Stream processing failed")

				// 发送错误事件
				errorData := map[string]interface{}{
					"error": map[string]interface{}{
						"message": "Stream processing failed",
						"type":    "server_error",
						"code":    "stream_error",
					},
				}

				jsonData, _ := json.Marshal(errorData)
				w.Write([]byte(fmt.Sprintf("data: %s\n\n", jsonData)))
				w.Flush()
			}
			return

		case <-c.Request.Context().Done():
			// 客户端断开连接
			h.logger.WithFields(map[string]interface{}{
				"request_id": requestID,
			}).Info("Client disconnected from stream")
			return
		}
	}
}

// ChatCompletions 处理聊天完成请求
// @Summary 聊天补全
// @Description 创建聊天补全请求，兼容OpenAI API格式。支持流式和非流式响应。
// @Tags AI接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body clients.ChatCompletionRequest true "聊天补全请求"
// @Success 200 {object} clients.AIResponse "聊天补全响应"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Failure 401 {object} dto.Response "认证失败"
// @Failure 429 {object} dto.Response "请求过于频繁"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /v1/chat/completions [post]
func (h *AIHandler) ChatCompletions(c *gin.Context) {
	// 获取认证信息
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"Authentication required",
			nil,
		))
		return
	}

	apiKeyID, exists := middleware.GetAPIKeyIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"API key required",
			nil,
		))
		return
	}

	// 解析请求体
	var chatRequest clients.ChatCompletionRequest
	if err := c.ShouldBindJSON(&chatRequest); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id":    userID,
			"api_key_id": apiKeyID,
			"error":      err.Error(),
		}).Warn("Invalid request body")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	// 验证必需字段
	if chatRequest.Model == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_MODEL",
			"Model is required",
			nil,
		))
		return
	}

	if len(chatRequest.Messages) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_MESSAGES",
			"Messages array is required for chat completions",
			nil,
		))
		return
	}

	// 获取请求ID
	requestID := middleware.GetRequestIDFromContext(c)

	// 转换为通用 AIRequest 结构
	aiRequest := &clients.AIRequest{
		Model:       chatRequest.Model,
		Messages:    chatRequest.Messages,
		MaxTokens:   chatRequest.MaxTokens,
		Temperature: chatRequest.Temperature,
		Stream:      chatRequest.Stream,
	}

	// 构造网关请求
	gatewayRequest := &gateway.GatewayRequest{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		ModelSlug: aiRequest.Model,
		Request:   aiRequest,
		RequestID: requestID,
	}

	// 检查是否为流式请求
	if aiRequest.Stream {
		h.handleStreamingRequest(c, gatewayRequest, requestID, userID, apiKeyID)
		return
	}

	// 处理非流式请求
	response, err := h.gatewayService.ProcessRequest(c.Request.Context(), gatewayRequest)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
			"api_key_id": apiKeyID,
			"model":      aiRequest.Model,
			"error":      err.Error(),
		}).Error("Failed to process AI request")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"REQUEST_FAILED",
			"Failed to process request",
			map[string]interface{}{
				"request_id": requestID,
			},
		))
		return
	}

	// 输出AI提供商响应结果
	h.logger.WithFields(map[string]interface{}{
		"request_id":    requestID,
		"user_id":       userID,
		"api_key_id":    apiKeyID,
		"model":         aiRequest.Model,
		"provider":      response.Provider,
		"duration_ms":   response.Duration.Milliseconds(),
		"total_tokens":  response.Usage.TotalTokens,
		"input_tokens":  response.Usage.InputTokens,
		"output_tokens": response.Usage.OutputTokens,
		"total_cost":    response.Cost.TotalCost,
		"response_data": response.Response,
	}).Info("AI provider response received successfully")

	// 设置使用量到上下文（用于配额中间件）
	c.Set("tokens_used", response.Usage.TotalTokens)
	c.Set("cost_used", response.Cost.TotalCost)

	// 设置响应头
	c.Header("X-Request-ID", requestID)
	c.Header("X-Provider", response.Provider)
	c.Header("X-Model", response.Model)
	c.Header("X-Duration-Ms", strconv.FormatInt(response.Duration.Milliseconds(), 10))

	// 返回AI响应（保持与OpenAI API兼容的格式）
	c.JSON(http.StatusOK, response.Response)
}

// Completions 处理文本完成请求
// @Summary 文本补全
// @Description 创建文本补全请求，兼容OpenAI API格式
// @Tags AI接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body clients.CompletionRequest true "文本补全请求"
// @Success 200 {object} clients.AIResponse "文本补全响应"
// @Failure 400 {object} dto.Response "请求参数错误"
// @Failure 401 {object} dto.Response "认证失败"
// @Failure 429 {object} dto.Response "请求过于频繁"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /v1/completions [post]
func (h *AIHandler) Completions(c *gin.Context) {
	// 获取认证信息
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"Authentication required",
			nil,
		))
		return
	}

	apiKeyID, exists := middleware.GetAPIKeyIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"API key required",
			nil,
		))
		return
	}

	// 解析请求体
	var completionRequest clients.CompletionRequest
	if err := c.ShouldBindJSON(&completionRequest); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id":    userID,
			"api_key_id": apiKeyID,
			"error":      err.Error(),
		}).Warn("Invalid request body")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	// 验证必需字段
	if completionRequest.Model == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_MODEL",
			"Model is required",
			nil,
		))
		return
	}

	if completionRequest.Prompt == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_PROMPT",
			"Prompt is required",
			nil,
		))
		return
	}

	// 获取请求ID
	requestID := middleware.GetRequestIDFromContext(c)

	// 转换为通用 AIRequest 结构
	aiRequest := &clients.AIRequest{
		Model:       completionRequest.Model,
		Prompt:      completionRequest.Prompt,
		MaxTokens:   completionRequest.MaxTokens,
		Temperature: completionRequest.Temperature,
		Stream:      completionRequest.Stream,
	}

	// 构造网关请求
	gatewayRequest := &gateway.GatewayRequest{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		ModelSlug: aiRequest.Model,
		Request:   aiRequest,
		RequestID: requestID,
	}

	// 处理请求
	response, err := h.gatewayService.ProcessRequest(c.Request.Context(), gatewayRequest)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
			"api_key_id": apiKeyID,
			"model":      aiRequest.Model,
			"error":      err.Error(),
		}).Error("Failed to process AI request")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"REQUEST_FAILED",
			"Failed to process request",
			map[string]interface{}{
				"request_id": requestID,
			},
		))
		return
	}

	// 输出AI提供商响应结果
	h.logger.WithFields(map[string]interface{}{
		"request_id":    requestID,
		"user_id":       userID,
		"api_key_id":    apiKeyID,
		"model":         aiRequest.Model,
		"provider":      response.Provider,
		"duration_ms":   response.Duration.Milliseconds(),
		"total_tokens":  response.Usage.TotalTokens,
		"input_tokens":  response.Usage.InputTokens,
		"output_tokens": response.Usage.OutputTokens,
		"total_cost":    response.Cost.TotalCost,
		"response_data": response.Response,
	}).Info("AI provider response received successfully")

	// 设置使用量到上下文（用于配额中间件）
	c.Set("tokens_used", response.Usage.TotalTokens)
	c.Set("cost_used", response.Cost.TotalCost)

	// 设置响应头
	c.Header("X-Request-ID", requestID)
	c.Header("X-Provider", response.Provider)
	c.Header("X-Model", response.Model)
	c.Header("X-Duration-Ms", strconv.FormatInt(response.Duration.Milliseconds(), 10))

	// 返回AI响应
	c.JSON(http.StatusOK, response.Response)
}

// Models 获取可用模型列表
// @Summary 列出模型
// @Description 获取可用的AI模型列表
// @Tags AI接口
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} clients.ModelsResponse "模型列表"
// @Failure 401 {object} dto.Response "认证失败"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /v1/models [get]
func (h *AIHandler) Models(c *gin.Context) {
	// TODO: 实现获取模型列表
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   []interface{}{},
	})
}

// Usage 获取使用情况
// @Summary 使用统计
// @Description 获取当前用户的API使用统计
// @Tags AI接口
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.UsageResponse "使用统计信息"
// @Failure 401 {object} dto.Response "认证失败"
// @Failure 500 {object} dto.Response "服务器内部错误"
// @Router /v1/usage [get]
func (h *AIHandler) Usage(c *gin.Context) {
	// 获取认证信息
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"Authentication required",
			nil,
		))
		return
	}

	// TODO: 实现获取使用情况
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"usage":   gin.H{},
	})
}
