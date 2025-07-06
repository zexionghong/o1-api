package handlers

import (
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ToolHandler 工具处理器
type ToolHandler struct {
	toolService *services.ToolService
	logger      logger.Logger
}

// NewToolHandler 创建工具处理器
func NewToolHandler(toolService *services.ToolService, logger logger.Logger) *ToolHandler {
	return &ToolHandler{
		toolService: toolService,
		logger:      logger,
	}
}

// GetTools 获取工具模板列表
func (h *ToolHandler) GetTools(c *gin.Context) {
	tools, err := h.toolService.GetTools(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tools",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
	})
}

// GetPublicTools 获取公开工具列表
func (h *ToolHandler) GetPublicTools(c *gin.Context) {
	// 解析分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 这里需要修改service方法来支持分页
	tools, err := h.toolService.GetUserToolInstances(c.Request.Context(), 0, "public")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get public tools",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
	})
}

// GetUserToolInstances 获取用户工具实例列表
func (h *ToolHandler) GetUserToolInstances(c *gin.Context) {
	userID := c.GetInt64("user_id")
	category := c.DefaultQuery("category", "all")

	tools, err := h.toolService.GetUserToolInstances(c.Request.Context(), userID, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get user tools",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
	})
}

// CreateUserToolInstance 创建用户工具实例
func (h *ToolHandler) CreateUserToolInstance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req entities.CreateUserToolInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	tool, err := h.toolService.CreateUserToolInstance(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to create tool",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    tool,
		"message": "Tool created successfully",
	})
}

// GetUserToolInstance 获取用户工具实例详情
func (h *ToolHandler) GetUserToolInstance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	tool, err := h.toolService.GetUserToolInstanceByID(c.Request.Context(), toolID, userID)
	if err != nil {
		if err.Error() == "tool not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tool not found",
			})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tool",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
	})
}

// UpdateUserToolInstance 更新用户工具实例
func (h *ToolHandler) UpdateUserToolInstance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	var req entities.UpdateUserToolInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	tool, err := h.toolService.UpdateUserToolInstance(c.Request.Context(), toolID, userID, &req)
	if err != nil {
		if err.Error() == "tool not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tool not found",
			})
			return
		}
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Access denied",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Failed to update tool",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
		"message": "Tool updated successfully",
	})
}

// DeleteUserToolInstance 删除用户工具实例
func (h *ToolHandler) DeleteUserToolInstance(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	err := h.toolService.DeleteUserToolInstance(c.Request.Context(), toolID, userID)
	if err != nil {
		if err.Error() == "tool not found or not owned by user" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tool not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete tool",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tool deleted successfully",
	})
}

// GetSharedToolInstance 获取分享的工具实例
func (h *ToolHandler) GetSharedToolInstance(c *gin.Context) {
	shareToken := c.Param("token")

	tool, err := h.toolService.GetSharedToolInstance(c.Request.Context(), shareToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get shared tool",
			"error":   err.Error(),
		})
		return
	}

	if tool == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Shared tool not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tool,
	})
}

// IncrementUsage 增加工具使用次数
func (h *ToolHandler) IncrementUsage(c *gin.Context) {
	toolID := c.Param("id")

	err := h.toolService.IncrementUsageCount(c.Request.Context(), toolID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to increment usage count",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Usage count incremented",
	})
}

// GetModels 获取可用模型列表
func (h *ToolHandler) GetModels(c *gin.Context) {
	// 获取聊天类型的活跃模型
	models, err := h.toolService.GetAvailableModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get models",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    models,
	})
}

// GetUserAPIKeys 获取用户API密钥列表
func (h *ToolHandler) GetUserAPIKeys(c *gin.Context) {
	userID := c.GetInt64("user_id")

	apiKeys, err := h.toolService.GetUserAPIKeys(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get API keys",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiKeys,
	})
}
