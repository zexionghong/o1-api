package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/application/services"
	"ai-api-gateway/internal/domain/entities"
	"ai-api-gateway/internal/infrastructure/logger"
)

// QuotaMiddleware 配额检查中间件
type QuotaMiddleware struct {
	quotaService services.QuotaService
	logger       logger.Logger
}

// NewQuotaMiddleware 创建配额检查中间件
func NewQuotaMiddleware(quotaService services.QuotaService, logger logger.Logger) *QuotaMiddleware {
	return &QuotaMiddleware{
		quotaService: quotaService,
		logger:       logger,
	}
}

// CheckQuota 配额检查中间件函数
func (m *QuotaMiddleware) CheckQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
				"AUTHENTICATION_REQUIRED",
				"Authentication required for quota check",
				nil,
			))
			c.Abort()
			return
		}
		
		// 检查请求配额
		allowed, err := m.quotaService.CheckQuota(c.Request.Context(), userID, entities.QuotaTypeRequests, 1)
		if err != nil {
			m.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("Failed to check request quota")
			
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
				"QUOTA_CHECK_ERROR",
				"Failed to check quota",
				nil,
			))
			c.Abort()
			return
		}
		
		if !allowed {
			m.logger.WithFields(map[string]interface{}{
				"user_id":    userID,
				"quota_type": entities.QuotaTypeRequests,
			}).Warn("Request quota exceeded")
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"QUOTA_EXCEEDED",
				"Request quota exceeded",
				map[string]interface{}{
					"quota_type": entities.QuotaTypeRequests,
				},
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// CheckTokenQuota 检查token配额
func (m *QuotaMiddleware) CheckTokenQuota(estimatedTokens int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
				"AUTHENTICATION_REQUIRED",
				"Authentication required for quota check",
				nil,
			))
			c.Abort()
			return
		}
		
		// 检查token配额
		allowed, err := m.quotaService.CheckQuota(c.Request.Context(), userID, entities.QuotaTypeTokens, float64(estimatedTokens))
		if err != nil {
			m.logger.WithFields(map[string]interface{}{
				"user_id":          userID,
				"estimated_tokens": estimatedTokens,
				"error":            err.Error(),
			}).Error("Failed to check token quota")
			
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
				"QUOTA_CHECK_ERROR",
				"Failed to check token quota",
				nil,
			))
			c.Abort()
			return
		}
		
		if !allowed {
			m.logger.WithFields(map[string]interface{}{
				"user_id":          userID,
				"estimated_tokens": estimatedTokens,
				"quota_type":       entities.QuotaTypeTokens,
			}).Warn("Token quota exceeded")
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"TOKEN_QUOTA_EXCEEDED",
				"Token quota exceeded",
				map[string]interface{}{
					"quota_type":       entities.QuotaTypeTokens,
					"estimated_tokens": estimatedTokens,
				},
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// CheckCostQuota 检查成本配额
func (m *QuotaMiddleware) CheckCostQuota(estimatedCost float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
				"AUTHENTICATION_REQUIRED",
				"Authentication required for quota check",
				nil,
			))
			c.Abort()
			return
		}
		
		// 检查成本配额
		allowed, err := m.quotaService.CheckQuota(c.Request.Context(), userID, entities.QuotaTypeCost, estimatedCost)
		if err != nil {
			m.logger.WithFields(map[string]interface{}{
				"user_id":        userID,
				"estimated_cost": estimatedCost,
				"error":          err.Error(),
			}).Error("Failed to check cost quota")
			
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(
				"QUOTA_CHECK_ERROR",
				"Failed to check cost quota",
				nil,
			))
			c.Abort()
			return
		}
		
		if !allowed {
			m.logger.WithFields(map[string]interface{}{
				"user_id":        userID,
				"estimated_cost": estimatedCost,
				"quota_type":     entities.QuotaTypeCost,
			}).Warn("Cost quota exceeded")
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"COST_QUOTA_EXCEEDED",
				"Cost quota exceeded",
				map[string]interface{}{
					"quota_type":     entities.QuotaTypeCost,
					"estimated_cost": estimatedCost,
				},
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ConsumeQuota 消费配额中间件（在请求完成后调用）
func (m *QuotaMiddleware) ConsumeQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // 先执行后续处理
		
		// 获取用户ID
		userID, exists := GetUserIDFromContext(c)
		if !exists {
			return
		}
		
		// 消费请求配额
		if err := m.quotaService.ConsumeQuota(c.Request.Context(), userID, entities.QuotaTypeRequests, 1); err != nil {
			m.logger.WithFields(map[string]interface{}{
				"user_id": userID,
				"error":   err.Error(),
			}).Error("Failed to consume request quota")
		}
		
		// 如果有token使用量，消费token配额
		if tokensUsed, exists := c.Get("tokens_used"); exists {
			if tokens, ok := tokensUsed.(int); ok && tokens > 0 {
				if err := m.quotaService.ConsumeQuota(c.Request.Context(), userID, entities.QuotaTypeTokens, float64(tokens)); err != nil {
					m.logger.WithFields(map[string]interface{}{
						"user_id":     userID,
						"tokens_used": tokens,
						"error":       err.Error(),
					}).Error("Failed to consume token quota")
				}
			}
		}
		
		// 如果有成本，消费成本配额
		if costUsed, exists := c.Get("cost_used"); exists {
			if cost, ok := costUsed.(float64); ok && cost > 0 {
				if err := m.quotaService.ConsumeQuota(c.Request.Context(), userID, entities.QuotaTypeCost, cost); err != nil {
					m.logger.WithFields(map[string]interface{}{
						"user_id":   userID,
						"cost_used": cost,
						"error":     err.Error(),
					}).Error("Failed to consume cost quota")
				}
			}
		}
	}
}

// CheckBalance 检查用户余额
func (m *QuotaMiddleware) CheckBalance(estimatedCost float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		user, exists := GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse(
				"AUTHENTICATION_REQUIRED",
				"Authentication required for balance check",
				nil,
			))
			c.Abort()
			return
		}
		
		// 检查余额是否足够
		if !user.CanMakeRequest() {
			m.logger.WithFields(map[string]interface{}{
				"user_id": user.ID,
				"balance": user.Balance,
			}).Warn("Insufficient balance")
			
			c.JSON(http.StatusPaymentRequired, dto.ErrorResponse(
				"INSUFFICIENT_BALANCE",
				"Insufficient account balance",
				map[string]interface{}{
					"current_balance":  user.Balance,
					"estimated_cost":   estimatedCost,
				},
			))
			c.Abort()
			return
		}
		
		// 检查余额是否足够支付预估成本
		if estimatedCost > 0 && user.Balance < estimatedCost {
			m.logger.WithFields(map[string]interface{}{
				"user_id":        user.ID,
				"balance":        user.Balance,
				"estimated_cost": estimatedCost,
			}).Warn("Insufficient balance for estimated cost")
			
			c.JSON(http.StatusPaymentRequired, dto.ErrorResponse(
				"INSUFFICIENT_BALANCE",
				"Insufficient balance for estimated cost",
				map[string]interface{}{
					"current_balance": user.Balance,
					"estimated_cost":  estimatedCost,
					"shortfall":       estimatedCost - user.Balance,
				},
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}
