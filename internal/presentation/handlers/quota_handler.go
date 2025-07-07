package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/services"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// QuotaHandler 配额管理处理器
type QuotaHandler struct {
	quotaService services.QuotaService
	logger       logger.Logger
}

// NewQuotaHandler 创建配额处理器
func NewQuotaHandler(quotaService services.QuotaService, logger logger.Logger) *QuotaHandler {
	return &QuotaHandler{
		quotaService: quotaService,
		logger:       logger,
	}
}

// CreateQuotaRequest 创建配额请求
type CreateQuotaRequest struct {
	QuotaType  entities.QuotaType    `json:"quota_type" binding:"required"`
	Period     *entities.QuotaPeriod `json:"period,omitempty"`
	LimitValue float64               `json:"limit_value" binding:"required,gt=0"`
	ResetTime  *string               `json:"reset_time,omitempty"`
}

// UpdateQuotaRequest 更新配额请求
type UpdateQuotaRequest struct {
	LimitValue *float64              `json:"limit_value,omitempty"`
	Status     *entities.QuotaStatus `json:"status,omitempty"`
}

// GetAPIKeyQuotas 获取API Key的配额列表
func (h *QuotaHandler) GetAPIKeyQuotas(c *gin.Context) {
	apiKeyIDStr := c.Param("id")
	apiKeyID, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	quotas, err := h.quotaService.GetAPIKeyQuotas(c.Request.Context(), apiKeyID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": apiKeyID,
			"error":      err.Error(),
		}).Error("Failed to get API key quotas")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"QUOTA_FETCH_ERROR",
			"Failed to fetch quotas",
			nil,
		))
		return
	}

	// 获取每个配额的使用情况
	var quotaResponses []map[string]interface{}
	for _, quota := range quotas {
		usageInfo, err := h.quotaService.GetQuotaUsage(c.Request.Context(), apiKeyID, quota.QuotaType, quota.Period)
		if err != nil {
			h.logger.WithFields(map[string]interface{}{
				"quota_id": quota.ID,
				"error":    err.Error(),
			}).Warn("Failed to get quota usage")
		}

		quotaResponse := map[string]interface{}{
			"id":          quota.ID,
			"quota_type":  quota.QuotaType,
			"period":      quota.Period,
			"limit_value": quota.LimitValue,
			"reset_time":  quota.ResetTime,
			"status":      quota.Status,
			"created_at":  quota.CreatedAt,
			"updated_at":  quota.UpdatedAt,
		}

		if usageInfo != nil {
			quotaResponse["used_value"] = usageInfo.Used
			quotaResponse["remaining"] = usageInfo.Remaining
			quotaResponse["percentage"] = usageInfo.Percentage
		}

		quotaResponses = append(quotaResponses, quotaResponse)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(quotaResponses, "API Key quotas retrieved successfully"))
}

// CreateAPIKeyQuota 为API Key创建配额
func (h *QuotaHandler) CreateAPIKeyQuota(c *gin.Context) {
	apiKeyIDStr := c.Param("id")
	apiKeyID, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	var req CreateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request data",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	quota, err := h.quotaService.CreateQuota(c.Request.Context(), apiKeyID, req.QuotaType, req.Period, req.LimitValue)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": apiKeyID,
			"quota_type": req.QuotaType,
			"period":     req.Period,
			"error":      err.Error(),
		}).Error("Failed to create quota")

		// 检查是否是重复配额错误
		if strings.Contains(err.Error(), "quota already exists") {
			c.JSON(http.StatusConflict, dto.ErrorResponse(
				"QUOTA_ALREADY_EXISTS",
				"Quota already exists for this API key with the same type and period",
				nil,
			))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"QUOTA_CREATE_ERROR",
			"Failed to create quota",
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(quota, "Quota created successfully"))
}

// UpdateQuota 更新配额
func (h *QuotaHandler) UpdateQuota(c *gin.Context) {
	quotaIDStr := c.Param("quota_id")
	quotaID, err := strconv.ParseInt(quotaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_QUOTA_ID",
			"Invalid quota ID",
			nil,
		))
		return
	}

	var req UpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request data",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	if req.LimitValue != nil {
		err = h.quotaService.UpdateQuota(c.Request.Context(), quotaID, *req.LimitValue)
		if err != nil {
			h.logger.WithFields(map[string]interface{}{
				"quota_id": quotaID,
				"error":    err.Error(),
			}).Error("Failed to update quota")

			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
				"QUOTA_UPDATE_ERROR",
				"Failed to update quota",
				nil,
			))
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(map[string]interface{}{
		"message": "Quota updated successfully",
	}, "Quota updated successfully"))
}

// DeleteQuota 删除配额
func (h *QuotaHandler) DeleteQuota(c *gin.Context) {
	quotaIDStr := c.Param("quota_id")
	quotaID, err := strconv.ParseInt(quotaIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_QUOTA_ID",
			"Invalid quota ID",
			nil,
		))
		return
	}

	err = h.quotaService.DeleteQuota(c.Request.Context(), quotaID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"quota_id": quotaID,
			"error":    err.Error(),
		}).Error("Failed to delete quota")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"QUOTA_DELETE_ERROR",
			"Failed to delete quota",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(map[string]interface{}{
		"message": "Quota deleted successfully",
	}, "Quota deleted successfully"))
}

// GetQuotaStatus 获取API Key的配额状态
func (h *QuotaHandler) GetQuotaStatus(c *gin.Context) {
	apiKeyIDStr := c.Param("id")
	apiKeyID, err := strconv.ParseInt(apiKeyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_API_KEY_ID",
			"Invalid API key ID",
			nil,
		))
		return
	}

	status, err := h.quotaService.GetQuotaStatus(c.Request.Context(), apiKeyID)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"api_key_id": apiKeyID,
			"error":      err.Error(),
		}).Error("Failed to get quota status")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"QUOTA_STATUS_ERROR",
			"Failed to get quota status",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(status, "Quota status retrieved successfully"))
}
