package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/gateway"
	"ai-api-gateway/internal/infrastructure/logger"
	"ai-api-gateway/internal/presentation/middleware"
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

// ChatCompletions 处理聊天完成请求
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
	var aiRequest clients.AIRequest
	if err := c.ShouldBindJSON(&aiRequest); err != nil {
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
	if aiRequest.Model == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_MODEL",
			"Model is required",
			nil,
		))
		return
	}
	
	if len(aiRequest.Messages) == 0 && aiRequest.Prompt == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_INPUT",
			"Either messages or prompt is required",
			nil,
		))
		return
	}
	
	// 获取请求ID
	requestID := middleware.GetRequestIDFromContext(c)
	
	// 构造网关请求
	gatewayRequest := &gateway.GatewayRequest{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		ModelSlug: aiRequest.Model,
		Request:   &aiRequest,
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
	var aiRequest clients.AIRequest
	if err := c.ShouldBindJSON(&aiRequest); err != nil {
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
	if aiRequest.Model == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_MODEL",
			"Model is required",
			nil,
		))
		return
	}
	
	if aiRequest.Prompt == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"MISSING_PROMPT",
			"Prompt is required",
			nil,
		))
		return
	}
	
	// 获取请求ID
	requestID := middleware.GetRequestIDFromContext(c)
	
	// 构造网关请求
	gatewayRequest := &gateway.GatewayRequest{
		UserID:    userID,
		APIKeyID:  apiKeyID,
		ModelSlug: aiRequest.Model,
		Request:   &aiRequest,
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
func (h *AIHandler) Models(c *gin.Context) {
	// TODO: 实现获取模型列表
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   []interface{}{},
	})
}

// Usage 获取使用情况
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
