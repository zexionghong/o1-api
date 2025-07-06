import { useState, useCallback, useRef, useEffect } from 'react';
import { api, type ApiResponse } from '../services/api';

// API状态类型
export interface ApiState<T = any> {
  data: T | null;
  loading: boolean;
  error: string | null;
  success: boolean;
}

// API选项类型
export interface ApiOptions {
  immediate?: boolean; // 是否立即执行
  onSuccess?: (data: any) => void;
  onError?: (error: string) => void;
  retries?: number; // 重试次数
  retryDelay?: number; // 重试延迟（毫秒）
}

// 通用API调用Hook
export function useApi<T = any>(
  apiCall: (...args: any[]) => Promise<ApiResponse<T>>,
  options: ApiOptions = {}
) {
  const {
    immediate = false,
    onSuccess,
    onError,
    retries = 0,
    retryDelay = 1000
  } = options;

  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: false,
    error: null,
    success: false,
  });

  const abortControllerRef = useRef<AbortController | null>(null);
  const retryCountRef = useRef(0);

  // 执行API调用
  const execute = useCallback(async (...args: any[]) => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // 创建新的AbortController
    abortControllerRef.current = new AbortController();

    setState(prev => ({
      ...prev,
      loading: true,
      error: null,
      success: false,
    }));

    const attemptRequest = async (attemptCount: number): Promise<void> => {
      try {
        const response = await apiCall(...args);
        
        // 检查请求是否被取消
        if (abortControllerRef.current?.signal.aborted) {
          return;
        }

        if (response.success) {
          setState({
            data: response.data || null,
            loading: false,
            error: null,
            success: true,
          });
          onSuccess?.(response.data);
        } else {
          const errorMessage = response.error?.message || 'API调用失败';
          setState({
            data: null,
            loading: false,
            error: errorMessage,
            success: false,
          });
          onError?.(errorMessage);
        }
      } catch (error: any) {
        // 检查请求是否被取消
        if (abortControllerRef.current?.signal.aborted) {
          return;
        }

        const errorMessage = error.message || '网络错误';
        
        // 重试逻辑
        if (attemptCount < retries) {
          retryCountRef.current = attemptCount + 1;
          setTimeout(() => {
            attemptRequest(attemptCount + 1);
          }, retryDelay);
          return;
        }

        setState({
          data: null,
          loading: false,
          error: errorMessage,
          success: false,
        });
        onError?.(errorMessage);
      }
    };

    retryCountRef.current = 0;
    await attemptRequest(0);
  }, [apiCall, onSuccess, onError, retries, retryDelay]);

  // 重置状态
  const reset = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    setState({
      data: null,
      loading: false,
      error: null,
      success: false,
    });
    retryCountRef.current = 0;
  }, []);

  // 取消请求
  const cancel = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    setState(prev => ({
      ...prev,
      loading: false,
    }));
  }, []);

  // 立即执行
  useEffect(() => {
    if (immediate) {
      execute();
    }
    
    // 清理函数
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [immediate, execute]);

  return {
    ...state,
    execute,
    reset,
    cancel,
    retryCount: retryCountRef.current,
  };
}

// 专门用于GET请求的Hook
export function useApiGet<T = any>(
  url: string,
  options: ApiOptions = {}
) {
  const apiCall = useCallback(() => api.get<T>(url), [url]);
  return useApi<T>(apiCall, options);
}

// 专门用于POST请求的Hook
export function useApiPost<T = any>(
  url: string,
  options: ApiOptions = {}
) {
  const apiCall = useCallback((data: any) => api.post<T>(url, data), [url]);
  return useApi<T>(apiCall, options);
}

// 专门用于PUT请求的Hook
export function useApiPut<T = any>(
  url: string,
  options: ApiOptions = {}
) {
  const apiCall = useCallback((data: any) => api.put<T>(url, data), [url]);
  return useApi<T>(apiCall, options);
}

// 专门用于DELETE请求的Hook
export function useApiDelete<T = any>(
  url: string,
  options: ApiOptions = {}
) {
  const apiCall = useCallback(() => api.delete<T>(url), [url]);
  return useApi<T>(apiCall, options);
}

// 分页数据获取Hook
export function useApiPagination<T = any>(
  baseUrl: string,
  options: ApiOptions & {
    pageSize?: number;
    initialPage?: number;
  } = {}
) {
  const { pageSize = 10, initialPage = 1, ...apiOptions } = options;
  const [page, setPage] = useState(initialPage);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);

  const apiCall = useCallback(
    (currentPage: number) => {
      const params = new URLSearchParams({
        page: currentPage.toString(),
        page_size: pageSize.toString(),
      });
      return api.get<{
        data: T[];
        total: number;
        page: number;
        page_size: number;
        total_pages: number;
      }>(`${baseUrl}?${params}`);
    },
    [baseUrl, pageSize]
  );

  const {
    data: response,
    loading,
    error,
    success,
    execute,
    reset,
    cancel
  } = useApi(apiCall, {
    ...apiOptions,
    onSuccess: (data) => {
      setTotal(data.total || 0);
      setTotalPages(data.total_pages || 0);
      apiOptions.onSuccess?.(data);
    },
  });

  const loadPage = useCallback((newPage: number) => {
    setPage(newPage);
    execute(newPage);
  }, [execute]);

  const nextPage = useCallback(() => {
    if (page < totalPages) {
      loadPage(page + 1);
    }
  }, [page, totalPages, loadPage]);

  const prevPage = useCallback(() => {
    if (page > 1) {
      loadPage(page - 1);
    }
  }, [page, loadPage]);

  const refresh = useCallback(() => {
    execute(page);
  }, [execute, page]);

  return {
    data: response?.data || [],
    loading,
    error,
    success,
    page,
    totalPages,
    total,
    pageSize,
    loadPage,
    nextPage,
    prevPage,
    refresh,
    reset,
    cancel,
  };
}
