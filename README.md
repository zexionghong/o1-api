# AI API Gateway

一个高性能的AI API网关/代理服务，支持多提供商、配额管理、计费系统和负载均衡。

## 功能特性

- **API请求转发**：智能路由到多个AI服务提供商
- **配额管理**：API密钥级别的使用配额和速率限制
- **计费系统**：详细的使用日志和成本计算
- **多提供商支持**：支持OpenAI、Anthropic等多个AI服务
- **模型管理**：维护模型信息和定价数据
- **负载均衡**：智能分配请求，支持故障转移
- **高可用性**：健康检查和自动恢复机制

## 技术架构

- **语言**：Go 1.21+
- **架构**：清洁架构（Clean Architecture）
- **数据库**：SQLite（开发）/ PostgreSQL（生产）
- **设计原则**：无外键约束，应用层关系管理

## 项目结构

```
ai-api-gateway/
├── cmd/                    # 应用程序入口
│   └── server/            # 服务器启动代码
├── internal/              # 私有应用代码
│   ├── domain/           # 领域层（实体、业务规则）
│   ├── application/      # 应用层（用例、服务）
│   ├── infrastructure/   # 基础设施层（数据库、外部API）
│   └── presentation/     # 表现层（HTTP处理）
├── pkg/                  # 公共库代码
├── configs/              # 配置文件
├── migrations/           # 数据库迁移
├── docs/                 # 文档
└── tests/                # 测试文件
```

## 快速开始

1. 克隆项目
2. 安装依赖：`go mod tidy`
3. 运行迁移：`go run cmd/migrate/main.go`
4. 启动服务：`go run cmd/server/main.go`

## API文档

详见 [docs/api.md](docs/api.md)

## 许可证

MIT License
