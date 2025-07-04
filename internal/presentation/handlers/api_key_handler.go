package handlers

import (
	"net/http"
	"strconv"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler API密钥处理器
type APIKeyHandler struct {
	apiKeyService services.APIKeyService
	logger        logger.Logger
}

// NewAPIKeyHandler 创建API密钥处理器
func NewAPIKeyHandler(apiKeyService services.APIKeyService, logger logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: apiKeyService,
		logger:        logger,
	}
}

// CreateAPIKey 创建API密钥
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req dto.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid create API key request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": req.UserID,
			"name":    req.Name,
			"error":   err.Error(),
		}).Error("Failed to create API key")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"CREATE_API_KEY_FAILED",
			"Failed to create API key",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(apiKey, "API key created successfully"))
}

// GetAPIKey 获取API密钥
func (h *APIKeyHandler) GetAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": id,
			"error":      err.Error(),
		}).Error("Failed to get API key")

		c.JSON(http.StatusNotFound, dto.ErrorResponse(
			"API_KEY_NOT_FOUND",
			"API key not found",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(apiKey, "API key retrieved successfully"))
}

// UpdateAPIKey 更新API密钥
func (h *APIKeyHandler) UpdateAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	var req dto.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid update API key request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	apiKey, err := h.apiKeyService.UpdateAPIKey(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": id,
			"error":      err.Error(),
		}).Error("Failed to update API key")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"UPDATE_API_KEY_FAILED",
			"Failed to update API key",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(apiKey, "API key updated successfully"))
}

// DeleteAPIKey 删除API密钥
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	err = h.apiKeyService.DeleteAPIKey(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": id,
			"error":      err.Error(),
		}).Error("Failed to delete API key")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DELETE_API_KEY_FAILED",
			"Failed to delete API key",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "API key deleted successfully"))
}

// RevokeAPIKey 撤销API密钥
func (h *APIKeyHandler) RevokeAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	err = h.apiKeyService.RevokeAPIKey(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": id,
			"error":      err.Error(),
		}).Error("Failed to revoke API key")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"REVOKE_API_KEY_FAILED",
			"Failed to revoke API key",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "API key revoked successfully"))
}

// ListAPIKeys 获取API密钥列表
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	// 解析分页参数
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid pagination parameters")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_PAGINATION",
			"Invalid pagination parameters",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	pagination.SetDefaults()

	apiKeys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), &pagination)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to list API keys")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"LIST_API_KEYS_FAILED",
			"Failed to list API keys",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(apiKeys, "API keys retrieved successfully"))
}

// GetUserAPIKeys 获取用户的API密钥列表
func (h *APIKeyHandler) GetUserAPIKeys(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_USER_ID",
			"Invalid user ID",
			nil,
		))
		return
	}

	apiKeys, err := h.apiKeyService.GetUserAPIKeys(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user API keys")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"GET_USER_API_KEYS_FAILED",
			"Failed to get user API keys",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(apiKeys, "User API keys retrieved successfully"))
}
