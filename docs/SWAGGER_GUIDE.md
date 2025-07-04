# AI API Gateway Swagger 文档使用指南

## 🚀 快速开始

### 访问 Swagger 文档

启动服务器后，您可以通过以下地址访问 Swagger 文档：

```
http://localhost:8080/swagger/index.html
```

### 主要功能

✅ **完整的 API 文档** - 所有接口的详细说明和参数
✅ **在线调试功能** - 直接在浏览器中测试 API
✅ **API Key 认证** - 支持标准的 Bearer Token 认证
✅ **请求/响应示例** - 完整的数据格式说明
✅ **错误代码说明** - 详细的错误处理信息

## 🔐 API Key 认证设置

### 1. 获取 API Key

首先需要创建用户并生成 API Key：

```bash
# 创建用户
curl -X POST http://localhost:8080/admin/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_user",
    "email": "test@example.com",
    "balance": 100.0
  }'

# 创建 API Key
curl -X POST http://localhost:8080/admin/api-keys \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "name": "Test API Key"
  }'
```

### 2. 在 Swagger UI 中设置认证

1. 打开 Swagger UI：`http://localhost:8080/swagger/index.html`
2. 点击页面右上角的 **"Authorize"** 按钮
3. 在弹出的对话框中输入：`Bearer YOUR_API_KEY`
4. 点击 **"Authorize"** 确认
5. 现在您可以测试需要认证的 API 接口

## 📋 API 接口分类

### AI 接口 (需要认证)
- `POST /v1/chat/completions` - 聊天补全
- `POST /v1/completions` - 文本补全  
- `GET /v1/models` - 列出可用模型
- `GET /v1/usage` - 获取使用统计

### 健康检查 (无需认证)
- `GET /health` - 整体健康检查
- `GET /health/ready` - 就绪检查
- `GET /health/live` - 存活检查
- `GET /health/stats` - 系统统计
- `GET /health/version` - 版本信息

### 监控 (无需认证)
- `GET /metrics` - Prometheus 监控指标

### 管理 API (无需认证，生产环境建议添加认证)
- 用户管理：`/admin/users/*`
- API Key 管理：`/admin/api-keys/*`

## 🧪 测试示例

### 聊天补全测试

1. 在 Swagger UI 中找到 `POST /v1/chat/completions`
2. 点击 **"Try it out"**
3. 输入请求体：

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 100,
  "temperature": 0.7
}
```

4. 点击 **"Execute"** 执行请求
5. 查看响应结果

### 流式响应测试

```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "user",
      "content": "Tell me a story"
    }
  ],
  "stream": true,
  "max_tokens": 200
}
```

## 🔧 开发者工具

### 生成 Swagger 文档

当您修改了 API 注释后，需要重新生成文档：

```bash
# 安装 swag 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g docs/swagger.go -o docs
```

### 添加新的 API 注释

在处理器方法上添加 Swagger 注释：

```go
// @Summary 接口摘要
// @Description 详细描述
// @Tags 标签名
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body RequestType true "请求参数"
// @Success 200 {object} ResponseType "成功响应"
// @Failure 400 {object} dto.Response "错误响应"
// @Router /api/path [post]
func (h *Handler) Method(c *gin.Context) {
    // 实现代码
}
```

## 📚 更多资源

- [Swagger 官方文档](https://swagger.io/docs/)
- [gin-swagger 文档](https://github.com/swaggo/gin-swagger)
- [swag 注释语法](https://github.com/swaggo/swag#declarative-comments-format)

## 🐛 常见问题

### Q: 为什么看不到某些接口？
A: 确保您已经添加了正确的 Swagger 注释并重新生成了文档。

### Q: 认证失败怎么办？
A: 检查 API Key 格式是否正确，应该是 `Bearer YOUR_API_KEY`。

### Q: 如何测试流式响应？
A: Swagger UI 不能很好地显示流式响应，建议使用 curl 或其他工具测试。

### Q: 如何添加新的响应模型？
A: 在相应的包中定义结构体，然后在注释中引用即可。
