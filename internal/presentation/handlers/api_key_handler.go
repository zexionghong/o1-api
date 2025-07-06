package handlers

import (
	"net/http"
	"strconv"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// APIKeyHandler API密钥处理器
type APIKeyHandler struct {
	apiKeyService     services.APIKeyService
	usageLogRepo      repositories.UsageLogRepository
	billingRecordRepo repositories.BillingRecordRepository
	modelRepo         repositories.ModelRepository
	logger            logger.Logger
}

// NewAPIKeyHandler 创建API密钥处理器
func NewAPIKeyHandler(apiKeyService services.APIKeyService, usageLogRepo repositories.UsageLogRepository, billingRecordRepo repositories.BillingRecordRepository, modelRepo repositories.ModelRepository, logger logger.Logger) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService:     apiKeyService,
		usageLogRepo:      usageLogRepo,
		billingRecordRepo: billingRecordRepo,
		modelRepo:         modelRepo,
		logger:            logger,
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

// GetAPIKeyUsageLogs 获取API密钥使用日志
func (h *APIKeyHandler) GetAPIKeyUsageLogs(c *gin.Context) {
	idStr := c.Param("id")
	apiKeyID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	// 绑定查询参数
	var req dto.UsageLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_QUERY_PARAMS",
			"Invalid query parameters",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	req.APIKeyID = apiKeyID

	// 验证参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	h.logger.WithFields(map[string]interface{}{
		"api_key_id": apiKeyID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"page":       req.Page,
		"page_size":  req.PageSize,
	}).Debug("Querying usage logs with date range")

	// 从数据库获取使用日志
	offset := (req.Page - 1) * req.PageSize
	usageLogEntities, err := h.usageLogRepo.GetByAPIKeyIDAndDateRange(c.Request.Context(), apiKeyID, req.StartDate, req.EndDate, offset, req.PageSize)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"api_key_id": apiKeyID,
		}).Error("Failed to get usage logs from database")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DATABASE_ERROR",
			"Failed to retrieve usage logs",
			nil,
		))
		return
	}

	// 获取总数
	total, err := h.usageLogRepo.CountByAPIKeyIDAndDateRange(c.Request.Context(), apiKeyID, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"api_key_id": apiKeyID,
		}).Error("Failed to count usage logs from database")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DATABASE_ERROR",
			"Failed to count usage logs",
			nil,
		))
		return
	}

	// 转换为响应DTO
	usageLogs := make([]dto.UsageLogResponse, len(usageLogEntities))
	for i, entity := range usageLogEntities {
		// 获取模型名称
		modelName := "unknown"
		if model, err := h.modelRepo.GetByID(c.Request.Context(), entity.ModelID); err == nil {
			modelName = model.GetDisplayName()
		}

		usageLogs[i] = dto.UsageLogResponse{
			ID:          entity.ID,
			APIKeyID:    entity.APIKeyID,
			UserID:      entity.UserID,
			Model:       modelName,
			TokensUsed:  entity.TotalTokens,
			Cost:        entity.Cost,
			RequestType: entity.Endpoint, // 暂时使用endpoint作为request_type
			Status:      getStatusFromCode(entity.StatusCode),
			RequestID:   entity.RequestID,
			IPAddress:   "", // TODO: 如果需要可以添加到entity
			UserAgent:   "", // TODO: 如果需要可以添加到entity
			Timestamp:   entity.CreatedAt,
		}
	}

	response := dto.PaginatedResponse{
		Data:       usageLogs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Usage logs retrieved successfully"))
}

// getStatusFromCode 根据HTTP状态码获取状态字符串
func getStatusFromCode(statusCode int) string {
	if statusCode >= 200 && statusCode < 300 {
		return "success"
	}
	return "error"
}

// getStringValue 获取字符串指针的值，如果为nil则返回空字符串
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetAPIKeyBillingRecords 获取API密钥扣费记录
func (h *APIKeyHandler) GetAPIKeyBillingRecords(c *gin.Context) {
	idStr := c.Param("id")
	apiKeyID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	// 绑定查询参数
	var req dto.BillingRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_QUERY_PARAMS",
			"Invalid query parameters",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	req.APIKeyID = apiKeyID

	// 验证参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	h.logger.WithFields(map[string]interface{}{
		"api_key_id": apiKeyID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"page":       req.Page,
		"page_size":  req.PageSize,
	}).Debug("Querying billing records with date range")

	// 从数据库获取扣费记录
	offset := (req.Page - 1) * req.PageSize
	billingRecordEntities, err := h.billingRecordRepo.GetByAPIKeyIDAndDateRange(c.Request.Context(), apiKeyID, req.StartDate, req.EndDate, offset, req.PageSize)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"api_key_id": apiKeyID,
		}).Error("Failed to get billing records from database")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DATABASE_ERROR",
			"Failed to retrieve billing records",
			nil,
		))
		return
	}

	// 获取总数
	total, err := h.billingRecordRepo.CountByAPIKeyIDAndDateRange(c.Request.Context(), apiKeyID, req.StartDate, req.EndDate)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error":      err.Error(),
			"api_key_id": apiKeyID,
		}).Error("Failed to count billing records from database")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DATABASE_ERROR",
			"Failed to count billing records",
			nil,
		))
		return
	}

	// 转换为响应DTO
	billingRecords := make([]dto.BillingRecordResponse, len(billingRecordEntities))
	for i, entity := range billingRecordEntities {
		billingRecords[i] = dto.BillingRecordResponse{
			ID:              entity.ID,
			UserID:          entity.UserID,
			Amount:          entity.Amount,
			Description:     getStringValue(entity.Description),
			TransactionType: string(entity.BillingType),
			BalanceBefore:   0.0, // TODO: 需要计算或存储余额变化
			BalanceAfter:    0.0, // TODO: 需要计算或存储余额变化
			Timestamp:       entity.CreatedAt,
		}
	}

	response := dto.PaginatedResponse{
		Data:       billingRecords,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int((total + int64(req.PageSize) - 1) / int64(req.PageSize)),
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Billing records retrieved successfully"))
}
