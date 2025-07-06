import { useState, useEffect, useCallback, useRef } from 'react';
import { useApiPagination, useApi } from './useApi';
import { api } from '../services/api';

// 数据状态类型
export interface DataState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  lastFetched: Date | null;
}

// 数据获取选项
export interface DataFetchOptions {
  immediate?: boolean;
  refreshInterval?: number; // 自动刷新间隔（毫秒）
  staleTime?: number; // 数据过期时间（毫秒）
  cacheKey?: string; // 缓存键
  onSuccess?: (data: any) => void;
  onError?: (error: string) => void;
  retries?: number;
  retryDelay?: number;
}

// 简单内存缓存
class SimpleCache {
  private cache = new Map<string, { data: any; timestamp: number }>();

  set(key: string, data: any) {
    this.cache.set(key, { data, timestamp: Date.now() });
  }

  get(key: string, staleTime = 5 * 60 * 1000): any | null {
    const cached = this.cache.get(key);
    if (!cached) return null;
    
    if (Date.now() - cached.timestamp > staleTime) {
      this.cache.delete(key);
      return null;
    }
    
    return cached.data;
  }

  clear(key?: string) {
    if (key) {
      this.cache.delete(key);
    } else {
      this.cache.clear();
    }
  }
}

const dataCache = new SimpleCache();

// 通用数据获取Hook
export function useData<T = any>(
  fetcher: () => Promise<any>,
  options: DataFetchOptions = {}
) {
  const {
    immediate = true,
    refreshInterval,
    staleTime = 5 * 60 * 1000, // 5分钟
    cacheKey,
    onSuccess,
    onError,
    retries = 0,
    retryDelay = 1000,
  } = options;

  const [state, setState] = useState<DataState<T>>({
    data: null,
    loading: false,
    error: null,
    lastFetched: null,
  });

  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const mountedRef = useRef(true);

  // 检查缓存
  const getCachedData = useCallback(() => {
    if (!cacheKey) return null;
    return dataCache.get(cacheKey, staleTime);
  }, [cacheKey, staleTime]);

  // 设置缓存
  const setCachedData = useCallback((data: T) => {
    if (cacheKey) {
      dataCache.set(cacheKey, data);
    }
  }, [cacheKey]);

  // 执行数据获取
  const fetchData = useCallback(async (force = false) => {
    // 如果不强制刷新，先检查缓存
    if (!force) {
      const cachedData = getCachedData();
      if (cachedData) {
        setState(prev => ({
          ...prev,
          data: cachedData,
          loading: false,
          error: null,
          lastFetched: new Date(),
        }));
        onSuccess?.(cachedData);
        return cachedData;
      }
    }

    setState(prev => ({ ...prev, loading: true, error: null }));

    let attempt = 0;
    const maxAttempts = retries + 1;

    while (attempt < maxAttempts) {
      try {
        const response = await fetcher();
        
        if (!mountedRef.current) return;

        const data = response.success ? response.data : response;
        
        setState({
          data,
          loading: false,
          error: null,
          lastFetched: new Date(),
        });

        setCachedData(data);
        onSuccess?.(data);
        return data;
      } catch (error: any) {
        attempt++;
        
        if (attempt >= maxAttempts) {
          if (!mountedRef.current) return;
          
          const errorMessage = error.message || 'Failed to fetch data';
          setState(prev => ({
            ...prev,
            loading: false,
            error: errorMessage,
          }));
          onError?.(errorMessage);
          throw error;
        }
        
        // 等待重试
        if (attempt < maxAttempts) {
          await new Promise(resolve => setTimeout(resolve, retryDelay));
        }
      }
    }
  }, [fetcher, getCachedData, setCachedData, onSuccess, onError, retries, retryDelay]);

  // 刷新数据
  const refresh = useCallback(() => {
    return fetchData(true);
  }, [fetchData]);

  // 清除缓存
  const clearCache = useCallback(() => {
    if (cacheKey) {
      dataCache.clear(cacheKey);
    }
  }, [cacheKey]);

  // 设置自动刷新
  useEffect(() => {
    if (refreshInterval && refreshInterval > 0) {
      refreshIntervalRef.current = setInterval(() => {
        fetchData(true);
      }, refreshInterval);

      return () => {
        if (refreshIntervalRef.current) {
          clearInterval(refreshIntervalRef.current);
        }
      };
    }
  }, [refreshInterval, fetchData]);

  // 初始数据获取
  useEffect(() => {
    if (immediate) {
      fetchData();
    }
  }, [immediate, fetchData]);

  // 清理
  useEffect(() => {
    return () => {
      mountedRef.current = false;
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
      }
    };
  }, []);

  return {
    ...state,
    refresh,
    clearCache,
    isStale: state.lastFetched ? Date.now() - state.lastFetched.getTime() > staleTime : true,
  };
}

// API密钥数据获取Hook
export function useApiKeys(userId?: number) {
  const fetcher = useCallback(async () => {
    if (!userId) throw new Error('User ID is required');
    return api.get(`/admin/users/${userId}/api-keys`);
  }, [userId]);

  return useData(fetcher, {
    immediate: !!userId,
    cacheKey: userId ? `api-keys-${userId}` : undefined,
    staleTime: 2 * 60 * 1000, // 2分钟
  });
}

// 用户工具实例数据获取Hook
export function useUserTools() {
  const fetcher = useCallback(async () => {
    return api.get('/admin/tools/');
  }, []);

  return useData(fetcher, {
    cacheKey: 'user-tools',
    staleTime: 5 * 60 * 1000, // 5分钟
  });
}

// 模型列表数据获取Hook
export function useModels() {
  const fetcher = useCallback(async () => {
    return api.get('/tools/models');
  }, []);

  return useData(fetcher, {
    cacheKey: 'models',
    staleTime: 10 * 60 * 1000, // 10分钟
  });
}

// 工具类型数据获取Hook
export function useToolTypes() {
  const fetcher = useCallback(async () => {
    return api.get('/tools/types');
  }, []);

  return useData(fetcher, {
    cacheKey: 'tool-types',
    staleTime: 30 * 60 * 1000, // 30分钟
  });
}

// 使用日志分页数据Hook
export function useUsageLogs(apiKeyId: number, options: {
  page?: number;
  pageSize?: number;
  startDate?: string;
  endDate?: string;
} = {}) {
  const { page = 1, pageSize = 10, startDate, endDate } = options;
  
  const baseUrl = `/admin/api-keys/${apiKeyId}/usage-logs`;
  
  return useApiPagination(baseUrl, {
    pageSize,
    initialPage: page,
    immediate: !!apiKeyId,
  });
}

// 计费记录分页数据Hook
export function useBillingRecords(apiKeyId: number, options: {
  page?: number;
  pageSize?: number;
  startDate?: string;
  endDate?: string;
} = {}) {
  const { page = 1, pageSize = 10, startDate, endDate } = options;
  
  const baseUrl = `/admin/api-keys/${apiKeyId}/billing-records`;
  
  return useApiPagination(baseUrl, {
    pageSize,
    initialPage: page,
    immediate: !!apiKeyId,
  });
}

// 数据变更Hook（用于创建、更新、删除操作）
export function useDataMutation<T = any>(
  mutationFn: (...args: any[]) => Promise<any>,
  options: {
    onSuccess?: (data: any) => void;
    onError?: (error: string) => void;
    invalidateCache?: string[]; // 需要清除的缓存键
  } = {}
) {
  const { onSuccess, onError, invalidateCache = [] } = options;
  
  const { execute, loading, error, data, success } = useApi(mutationFn, {
    onSuccess: (data) => {
      // 清除相关缓存
      invalidateCache.forEach(key => {
        dataCache.clear(key);
      });
      onSuccess?.(data);
    },
    onError,
  });

  return {
    mutate: execute,
    loading,
    error,
    data,
    success,
  };
}
