package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"ai-api-gateway/internal/infrastructure/logger"
)

var (
	// ErrLockNotObtained 获取锁失败
	ErrLockNotObtained = errors.New("lock not obtained")
	// ErrLockNotHeld 锁未持有
	ErrLockNotHeld = errors.New("lock not held")
)

// DistributedLock 分布式锁
type DistributedLock struct {
	client     *RedisClient
	logger     logger.Logger
	key        string
	value      string
	ttl        time.Duration
	retryDelay time.Duration
	maxRetries int
}

// LockOptions 锁选项
type LockOptions struct {
	TTL        time.Duration
	RetryDelay time.Duration
	MaxRetries int
}

// DistributedLockService 分布式锁服务
type DistributedLockService struct {
	client *RedisClient
	logger logger.Logger
}

// NewDistributedLockService 创建分布式锁服务
func NewDistributedLockService(client *RedisClient, logger logger.Logger) *DistributedLockService {
	return &DistributedLockService{
		client: client,
		logger: logger,
	}
}

// NewLock 创建新锁
func (s *DistributedLockService) NewLock(key string, options *LockOptions) *DistributedLock {
	if options == nil {
		options = &LockOptions{
			TTL:        viper.GetDuration("distributed_lock.default_ttl"),
			RetryDelay: viper.GetDuration("distributed_lock.retry_interval"),
			MaxRetries: viper.GetInt("distributed_lock.max_retries"),
		}
	}

	// 生成唯一值
	value := generateLockValue()

	return &DistributedLock{
		client:     s.client,
		logger:     s.logger,
		key:        key,
		value:      value,
		ttl:        options.TTL,
		retryDelay: options.RetryDelay,
		maxRetries: options.MaxRetries,
	}
}

// Lock 获取锁
func (l *DistributedLock) Lock(ctx context.Context) error {
	for i := 0; i <= l.maxRetries; i++ {
		acquired, err := l.tryLock(ctx)
		if err != nil {
			l.logger.WithFields(map[string]interface{}{
				"key":   l.key,
				"error": err.Error(),
			}).Error("Failed to try lock")
			return err
		}

		if acquired {
			l.logger.WithFields(map[string]interface{}{
				"key":     l.key,
				"value":   l.value,
				"ttl":     l.ttl,
				"retries": i,
			}).Debug("Lock acquired successfully")
			return nil
		}

		if i < l.maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(l.retryDelay):
				// 继续重试
			}
		}
	}

	l.logger.WithFields(map[string]interface{}{
		"key":         l.key,
		"max_retries": l.maxRetries,
	}).Warn("Failed to acquire lock after max retries")

	return ErrLockNotObtained
}

// TryLock 尝试获取锁（不重试）
func (l *DistributedLock) TryLock(ctx context.Context) error {
	acquired, err := l.tryLock(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		return ErrLockNotObtained
	}
	return nil
}

// Unlock 释放锁
func (l *DistributedLock) Unlock(ctx context.Context) error {
	// 使用Lua脚本确保原子性
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result := l.client.Eval(ctx, script, []string{l.key}, l.value)
	if err := result.Err(); err != nil {
		l.logger.WithFields(map[string]interface{}{
			"key":   l.key,
			"error": err.Error(),
		}).Error("Failed to unlock")
		return err
	}

	deleted, ok := result.Val().(int64)
	if !ok || deleted == 0 {
		l.logger.WithFields(map[string]interface{}{
			"key":   l.key,
			"value": l.value,
		}).Warn("Lock not held during unlock")
		return ErrLockNotHeld
	}

	l.logger.WithFields(map[string]interface{}{
		"key":   l.key,
		"value": l.value,
	}).Debug("Lock released successfully")

	return nil
}

// Extend 延长锁的过期时间
func (l *DistributedLock) Extend(ctx context.Context, ttl time.Duration) error {
	// 使用Lua脚本确保原子性
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result := l.client.Eval(ctx, script, []string{l.key}, l.value, int64(ttl.Seconds()))
	if err := result.Err(); err != nil {
		l.logger.WithFields(map[string]interface{}{
			"key":   l.key,
			"ttl":   ttl,
			"error": err.Error(),
		}).Error("Failed to extend lock")
		return err
	}

	extended, ok := result.Val().(int64)
	if !ok || extended == 0 {
		l.logger.WithFields(map[string]interface{}{
			"key":   l.key,
			"value": l.value,
		}).Warn("Lock not held during extend")
		return ErrLockNotHeld
	}

	l.ttl = ttl
	l.logger.WithFields(map[string]interface{}{
		"key": l.key,
		"ttl": ttl,
	}).Debug("Lock extended successfully")

	return nil
}

// IsHeld 检查锁是否被持有
func (l *DistributedLock) IsHeld(ctx context.Context) (bool, error) {
	value, err := l.client.Get(ctx, l.key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return value == l.value, nil
}

// TTL 获取锁的剩余时间
func (l *DistributedLock) TTL(ctx context.Context) (time.Duration, error) {
	return l.client.TTL(ctx, l.key)
}

// tryLock 尝试获取锁的内部方法
func (l *DistributedLock) tryLock(ctx context.Context) (bool, error) {
	// 使用SET命令的NX和EX选项实现原子操作
	result := l.client.GetClient().SetNX(ctx, l.key, l.value, l.ttl)
	if err := result.Err(); err != nil {
		return false, err
	}

	return result.Val(), nil
}

// generateLockValue 生成锁的唯一值
func generateLockValue() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}

// WithLock 使用锁执行函数
func (s *DistributedLockService) WithLock(ctx context.Context, key string, options *LockOptions, fn func() error) error {
	lock := s.NewLock(key, options)

	// 获取锁
	if err := lock.Lock(ctx); err != nil {
		return err
	}

	// 确保释放锁
	defer func() {
		if err := lock.Unlock(ctx); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"key":   key,
				"error": err.Error(),
			}).Error("Failed to unlock in defer")
		}
	}()

	// 执行函数
	return fn()
}

// GetLockKey 生成锁的键名
func GetLockKey(prefix, identifier string) string {
	return "lock:" + prefix + ":" + identifier
}

// GetBillingLockKey 生成计费锁的键名
func GetBillingLockKey(userID int64) string {
	return GetLockKey("billing", fmt.Sprintf("%d", userID))
}

// GetQuotaLockKey 生成配额锁的键名
func GetQuotaLockKey(userID int64, quotaType string) string {
	return GetLockKey("quota", fmt.Sprintf("%d:%s", userID, quotaType))
}

// GetUserLockKey 生成用户锁的键名
func GetUserLockKey(userID int64) string {
	return GetLockKey("user", fmt.Sprintf("%d", userID))
}
