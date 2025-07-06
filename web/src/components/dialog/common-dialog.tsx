import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  IconButton,
  Typography,
  Box,
  CircularProgress,
  Divider,
} from '@mui/material';
import { Close as CloseIcon } from '@mui/icons-material';
import { useTranslation } from 'react-i18next';

// 对话框大小类型
export type DialogSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

// 按钮配置类型
export interface DialogButton {
  label: string;
  onClick: () => void | Promise<void> | Promise<boolean> | Promise<any>;
  color?: 'inherit' | 'primary' | 'secondary' | 'success' | 'error' | 'info' | 'warning';
  variant?: 'text' | 'outlined' | 'contained';
  disabled?: boolean;
  loading?: boolean;
  startIcon?: React.ReactNode;
  endIcon?: React.ReactNode;
}

// 通用对话框属性类型
export interface CommonDialogProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  subtitle?: string;
  children: React.ReactNode;
  actions?: DialogButton[];
  size?: DialogSize;
  fullScreen?: boolean;
  fullWidth?: boolean;
  disableEscapeKeyDown?: boolean;
  disableBackdropClick?: boolean;
  showCloseButton?: boolean;
  loading?: boolean;
  dividers?: boolean;
  maxWidth?: DialogSize | false;
  PaperProps?: any;
  TransitionProps?: any;
}

// 通用对话框组件
export function CommonDialog({
  open,
  onClose,
  title,
  subtitle,
  children,
  actions = [],
  size = 'sm',
  fullScreen = false,
  fullWidth = true,
  disableEscapeKeyDown = false,
  disableBackdropClick = false,
  showCloseButton = true,
  loading = false,
  dividers = false,
  maxWidth,
  PaperProps,
  TransitionProps,
}: CommonDialogProps) {
  const { t } = useTranslation();

  const handleClose = (event: any, reason: string) => {
    if (disableBackdropClick && reason === 'backdropClick') {
      return;
    }
    if (disableEscapeKeyDown && reason === 'escapeKeyDown') {
      return;
    }
    onClose();
  };

  const handleActionClick = async (action: DialogButton) => {
    if (action.disabled || action.loading) {
      return;
    }
    
    try {
      await action.onClick();
    } catch (error) {
      console.error('Dialog action error:', error);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth={maxWidth !== undefined ? maxWidth : size}
      fullWidth={fullWidth}
      fullScreen={fullScreen}
      PaperProps={PaperProps}
      TransitionProps={TransitionProps}
    >
      {/* 标题栏 */}
      {(title || showCloseButton) && (
        <DialogTitle
          sx={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            pb: subtitle ? 1 : 2,
          }}
        >
          <Box>
            {title && (
              <Typography variant="h6" component="div">
                {title}
              </Typography>
            )}
            {subtitle && (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                {subtitle}
              </Typography>
            )}
          </Box>
          
          {showCloseButton && (
            <IconButton
              aria-label={t('common.close')}
              onClick={onClose}
              sx={{ ml: 1 }}
              disabled={loading}
            >
              <CloseIcon />
            </IconButton>
          )}
        </DialogTitle>
      )}

      {/* 分割线 */}
      {dividers && (title || showCloseButton) && <Divider />}

      {/* 内容区域 */}
      <DialogContent
        dividers={dividers}
        sx={{
          position: 'relative',
          minHeight: loading ? 200 : 'auto',
        }}
      >
        {loading ? (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: 200,
            }}
          >
            <CircularProgress />
          </Box>
        ) : (
          children
        )}
      </DialogContent>

      {/* 操作按钮 */}
      {actions.length > 0 && (
        <>
          {dividers && <Divider />}
          <DialogActions sx={{ px: 3, py: 2 }}>
            {actions.map((action, index) => (
              <Button
                key={index}
                onClick={() => handleActionClick(action)}
                color={action.color || 'primary'}
                variant={action.variant || 'text'}
                disabled={action.disabled || loading}
                startIcon={action.loading ? <CircularProgress size={16} /> : action.startIcon}
                endIcon={!action.loading ? action.endIcon : undefined}
              >
                {action.label}
              </Button>
            ))}
          </DialogActions>
        </>
      )}
    </Dialog>
  );
}

// 确认对话框组件
export interface ConfirmDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  confirmColor?: 'primary' | 'secondary' | 'error' | 'warning';
  loading?: boolean;
}

export function ConfirmDialog({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmText,
  cancelText,
  confirmColor = 'primary',
  loading = false,
}: ConfirmDialogProps) {
  const { t } = useTranslation();

  const actions: DialogButton[] = [
    {
      label: cancelText || t('common.cancel'),
      onClick: onClose,
      variant: 'text',
      disabled: loading,
    },
    {
      label: confirmText || t('common.confirm'),
      onClick: onConfirm,
      color: confirmColor,
      variant: 'contained',
      loading,
    },
  ];

  return (
    <CommonDialog
      open={open}
      onClose={onClose}
      title={title || t('common.confirm')}
      actions={actions}
      size="xs"
      disableBackdropClick={loading}
      disableEscapeKeyDown={loading}
    >
      <Typography>{message}</Typography>
    </CommonDialog>
  );
}

// 表单对话框组件
export interface FormDialogProps {
  open: boolean;
  onClose: () => void;
  onSubmit: () => void | Promise<void> | Promise<boolean>;
  title: string;
  children: React.ReactNode;
  submitText?: string;
  cancelText?: string;
  loading?: boolean;
  disabled?: boolean;
  size?: DialogSize;
}

export function FormDialog({
  open,
  onClose,
  onSubmit,
  title,
  children,
  submitText,
  cancelText,
  loading = false,
  disabled = false,
  size = 'sm',
}: FormDialogProps) {
  const { t } = useTranslation();

  const actions: DialogButton[] = [
    {
      label: cancelText || t('common.cancel'),
      onClick: onClose,
      variant: 'text',
      disabled: loading,
    },
    {
      label: submitText || t('common.submit'),
      onClick: onSubmit,
      color: 'primary',
      variant: 'contained',
      loading,
      disabled: disabled || loading,
    },
  ];

  return (
    <CommonDialog
      open={open}
      onClose={onClose}
      title={title}
      actions={actions}
      size={size}
      disableBackdropClick={loading}
      disableEscapeKeyDown={loading}
      dividers
    >
      {children}
    </CommonDialog>
  );
}

// 对话框Hook
export interface UseDialogReturn {
  open: boolean;
  openDialog: () => void;
  closeDialog: () => void;
  toggleDialog: () => void;
}

export function useDialog(initialOpen = false): UseDialogReturn {
  const [open, setOpen] = React.useState(initialOpen);

  const openDialog = React.useCallback(() => {
    setOpen(true);
  }, []);

  const closeDialog = React.useCallback(() => {
    setOpen(false);
  }, []);

  const toggleDialog = React.useCallback(() => {
    setOpen(prev => !prev);
  }, []);

  return {
    open,
    openDialog,
    closeDialog,
    toggleDialog,
  };
}
