import React from 'react';
import {
  Box,
  Pagination as MuiPagination,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Typography,
  Stack,
} from '@mui/material';
import { useTranslation } from 'react-i18next';

// 分页组件属性类型
export interface PaginationProps {
  page: number;
  totalPages: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  showPageSizeSelector?: boolean;
  showTotal?: boolean;
  showPageInfo?: boolean;
  disabled?: boolean;
  size?: 'small' | 'medium' | 'large';
  color?: 'primary' | 'secondary' | 'standard';
  variant?: 'text' | 'outlined';
  shape?: 'circular' | 'rounded';
}

// 通用分页组件
export function Pagination({
  page,
  totalPages,
  total,
  pageSize,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 50, 100],
  showPageSizeSelector = true,
  showTotal = true,
  showPageInfo = true,
  disabled = false,
  size = 'medium',
  color = 'primary',
  variant = 'outlined',
  shape = 'rounded',
}: PaginationProps) {
  const { t } = useTranslation();

  // 计算当前显示的数据范围
  const startItem = totalPages > 0 ? (page - 1) * pageSize + 1 : 0;
  const endItem = Math.min(page * pageSize, total);

  const handlePageChange = (event: React.ChangeEvent<unknown>, newPage: number) => {
    onPageChange(newPage);
  };

  const handlePageSizeChange = (event: any) => {
    const newPageSize = parseInt(event.target.value, 10);
    onPageSizeChange?.(newPageSize);
  };

  if (totalPages <= 1 && !showTotal && !showPageSizeSelector) {
    return null;
  }

  return (
    <Box
      sx={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: 2,
        py: 2,
      }}
    >
      {/* 左侧：总数和页面大小选择器 */}
      <Stack direction="row" spacing={2} alignItems="center">
        {showTotal && (
          <Typography variant="body2" color="text.secondary">
            {t('pagination.total_items', { total })}
          </Typography>
        )}

        {showPageSizeSelector && onPageSizeChange && (
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>{t('pagination.page_size')}</InputLabel>
            <Select
              value={pageSize}
              label={t('pagination.page_size')}
              onChange={handlePageSizeChange}
              disabled={disabled}
            >
              {pageSizeOptions.map((option) => (
                <MenuItem key={option} value={option}>
                  {option} {t('pagination.items_per_page')}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        )}
      </Stack>

      {/* 中间：分页控件 */}
      {totalPages > 1 && (
        <MuiPagination
          count={totalPages}
          page={page}
          onChange={handlePageChange}
          disabled={disabled}
          size={size}
          color={color}
          variant={variant}
          shape={shape}
          showFirstButton
          showLastButton
        />
      )}

      {/* 右侧：页面信息 */}
      {showPageInfo && totalPages > 0 && (
        <Typography variant="body2" color="text.secondary">
          {t('pagination.page_info', {
            start: startItem,
            end: endItem,
            total,
            page,
            totalPages,
          })}
        </Typography>
      )}
    </Box>
  );
}

// 分页Hook
export interface UsePaginationOptions {
  initialPage?: number;
  initialPageSize?: number;
  total?: number;
  onPageChange?: (page: number, pageSize: number) => void;
}

export interface UsePaginationReturn {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  setPage: (page: number) => void;
  setPageSize: (pageSize: number) => void;
  setTotal: (total: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  firstPage: () => void;
  lastPage: () => void;
  canGoNext: boolean;
  canGoPrev: boolean;
  reset: () => void;
}

export function usePagination({
  initialPage = 1,
  initialPageSize = 10,
  total = 0,
  onPageChange,
}: UsePaginationOptions = {}): UsePaginationReturn {
  const [page, setPageState] = React.useState(initialPage);
  const [pageSize, setPageSizeState] = React.useState(initialPageSize);
  const [totalItems, setTotalItems] = React.useState(total);

  const totalPages = Math.ceil(totalItems / pageSize);
  const canGoNext = page < totalPages;
  const canGoPrev = page > 1;

  const setPage = React.useCallback((newPage: number) => {
    const validPage = Math.max(1, Math.min(newPage, totalPages));
    setPageState(validPage);
    onPageChange?.(validPage, pageSize);
  }, [totalPages, pageSize, onPageChange]);

  const setPageSize = React.useCallback((newPageSize: number) => {
    setPageSizeState(newPageSize);
    // 重新计算当前页，确保不超出范围
    const newTotalPages = Math.ceil(totalItems / newPageSize);
    const newPage = Math.min(page, newTotalPages || 1);
    setPageState(newPage);
    onPageChange?.(newPage, newPageSize);
  }, [page, totalItems, onPageChange]);

  const setTotal = React.useCallback((newTotal: number) => {
    setTotalItems(newTotal);
    // 如果当前页超出了新的总页数，调整到最后一页
    const newTotalPages = Math.ceil(newTotal / pageSize);
    if (page > newTotalPages && newTotalPages > 0) {
      setPage(newTotalPages);
    }
  }, [page, pageSize, setPage]);

  const nextPage = React.useCallback(() => {
    if (canGoNext) {
      setPage(page + 1);
    }
  }, [canGoNext, page, setPage]);

  const prevPage = React.useCallback(() => {
    if (canGoPrev) {
      setPage(page - 1);
    }
  }, [canGoPrev, page, setPage]);

  const firstPage = React.useCallback(() => {
    setPage(1);
  }, [setPage]);

  const lastPage = React.useCallback(() => {
    if (totalPages > 0) {
      setPage(totalPages);
    }
  }, [totalPages, setPage]);

  const reset = React.useCallback(() => {
    setPageState(initialPage);
    setPageSizeState(initialPageSize);
    setTotalItems(total);
  }, [initialPage, initialPageSize, total]);

  return {
    page,
    pageSize,
    total: totalItems,
    totalPages,
    setPage,
    setPageSize,
    setTotal,
    nextPage,
    prevPage,
    firstPage,
    lastPage,
    canGoNext,
    canGoPrev,
    reset,
  };
}

// 表格分页组件（专门用于表格底部）
export interface TablePaginationProps extends Omit<PaginationProps, 'showPageInfo'> {
  dense?: boolean;
}

export function TablePagination({
  dense = false,
  ...props
}: TablePaginationProps) {
  return (
    <Box
      sx={{
        borderTop: 1,
        borderColor: 'divider',
        bgcolor: 'background.paper',
        px: dense ? 1 : 2,
        py: dense ? 1 : 1.5,
      }}
    >
      <Pagination
        {...props}
        showPageInfo={true}
        size={dense ? 'small' : 'medium'}
      />
    </Box>
  );
}
