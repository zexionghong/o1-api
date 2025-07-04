package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"ai-api-gateway/internal/application/dto"
	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/logger"
)

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	config   *config.RateLimitConfig
	logger   logger.Logger
}

// NewRateLimitMiddleware 创建速率限制中间件
func NewRateLimitMiddleware(config *config.RateLimitConfig, logger logger.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rate.Limiter),
		config:   config,
		logger:   logger,
	}
}

// RateLimit 速率限制中间件函数
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID和API密钥ID
		userID, hasUserID := GetUserIDFromContext(c)
		apiKeyID, hasAPIKeyID := GetAPIKeyIDFromContext(c)
		
		var identifier string
		if hasAPIKeyID {
			identifier = fmt.Sprintf("api_key_%d", apiKeyID)
		} else if hasUserID {
			identifier = fmt.Sprintf("user_%d", userID)
		} else {
			// 如果没有认证信息，使用IP地址
			identifier = fmt.Sprintf("ip_%s", c.ClientIP())
		}
		
		// 获取或创建限流器
		limiter := m.getLimiter(identifier)
		
		// 检查是否允许请求
		if !limiter.Allow() {
			m.logger.WithFields(map[string]interface{}{
				"identifier": identifier,
				"user_id":    userID,
				"api_key_id": apiKeyID,
				"ip":         c.ClientIP(),
			}).Warn("Rate limit exceeded")
			
			// 计算重试时间
			reservation := limiter.Reserve()
			retryAfter := int(reservation.Delay().Seconds())
			reservation.Cancel() // 取消预约
			
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.DefaultRequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"RATE_LIMIT_EXCEEDED",
				"Rate limit exceeded",
				map[string]interface{}{
					"retry_after_seconds": retryAfter,
					"limit_per_minute":    m.config.DefaultRequestsPerMinute,
				},
			))
			c.Abort()
			return
		}
		
		// 设置速率限制响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.DefaultRequestsPerMinute))
		// 这里简化处理，实际应该计算剩余请求数
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", m.config.DefaultRequestsPerMinute-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		
		c.Next()
	}
}

// getLimiter 获取或创建限流器
func (m *RateLimitMiddleware) getLimiter(identifier string) *rate.Limiter {
	m.mu.RLock()
	limiter, exists := m.limiters[identifier]
	m.mu.RUnlock()
	
	if exists {
		return limiter
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 双重检查
	if limiter, exists := m.limiters[identifier]; exists {
		return limiter
	}
	
	// 创建新的限流器
	// 每分钟允许的请求数，突发容量为请求数的2倍
	limiter = rate.NewLimiter(
		rate.Every(time.Minute/time.Duration(m.config.DefaultRequestsPerMinute)),
		m.config.DefaultRequestsPerMinute*2,
	)
	
	m.limiters[identifier] = limiter
	return limiter
}

// CleanupLimiters 清理不活跃的限流器
func (m *RateLimitMiddleware) CleanupLimiters() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mu.Lock()
		// 简化处理：清理所有限流器
		// 实际应该根据最后使用时间来清理
		if len(m.limiters) > 1000 {
			m.limiters = make(map[string]*rate.Limiter)
			m.logger.Info("Cleaned up rate limiters")
		}
		m.mu.Unlock()
	}
}

// CustomRateLimit 自定义速率限制
func (m *RateLimitMiddleware) CustomRateLimit(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户ID和API密钥ID
		userID, hasUserID := GetUserIDFromContext(c)
		apiKeyID, hasAPIKeyID := GetAPIKeyIDFromContext(c)
		
		var identifier string
		if hasAPIKeyID {
			identifier = fmt.Sprintf("custom_api_key_%d", apiKeyID)
		} else if hasUserID {
			identifier = fmt.Sprintf("custom_user_%d", userID)
		} else {
			identifier = fmt.Sprintf("custom_ip_%s", c.ClientIP())
		}
		
		// 创建自定义限流器
		limiter := rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(requestsPerMinute)),
			requestsPerMinute*2,
		)
		
		// 检查是否允许请求
		if !limiter.Allow() {
			m.logger.WithFields(map[string]interface{}{
				"identifier":         identifier,
				"user_id":           userID,
				"api_key_id":        apiKeyID,
				"ip":                c.ClientIP(),
				"requests_per_minute": requestsPerMinute,
			}).Warn("Custom rate limit exceeded")
			
			reservation := limiter.Reserve()
			retryAfter := int(reservation.Delay().Seconds())
			reservation.Cancel()
			
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"RATE_LIMIT_EXCEEDED",
				"Rate limit exceeded",
				map[string]interface{}{
					"retry_after_seconds": retryAfter,
					"limit_per_minute":    requestsPerMinute,
				},
			))
			c.Abort()
			return
		}
		
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", requestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", requestsPerMinute-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		
		c.Next()
	}
}

// IPRateLimit IP级别的速率限制
func (m *RateLimitMiddleware) IPRateLimit(requestsPerMinute int) gin.HandlerFunc {
	ipLimiters := make(map[string]*rate.Limiter)
	var mu sync.RWMutex
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		mu.RLock()
		limiter, exists := ipLimiters[ip]
		mu.RUnlock()
		
		if !exists {
			mu.Lock()
			if limiter, exists = ipLimiters[ip]; !exists {
				limiter = rate.NewLimiter(
					rate.Every(time.Minute/time.Duration(requestsPerMinute)),
					requestsPerMinute,
				)
				ipLimiters[ip] = limiter
			}
			mu.Unlock()
		}
		
		if !limiter.Allow() {
			m.logger.WithFields(map[string]interface{}{
				"ip":                ip,
				"requests_per_minute": requestsPerMinute,
			}).Warn("IP rate limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse(
				"IP_RATE_LIMIT_EXCEEDED",
				"IP rate limit exceeded",
				map[string]interface{}{
					"limit_per_minute": requestsPerMinute,
				},
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}
