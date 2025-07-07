package gateway

import (
	"context"
	"fmt"
	"time"

	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/domain/repositories"
	"ai-api-gateway/internal/domain/values"
	"ai-api-gateway/internal/infrastructure/clients"
	"ai-api-gateway/internal/infrastructure/logger"
)

// RequestRouter 请求路由器接口
type RequestRouter interface {
	// RouteRequest 路由请求
	RouteRequest(ctx context.Context, request *RouteRequest) (*RouteResponse, error)

	// RouteStreamRequest 路由流式请求
	RouteStreamRequest(ctx context.Context, request *GatewayRequest, streamChan chan<- *StreamChunk) (*RouteResponse, error)

	// GetAvailableProviders 获取可用的提供商
	GetAvailableProviders(ctx context.Context, modelSlug string) ([]*entities.Provider, error)
}

// RouteRequest 路由请求
type RouteRequest struct {
	UserID     int64              `json:"user_id"`
	APIKeyID   int64              `json:"api_key_id"`
	ModelSlug  string             `json:"model_slug"`
	Request    *clients.AIRequest `json:"request"`
	MaxRetries int                `json:"max_retries"`
	Timeout    time.Duration      `json:"timeout"`
	RequestID  string             `json:"request_id"`
}

// RouteResponse 路由响应
type RouteResponse struct {
	Response  *clients.AIResponse `json:"response"`
	Provider  *entities.Provider  `json:"provider"`
	Model     *entities.Model     `json:"model"`
	Duration  time.Duration       `json:"duration"`
	Retries   int                 `json:"retries"`
	RequestID string              `json:"request_id"`
	Error     error               `json:"error,omitempty"`
}

// requestRouterImpl 请求路由器实现
type requestRouterImpl struct {
	providerService          services.ProviderService
	modelService             services.ModelService
	providerModelSupportRepo repositories.ProviderModelSupportRepository
	loadBalancer             LoadBalancer
	aiClient                 clients.AIProviderClient
	logger                   logger.Logger
	requestIDGen             *values.RequestIDGenerator
}

// NewRequestRouter 创建请求路由器
func NewRequestRouter(
	providerService services.ProviderService,
	modelService services.ModelService,
	providerModelSupportRepo repositories.ProviderModelSupportRepository,
	loadBalancer LoadBalancer,
	aiClient clients.AIProviderClient,
	logger logger.Logger,
) RequestRouter {
	return &requestRouterImpl{
		providerService:          providerService,
		modelService:             modelService,
		providerModelSupportRepo: providerModelSupportRepo,
		loadBalancer:             loadBalancer,
		aiClient:                 aiClient,
		logger:                   logger,
		requestIDGen:             values.NewRequestIDGenerator(),
	}
}

// RouteRequest 路由请求
func (r *requestRouterImpl) RouteRequest(ctx context.Context, request *RouteRequest) (*RouteResponse, error) {
	start := time.Now()

	// 生成请求ID（如果没有提供）
	if request.RequestID == "" {
		var err error
		request.RequestID, err = r.requestIDGen.Generate()
		if err != nil {
			r.logger.WithField("error", err.Error()).Error("Failed to generate request ID")
			request.RequestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"request_id": request.RequestID,
		"user_id":    request.UserID,
		"api_key_id": request.APIKeyID,
		"model_slug": request.ModelSlug,
	}).Info("Routing AI request")

	// 获取可用的提供商
	providers, err := r.GetAvailableProviders(ctx, request.ModelSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get available providers: %w", err)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no available providers for model: %s", request.ModelSlug)
	}

	// 尝试发送请求，支持重试和故障转移
	var lastError error
	retries := 0
	maxRetries := request.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for retries <= maxRetries {
		// 选择提供商
		provider, err := r.loadBalancer.SelectProvider(ctx, providers)
		if err != nil {
			lastError = fmt.Errorf("failed to select provider: %w", err)
			break
		}

		// 获取模型信息
		model, err := r.modelService.GetModelBySlug(ctx, provider.ID, request.ModelSlug)
		if err != nil {
			r.logger.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"provider_id": provider.ID,
				"model_slug":  request.ModelSlug,
				"error":       err.Error(),
			}).Warn("Model not found for provider, trying next provider")

			// 从可用提供商列表中移除这个提供商
			providers = r.removeProvider(providers, provider.ID)
			if len(providers) == 0 {
				lastError = fmt.Errorf("no providers support model: %s", request.ModelSlug)
				break
			}
			retries++
			continue
		}

		// 发送请求
		response, err := r.sendRequest(ctx, provider, model, request)
		if err != nil {
			r.logger.WithFields(map[string]interface{}{
				"request_id":  request.RequestID,
				"provider_id": provider.ID,
				"model_slug":  request.ModelSlug,
				"retry":       retries,
				"error":       err.Error(),
			}).Warn("Request failed, retrying with next provider")

			// 记录失败
			r.loadBalancer.RecordResponse(ctx, provider.ID, false, time.Since(start))

			lastError = err
			retries++

			// 如果还有重试机会，从列表中移除失败的提供商
			if retries <= maxRetries {
				providers = r.removeProvider(providers, provider.ID)
				if len(providers) == 0 {
					break
				}
			}
			continue
		}

		// 请求成功
		duration := time.Since(start)
		r.loadBalancer.RecordResponse(ctx, provider.ID, true, duration)

		r.logger.WithFields(map[string]interface{}{
			"request_id":  request.RequestID,
			"provider_id": provider.ID,
			"model_slug":  request.ModelSlug,
			"duration":    duration,
			"retries":     retries,
		}).Info("Request completed successfully")

		return &RouteResponse{
			Response:  response,
			Provider:  provider,
			Model:     model,
			Duration:  duration,
			Retries:   retries,
			RequestID: request.RequestID,
		}, nil
	}

	// 所有重试都失败了
	r.logger.WithFields(map[string]interface{}{
		"request_id": request.RequestID,
		"retries":    retries,
		"error":      lastError.Error(),
	}).Error("Request failed after all retries")

	return &RouteResponse{
		RequestID: request.RequestID,
		Duration:  time.Since(start),
		Retries:   retries,
		Error:     lastError,
	}, lastError
}

// GetAvailableProviders 获取可用的提供商
func (r *requestRouterImpl) GetAvailableProviders(ctx context.Context, modelSlug string) ([]*entities.Provider, error) {
	// 直接从 provider_model_support 表查询支持指定模型的提供商
	supportInfos, err := r.providerModelSupportRepo.GetSupportingProviders(ctx, modelSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get supporting providers for model %s: %w", modelSlug, err)
	}

	if len(supportInfos) == 0 {
		r.logger.WithField("model_slug", modelSlug).Info("No providers support this model")
		return []*entities.Provider{}, nil
	}

	// 提取提供商列表，按优先级排序（GetSupportingProviders 已经排序了）
	var availableProviders []*entities.Provider
	for _, supportInfo := range supportInfos {
		if supportInfo.IsAvailable() {
			availableProviders = append(availableProviders, supportInfo.Provider)
		}
	}

	r.logger.WithFields(map[string]interface{}{
		"model_slug":          modelSlug,
		"total_supports":      len(supportInfos),
		"available_providers": len(availableProviders),
	}).Debug("Found supporting providers for model")

	return availableProviders, nil
}

// sendRequest 发送请求到提供商
func (r *requestRouterImpl) sendRequest(ctx context.Context, provider *entities.Provider, model *entities.Model, request *RouteRequest) (*clients.AIResponse, error) {
	// 设置超时
	timeout := request.Timeout
	if timeout <= 0 {
		timeout = provider.GetTimeout()
	}

	fmt.Printf("DEBUG: Provider timeout_seconds=%d, calculated timeout=%v\n", provider.TimeoutSeconds, timeout)

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 发送请求
	response, err := r.aiClient.SendRequest(requestCtx, provider, request.Request)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to provider %s: %w", provider.Name, err)
	}

	return response, nil
}

// removeProvider 从提供商列表中移除指定的提供商
func (r *requestRouterImpl) removeProvider(providers []*entities.Provider, providerID int64) []*entities.Provider {
	var result []*entities.Provider
	for _, provider := range providers {
		if provider.ID != providerID {
			result = append(result, provider)
		}
	}
	return result
}

// GetProviderStats 获取提供商统计信息
func (r *requestRouterImpl) GetProviderStats(ctx context.Context, providerID int64) (*ProviderStats, error) {
	return r.loadBalancer.GetProviderStats(ctx, providerID)
}

// GetAllProviderStats 获取所有提供商统计信息
func (r *requestRouterImpl) GetAllProviderStats() map[int64]*ProviderStats {
	if lb, ok := r.loadBalancer.(*loadBalancerImpl); ok {
		return lb.GetAllStats()
	}
	return make(map[int64]*ProviderStats)
}

// RouteStreamRequest 路由流式请求
func (r *requestRouterImpl) RouteStreamRequest(ctx context.Context, request *GatewayRequest, streamChan chan<- *StreamChunk) (*RouteResponse, error) {
	// 暂时返回错误，表示流式功能尚未完全实现
	// TODO: 实现完整的流式请求路由

	// 使用 channel 来等待流式数据发送完成
	done := make(chan struct{})

	// 模拟流式响应
	go func() {
		defer func() {
			// 捕获panic，防止向已关闭的channel发送数据
			if r := recover(); r != nil {
				// 记录panic但不重新抛出
				// 这通常发生在客户端断开连接后channel被关闭的情况
			}
			// 通知完成
			close(done)
		}()

		// 发送一些模拟的流式数据块
		chunks := []string{
			"Hello",
			" there!",
			" This",
			" is",
			" a",
			" simulated",
			" streaming",
			" response.",
		}

		for i, content := range chunks {
			select {
			case <-ctx.Done():
				// 上下文取消，停止发送
				return
			default:
				// 尝试发送数据，如果channel已关闭则会panic，被defer捕获
				select {
				case streamChan <- &StreamChunk{
					ID:      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   request.ModelSlug,
					Content: content,
					FinishReason: func() *string {
						if i == len(chunks)-1 {
							reason := "stop"
							return &reason
						}
						return nil
					}(),
				}:
					time.Sleep(100 * time.Millisecond) // 模拟延迟
				case <-ctx.Done():
					// 在发送过程中上下文被取消
					return
				}
			}
		}
	}()

	// 等待流式数据发送完成或上下文取消
	select {
	case <-done:
		// 流式数据发送完成
	case <-ctx.Done():
		// 上下文取消
		return nil, ctx.Err()
	}

	// 构造基本响应
	response := &RouteResponse{
		Provider: &entities.Provider{
			Name: "simulated",
		},
		Model: &entities.Model{
			Name: request.ModelSlug,
		},
		Response: &clients.AIResponse{
			Usage: clients.AIUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		},
		Duration:  100 * time.Millisecond,
		RequestID: request.RequestID,
	}

	return response, nil
}
