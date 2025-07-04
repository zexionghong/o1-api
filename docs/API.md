# AI API Gateway API 文档

## 概述

AI API Gateway 提供了统一的AI服务接口，兼容OpenAI API格式，支持多个AI提供商的负载均衡和故障转移。

## 认证

所有API请求都需要提供有效的API密钥。API密钥可以通过以下方式提供：

1. **Authorization头** (推荐)
```
Authorization: Bearer YOUR_API_KEY
```

2. **X-API-Key头**
```
X-API-Key: YOUR_API_KEY
```

3. **查询参数**
```
?api_key=YOUR_API_KEY
```

## AI API 接口

### 聊天完成 (Chat Completions)

创建聊天完成请求。

**端点**: `POST /v1/chat/completions`

**请求体**:
```json
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user", 
      "content": "Hello!"
    }
  ],
  "max_tokens": 100,
  "temperature": 0.7,
  "stream": false
}
```

**响应**:
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-3.5-turbo",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 10,
    "total_tokens": 30
  }
}
```

### 文本完成 (Completions)

创建文本完成请求。

**端点**: `POST /v1/completions`

**请求体**:
```json
{
  "model": "text-davinci-003",
  "prompt": "Say this is a test",
  "max_tokens": 7,
  "temperature": 0
}
```

**响应**:
```json
{
  "id": "cmpl-123",
  "object": "text_completion",
  "created": 1677652288,
  "model": "text-davinci-003",
  "choices": [
    {
      "text": "\n\nThis is a test.",
      "index": 0,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 7,
    "total_tokens": 12
  }
}
```

### 获取模型列表

获取可用的AI模型列表。

**端点**: `GET /v1/models`

**响应**:
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-3.5-turbo",
      "object": "model",
      "created": 1677610602,
      "owned_by": "openai"
    },
    {
      "id": "gpt-4",
      "object": "model", 
      "created": 1687882411,
      "owned_by": "openai"
    }
  ]
}
```

## 管理 API 接口

### 用户管理

#### 创建用户

**端点**: `POST /admin/users`

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "full_name": "Test User"
}
```

**响应**:
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "status": "active",
    "balance": 0.0,
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z"
  }
}
```

#### 获取用户

**端点**: `GET /admin/users/{id}`

#### 更新用户

**端点**: `PUT /admin/users/{id}`

#### 删除用户

**端点**: `DELETE /admin/users/{id}`

#### 更新用户余额

**端点**: `POST /admin/users/{id}/balance`

**请求体**:
```json
{
  "amount": 100.0,
  "operation": "add",
  "description": "充值"
}
```

### API密钥管理

#### 创建API密钥

**端点**: `POST /admin/api-keys`

**请求体**:
```json
{
  "user_id": 1,
  "name": "My API Key",
  "permissions": {
    "allowed_providers": ["openai", "anthropic"],
    "allowed_models": ["gpt-3.5-turbo", "claude-3"]
  },
  "expires_at": "2024-12-01T00:00:00Z"
}
```

**响应**:
```json
{
  "success": true,
  "message": "API key created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "key": "ak_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
    "key_prefix": "ak_12345",
    "name": "My API Key",
    "status": "active",
    "permissions": {
      "allowed_providers": ["openai", "anthropic"],
      "allowed_models": ["gpt-3.5-turbo", "claude-3"]
    },
    "expires_at": "2024-12-01T00:00:00Z",
    "created_at": "2023-12-01T10:00:00Z",
    "updated_at": "2023-12-01T10:00:00Z"
  }
}
```

#### 撤销API密钥

**端点**: `POST /admin/api-keys/{id}/revoke`

## 健康检查接口

### 健康检查

**端点**: `GET /health`

**响应**:
```json
{
  "success": true,
  "message": "Health check passed",
  "data": {
    "status": "healthy",
    "timestamp": "2023-12-01T10:00:00Z",
    "providers": {
      "openai": {
        "status": "healthy",
        "response_time": "100ms",
        "last_check": "2023-12-01T09:59:00Z"
      }
    },
    "database": {
      "status": "healthy",
      "response_time": "5ms"
    }
  }
}
```

### 就绪检查

**端点**: `GET /health/ready`

### 存活检查

**端点**: `GET /health/live`

### 统计信息

**端点**: `GET /health/stats`

### 监控指标

**端点**: `GET /metrics`

返回Prometheus格式的监控指标。

## 错误处理

所有API错误都遵循统一的格式：

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description",
    "details": {
      "additional": "information"
    }
  },
  "timestamp": "2023-12-01T10:00:00Z"
}
```

### 常见错误码

- `MISSING_API_KEY`: 缺少API密钥
- `INVALID_API_KEY`: 无效的API密钥
- `API_KEY_EXPIRED`: API密钥已过期
- `API_KEY_INACTIVE`: API密钥未激活
- `USER_INACTIVE`: 用户账户未激活
- `RATE_LIMIT_EXCEEDED`: 超过速率限制
- `QUOTA_EXCEEDED`: 超过配额限制
- `INSUFFICIENT_BALANCE`: 余额不足
- `PROVIDER_PERMISSION_DENIED`: 提供商权限被拒绝
- `MODEL_PERMISSION_DENIED`: 模型权限被拒绝
- `REQUEST_FAILED`: 请求处理失败

## 速率限制

API实施多级速率限制：

- **IP级别**: 每分钟100请求
- **用户级别**: 每分钟60请求（默认）
- **API密钥级别**: 根据配置的限制

当达到速率限制时，响应会包含以下头部：

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1677652348
Retry-After: 30
```

## 配额管理

系统支持多维度配额控制：

1. **请求数配额**: 限制请求总数
2. **Token配额**: 限制Token使用量
3. **成本配额**: 限制总成本

配额信息会在响应头中返回：

```
X-Quota-Requests-Limit: 1000
X-Quota-Requests-Remaining: 950
X-Quota-Tokens-Limit: 100000
X-Quota-Tokens-Remaining: 95000
```

## SDK和示例

### cURL示例

```bash
# 聊天完成
curl -X POST https://api.example.com/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Python示例

```python
import requests

headers = {
    "Authorization": "Bearer YOUR_API_KEY",
    "Content-Type": "application/json"
}

data = {
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
}

response = requests.post(
    "https://api.example.com/v1/chat/completions",
    headers=headers,
    json=data
)

print(response.json())
```
