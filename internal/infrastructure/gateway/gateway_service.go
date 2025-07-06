package gateway

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/domain/values"
	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/logger"
)

// GatewayService API网关服务接口
type GatewayService interface {
	// ProcessRequest 处理AI请求
	ProcessRequest(ctx context.Context, request *GatewayRequest) (*GatewayResponse, error)

	// ProcessStreamRequest 处理流式AI请求
	ProcessStreamRequest(ctx context.Context, request *GatewayRequest, streamChan chan<- *StreamChunk) error

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) (*HealthCheckResult, error)

	// GetStats 获取统计信息
	GetStats(ctx context.Context) (*GatewayStats, error)
}

// StreamChunk 流式响应块
type StreamChunk struct {
	ID           string           `json:"id"`
	Object       string           `json:"object"`
	Created      int64            `json:"created"`
	Model        string           `json:"model"`
	Content      string           `json:"content"`
	FinishReason *string          `json:"finish_reason"`
	Usage        *clients.AIUsage `json:"usage,omitempty"`
	Cost         *CostInfo        `json:"cost,omitempty"`
}

// GatewayRequest 网关请求
type GatewayRequest struct {
	UserID    int64              `json:"user_id"`
	APIKeyID  int64              `json:"api_key_id"`
	ModelSlug string             `json:"model_slug"`
	Request   *clients.AIRequest `json:"request"`
	RequestID string             `json:"request_id"`
}

// GatewayResponse 网关响应
type GatewayResponse struct {
	Response  *clients.AIResponse `json:"response"`
	Usage     *UsageInfo          `json:"usage"`
	Cost      *CostInfo           `json:"cost"`
	Provider  string              `json:"provider"`
	Model     string              `json:"model"`
	Duration  time.Duration       `json:"duration"`
	RequestID string              `json:"request_id"`
}

// UsageInfo 使用信息
type UsageInfo struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// CostInfo 成本信息
type CostInfo struct {
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	TotalCost  float64 `json:"total_cost"`
	Currency   string  `json:"currency"`
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    string                    `json:"status"`
	Timestamp time.Time                 `json:"timestamp"`
	Providers map[string]ProviderHealth `json:"providers"`
	Database  DatabaseHealth            `json:"database"`
}

// ProviderHealth 提供商健康状态
type ProviderHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	Error        string        `json:"error,omitempty"`
}

// DatabaseHealth 数据库健康状态
type DatabaseHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// GatewayStats 网关统计信息
type GatewayStats struct {
	TotalRequests      int64                    `json:"total_requests"`
	SuccessfulRequests int64                    `json:"successful_requests"`
	FailedRequests     int64                    `json:"failed_requests"`
	SuccessRate        float64                  `json:"success_rate"`
	AvgResponseTime    time.Duration            `json:"avg_response_time"`
	ProvidersStats     map[int64]*ProviderStats `json:"providers_stats"`
	Uptime             time.Duration            `json:"uptime"`
}

// gatewayServiceImpl 网关服务实现
type gatewayServiceImpl struct {
	router          RequestRouter
	userService     services.UserService
	apiKeyService   services.APIKeyService
	quotaService    services.QuotaService
	billingService  services.BillingService
	usageLogService services.UsageLogService
	billingRepo     repositories.BillingRecordRepository
	logger          logger.Logger
	startTime       time.Time
	requestIDGen    *values.RequestIDGenerator
}

// NewGatewayService 创建网关服务
func NewGatewayService(
	router RequestRouter,
	userService services.UserService,
	apiKeyService services.APIKeyService,
	quotaService services.QuotaService,
	billingService services.BillingService,
	usageLogService services.UsageLogService,
	billingRepo repositories.BillingRecordRepository,
	logger logger.Logger,
) GatewayService {
	return &gatewayServiceImpl{
		router:          router,
		userService:     userService,
		apiKeyService:   apiKeyService,
		quotaService:    quotaService,
		billingService:  billingService,
		usageLogService: usageLogService,
		billingRepo:     billingRepo,
		logger:          logger,
		startTime:       time.Now(),
		requestIDGen:    values.NewRequestIDGenerator(),
	}
}

// ProcessRequest 处理AI请求
func (g *gatewayServiceImpl) ProcessRequest(ctx context.Context, request *GatewayRequest) (*GatewayResponse, error) {
	start := time.Now()

	// 生成请求ID（如果没有提供）
	if request.RequestID == "" {
		var err error
		request.RequestID, err = g.requestIDGen.Generate()
		if err != nil {
			g.logger.WithField("error", err.Error()).Error("Failed to generate request ID")
			request.RequestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
	}

	g.logger.WithFields(map[string]interface{}{
		"request_id": request.RequestID,
		"user_id":    request.UserID,
		"api_key_id": request.APIKeyID,
		"model_slug": request.ModelSlug,
	}).Info("Processing AI request")

	// 路由请求
	routeRequest := &RouteRequest{
		UserID:    request.UserID,
		APIKeyID:  request.APIKeyID,
		ModelSlug: request.ModelSlug,
		Request:   request.Request,
		RequestID: request.RequestID,
	}
	g.logger.WithFields(map[string]interface{}{
		"request_id": routeRequest.RequestID,
		"user_id":    routeRequest.UserID,
		"api_key_id": routeRequest.Request,
		"model_slug": routeRequest.ModelSlug,
	}).Info("Routing AI request")
	routeResponse, err := g.router.RouteRequest(ctx, routeRequest)
	if err != nil {
		// 记录失败的使用日志
		g.recordUsageLog(ctx, request, nil, nil, nil, nil, time.Since(start), err)
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	// 计算使用量和成本
	usage := &UsageInfo{
		InputTokens:  routeResponse.Response.Usage.PromptTokens,
		OutputTokens: routeResponse.Response.Usage.CompletionTokens,
		TotalTokens:  routeResponse.Response.Usage.TotalTokens,
	}

	// 计算成本
	cost, err := g.calculateCost(ctx, routeResponse.Model.ID, usage.InputTokens, usage.OutputTokens)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"request_id": request.RequestID,
			"model_id":   routeResponse.Model.ID,
			"error":      err.Error(),
		}).Warn("Failed to calculate cost")
		cost = &CostInfo{Currency: "USD"} // 默认值
	}

	// 记录使用日志
	usageLog := g.recordUsageLog(ctx, request, routeResponse.Provider, routeResponse.Model, usage, cost, routeResponse.Duration, nil)

	// 注意：配额消费已在中间件中处理，这里不再重复消费

	// 处理计费
	if usageLog != nil {
		g.processBilling(ctx, request.UserID, usageLog.ID, cost.TotalCost)
	}

	response := &GatewayResponse{
		Response:  routeResponse.Response,
		Usage:     usage,
		Cost:      cost,
		Provider:  routeResponse.Provider.Name,
		Model:     routeResponse.Model.Name,
		Duration:  routeResponse.Duration,
		RequestID: request.RequestID,
	}

	g.logger.WithFields(map[string]interface{}{
		"request_id":   request.RequestID,
		"provider":     response.Provider,
		"model":        response.Model,
		"total_tokens": usage.TotalTokens,
		"total_cost":   cost.TotalCost,
		"duration":     response.Duration,
	}).Info("Request processed successfully")

	return response, nil
}

// calculateCost 计算成本
func (g *gatewayServiceImpl) calculateCost(ctx context.Context, modelID int64, inputTokens, outputTokens int) (*CostInfo, error) {
	totalCost, err := g.billingService.CalculateCost(ctx, modelID, inputTokens, outputTokens)
	if err != nil {
		return nil, err
	}

	// 简化处理，实际应该根据定价类型分别计算
	inputCost := totalCost * 0.6  // 假设输入占60%
	outputCost := totalCost * 0.4 // 假设输出占40%

	return &CostInfo{
		InputCost:  inputCost,
		OutputCost: outputCost,
		TotalCost:  totalCost,
		Currency:   "USD",
	}, nil
}

// recordUsageLog 记录使用日志
func (g *gatewayServiceImpl) recordUsageLog(ctx context.Context, request *GatewayRequest, provider *entities.Provider, model *entities.Model, usage *UsageInfo, cost *CostInfo, duration time.Duration, requestError error) *entities.UsageLog {
	usageLog := &entities.UsageLog{
		UserID:     request.UserID,
		APIKeyID:   request.APIKeyID,
		RequestID:  request.RequestID,
		Method:     "POST",
		Endpoint:   "/v1/chat/completions",
		DurationMs: int(duration.Milliseconds()),
		CreatedAt:  time.Now(),
	}

	if provider != nil {
		usageLog.ProviderID = provider.ID
	}

	if model != nil {
		usageLog.ModelID = model.ID
	}

	if usage != nil {
		usageLog.InputTokens = usage.InputTokens
		usageLog.OutputTokens = usage.OutputTokens
		usageLog.TotalTokens = usage.TotalTokens
	}

	if cost != nil {
		usageLog.Cost = cost.TotalCost
	}

	if requestError != nil {
		usageLog.StatusCode = 500
		errorMsg := requestError.Error()
		usageLog.ErrorMessage = &errorMsg
	} else {
		usageLog.StatusCode = 200
	}

	if err := g.usageLogService.CreateUsageLog(ctx, usageLog); err != nil {
		g.logger.WithFields(map[string]interface{}{
			"request_id": request.RequestID,
			"error":      err.Error(),
		}).Error("Failed to create usage log")
		return nil
	}

	return usageLog
}

// 注意：consumeQuotas 函数已删除，配额消费现在只在中间件中处理，避免重复消费

// processBilling 处理计费
func (g *gatewayServiceImpl) processBilling(ctx context.Context, userID int64, usageLogID int64, cost float64) {
	if cost <= 0 {
		return
	}

	// 获取用户信息
	user, err := g.userService.GetUser(ctx, userID)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"cost":    cost,
			"error":   err.Error(),
		}).Error("Failed to get user for billing")
		return
	}

	// 记录扣费信息（允许余额变负数）
	g.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"balance": user.Balance,
		"cost":    cost,
	}).Info("Processing billing - balance may become negative")

	// 扣减用户余额
	updateReq := &dto.BalanceUpdateRequest{
		Amount:      cost,
		Operation:   "deduct",
		Description: fmt.Sprintf("API usage cost: %.8f USD", cost),
	}

	_, err = g.userService.UpdateBalance(ctx, userID, updateReq)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"cost":    cost,
			"error":   err.Error(),
		}).Error("Failed to deduct user balance")
		return
	}

	// 创建计费记录
	description := fmt.Sprintf("API usage cost: %.8f USD", cost)
	processedAt := time.Now()
	billingRecord := &entities.BillingRecord{
		UserID:      userID,
		UsageLogID:  usageLogID,
		Amount:      cost,
		Currency:    "USD",
		BillingType: entities.BillingTypeUsage,
		Description: &description,
		ProcessedAt: &processedAt,
		Status:      entities.BillingStatusProcessed,
	}

	// 创建计费记录
	if err := g.billingRepo.Create(ctx, billingRecord); err != nil {
		g.logger.WithFields(map[string]interface{}{
			"user_id": userID,
			"cost":    cost,
			"error":   err.Error(),
		}).Error("Failed to create billing record")
		// 不返回错误，因为余额已经扣减成功
	} else {
		g.logger.WithFields(map[string]interface{}{
			"user_id":           userID,
			"billing_record_id": billingRecord.ID,
			"amount":            cost,
		}).Info("Billing record created successfully")
	}

	g.logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"cost":        cost,
		"new_balance": user.Balance - cost,
	}).Info("Billing processed successfully")
}

// 注意：processQuotaConsumption 函数已删除，配额消费现在只在中间件中处理，避免重复消费

// HealthCheck 健康检查
func (g *gatewayServiceImpl) HealthCheck(ctx context.Context) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		Status:    "healthy",
		Timestamp: time.Now(),
		Providers: make(map[string]ProviderHealth),
		Database: DatabaseHealth{
			Status: "healthy",
		},
	}

	// TODO: 实现提供商健康检查
	// TODO: 实现数据库健康检查

	return result, nil
}

// ProcessStreamRequest 处理流式AI请求
func (g *gatewayServiceImpl) ProcessStreamRequest(ctx context.Context, request *GatewayRequest, streamChan chan<- *StreamChunk) error {
	// 生成请求ID
	if request.RequestID == "" {
		var err error
		request.RequestID, err = g.requestIDGen.Generate()
		if err != nil {
			g.logger.WithField("error", err.Error()).Error("Failed to generate request ID")
			request.RequestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
	}

	// 路由请求到提供商
	routeResponse, err := g.router.RouteStreamRequest(ctx, request, streamChan)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"request_id": request.RequestID,
			"user_id":    request.UserID,
			"model":      request.ModelSlug,
			"error":      err.Error(),
		}).Error("Failed to route stream request")
		return err
	}

	// 记录使用日志（异步）
	go g.recordStreamUsage(ctx, request, routeResponse)

	return nil
}

// recordStreamUsage 记录流式请求的使用日志
func (g *gatewayServiceImpl) recordStreamUsage(ctx context.Context, request *GatewayRequest, routeResponse *RouteResponse) {
	// 计算使用量和成本
	usage := &UsageInfo{
		InputTokens:  routeResponse.Response.Usage.PromptTokens,
		OutputTokens: routeResponse.Response.Usage.CompletionTokens,
		TotalTokens:  routeResponse.Response.Usage.TotalTokens,
	}

	// 计算成本
	cost, err := g.calculateCost(ctx, routeResponse.Model.ID, usage.InputTokens, usage.OutputTokens)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"request_id": request.RequestID,
			"model_id":   routeResponse.Model.ID,
			"error":      err.Error(),
		}).Warn("Failed to calculate cost for stream request")
		cost = &CostInfo{Currency: "USD"} // 默认值
	}

	// 记录使用日志
	usageLog := g.recordUsageLog(ctx, request, routeResponse.Provider, routeResponse.Model, usage, cost, routeResponse.Duration, nil)

	// 注意：配额消费已在中间件中处理，这里不再重复消费

	// 处理计费
	if usageLog != nil {
		g.processBilling(ctx, request.UserID, usageLog.ID, cost.TotalCost)
	}
}

// GetStats 获取统计信息
func (g *gatewayServiceImpl) GetStats(ctx context.Context) (*GatewayStats, error) {
	// TODO: 实现统计信息收集
	stats := &GatewayStats{
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		SuccessRate:        0.0,
		AvgResponseTime:    0,
		ProvidersStats:     make(map[int64]*ProviderStats),
		Uptime:             time.Since(g.startTime),
	}

	return stats, nil
}
