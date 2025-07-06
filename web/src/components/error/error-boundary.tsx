import React from 'react';
import {
  Box,
  Typography,
  Button,
  Alert,
  AlertTitle,
  Paper,
  Stack,
} from '@mui/material';
import { Refresh as RefreshIcon, Home as HomeIcon } from '@mui/icons-material';
import { useTranslation } from 'react-i18next';

// 错误信息类型
export interface ErrorInfo {
  componentStack: string;
  errorBoundary?: string;
  errorBoundaryStack?: string;
}

// 错误边界状态类型
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

// 错误边界属性类型
interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<ErrorFallbackProps>;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  resetOnPropsChange?: boolean;
  resetKeys?: Array<string | number>;
}

// 错误回退组件属性类型
export interface ErrorFallbackProps {
  error: Error;
  errorInfo: ErrorInfo;
  resetError: () => void;
}

// 默认错误回退组件
function DefaultErrorFallback({ error, resetError }: ErrorFallbackProps) {
  const { t } = useTranslation();

  const handleGoHome = () => {
    window.location.href = '/';
  };

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '50vh',
        p: 3,
      }}
    >
      <Paper
        elevation={3}
        sx={{
          p: 4,
          maxWidth: 600,
          width: '100%',
          textAlign: 'center',
        }}
      >
        <Typography variant="h4" color="error" gutterBottom>
          {t('error.something_went_wrong')}
        </Typography>
        
        <Typography variant="body1" color="text.secondary" paragraph>
          {t('error.unexpected_error_occurred')}
        </Typography>

        <Alert severity="error" sx={{ mb: 3, textAlign: 'left' }}>
          <AlertTitle>{t('error.error_details')}</AlertTitle>
          <Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap' }}>
            {error.message}
          </Typography>
        </Alert>

        <Stack direction="row" spacing={2} justifyContent="center">
          <Button
            variant="contained"
            startIcon={<RefreshIcon />}
            onClick={resetError}
          >
            {t('error.try_again')}
          </Button>
          
          <Button
            variant="outlined"
            startIcon={<HomeIcon />}
            onClick={handleGoHome}
          >
            {t('error.go_home')}
          </Button>
        </Stack>
      </Paper>
    </Box>
  );
}

// 错误边界组件
export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  private resetTimeoutId: number | null = null;

  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.setState({
      error,
      errorInfo,
    });

    // 调用错误回调
    this.props.onError?.(error, errorInfo);

    // 记录错误到控制台
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  componentDidUpdate(prevProps: ErrorBoundaryProps) {
    const { resetKeys, resetOnPropsChange } = this.props;
    const { hasError } = this.state;

    if (hasError && prevProps.resetKeys !== resetKeys) {
      if (resetKeys?.some((key, index) => key !== prevProps.resetKeys?.[index])) {
        this.resetError();
      }
    }

    if (hasError && resetOnPropsChange && prevProps.children !== this.props.children) {
      this.resetError();
    }
  }

  componentWillUnmount() {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId);
    }
  }

  resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render() {
    const { hasError, error, errorInfo } = this.state;
    const { children, fallback: Fallback = DefaultErrorFallback } = this.props;

    if (hasError && error && errorInfo) {
      return (
        <Fallback
          error={error}
          errorInfo={errorInfo}
          resetError={this.resetError}
        />
      );
    }

    return children;
  }
}

// 错误处理Hook
export interface UseErrorHandlerReturn {
  error: string | null;
  setError: (error: string | null) => void;
  clearError: () => void;
  handleError: (error: any) => void;
}

export function useErrorHandler(): UseErrorHandlerReturn {
  const [error, setError] = React.useState<string | null>(null);

  const clearError = React.useCallback(() => {
    setError(null);
  }, []);

  const handleError = React.useCallback((err: any) => {
    let errorMessage = 'An unexpected error occurred';

    if (typeof err === 'string') {
      errorMessage = err;
    } else if (err?.message) {
      errorMessage = err.message;
    } else if (err?.response?.data?.error?.message) {
      errorMessage = err.response.data.error.message;
    } else if (err?.response?.data?.message) {
      errorMessage = err.response.data.message;
    }

    setError(errorMessage);
    console.error('Error handled:', err);
  }, []);

  return {
    error,
    setError,
    clearError,
    handleError,
  };
}

// 错误显示组件
export interface ErrorDisplayProps {
  error: string | null;
  onClose?: () => void;
  severity?: 'error' | 'warning' | 'info';
  variant?: 'filled' | 'outlined' | 'standard';
  sx?: any;
}

export function ErrorDisplay({
  error,
  onClose,
  severity = 'error',
  variant = 'filled',
  sx,
}: ErrorDisplayProps) {
  if (!error) {
    return null;
  }

  return (
    <Alert
      severity={severity}
      variant={variant}
      onClose={onClose}
      sx={{ mb: 2, ...sx }}
    >
      {error}
    </Alert>
  );
}

// 全局错误处理器
export class GlobalErrorHandler {
  private static instance: GlobalErrorHandler;
  private errorHandlers: Array<(error: any) => void> = [];

  private constructor() {
    // 监听未捕获的Promise错误
    window.addEventListener('unhandledrejection', (event) => {
      this.handleError(event.reason);
      event.preventDefault();
    });

    // 监听未捕获的JavaScript错误
    window.addEventListener('error', (event) => {
      this.handleError(event.error);
    });
  }

  static getInstance(): GlobalErrorHandler {
    if (!GlobalErrorHandler.instance) {
      GlobalErrorHandler.instance = new GlobalErrorHandler();
    }
    return GlobalErrorHandler.instance;
  }

  addErrorHandler(handler: (error: any) => void) {
    this.errorHandlers.push(handler);
  }

  removeErrorHandler(handler: (error: any) => void) {
    const index = this.errorHandlers.indexOf(handler);
    if (index > -1) {
      this.errorHandlers.splice(index, 1);
    }
  }

  handleError(error: any) {
    console.error('Global error:', error);
    this.errorHandlers.forEach(handler => {
      try {
        handler(error);
      } catch (err) {
        console.error('Error in error handler:', err);
      }
    });
  }
}

// 获取全局错误处理器实例
export const globalErrorHandler = GlobalErrorHandler.getInstance();
