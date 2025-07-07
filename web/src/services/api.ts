import type { AxiosInstance, AxiosResponse, AxiosRequestConfig } from 'axios';

import axios from 'axios';

import { API_CONFIG } from 'src/config/api';

// 扩展AxiosRequestConfig以支持skipAuth选项
export interface ApiRequestConfig extends AxiosRequestConfig {
  skipAuth?: boolean; // 跳过token注入
}

// 创建axios实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  timeout: API_CONFIG.TIMEOUT,
  headers: API_CONFIG.DEFAULT_HEADERS,
});

// 检查是否需要跳过认证
const shouldSkipAuth = (url: string, config?: ApiRequestConfig): boolean => {
  // 如果明确指定跳过认证
  if (config?.skipAuth) {
    return true;
  }

  // 检查是否是不需要认证的接口
  return API_CONFIG.NO_AUTH_ENDPOINTS.some((endpoint: string) => url.startsWith(endpoint));
};

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    const url = config.url || '';

    // 检查是否需要跳过认证
    if (!shouldSkipAuth(url, config as ApiRequestConfig)) {
      // 从localStorage获取token
      const token = localStorage.getItem('access_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }

    // 添加请求ID用于追踪
    config.headers['X-Request-ID'] = `web-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;

    console.log('API Request:', {
      method: config.method?.toUpperCase(),
      url: config.url,
      baseURL: config.baseURL,
      skipAuth: shouldSkipAuth(url, config as ApiRequestConfig),
      headers: config.headers,
    });

    return config;
  },
  (error) => {
    console.error('Request interceptor error:', error);
    return Promise.reject(error);
  }
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    console.log('API Response:', {
      status: response.status,
      url: response.config.url,
      data: response.data,
    });
    return response;
  },
  async (error) => {
    console.error('API Error:', {
      status: error.response?.status,
      url: error.config?.url,
      message: error.message,
      data: error.response?.data,
    });

    // 处理401错误（未授权）
    if (error.response?.status === 401) {
      // 尝试刷新token
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken && !error.config._retry) {
        error.config._retry = true;
        
        try {
          const response = await axios.post(`${API_CONFIG.BASE_URL}/auth/refresh`, {
            refresh_token: refreshToken,
          });
          
          const { access_token, refresh_token: newRefreshToken } = response.data;
          
          // 更新存储的token
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', newRefreshToken);
          
          // 重新发送原始请求
          error.config.headers.Authorization = `Bearer ${access_token}`;
          return apiClient.request(error.config);
        } catch (refreshError) {
          // 刷新失败，清除token并跳转到登录页
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          window.location.href = '/sign-in';
          return Promise.reject(refreshError);
        }
      } else {
        // 没有refresh token或已经重试过，跳转到登录页
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/sign-in';
      }
    }

    return Promise.reject(error);
  }
);

// API响应类型定义
export interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: Record<string, any>;
  };
  timestamp: string;
}

// API请求方法封装
export const api = {
  // GET请求
  get: <T = any>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
    apiClient.get(url, config).then(response => response.data),

  // POST请求
  post: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
    apiClient.post(url, data, config).then(response => response.data),

  // PUT请求
  put: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
    apiClient.put(url, data, config).then(response => response.data),

  // DELETE请求
  delete: <T = any>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
    apiClient.delete(url, config).then(response => response.data),

  // PATCH请求
  patch: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
    apiClient.patch(url, data, config).then(response => response.data),

  // 不需要认证的请求方法（明确跳过token注入）
  noAuth: {
    get: <T = any>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
      apiClient.get(url, { ...config, skipAuth: true } as ApiRequestConfig).then(response => response.data),

    post: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
      apiClient.post(url, data, { ...config, skipAuth: true } as ApiRequestConfig).then(response => response.data),

    put: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
      apiClient.put(url, data, { ...config, skipAuth: true } as ApiRequestConfig).then(response => response.data),

    delete: <T = any>(url: string, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
      apiClient.delete(url, { ...config, skipAuth: true } as ApiRequestConfig).then(response => response.data),

    patch: <T = any>(url: string, data?: any, config?: ApiRequestConfig): Promise<ApiResponse<T>> =>
      apiClient.patch(url, data, { ...config, skipAuth: true } as ApiRequestConfig).then(response => response.data),
  }
};

// 导出axios实例供特殊情况使用
export { apiClient };
export default api;
