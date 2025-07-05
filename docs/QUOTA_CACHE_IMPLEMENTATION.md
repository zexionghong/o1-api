# 配额缓存实现文档

## 概述

本文档描述了为配额系统添加的完整缓存机制，包括配额检查、配额消费和配额状态查询的缓存优化。

## 🎯 实现目标

### 性能优化
- **减少数据库查询**：配额检查是每个API请求都需要的操作，通过缓存可以显著减少数据库压力
- **提升响应速度**：将数据库查询时间从毫秒级降低到微秒级
- **支持高并发**：缓存可以支持更高的并发请求量

### 数据一致性
- **短期缓存**：配额使用情况使用短期缓存（1-2分钟），保证实时性
- **智能失效**：配额消费后立即失效相关缓存
- **降级策略**：缓存失败时自动降级到数据库查询

## 🏗️ 架构设计

### 缓存层次
```
API请求 → QuotaService → CacheService → Redis
                    ↓
                 Repository → Database
```

### 缓存策略
1. **配额设置缓存**：TTL 5分钟（配额设置不经常变化）
2. **配额使用缓存**：TTL 2分钟（需要保证实时性）
3. **用户配额列表缓存**：TTL 5分钟（中等频率更新）

## 📋 实现详情

### 1. 扩展的QuotaService

#### 新增构造函数
```go
// NewQuotaServiceWithCache 创建带缓存的配额服务
func NewQuotaServiceWithCache(
    quotaRepo repositories.QuotaRepository,
    quotaUsageRepo repositories.QuotaUsageRepository,
    userRepo repositories.UserRepository,
    cache *redisInfra.CacheService,
    invalidationService *redisInfra.CacheInvalidationService,
    logger logger.Logger,
) QuotaService
```

#### 缓存集成的方法
- `CheckQuota()` - 配额检查（带缓存）
- `ConsumeQuota()` - 配额消费（带缓存失效）
- `GetQuotaStatus()` - 配额状态查询（带缓存）

### 2. 缓存键设计

#### 用户配额列表
- **键格式**：`user_quotas:{userID}`
- **TTL**：5分钟
- **用途**：缓存用户的所有配额设置

#### 配额使用情况
- **键格式**：`quota_usage:{userID}:{quotaType}:{period}`
- **TTL**：2分钟
- **用途**：缓存特定周期的配额使用情况

#### 活跃配额列表
- **键格式**：`user:{userID}:quotas:active`
- **TTL**：5分钟
- **用途**：缓存用户的活跃配额列表

### 3. 缓存辅助方法

#### getUserQuotasWithCache
```go
func (s *quotaServiceImpl) getUserQuotasWithCache(ctx context.Context, userID int64) ([]*entities.Quota, error) {
    // 1. 尝试从缓存获取
    // 2. 缓存未命中时查询数据库
    // 3. 缓存查询结果
}
```

#### getQuotaUsageWithCache
```go
func (s *quotaServiceImpl) getQuotaUsageWithCache(ctx context.Context, userID, quotaID int64, period entities.QuotaPeriod, now time.Time) (*entities.QuotaUsage, error) {
    // 1. 生成基于时间周期的缓存键
    // 2. 尝试从缓存获取使用情况
    // 3. 缓存未命中时查询数据库
    // 4. 缓存查询结果（短期TTL）
}
```

#### invalidateUserQuotaCache
```go
func (s *quotaServiceImpl) invalidateUserQuotaCache(ctx context.Context, userID int64) {
    // 失效用户相关的所有配额缓存
}
```

### 4. ServiceFactory集成

#### 自动缓存检测
```go
func (f *ServiceFactory) QuotaService() QuotaService {
    // 如果有Redis工厂，创建带缓存的配额服务
    if f.redisFactory != nil {
        return NewQuotaServiceWithCache(...)
    }
    
    // 否则创建普通的配额服务
    return NewQuotaService(...)
}
```

## 🔄 缓存失效策略

### 主动失效
- **配额消费后**：立即失效用户配额相关缓存
- **配额设置更新**：失效相关的配额和使用情况缓存
- **用户状态变更**：失效用户相关的所有缓存

### 模式匹配失效
```go
// 失效用户的所有配额使用缓存
pattern := fmt.Sprintf("quota_usage:user:%d:*", userID)
invalidationService.deleteByPattern(ctx, pattern)
```

### 级联失效
- 更新配额设置时，同时失效相关的使用情况缓存
- 用户状态变更时，失效用户的所有配额相关缓存

## 📊 性能提升效果

### 预期性能改进
1. **配额检查延迟**：从 5-10ms 降低到 0.1-0.5ms
2. **数据库查询减少**：配额相关查询减少 80-90%
3. **并发处理能力**：提升 3-5 倍
4. **系统响应时间**：整体API响应时间减少 20-30%

### 缓存命中率目标
- **配额设置查询**：目标命中率 > 95%
- **配额使用查询**：目标命中率 > 85%
- **用户配额列表**：目标命中率 > 90%

## 🛠️ 使用方法

### 1. 基本使用
```go
// 配额服务会自动使用缓存（如果可用）
quotaService := serviceFactory.QuotaService()

// 检查配额（带缓存）
allowed, err := quotaService.CheckQuota(ctx, userID, quotaType, value)

// 消费配额（自动失效缓存）
err = quotaService.ConsumeQuota(ctx, userID, quotaType, value)
```

### 2. 缓存监控
```go
// 获取缓存统计
cacheManager := redisFactory.GetCacheManager()
stats, err := cacheManager.GetCacheStats(ctx)

// 检查配额相关缓存
quotaKeys := stats.KeysByPattern["quota:*"]
usageKeys := stats.KeysByPattern["quota_usage:*"]
```

### 3. 缓存管理
```go
// 清除特定用户的配额缓存
err := cacheManager.ClearCache(ctx, fmt.Sprintf("user:%d:*", userID))

// 刷新配额缓存
err := cacheManager.RefreshCache(ctx, "quotas")
```

## 🔍 监控和调试

### 日志记录
- **缓存命中**：DEBUG级别记录缓存命中情况
- **缓存失效**：INFO级别记录缓存失效操作
- **缓存错误**：WARN级别记录缓存操作失败

### 监控指标
- 配额缓存命中率
- 配额查询响应时间
- 缓存失效频率
- Redis内存使用情况

### 故障排查
1. **缓存未命中**：检查TTL设置和缓存键格式
2. **数据不一致**：检查缓存失效机制是否正确触发
3. **性能问题**：监控Redis连接池和内存使用

## 🚀 部署和配置

### 配置示例
```yaml
cache:
  enabled: true
  
  # 配额相关缓存TTL
  entity:
    quota_ttl: "1m"              # 配额信息缓存时间
  
  query:
    quota_usage_ttl: "2m"        # 配额使用情况缓存时间
    user_quota_list_ttl: "5m"    # 用户配额列表缓存时间
  
  # 功能开关
  features:
    entity_cache: true           # 实体缓存
    query_cache: true            # 查询缓存
    auto_invalidation: true      # 自动缓存失效
```

### 部署注意事项
1. **Redis高可用**：确保Redis集群的高可用性
2. **内存监控**：监控Redis内存使用，设置合理的过期策略
3. **网络延迟**：确保应用服务器与Redis之间的网络延迟较低
4. **备份策略**：虽然是缓存数据，但建议定期备份Redis数据

## 📈 未来优化方向

### 1. 智能预热
- 系统启动时预加载热点用户的配额信息
- 基于历史访问模式预测需要缓存的数据

### 2. 分层缓存
- 本地内存缓存 + Redis缓存的二级缓存架构
- 进一步减少网络延迟

### 3. 缓存压缩
- 对大型配额列表进行压缩存储
- 减少Redis内存使用

### 4. 动态TTL
- 根据数据访问频率动态调整TTL
- 热点数据使用更长的TTL
