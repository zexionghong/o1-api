package handlers

import (
	"net/http"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService services.AuthService
	logger      logger.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService services.AuthService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码进行登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} dto.LoginResponse "登录成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "认证失败"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Invalid login request")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request format",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	// 调用认证服务进行登录
	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("Login failed")

		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"LOGIN_FAILED",
			"Invalid username or password",
			nil,
		))
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"username": req.Username,
		"user_id":  response.User.ID,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Login successful"))
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册请求"
// @Success 201 {object} dto.RegisterResponse "注册成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 409 {object} dto.ErrorResponse "用户名或邮箱已存在"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Invalid register request")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request format",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	// 调用认证服务进行注册
	response, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"username": req.Username,
			"email":    req.Email,
			"error":    err.Error(),
		}).Warn("Registration failed")

		// 根据错误类型返回不同的状态码
		statusCode := http.StatusInternalServerError
		errorCode := "REGISTRATION_FAILED"
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			statusCode = http.StatusConflict
			errorCode = "USER_EXISTS"
		}

		c.JSON(statusCode, dto.ErrorResponse(
			errorCode,
			err.Error(),
			nil,
		))
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"username": req.Username,
		"user_id":  response.ID,
	}).Info("User registered successfully")

	c.JSON(http.StatusCreated, dto.SuccessResponse(response, "User registered successfully"))
}

// RefreshToken 刷新令牌
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} dto.RefreshTokenResponse "刷新成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "刷新令牌无效"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Invalid refresh token request")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request format",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	// 调用认证服务刷新令牌
	response, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Token refresh failed")

		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"TOKEN_REFRESH_FAILED",
			"Invalid refresh token",
			nil,
		))
		return
	}

	h.logger.Debug("Token refreshed successfully")
	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Token refreshed successfully"))
}

// GetProfile 获取用户资料
// @Summary 获取当前用户资料
// @Description 获取当前登录用户的详细信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.GetUserProfileResponse "获取成功"
// @Failure 401 {object} dto.ErrorResponse "未认证"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"User authentication required",
			nil,
		))
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"INTERNAL_ERROR",
			"Invalid user ID format",
			nil,
		))
		return
	}

	// 获取用户资料
	response, err := h.authService.GetUserProfile(c.Request.Context(), userIDInt64)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": userIDInt64,
			"error":   err.Error(),
		}).Error("Failed to get user profile")

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"PROFILE_FETCH_FAILED",
			"Failed to get user profile",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "User profile retrieved successfully"))
}

// ChangePassword 修改密码
// @Summary 修改用户密码
// @Description 修改当前用户的密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} dto.SuccessResponse "修改成功"
// @Failure 400 {object} dto.ErrorResponse "请求参数错误"
// @Failure 401 {object} dto.ErrorResponse "未认证或旧密码错误"
// @Failure 500 {object} dto.ErrorResponse "服务器内部错误"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
			"AUTHENTICATION_REQUIRED",
			"User authentication required",
			nil,
		))
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"INTERNAL_ERROR",
			"Invalid user ID format",
			nil,
		))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Invalid change password request")

		c.JSON(http.StatusBadRequest, dto.ErrorResponse(
			"INVALID_REQUEST",
			"Invalid request format",
			map[string]interface{}{"details": err.Error()},
		))
		return
	}

	// 调用认证服务修改密码
	err := h.authService.ChangePassword(c.Request.Context(), userIDInt64, &req)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"user_id": userIDInt64,
			"error":   err.Error(),
		}).Warn("Password change failed")

		statusCode := http.StatusInternalServerError
		if err.Error() == "old password is incorrect" {
			statusCode = http.StatusUnauthorized
		}

		c.JSON(statusCode, dto.ErrorResponse(
			"PASSWORD_CHANGE_FAILED",
			err.Error(),
			nil,
		))
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id": userIDInt64,
	}).Info("Password changed successfully")

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Password changed successfully"))
}
