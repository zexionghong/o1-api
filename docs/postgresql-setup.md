# PostgreSQL 配置指南

本文档介绍如何配置AI API Gateway使用PostgreSQL数据库。

## 📋 目录

- [配置选项](#配置选项)
- [本地安装PostgreSQL](#本地安装postgresql)
- [Docker方式运行](#docker方式运行)
- [配置文件设置](#配置文件设置)
- [数据库迁移](#数据库迁移)
- [常见问题](#常见问题)

## 🔧 配置选项

### 方案1：修改现有配置文件

编辑 `configs/config.yaml`：

```yaml
database:
  # 数据库驱动: sqlite 或 postgres
  driver: "postgres"
  # PostgreSQL连接字符串
  dsn: "host=localhost port=5432 user=gateway password=gateway_password dbname=gateway sslmode=disable"
  # 连接池配置
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s
```

### 方案2：使用环境特定配置文件

- **开发环境**: 使用 `configs/config-dev.yaml` (SQLite)
- **生产环境**: 使用 `configs/config-prod.yaml` (PostgreSQL)

启动时指定配置文件：
```bash
# 开发环境
go run cmd/server/main.go -config configs/config-dev.yaml

# 生产环境
go run cmd/server/main.go -config configs/config-prod.yaml
```

## 🐘 本地安装PostgreSQL

### Windows

1. 下载PostgreSQL安装程序：https://www.postgresql.org/download/windows/
2. 运行安装程序，设置密码
3. 创建数据库和用户：

```sql
-- 连接到PostgreSQL (使用psql或pgAdmin)
CREATE USER gateway WITH PASSWORD 'gateway_password';
CREATE DATABASE gateway OWNER gateway;
GRANT ALL PRIVILEGES ON DATABASE gateway TO gateway;
```

### macOS

```bash
# 使用Homebrew安装
brew install postgresql
brew services start postgresql

# 创建数据库
createdb gateway
psql gateway -c "CREATE USER gateway WITH PASSWORD 'gateway_password';"
psql gateway -c "GRANT ALL PRIVILEGES ON DATABASE gateway TO gateway;"
```

### Linux (Ubuntu/Debian)

```bash
# 安装PostgreSQL
sudo apt update
sudo apt install postgresql postgresql-contrib

# 切换到postgres用户
sudo -u postgres psql

# 在PostgreSQL shell中执行
CREATE USER gateway WITH PASSWORD 'gateway_password';
CREATE DATABASE gateway OWNER gateway;
GRANT ALL PRIVILEGES ON DATABASE gateway TO gateway;
\q
```

## 🐳 Docker方式运行

### 使用Docker Compose（推荐）

项目已包含完整的Docker Compose配置：

```bash
# 启动所有服务（包括PostgreSQL）
docker-compose up -d

# 仅启动PostgreSQL
docker-compose up -d postgres
```

### 单独运行PostgreSQL容器

```bash
# 运行PostgreSQL容器
docker run --name postgres-gateway \
  -e POSTGRES_DB=gateway \
  -e POSTGRES_USER=gateway \
  -e POSTGRES_PASSWORD=gateway_password \
  -p 5432:5432 \
  -d postgres:15-alpine

# 验证连接
docker exec -it postgres-gateway psql -U gateway -d gateway -c "SELECT version();"
```

## ⚙️ 配置文件设置

### DSN连接字符串格式

```
host=主机地址 port=端口 user=用户名 password=密码 dbname=数据库名 sslmode=SSL模式
```

### 常用DSN示例

```yaml
# 本地开发（无SSL）
dsn: "host=localhost port=5432 user=gateway password=gateway_password dbname=gateway sslmode=disable"

# 生产环境（启用SSL）
dsn: "host=prod-db.example.com port=5432 user=gateway password=your_secure_password dbname=gateway sslmode=require"

# 使用连接池和超时设置
dsn: "host=localhost port=5432 user=gateway password=gateway_password dbname=gateway sslmode=disable connect_timeout=10"
```

### 连接池配置建议

```yaml
database:
  driver: "postgres"
  dsn: "your_dsn_here"
  # 生产环境建议配置
  max_open_conns: 50      # 最大打开连接数
  max_idle_conns: 10      # 最大空闲连接数
  conn_max_lifetime: 600s # 连接最大生存时间
```

## 🔄 数据库迁移

### 自动迁移（推荐）

使用智能迁移工具，自动根据配置选择数据库类型：

```bash
# 执行迁移（自动检测数据库类型）
make migrate-up

# 或者直接运行
go run cmd/migrate-auto/main.go -direction=up -config=configs/config.yaml
```

### 手动PostgreSQL迁移

```bash
# 使用PostgreSQL专用迁移工具
make migrate-postgres-up

# 或者直接运行
go run cmd/migrate-postgres/main.go -direction=up -dsn="your_dsn_here"
```

### 迁移文件位置

- SQLite迁移文件：`migrations/`
- PostgreSQL迁移文件：`migrations-postgres/`

## 🔍 验证配置

### 1. 测试数据库连接

```bash
# 运行数据库检查工具
go run cmd/checkdb/main.go -config configs/config.yaml
```

### 2. 启动服务

```bash
# 启动服务
go run cmd/server/main.go -config configs/config.yaml

# 检查健康状态
curl http://localhost:8080/health
```

## ❓ 常见问题

### Q: 连接被拒绝 (connection refused)

**A:** 检查PostgreSQL是否正在运行：
```bash
# 检查PostgreSQL状态
sudo systemctl status postgresql  # Linux
brew services list | grep postgresql  # macOS
```

### Q: 认证失败 (authentication failed)

**A:** 检查用户名和密码是否正确：
```bash
# 测试连接
psql -h localhost -p 5432 -U gateway -d gateway
```

### Q: 数据库不存在 (database does not exist)

**A:** 创建数据库：
```sql
CREATE DATABASE gateway OWNER gateway;
```

### Q: SSL连接问题

**A:** 根据环境调整SSL模式：
- 开发环境：`sslmode=disable`
- 生产环境：`sslmode=require` 或 `sslmode=verify-full`

### Q: 迁移失败

**A:** 检查迁移文件和权限：
```bash
# 检查迁移状态
go run cmd/migrate-auto/main.go -direction=version -config=configs/config.yaml

# 重置迁移（谨慎使用）
go run cmd/migrate-auto/main.go -direction=down -config=configs/config.yaml
go run cmd/migrate-auto/main.go -direction=up -config=configs/config.yaml
```

## 🚀 生产环境建议

1. **使用连接池**：合理设置连接池参数
2. **启用SSL**：生产环境必须使用SSL连接
3. **定期备份**：设置自动备份策略
4. **监控连接**：监控数据库连接数和性能
5. **使用专用用户**：不要使用超级用户连接
6. **网络安全**：限制数据库访问IP范围

## 📚 相关文档

- [PostgreSQL官方文档](https://www.postgresql.org/docs/)
- [Go pq驱动文档](https://pkg.go.dev/github.com/lib/pq)
- [Docker Compose配置](../docker-compose.yml)
