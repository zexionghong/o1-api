# 缓存使用指南

## 概述

本项目已实现了完整的Redis缓存管理系统，包括：

- **基础缓存服务**：提供基本的缓存操作（Set、Get、Delete等）
- **缓存键管理**：统一的缓存键命名规范
- **缓存失效管理**：自动和手动的缓存失效机制
- **缓存监控**：缓存统计和健康检查
- **Repository层缓存集成**：透明的数据库查询缓存

## 缓存键命名规范

### 单实体缓存
- `user:{userID}` - 用户基本信息
- `user:username:{username}` - 按用户名查询
- `user:email:{email}` - 按邮箱查询
- `apikey:{apiKeyStr}` - API密钥信息
- `model:{modelID}` - 模型信息
- `provider:{providerID}` - 提供商信息
- `quota:{quotaID}` - 配额信息

### 列表缓存
- `models:active` - 活跃模型列表
- `models:type:{modelType}` - 按类型的模型列表
- `models:available` - 可用模型列表
- `providers:available` - 可用提供商列表
- `user:{userID}:quotas` - 用户配额列表
- `user:{userID}:apikeys` - 用户API密钥列表

### 复合查询缓存
- `quota:user:{userID}:type:{quotaType}:period:{period}` - 特定配额查询
- `quota_usage:user:{userID}:quota:{quotaID}:period:{start}:{end}` - 配额使用情况

### 统计缓存
- `stats:users:count` - 用户总数
- `stats:quotas:count` - 配额总数
- `users:active:page:{offset}:{limit}` - 分页的活跃用户列表

## TTL策略

### 长期缓存（30分钟）
- 模型信息：相对稳定，更新频率低
- 提供商信息：配置类数据，变化较少
- 模型列表：系统级配置，更新不频繁

### 中期缓存（10-15分钟）
- 用户基本信息：个人信息变化不频繁
- API密钥信息：安全相关，需要一定实时性

### 短期缓存（1-5分钟）
- 配额信息：需要较高实时性
- 配额使用情况：频繁更新，短期缓存
- 用户查询（按用户名/邮箱）：登录场景使用

## 使用方法

### 1. 基础缓存操作

```go
// 获取缓存服务
cache := redisFactory.GetCache()

// 设置缓存
err := cache.Set(ctx, "key", value, 5*time.Minute)

// 获取缓存
var result SomeType
err := cache.Get(ctx, "key", &result)

// 删除缓存
err := cache.Delete(ctx, "key1", "key2")
```

### 2. 实体缓存操作

```go
// 用户缓存
err := cache.SetUser(ctx, user)
user, err := cache.GetUser(ctx, userID)
err := cache.DeleteUser(ctx, userID)

// API密钥缓存
err := cache.SetAPIKey(ctx, apiKey)
apiKey, err := cache.GetAPIKey(ctx, apiKeyStr)
err := cache.DeleteAPIKey(ctx, apiKeyStr)

// 模型缓存
err := cache.SetModel(ctx, model)
model, err := cache.GetModel(ctx, modelID)
err := cache.DeleteModel(ctx, modelID)
```

### 3. 列表缓存操作

```go
// 活跃模型列表
err := cache.SetActiveModels(ctx, models)
models, err := cache.GetActiveModels(ctx)

// 可用提供商列表
err := cache.SetAvailableProviders(ctx, providers)
providers, err := cache.GetAvailableProviders(ctx)

// 用户配额列表
err := cache.SetUserQuotas(ctx, userID, quotas)
quotas, err := cache.GetUserQuotas(ctx, userID)
```

### 4. 缓存失效操作

```go
// 获取缓存失效服务
invalidationService := redisFactory.GetInvalidationService()

// 失效用户相关缓存
err := invalidationService.InvalidateUserCache(ctx, userID, username, email)

// 失效模型相关缓存
err := invalidationService.InvalidateModelCache(ctx, modelID, modelType)

// 失效API密钥相关缓存
err := invalidationService.InvalidateAPIKeyCache(ctx, userID, apiKeyStr)

// 批量失效操作
operations := []InvalidationOperation{
    NewKeysInvalidation("key1", "key2"),
    NewPatternInvalidation("user:*"),
}
err := invalidationService.BatchInvalidate(ctx, operations)
```

### 5. 缓存管理操作

```go
// 获取缓存管理器
cacheManager := redisFactory.GetCacheManager()

// 获取缓存统计
stats, err := cacheManager.GetCacheStats(ctx)

// 清除缓存
err := cacheManager.ClearCache(ctx, "user:*", "model:*")

// 刷新特定类型缓存
err := cacheManager.RefreshCache(ctx, "models")

// 获取缓存健康状态
health, err := cacheManager.GetCacheHealth(ctx)

// 预热缓存
err := cacheManager.WarmupCache(ctx)
```

## Repository层缓存集成

### 使用带缓存的Repository

```go
// 创建带缓存的Repository
cachedModelRepo := NewCachedModelRepository(
    db,
    redisFactory.GetCache(),
    redisFactory.GetInvalidationService(),
)

// 使用方式与普通Repository相同
model, err := cachedModelRepo.GetByID(ctx, modelID)
models, err := cachedModelRepo.GetActiveModels(ctx)

// 更新操作会自动失效相关缓存
err := cachedModelRepo.Update(ctx, model)
```

## 配置说明

### 缓存配置示例

```yaml
cache:
  enabled: true
  
  # 实体缓存TTL
  entity:
    user_ttl: "10m"
    api_key_ttl: "15m"
    model_ttl: "30m"
    provider_ttl: "30m"
    quota_ttl: "1m"
  
  # 查询缓存TTL
  query:
    user_lookup_ttl: "5m"
    model_list_ttl: "30m"
    provider_list_ttl: "30m"
    quota_usage_ttl: "2m"
    user_quota_list_ttl: "5m"
    api_key_list_ttl: "10m"
  
  # 统计缓存TTL
  stats:
    count_ttl: "10m"
    pagination_ttl: "5m"
  
  # 功能开关
  features:
    entity_cache: true
    list_cache: true
    query_cache: true
    stats_cache: true
    auto_invalidation: true
  
  # 性能配置
  performance:
    batch_invalidation: true
    preload_on_startup: true
    max_key_length: 250
```

## 最佳实践

### 1. 缓存策略选择
- **高频查询**：积极缓存，较长TTL
- **实时性要求高**：短TTL或主动失效
- **大数据量查询**：分页缓存，避免内存压力

### 2. 缓存失效策略
- **数据更新时**：立即失效相关缓存
- **批量操作**：使用批量失效避免性能问题
- **定期清理**：设置合理的TTL，避免内存泄漏

### 3. 错误处理
- **缓存错误不影响主流程**：缓存失败时继续查询数据库
- **记录缓存错误**：便于监控和调试
- **降级策略**：缓存不可用时自动降级到数据库查询

### 4. 监控和维护
- **定期监控缓存命中率**：优化缓存策略
- **监控内存使用**：避免Redis内存溢出
- **定期清理过期数据**：保持缓存健康

## 故障排查

### 常见问题
1. **缓存未命中**：检查键名是否正确，TTL是否过期
2. **内存使用过高**：检查TTL设置，清理无用缓存
3. **缓存不一致**：检查失效机制是否正确触发
4. **性能问题**：检查缓存键数量，优化查询模式

### 调试工具
- 使用`GetCacheStats`查看缓存统计
- 使用`GetCacheHealth`检查缓存健康状态
- 查看Redis日志分析问题
- 使用Redis CLI直接查看缓存内容
