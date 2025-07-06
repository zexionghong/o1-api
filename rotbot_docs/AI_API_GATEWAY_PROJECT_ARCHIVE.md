# AI API Gateway 项目完整归档文档

## 📋 项目概述

### 基本信息
- **项目名称**: AI API Gateway
- **开发语言**: Go 1.23+
- **架构模式**: 清洁架构 (Clean Architecture)
- **主要功能**: AI API网关、负载均衡、配额管理、计费系统
- **部署方式**: Docker容器化 + Docker Compose编排

### 核心特性
1. **多提供商支持**: 统一接口访问OpenAI、Anthropic等多个AI服务
2. **智能负载均衡**: 支持轮询、加权、最少连接等多种策略
3. **精确配额管理**: 多维度配额控制（请求数、Token数、成本）
4. **完整计费系统**: 实时成本计算和详细账单记录
5. **企业级安全**: API密钥管理、权限控制、速率限制
6. **高可用性**: 健康检查、故障转移、自动恢复
7. **实时监控**: Prometheus指标、详细日志、性能统计

## 🏗️ 架构设计

### 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │───▶│  API Gateway    │───▶│  AI Providers   │
│                 │    │                 │    │                 │
│ • Web Apps      │    │ • Load Balance  │    │ • OpenAI        │
│ • Mobile Apps   │    │ • Rate Limiting │    │ • Anthropic     │
│ • CLI Tools     │    │ • Auth & Quota  │    │ • Others        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   Database      │
                    │                 │
                    │ • SQLite/PG     │
                    │ • Redis Cache   │
                    └─────────────────┘
```

### 清洁架构分层
```
internal/
├── domain/           # 领域层 - 业务实体和规则
│   ├── entities/     # 业务实体 (User, APIKey, Provider, Model等)
│   ├── repositories/ # 仓储接口定义
│   ├── services/     # 领域服务接口
│   └── values/       # 值对象
├── application/      # 应用层 - 用例和服务实现
│   ├── dto/          # 数据传输对象
│   └── services/     # 应用服务实现
├── infrastructure/   # 基础设施层 - 外部依赖
│   ├── clients/      # 外部API客户端
│   ├── config/       # 配置管理
│   ├── database/     # 数据库连接
│   ├── gateway/      # 网关核心逻辑
│   ├── logger/       # 日志系统
│   ├── redis/        # Redis缓存
│   └── repositories/ # 仓储实现
└── presentation/     # 表现层 - HTTP接口
    ├── handlers/     # HTTP处理器
    ├── middleware/   # 中间件
    └── routes/       # 路由配置
```

## 🔧 核心功能模块

### 1. 用户管理系统
- **用户实体**: 用户名、邮箱、余额、状态管理
- **余额控制**: 充值、扣费、余额不足检查
- **状态管理**: active(活跃)、suspended(暂停)、deleted(已删除)

### 2. API密钥管理
- **密钥生成**: 64字符随机密钥，ak_前缀标识
- **权限控制**: 细粒度的提供商和模型访问权限
- **生命周期**: 创建、激活、暂停、撤销、过期管理
- **安全存储**: 直接存储完整密钥，支持前缀快速查找

### 3. AI提供商集成
- **支持的提供商**: OpenAI、Anthropic、302.AI等
- **统一接口**: OpenAI兼容的API格式
- **配置管理**: 基础URL、API密钥、超时设置、重试策略
- **健康检查**: 定期检查提供商可用性

### 4. 负载均衡和路由
- **负载均衡策略**:
  - Round Robin (轮询)
  - Weighted (加权)
  - Least Connections (最少连接)
  - Random (随机)
- **故障转移**: 自动切换到健康的提供商
- **重试机制**: 可配置的重试次数和间隔

### 5. 配额系统
- **配额类型**:
  - requests: 请求次数限制
  - tokens: Token使用量限制
  - cost: 成本限制
- **时间周期**: 分钟、小时、天、月
- **实时检查**: 请求前检查配额，请求后消费配额

### 6. 计费系统
- **成本计算**: 基于Token使用量和模型定价
- **价格倍率**: 支持在原价基础上设置倍率（默认1.5倍）
- **计费记录**: 详细的使用记录和成本明细
- **余额扣减**: 自动从用户余额中扣除费用

## 💾 数据库设计

### 核心表结构
1. **users**: 用户信息和余额
2. **api_keys**: API密钥和权限配置
3. **providers**: AI服务提供商配置
4. **models**: AI模型信息和定价
5. **quotas**: 配额设置
6. **quota_usage**: 配额使用情况
7. **usage_logs**: 详细的使用日志
8. **billing_records**: 计费记录

### 设计原则
- **无外键约束**: 关系在应用层维护，提高性能和灵活性
- **索引优化**: 针对查询频繁的字段建立索引
- **数据精度**: 金额字段使用DECIMAL(15,6)保证精度

## 🌐 API接口规范

### OpenAI兼容接口
- `POST /v1/chat/completions` - 聊天完成
- `POST /v1/completions` - 文本完成
- `GET /v1/models` - 获取模型列表
- `GET /v1/usage` - 获取使用情况

### 管理接口
- `POST /admin/users/` - 创建用户
- `GET /admin/users/{id}` - 获取用户信息
- `POST /admin/users/{id}/balance` - 更新用户余额
- `POST /admin/api-keys/` - 创建API密钥
- `POST /admin/api-keys/{id}/revoke` - 撤销API密钥

### 健康检查接口
- `GET /health` - 健康检查
- `GET /health/ready` - 就绪检查
- `GET /health/stats` - 统计信息
- `GET /metrics` - Prometheus监控指标

## 🔒 安全和中间件

### 认证中间件
- API密钥验证
- 用户状态检查
- 权限验证

### 速率限制中间件
- IP级别限流
- 用户级别限流
- API密钥级别限流

### 配额中间件
- 请求前配额检查
- 请求后配额消费
- 配额超限处理

### 通用中间件
- CORS跨域处理
- 安全头设置
- 请求ID生成
- 超时控制
- 异常恢复

## 📊 监控和日志

### 日志系统
- 结构化日志记录
- 多级别日志输出
- 请求链路追踪
- 错误详情记录

### 监控指标
- 请求总数和成功率
- 响应时间分布
- 提供商健康状态
- 配额使用情况
- 系统资源使用

### 健康检查
- 数据库连接检查
- Redis连接检查
- 提供商可用性检查
- 系统资源检查

## 🚀 部署和运维

### 环境要求
- Go 1.23+
- SQLite 3.x (开发) / PostgreSQL 12+ (生产)
- Redis 6.x+ (可选，用于缓存)
- Docker 20.x+ (容器化部署)

### 配置文件
主要配置在 `configs/config.yaml`:
```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  dsn: "./data/gateway.db"

providers:
  openai:
    name: "OpenAI"
    base_url: "https://api.openai.com/v1"
    enabled: true
    priority: 1

redis:
  addr: "localhost:6379"
  password: ""
  db: 0
```

### 部署方式

#### 1. 直接运行
```bash
# 安装依赖
go mod tidy

# 运行数据库迁移
go run cmd/migrate/main.go -direction=up

# 启动服务
go run cmd/server/main.go -config configs/config.yaml
```

#### 2. Docker部署
```bash
# 构建镜像
docker build -t ai-api-gateway .

# 运行容器
docker run -p 8080:8080 -v ./data:/app/data ai-api-gateway
```

#### 3. Docker Compose部署
```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 🛠️ 开发指南

### 项目初始化
```bash
# 克隆项目
git clone <repository-url>
cd ai-api-gateway

# 安装依赖
go mod tidy

# 运行数据库迁移
go run cmd/migrate/main.go -direction=up

# 设置测试数据
go run cmd/e2etest/main.go -action=setup
```

### 测试方法
```bash
# 运行健康检查
curl http://localhost:8080/health

# 运行Python测试脚本
python test_service.py

# 运行E2E测试
go run cmd/e2etest/main.go -action=test -apikey=YOUR_API_KEY
```

### 常用命令工具
- `cmd/checkapi/` - API连通性检查
- `cmd/checkdb/` - 数据库状态检查
- `cmd/setupquotas/` - 配额初始化
- `cmd/pricing/` - 定价数据管理
- `cmd/modelsupport/` - 模型支持配置

## 📈 性能和扩展

### 性能优化
- Redis缓存用户和API密钥信息
- 数据库连接池管理
- 异步日志记录
- 请求去重和幂等性

### 扩展能力
- 支持添加新的AI提供商
- 灵活的负载均衡策略
- 可配置的配额策略
- 插件化的中间件系统

### 高可用部署
- 多实例负载均衡
- 数据库主从复制
- Redis集群部署
- 健康检查和自动恢复

## 🔍 故障排查

### 常见问题
1. **服务启动失败**: 检查配置文件和数据库连接
2. **API调用401错误**: 验证API密钥有效性
3. **提供商调用失败**: 检查提供商API密钥配置
4. **配额超限**: 检查用户配额设置和使用情况

### 日志查看
```bash
# 查看服务日志
docker-compose logs -f gateway

# 查看数据库
sqlite3 data/gateway.db
.tables
SELECT * FROM users;
```

## 💻 关键代码示例

### 1. 网关服务核心逻辑
```go
// ProcessRequest 处理AI请求的核心流程
func (g *gatewayServiceImpl) ProcessRequest(ctx context.Context, request *GatewayRequest) (*GatewayResponse, error) {
    // 1. 路由请求到合适的提供商
    routeResponse, err := g.router.RouteRequest(ctx, routeRequest)
    if err != nil {
        return nil, fmt.Errorf("failed to route request: %w", err)
    }

    // 2. 计算使用量和成本
    usage := g.calculateUsage(routeResponse.Response)
    cost := g.calculateCost(routeResponse.Model, usage)

    // 3. 记录使用日志
    usageLog := g.recordUsageLog(ctx, request, routeResponse.Provider, routeResponse.Model, usage, cost, routeResponse.Duration, nil)

    // 4. 消费配额
    g.consumeQuotas(ctx, request.UserID, usage, cost)

    // 5. 处理计费
    if usageLog != nil {
        g.processBilling(ctx, request.UserID, usageLog.ID, cost.TotalCost)
    }

    return &GatewayResponse{
        Response:  routeResponse.Response,
        Usage:     usage,
        Cost:      cost,
        Provider:  routeResponse.Provider.Name,
        Model:     routeResponse.Model.Name,
        Duration:  routeResponse.Duration,
        RequestID: request.RequestID,
    }, nil
}
```

### 2. 负载均衡器实现
```go
// SelectProvider 根据策略选择提供商
func (lb *loadBalancerImpl) SelectProvider(ctx context.Context, providers []*entities.Provider) (*entities.Provider, error) {
    // 过滤健康的提供商
    healthyProviders := make([]*entities.Provider, 0, len(providers))
    for _, provider := range providers {
        if provider.IsAvailable() {
            healthyProviders = append(healthyProviders, provider)
        }
    }

    if len(healthyProviders) == 0 {
        return nil, fmt.Errorf("no healthy providers available")
    }

    // 根据策略选择提供商
    switch lb.strategy {
    case StrategyRoundRobin:
        return lb.selectRoundRobin(healthyProviders), nil
    case StrategyWeighted:
        return lb.selectWeighted(healthyProviders), nil
    case StrategyLeastConnections:
        return lb.selectLeastConnections(healthyProviders), nil
    default:
        return lb.selectRoundRobin(healthyProviders), nil
    }
}
```

### 3. 配额检查中间件
```go
// CheckQuota 检查用户配额
func (m *QuotaMiddleware) CheckQuota() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetInt64("user_id")

        // 检查各种类型的配额
        quotaTypes := []entities.QuotaType{
            entities.QuotaTypeRequests,
            entities.QuotaTypeTokens,
            entities.QuotaTypeCost,
        }

        for _, quotaType := range quotaTypes {
            canProceed, err := m.quotaService.CheckQuota(c.Request.Context(), userID, quotaType, 1.0)
            if err != nil {
                c.JSON(http.StatusInternalServerError, dto.ErrorResponse("QUOTA_CHECK_FAILED", "Failed to check quota", nil))
                c.Abort()
                return
            }

            if !canProceed {
                c.JSON(http.StatusTooManyRequests, dto.ErrorResponse("QUOTA_EXCEEDED", fmt.Sprintf("%s quota exceeded", quotaType), nil))
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

## 🗂️ 项目文件结构详解

### 命令行工具 (cmd/)
```
cmd/
├── server/           # 主服务器启动程序
├── migrate/          # 数据库迁移工具
├── e2etest/          # 端到端测试工具
├── checkapi/         # API连通性检查
├── checkdb/          # 数据库状态检查
├── setupquotas/      # 配额初始化工具
├── pricing/          # 定价数据管理
├── modelsupport/     # 模型支持配置
└── addprovider/      # 添加提供商工具
```

### 数据库迁移 (migrations/)
```
migrations/
├── 001_initial_schema.up.sql      # 初始数据库架构
├── 002_simplify_api_key_storage.up.sql  # 简化API密钥存储
├── 003_insert_model_pricing_data.up.sql # 插入模型定价数据
├── 004_provider_model_support.up.sql    # 提供商模型支持
├── 005_add_table_comments.up.sql        # 添加表注释
└── 007_add_pricing_multiplier.up.sql    # 添加价格倍率
```

### 配置和文档 (configs/, docs/)
```
configs/
└── config.yaml       # 主配置文件

docs/
├── API.md            # API文档
├── SWAGGER_GUIDE.md  # Swagger使用指南
├── swagger.json      # Swagger规范文件
└── swagger.yaml      # Swagger YAML格式
```

## 🔧 实用工具和脚本

### 测试脚本
- `test_service.py` - Python测试脚本，全面测试各个API端点
- `test_service.sh` - Shell测试脚本
- `test_chat_completions.sh` - 聊天完成API测试

### 管理工具使用示例
```bash
# 检查系统状态
go run cmd/checkapi/main.go
go run cmd/checkdb/main.go
go run cmd/checkproviders/main.go

# 初始化配额
go run cmd/setupquotas/main.go

# 添加新提供商
go run cmd/addprovider/main.go

# 配置模型支持
go run cmd/modelsupport/main.go

# 运行E2E测试
go run cmd/e2etest/main.go -action=setup
go run cmd/e2etest/main.go -action=test -apikey=YOUR_API_KEY
```

## 📊 监控和指标

### Prometheus指标示例
```
# 请求总数
gateway_requests_total{method="POST",endpoint="/v1/chat/completions",status="200"} 1234

# 响应时间
gateway_request_duration_seconds{method="POST",endpoint="/v1/chat/completions"} 0.123

# 提供商健康状态
gateway_provider_health{provider="openai"} 1

# 配额使用情况
gateway_quota_usage{user_id="1",quota_type="requests"} 45
```

### 日志格式示例
```json
{
  "timestamp": "2025-01-04T10:00:00Z",
  "level": "info",
  "message": "HTTP Request",
  "request_id": "req_1234567890",
  "user_id": 1,
  "api_key_id": 1,
  "method": "POST",
  "path": "/v1/chat/completions",
  "status": 200,
  "latency": "123ms",
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "tokens": 150,
  "cost": 0.0003
}
```

## 📝 后续开发建议

### 功能增强
- [ ] 流式响应支持优化
- [ ] 更多AI提供商集成 (Claude, Gemini, 文心一言等)
- [ ] 高级路由策略 (基于成本、延迟的智能路由)
- [ ] 缓存机制优化 (响应缓存、模型信息缓存)
- [ ] 批量请求支持
- [ ] 图像和音频模型支持

### 运维改进
- [ ] 完整的监控仪表板 (Grafana Dashboard)
- [ ] 告警系统集成 (AlertManager)
- [ ] 日志聚合和分析 (ELK Stack)
- [ ] 自动化部署流程 (CI/CD Pipeline)
- [ ] 性能基准测试和压力测试
- [ ] 蓝绿部署支持

### 安全加固
- [ ] API密钥加密存储 (AES加密)
- [ ] 请求签名验证 (HMAC签名)
- [ ] IP白名单功能
- [ ] 审计日志完善
- [ ] OAuth2.0集成
- [ ] 请求频率异常检测

### 性能优化
- [ ] 连接池优化
- [ ] 内存缓存优化
- [ ] 数据库查询优化
- [ ] 异步处理优化
- [ ] 负载测试和性能调优

---

**文档版本**: v1.0
**最后更新**: 2025-01-04
**维护者**: AI API Gateway Team

> 这份文档提供了AI API Gateway项目的完整技术概览和实施指南。项目采用现代化的Go开发实践，具有良好的可扩展性和可维护性，已具备生产环境部署的基础条件。如需了解具体实现细节，请参考源代码和相关技术文档。
