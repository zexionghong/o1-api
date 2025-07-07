import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import FormControl from '@mui/material/FormControl';
import InputLabel from '@mui/material/InputLabel';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import IconButton from '@mui/material/IconButton';

import { Iconify } from 'src/components/iconify';
import api from 'src/services/api';

// ----------------------------------------------------------------------

interface Quota {
  id: number;
  quota_type: 'requests' | 'tokens' | 'cost';
  period?: 'minute' | 'hour' | 'day' | 'month' | null;
  limit_value: number;
  used_value?: number;
  remaining?: number;
  percentage?: number;
  reset_time?: string;
  status: 'active' | 'inactive';
  created_at: string;
  updated_at: string;
}



interface QuotaConfigDialogProps {
  open: boolean;
  onClose: () => void;
  apiKeyId: number;
  apiKeyName: string;
}

interface CreateQuotaForm {
  quota_type: 'requests' | 'tokens' | 'cost';
  period?: 'minute' | 'hour' | 'day' | 'month' | 'total' | null;
  limit_value: number;
}

// ----------------------------------------------------------------------

export function QuotaConfigDialog({ open, onClose, apiKeyId, apiKeyName }: QuotaConfigDialogProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [quotas, setQuotas] = useState<Quota[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createForm, setCreateForm] = useState<CreateQuotaForm>({
    quota_type: 'requests',
    period: 'total', // 默认选择总限额
    limit_value: 1000,
  });

  // 获取配额列表（包含使用情况）
  const fetchQuotas = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/admin/api-keys/${apiKeyId}/quotas`);

      if (response.success && response.data) {
        setQuotas(response.data);
      }
    } catch (err) {
      console.error('Error fetching quotas:', err);
      setError(t('quota.errors.fetch_failed'));
    } finally {
      setLoading(false);
    }
  };

  // 检查是否已存在相同类型和周期的配额
  const checkDuplicateQuota = (quotaType: string, period: string | null) => {
    return quotas.some(quota =>
      quota.quota_type === quotaType && quota.period === period
    );
  };

  // 检查当前选择的组合是否已存在
  const isCurrentCombinationDuplicate = () => {
    const period = createForm.period === 'total' ? null : createForm.period;
    return checkDuplicateQuota(createForm.quota_type, period ?? null);
  };

  // 创建配额
  const handleCreateQuota = async () => {
    try {
      setLoading(true);
      setError(null);

      // 检查是否已存在相同类型和周期的配额
      const period = createForm.period === 'total' ? null : createForm.period;
      if (checkDuplicateQuota(createForm.quota_type, period ?? null)) {
        const periodLabel = getPeriodLabel(period);
        const typeLabel = getQuotaTypeLabel(createForm.quota_type);
        setError(`${periodLabel}${t('quota.errors.duplicate_quota')}`);
        setLoading(false);
        return;
      }

      const payload = {
        quota_type: createForm.quota_type,
        limit_value: createForm.limit_value,
        ...(period && { period }),
      };

      const response = await api.post(`/admin/api-keys/${apiKeyId}/quotas`, payload);

      // 如果创建成功，直接将新配额添加到本地状态中
      if (response.success && response.data) {
        setQuotas(prevQuotas => [...prevQuotas, response.data]);
      } else {
        // 如果响应中没有返回新配额数据，则重新获取列表
        await fetchQuotas();
      }

      // 重置表单
      setCreateForm({
        quota_type: 'requests',
        period: 'total', // 重置为总限额
        limit_value: 1000,
      });
      setShowCreateForm(false);
    } catch (err) {
      console.error('Error creating quota:', err);
      setError(t('quota.errors.create_failed'));
    } finally {
      setLoading(false);
    }
  };

  // 删除配额
  const handleDeleteQuota = async (quotaId: number) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/admin/quotas/${quotaId}`);

      // 直接从本地状态中移除该配额，无需重新获取整个列表
      setQuotas(prevQuotas => prevQuotas.filter(quota => quota.id !== quotaId));
    } catch (err) {
      console.error('Error deleting quota:', err);
      setError(t('quota.errors.delete_failed'));
    } finally {
      setLoading(false);
    }
  };

  // 获取配额类型标签
  const getQuotaTypeLabel = (type: string) => {
    switch (type) {
      case 'requests': return t('quota.quota_types.requests');
      case 'tokens': return t('quota.quota_types.tokens');
      case 'cost': return t('quota.quota_types.cost');
      default: return type;
    }
  };

  // 获取周期标签
  const getPeriodLabel = (period: string | null | undefined) => {
    if (!period) return t('quota.periods.total');
    switch (period) {
      case 'minute': return t('quota.periods.minute');
      case 'hour': return t('quota.periods.hour');
      case 'day': return t('quota.periods.day');
      case 'month': return t('quota.periods.month');
      default: return period;
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'success';
      case 'inactive': return 'default';
      default: return 'default';
    }
  };



  useEffect(() => {
    if (open) {
      fetchQuotas();
    }
  }, [open, apiKeyId]);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Typography variant="h6">
            {t('quota.title')} - {apiKeyName}
          </Typography>
          <IconButton onClick={onClose} size="small">
            <Iconify icon="solar:close-circle-bold" />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}



        {/* 配额列表 */}
        <Box sx={{ mb: 3 }}>
          <Box display="flex" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
            <Typography variant="h6">{t('quota.current_quotas')}</Typography>
            <Button
              variant="contained"
              startIcon={<Iconify icon="solar:check-circle-bold" />}
              onClick={() => setShowCreateForm(true)}
              disabled={loading}
            >
              {t('quota.add_quota')}
            </Button>
          </Box>

          {loading && quotas.length === 0 ? (
            <Box display="flex" justifyContent="center" py={3}>
              <CircularProgress />
            </Box>
          ) : (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>{t('quota.quota_type')}</TableCell>
                    <TableCell>{t('quota.period')}</TableCell>
                    <TableCell align="right">{t('quota.limit_value')}</TableCell>
                    <TableCell align="right">{t('quota.used_value')}</TableCell>
                    <TableCell align="right">{t('quota.remaining')}</TableCell>
                    <TableCell align="center">{t('quota.usage_rate')}</TableCell>
                    <TableCell>{t('common.status')}</TableCell>
                    <TableCell align="center">{t('common.actions')}</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {quotas.map((quota) => (
                    <TableRow key={quota.id}>
                      <TableCell>{getQuotaTypeLabel(quota.quota_type)}</TableCell>
                      <TableCell>{getPeriodLabel(quota.period)}</TableCell>
                      <TableCell align="right">{quota.limit_value.toLocaleString()}</TableCell>
                      <TableCell align="right">
                        {quota.used_value !== undefined ? quota.used_value.toLocaleString() : '-'}
                      </TableCell>
                      <TableCell align="right">
                        {quota.remaining !== undefined ? quota.remaining.toLocaleString() : '-'}
                      </TableCell>
                      <TableCell align="center">
                        {quota.percentage !== undefined ? (
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Box
                              sx={{
                                width: 60,
                                height: 6,
                                backgroundColor: 'grey.300',
                                borderRadius: 3,
                                overflow: 'hidden',
                              }}
                            >
                              <Box
                                sx={{
                                  width: `${Math.min(quota.percentage, 100)}%`,
                                  height: '100%',
                                  backgroundColor: quota.percentage > 90 ? 'error.main' :
                                                 quota.percentage > 70 ? 'warning.main' : 'success.main',
                                  transition: 'width 0.3s ease',
                                }}
                              />
                            </Box>
                            <Typography variant="caption" color="text.secondary">
                              {quota.percentage.toFixed(1)}%
                            </Typography>
                          </Box>
                        ) : '-'}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={quota.status === 'active' ? t('quota.status.active') : t('quota.status.inactive')}
                          color={getStatusColor(quota.status) as any}
                          size="small"
                        />
                      </TableCell>
                      <TableCell align="center">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDeleteQuota(quota.id)}
                          disabled={loading}
                        >
                          <Iconify icon="solar:trash-bin-trash-bold" />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                  {quotas.length === 0 && !loading && (
                    <TableRow>
                      <TableCell colSpan={8} align="center" sx={{ py: 3 }}>
                        <Typography variant="body2" color="text.secondary">
                          {t('quota.no_quotas')}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>

        {/* 创建配额表单 */}
        {showCreateForm && (
          <Box sx={{ p: 2, border: 1, borderColor: 'divider', borderRadius: 1 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              {t('quota.add_new_quota')}
            </Typography>
            
            <Box display="flex" gap={2} sx={{ mb: 2 }}>
              <FormControl size="small" sx={{ minWidth: 120 }}>
                <InputLabel>{t('quota.quota_type')}</InputLabel>
                <Select
                  value={createForm.quota_type}
                  label={t('quota.quota_type')}
                  onChange={(e) => setCreateForm({ ...createForm, quota_type: e.target.value as any })}
                >
                  <MenuItem value="requests">{t('quota.quota_types.requests')}</MenuItem>
                  <MenuItem value="tokens">{t('quota.quota_types.tokens')}</MenuItem>
                  <MenuItem value="cost">{t('quota.quota_types.cost')}</MenuItem>
                </Select>
              </FormControl>

              <FormControl size="small" sx={{ minWidth: 120 }}>
                <InputLabel>{t('quota.period')}</InputLabel>
                <Select
                  value={createForm.period || 'total'}
                  label={t('quota.period')}
                  onChange={(e) => {
                    const newPeriod = e.target.value as 'minute' | 'hour' | 'day' | 'month' | 'total';
                    setCreateForm({ ...createForm, period: newPeriod });
                  }}
                >
                  <MenuItem value="total">{t('quota.periods.total')}</MenuItem>
                  <MenuItem value="minute">{t('quota.periods.minute')}</MenuItem>
                  <MenuItem value="hour">{t('quota.periods.hour')}</MenuItem>
                  <MenuItem value="day">{t('quota.periods.day')}</MenuItem>
                  <MenuItem value="month">{t('quota.periods.month')}</MenuItem>
                </Select>
              </FormControl>

              <TextField
                size="small"
                label={t('quota.limit_value')}
                type="number"
                value={createForm.limit_value}
                onChange={(e) => setCreateForm({ ...createForm, limit_value: Number(e.target.value) })}
                sx={{ minWidth: 120 }}
                helperText={createForm.period === 'total' ? t('quota.helper_text.total_quota') : ''}
              />
            </Box>

            {/* 重复配额警告 */}
            {isCurrentCombinationDuplicate() && (
              <Alert severity="warning" sx={{ mb: 2 }}>
                {getPeriodLabel(createForm.period === 'total' ? null : createForm.period)}{t('quota.errors.duplicate_quota')}
              </Alert>
            )}

            <Box display="flex" gap={1}>
              <Button
                variant="contained"
                onClick={handleCreateQuota}
                disabled={loading || isCurrentCombinationDuplicate()}
                size="small"
              >
                {t('common.create')}
              </Button>
              <Button
                variant="outlined"
                onClick={() => setShowCreateForm(false)}
                size="small"
              >
                {t('common.cancel')}
              </Button>
            </Box>
          </Box>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>{t('common.close')}</Button>
      </DialogActions>
    </Dialog>
  );
}
