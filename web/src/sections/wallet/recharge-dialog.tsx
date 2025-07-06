import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import DialogTitle from '@mui/material/DialogTitle';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import CircularProgress from '@mui/material/CircularProgress';
import InputAdornment from '@mui/material/InputAdornment';

import AuthService from 'src/services/auth';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

type Props = {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
  currentBalance: number;
};

export function RechargeDialog({ open, onClose, onSuccess, currentBalance }: Props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [amount, setAmount] = useState('');
  const [error, setError] = useState('');

  const handleAmountChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    // 只允许数字和小数点
    if (value === '' || /^\d*\.?\d*$/.test(value)) {
      setAmount(value);
      setError('');
    }
  }, []);

  const validateAmount = useCallback(() => {
    const numAmount = parseFloat(amount);
    
    if (!amount || isNaN(numAmount)) {
      setError(t('wallet.amount_required'));
      return false;
    }
    
    if (numAmount <= 0) {
      setError(t('wallet.amount_must_positive'));
      return false;
    }
    
    if (numAmount > 10000) {
      setError(t('wallet.amount_too_large'));
      return false;
    }
    
    return true;
  }, [amount, t]);

  const handleSubmit = useCallback(async (event: React.FormEvent) => {
    event.preventDefault();

    if (!validateAmount()) {
      return;
    }

    setLoading(true);
    try {
      await AuthService.recharge({
        amount: parseFloat(amount),
        description: `用户充值 $${amount}`
      });

      // 重置表单
      setAmount('');
      setError('');
      onSuccess();
    } catch (error) {
      console.error('Recharge failed:', error);
      setError(error instanceof Error ? error.message : t('wallet.recharge_failed'));
    } finally {
      setLoading(false);
    }
  }, [amount, validateAmount, onSuccess, t]);

  const handleClose = useCallback(() => {
    if (!loading) {
      setAmount('');
      setError('');
      onClose();
    }
  }, [loading, onClose]);

  const handleQuickAmount = useCallback((quickAmount: number) => {
    setAmount(quickAmount.toString());
    setError('');
  }, []);

  const quickAmounts = [10, 50, 100, 500];

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Iconify icon="solar:cart-3-bold" />
          {t('wallet.recharge')}
        </Box>
      </DialogTitle>

      <form onSubmit={handleSubmit}>
        <DialogContent>
          <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
            {t('wallet.recharge_desc')}
          </Typography>

          {/* 当前余额显示 */}
          <Box 
            sx={{ 
              p: 2, 
              bgcolor: 'background.neutral',
              borderRadius: 1,
              mb: 3,
              textAlign: 'center'
            }}
          >
            <Typography variant="caption" color="text.secondary">
              {t('wallet.current_balance')}
            </Typography>
            <Typography variant="h6" color="primary.main">
              ${currentBalance.toFixed(6)}
            </Typography>
          </Box>

          {/* 充值金额输入 */}
          <TextField
            fullWidth
            label={t('wallet.recharge_amount')}
            value={amount}
            onChange={handleAmountChange}
            error={!!error}
            helperText={error}
            disabled={loading}
            placeholder="0.00"
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <Typography variant="body2" color="text.secondary">
                    $
                  </Typography>
                </InputAdornment>
              ),
            }}
            sx={{ mb: 3 }}
          />

          {/* 快捷金额按钮 */}
          <Box sx={{ mb: 3 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>
              {t('wallet.quick_amounts')}
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
              {quickAmounts.map((quickAmount) => (
                <Button
                  key={quickAmount}
                  variant="outlined"
                  size="small"
                  onClick={() => handleQuickAmount(quickAmount)}
                  disabled={loading}
                  sx={{ minWidth: 60 }}
                >
                  ${quickAmount}
                </Button>
              ))}
            </Box>
          </Box>

          {/* 充值后余额预览 */}
          {amount && !error && (
            <Box 
              sx={{ 
                p: 2, 
                bgcolor: 'success.lighter',
                borderRadius: 1,
                border: '1px solid',
                borderColor: 'success.light'
              }}
            >
              <Typography variant="caption" color="success.dark">
                {t('wallet.balance_after_recharge')}
              </Typography>
              <Typography variant="h6" color="success.dark">
                ${(currentBalance + parseFloat(amount || '0')).toFixed(6)}
              </Typography>
            </Box>
          )}
        </DialogContent>

        <DialogActions sx={{ px: 3, pb: 3 }}>
          <Button onClick={handleClose} disabled={loading}>
            {t('common.cancel')}
          </Button>
          <Button
            type="submit"
            variant="contained"
            disabled={!amount || !!error || loading}
            startIcon={loading ? <CircularProgress size={20} /> : null}
          >
            {loading ? t('common.loading') : t('wallet.confirm_recharge')}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}
