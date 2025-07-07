// API配置文件
export const API_CONFIG = {
  // 基础URL
  BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  
  // 超时时间
  TIMEOUT: Number(import.meta.env.VITE_API_TIMEOUT) || 30000,
  
  // 不需要认证的接口列表
  NO_AUTH_ENDPOINTS: [
    '/auth/login',
    '/auth/register',
    '/auth/refresh',
    '/health',
    '/swagger',
    '/docs'
  ],
  
  // 默认headers
  DEFAULT_HEADERS: {
    'Content-Type': 'application/json',
  },
  
  // 重试配置
  RETRY: {
    MAX_RETRIES: 3,
    RETRY_DELAY: 1000, // 毫秒
  },
  
  // 缓存配置
  CACHE: {
    DEFAULT_STALE_TIME: 5 * 60 * 1000, // 5分钟
    API_KEYS_STALE_TIME: 2 * 60 * 1000, // 2分钟
    MODELS_STALE_TIME: 10 * 60 * 1000, // 10分钟
    TOOL_TYPES_STALE_TIME: 30 * 60 * 1000, // 30分钟
  },
  
  // 分页配置
  PAGINATION: {
    DEFAULT_PAGE_SIZE: 10,
    MAX_PAGE_SIZE: 100,
  },
  
  // 环境相关配置
  IS_DEVELOPMENT: import.meta.env.DEV,
  IS_PRODUCTION: import.meta.env.PROD,
  DEBUG: import.meta.env.VITE_DEBUG === 'true',
  LOG_LEVEL: import.meta.env.VITE_LOG_LEVEL || 'info',
};

// API端点常量
export const API_ENDPOINTS = {
  // 认证相关
  AUTH: {
    LOGIN: '/auth/login',
    REGISTER: '/auth/register',
    REFRESH: '/auth/refresh',
    PROFILE: '/auth/profile',
    CHANGE_PASSWORD: '/auth/change-password',
    RECHARGE: '/auth/recharge',
  },
  
  // 用户管理
  USERS: {
    LIST: '/admin/users',
    DETAIL: (id: number) => `/admin/users/${id}`,
    API_KEYS: (id: number) => `/admin/users/${id}/api-keys`,
  },
  
  // API密钥管理
  API_KEYS: {
    LIST: '/admin/api-keys',
    DETAIL: (id: string) => `/admin/api-keys/${id}`,
    USAGE_LOGS: (id: string) => `/admin/api-keys/${id}/usage-logs`,
    BILLING_RECORDS: (id: string) => `/admin/api-keys/${id}/billing-records`,
  },
  
  // 工具管理
  TOOLS: {
    LIST: '/admin/tools',
    MODELS: '/tools/models',
    TYPES: '/tools/types',
  },
  
  // 系统相关
  SYSTEM: {
    HEALTH: '/health',
    USAGE: '/v1/usage',
  },
} as const;

// 错误码映射
export const ERROR_CODES = {
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  INSUFFICIENT_BALANCE: 'INSUFFICIENT_BALANCE',
  QUOTA_EXCEEDED: 'QUOTA_EXCEEDED',
  RATE_LIMIT_EXCEEDED: 'RATE_LIMIT_EXCEEDED',
} as const;

// 错误消息映射
export const ERROR_MESSAGES = {
  [ERROR_CODES.INVALID_CREDENTIALS]: '用户名或密码错误',
  [ERROR_CODES.TOKEN_EXPIRED]: '登录已过期，请重新登录',
  [ERROR_CODES.INSUFFICIENT_BALANCE]: '余额不足',
  [ERROR_CODES.QUOTA_EXCEEDED]: '配额已用完',
  [ERROR_CODES.RATE_LIMIT_EXCEEDED]: '请求过于频繁，请稍后再试',
} as const;
