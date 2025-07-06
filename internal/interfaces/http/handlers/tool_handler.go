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

// GetToolTypes 获取工具类型列表
func (h *ToolHandler) GetToolTypes(c *gin.Context) {
	toolTypes, err := h.toolService.GetToolTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get tool types",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    toolTypes,
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
	tools, err := h.toolService.GetUserTools(c.Request.Context(), 0, "public")
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

// GetUserTools 获取用户工具列表
func (h *ToolHandler) GetUserTools(c *gin.Context) {
	userID := c.GetInt64("user_id")
	category := c.DefaultQuery("category", "all")

	tools, err := h.toolService.GetUserTools(c.Request.Context(), userID, category)
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

// CreateUserTool 创建用户工具
func (h *ToolHandler) CreateUserTool(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req entities.CreateUserToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	tool, err := h.toolService.CreateUserTool(c.Request.Context(), userID, &req)
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

// GetUserTool 获取用户工具详情
func (h *ToolHandler) GetUserTool(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	tool, err := h.toolService.GetUserToolByID(c.Request.Context(), toolID, userID)
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

// UpdateUserTool 更新用户工具
func (h *ToolHandler) UpdateUserTool(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	var req entities.UpdateUserToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	tool, err := h.toolService.UpdateUserTool(c.Request.Context(), toolID, userID, &req)
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

// DeleteUserTool 删除用户工具
func (h *ToolHandler) DeleteUserTool(c *gin.Context) {
	userID := c.GetInt64("user_id")
	toolID := c.Param("id")

	err := h.toolService.DeleteUserTool(c.Request.Context(), toolID, userID)
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

// GetSharedTool 获取分享的工具
func (h *ToolHandler) GetSharedTool(c *gin.Context) {
	shareToken := c.Param("token")

	tool, err := h.toolService.GetSharedTool(c.Request.Context(), shareToken)
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
