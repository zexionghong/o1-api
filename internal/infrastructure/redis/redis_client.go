package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"ai-api-gateway/internal/infrastructure/logger"
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
	logger logger.Logger
}

// NewRedisClient 创建Redis客户端
func NewRedisClient(log logger.Logger) (*RedisClient, error) {
	// 从配置中读取Redis设置
	addr := viper.GetString("redis.addr")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")
	poolSize := viper.GetInt("redis.pool_size")
	minIdleConns := viper.GetInt("redis.min_idle_conns")
	dialTimeout := viper.GetDuration("redis.dial_timeout")
	readTimeout := viper.GetDuration("redis.read_timeout")
	writeTimeout := viper.GetDuration("redis.write_timeout")
	poolTimeout := viper.GetDuration("redis.pool_timeout")
	idleCheckFrequency := viper.GetDuration("redis.idle_check_frequency")
	idleTimeout := viper.GetDuration("redis.idle_timeout")
	maxConnAge := viper.GetDuration("redis.max_conn_age")

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:               addr,
		Password:           password,
		DB:                 db,
		PoolSize:           poolSize,
		MinIdleConns:       minIdleConns,
		DialTimeout:        dialTimeout,
		ReadTimeout:        readTimeout,
		WriteTimeout:       writeTimeout,
		PoolTimeout:        poolTimeout,
		IdleCheckFrequency: idleCheckFrequency,
		IdleTimeout:        idleTimeout,
		MaxConnAge:         maxConnAge,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.WithFields(map[string]interface{}{
			"addr":  addr,
			"error": err.Error(),
		}).Error("Failed to connect to Redis")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"addr": addr,
		"db":   db,
	}).Info("Connected to Redis successfully")

	return &RedisClient{
		client: rdb,
		logger: log,
	}, nil
}

// GetClient 获取原生Redis客户端
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set 设置键值对
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除键
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取键的剩余生存时间
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Keys 根据模式获取键列表
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Incr 递增
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// IncrByFloat 按指定浮点值递增
func (r *RedisClient) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	return r.client.IncrByFloat(ctx, key, value).Result()
}

// Decr 递减
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减
func (r *RedisClient) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}

// HSet 设置哈希字段
func (r *RedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取哈希所有字段
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// SAdd 添加集合成员
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SRem 移除集合成员
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

// ZAdd 添加有序集合成员
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合范围内的成员
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 移除有序集合成员
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// Pipeline 创建管道
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// TxPipeline 创建事务管道
func (r *RedisClient) TxPipeline() redis.Pipeliner {
	return r.client.TxPipeline()
}

// Eval 执行Lua脚本
func (r *RedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.Eval(ctx, script, keys, args...)
}

// EvalSha 执行已缓存的Lua脚本
func (r *RedisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.client.EvalSha(ctx, sha1, keys, args...)
}

// Publish 发布消息
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// PSubscribe 模式订阅
func (r *RedisClient) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return r.client.PSubscribe(ctx, patterns...)
}

// FlushDB 清空当前数据库
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// FlushAll 清空所有数据库
func (r *RedisClient) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// Info 获取Redis信息
func (r *RedisClient) Info(ctx context.Context, section ...string) (string, error) {
	return r.client.Info(ctx, section...).Result()
}

// Ping 测试连接
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
