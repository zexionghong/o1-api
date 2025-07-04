package handlers

import (
	"net/http"
	"strconv"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService services.UserService
	logger      logger.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService services.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid create user request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
			"error":    err.Error(),
		}).Error("Failed to create user")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"CREATE_USER_FAILED",
			"Failed to create user",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(user, "User created successfully"))
}

// GetUser 获取用户
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_USER_ID",
			"Invalid user ID",
			nil,
		))
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}).Error("Failed to get user")

		c.JSON(http.StatusNotFound, dto.ErrorResponse(
			"USER_NOT_FOUND",
			"User not found",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user, "User retrieved successfully"))
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_USER_ID",
			"Invalid user ID",
			nil,
		))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid update user request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}).Error("Failed to update user")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"UPDATE_USER_FAILED",
			"Failed to update user",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user, "User updated successfully"))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_USER_ID",
			"Invalid user ID",
			nil,
		))
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": id,
			"error":   err.Error(),
		}).Error("Failed to delete user")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"DELETE_USER_FAILED",
			"Failed to delete user",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User deleted successfully"))
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
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

	users, err := h.userService.ListUsers(c.Request.Context(), &pagination)
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to list users")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"LIST_USERS_FAILED",
			"Failed to list users",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(users, "Users retrieved successfully"))
}

// UpdateBalance 更新用户余额
func (h *UserHandler) UpdateBalance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_USER_ID",
			"Invalid user ID",
			nil,
		))
		return
	}

	var req dto.BalanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithField("error", err.Error()).Warn("Invalid balance update request")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	user, err := h.userService.UpdateBalance(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id":   id,
			"operation": req.Operation,
			"amount":    req.Amount,
			"error":     err.Error(),
		}).Error("Failed to update user balance")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"UPDATE_BALANCE_FAILED",
			"Failed to update balance",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user, "Balance updated successfully"))
}
