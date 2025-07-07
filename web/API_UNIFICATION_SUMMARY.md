# API统一实现总结

## 概述

本次更新统一了前端项目的API实现，默认注入token到header，并为登录注册等接口提供了特殊处理。

## 主要改进

### 1. 统一API服务 (`src/services/api.ts`)

#### 新增功能：
- **自动token注入**：默认为所有请求注入Authorization header
- **智能跳过认证**：自动识别不需要认证的接口
- **配置化管理**：使用配置文件管理API相关设置
- **类型安全**：扩展了AxiosRequestConfig以支持skipAuth选项

#### 不需要认证的接口列表：
```typescript
const NO_AUTH_ENDPOINTS = [
  '/auth/login',
  '/auth/register', 
  '/auth/refresh',
  '/health',
  '/swagger',
  '/docs'
];
```

#### 新增API方法：
```typescript
// 普通API调用（默认注入token）
api.get(url, config)
api.post(url, data, config)
api.put(url, data, config)
api.delete(url, config)
api.patch(url, data, config)

// 明确跳过认证的API调用
api.noAuth.get(url, config)
api.noAuth.post(url, data, config)
// ... 其他方法
```

### 2. 认证服务优化 (`src/services/auth.ts`)

#### 更新内容：
- 登录、注册、刷新token接口使用`api.noAuth`方法
- 确保认证相关接口不会意外注入token
- 保持向后兼容性

### 3. 环境配置

#### 新增配置文件：
- `.env.development` - 开发环境配置
- `.env.production` - 生产环境配置  
- `src/config/api.ts` - API配置管理

#### 配置内容：
```typescript
// API配置
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL,
  TIMEOUT: 30000,
  NO_AUTH_ENDPOINTS: [...],
  DEFAULT_HEADERS: { 'Content-Type': 'application/json' },
  // ... 其他配置
};

// API端点常量
export const API_ENDPOINTS = {
  AUTH: { LOGIN: '/auth/login', ... },
  USERS: { LIST: '/admin/users', ... },
  // ... 其他端点
};
```

### 4. 全面替换fetch调用

#### 已替换的文件：
- `src/sections/api-keys/view/api-keys-view.tsx`
- `src/sections/api-keys/api-key-detail-dialog.tsx`
- `src/sections/overview/view/real-dashboard-view.tsx`
- `src/sections/tools/view/tools-view.tsx`
- `src/sections/tools/tool-create-dialog.tsx`
- `src/sections/tools/tool-edit-dialog.tsx`
- `src/sections/tools/tool-launch-dialog.tsx`

#### 替换示例：
```typescript
// 之前的写法
const response = await fetch('http://localhost:8080/api/endpoint', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
});

// 现在的写法
const response = await api.get('/api/endpoint');
```

## 使用指南

### 1. 普通API调用（需要认证）
```typescript
import api from 'src/services/api';

// GET请求
const response = await api.get('/admin/users');

// POST请求
const response = await api.post('/admin/tools', { name: 'Tool Name' });
```

### 2. 不需要认证的API调用
```typescript
// 方式1：使用noAuth方法（推荐）
const response = await api.noAuth.get('/tools/models');

// 方式2：使用skipAuth配置
const response = await api.get('/tools/models', { skipAuth: true });
```

### 3. 错误处理
```typescript
try {
  const response = await api.get('/api/endpoint');
  if (response.success && response.data) {
    // 处理成功响应
    console.log(response.data);
  } else {
    // 处理业务错误
    console.error(response.error?.message);
  }
} catch (error) {
  // 处理网络错误等
  console.error('API调用失败:', error);
}
```

## 向后兼容性

- 保持了原有的API响应格式
- 现有的错误处理逻辑无需修改
- 自动token刷新机制继续工作
- 401错误处理和重定向逻辑保持不变

## 环境配置

### 开发环境
- 自动使用`.env.development`配置
- API_BASE_URL默认为`http://localhost:8080`
- 启用调试模式

### 生产环境  
- 自动使用`.env.production`配置
- 需要配置正确的生产API地址
- 禁用调试模式

## 构建验证

✅ TypeScript编译通过
✅ Vite构建成功
✅ 所有fetch调用已替换为统一API
✅ 认证逻辑正常工作
✅ 环境配置自动切换

## 下一步建议

1. **测试验证**：建议编写单元测试验证API调用逻辑
2. **文档更新**：更新API使用文档和开发指南
3. **监控添加**：考虑添加API调用监控和错误追踪
4. **性能优化**：可以考虑添加请求缓存和去重机制
