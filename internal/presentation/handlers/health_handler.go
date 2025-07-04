package handlers

import (
	"net/http"
	"time"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/infrastructure/gateway"
	"ai-api-gateway/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	gatewayService gateway.GatewayService
	logger         logger.Logger
	startTime      time.Time
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(gatewayService gateway.GatewayService, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		gatewayService: gatewayService,
		logger:         logger,
		startTime:      time.Now(),
	}
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务整体健康状态，包括数据库和AI提供商连接状态
// @Tags 健康检查
// @Produce json
// @Success 200 {object} dto.Response "健康检查通过"
// @Failure 503 {object} dto.Response "健康检查失败"
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	result, err := h.gatewayService.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Health check failed")
		c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse(
			"HEALTH_CHECK_FAILED",
			"Health check failed",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Health check passed"))
}

// ReadinessCheck 就绪检查
// @Summary 就绪检查
// @Description 检查服务是否已准备好接收请求
// @Tags 健康检查
// @Produce json
// @Success 200 {object} dto.Response "服务就绪"
// @Failure 503 {object} dto.Response "服务未就绪"
// @Router /health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// 简单的就绪检查
	response := &dto.HealthCheckResponse{
		Status:    "ready",
		Version:   "1.0.0", // TODO: 从配置或构建信息中获取
		Timestamp: time.Now(),
		Services: map[string]string{
			"gateway":  "ready",
			"database": "ready",
		},
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Service is ready"))
}

// LivenessCheck 存活检查
// @Summary 存活检查
// @Description 检查服务是否正在运行
// @Tags 健康检查
// @Produce json
// @Success 200 {object} dto.Response "服务存活"
// @Router /health/live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// 简单的存活检查
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(h.startTime).String(),
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(response, "Service is alive"))
}

// GetStats 获取统计信息
// @Summary 系统统计
// @Description 获取系统运行统计信息
// @Tags 健康检查
// @Produce json
// @Success 200 {object} dto.Response "统计信息"
// @Failure 500 {object} dto.Response "获取统计失败"
// @Router /health/stats [get]
func (h *HealthHandler) GetStats(c *gin.Context) {
	stats, err := h.gatewayService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.WithField("error", err.Error()).Error("Failed to get stats")
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
			"GET_STATS_FAILED",
			"Failed to get statistics",
			map[string]interface{}{
				"details": err.Error(),
			},
		))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, "Statistics retrieved successfully"))
}

// GetMetrics 获取监控指标（Prometheus格式）
// @Summary 监控指标
// @Description 获取Prometheus格式的监控指标
// @Tags 监控
// @Produce text/plain
// @Success 200 {string} string "Prometheus指标"
// @Router /metrics [get]
func (h *HealthHandler) GetMetrics(c *gin.Context) {
	// TODO: 实现Prometheus指标输出
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, `# HELP ai_gateway_requests_total Total number of requests
# TYPE ai_gateway_requests_total counter
ai_gateway_requests_total 0

# HELP ai_gateway_request_duration_seconds Request duration in seconds
# TYPE ai_gateway_request_duration_seconds histogram
ai_gateway_request_duration_seconds_bucket{le="0.1"} 0
ai_gateway_request_duration_seconds_bucket{le="0.5"} 0
ai_gateway_request_duration_seconds_bucket{le="1.0"} 0
ai_gateway_request_duration_seconds_bucket{le="2.5"} 0
ai_gateway_request_duration_seconds_bucket{le="5.0"} 0
ai_gateway_request_duration_seconds_bucket{le="10.0"} 0
ai_gateway_request_duration_seconds_bucket{le="+Inf"} 0
ai_gateway_request_duration_seconds_sum 0
ai_gateway_request_duration_seconds_count 0

# HELP ai_gateway_uptime_seconds Uptime in seconds
# TYPE ai_gateway_uptime_seconds gauge
ai_gateway_uptime_seconds %f
`, time.Since(h.startTime).Seconds())
}

// GetVersion 获取版本信息
// @Summary 版本信息
// @Description 获取服务版本信息
// @Tags 健康检查
// @Produce json
// @Success 200 {object} dto.Response "版本信息"
// @Router /health/version [get]
func (h *HealthHandler) GetVersion(c *gin.Context) {
	version := map[string]interface{}{
		"version":    "1.0.0",   // TODO: 从构建信息中获取
		"build_time": "unknown", // TODO: 从构建信息中获取
		"git_commit": "unknown", // TODO: 从构建信息中获取
		"go_version": "unknown", // TODO: 从运行时信息中获取
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(version, "Version information"))
}
