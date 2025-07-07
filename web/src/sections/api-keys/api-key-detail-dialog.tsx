import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Card from '@mui/material/Card';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import Dialog from '@mui/material/Dialog';
import Button from '@mui/material/Button';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import Typography from '@mui/material/Typography';
import DialogTitle from '@mui/material/DialogTitle';
import CardContent from '@mui/material/CardContent';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import TableContainer from '@mui/material/TableContainer';
import TablePagination from '@mui/material/TablePagination';
import CircularProgress from '@mui/material/CircularProgress';

import api from 'src/services/api';

import { Iconify } from 'src/components/iconify';
import { DateRangePicker } from 'src/components/date-range-picker';

// ----------------------------------------------------------------------

interface ApiKey {
  id: number;
  name: string;
  key?: string; // 完整的API密钥
  key_prefix: string;
  status: 'active' | 'inactive' | 'revoked';
  permissions?: {
    allowed_providers?: string[];
    allowed_models?: string[];
  };
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

interface UsageLog {
  id: number;
  timestamp: string;
  model: string;
  tokens_used: number;
  cost: number;
  request_type: string;
  status: string;
}

interface BillingRecord {
  id: number;
  timestamp: string;
  amount: number;
  description: string;
  transaction_type: string;
  balance_before: number;
  balance_after: number;
}

interface ApiKeyDetailDialogProps {
  open: boolean;
  apiKey: ApiKey | null;
  onClose: () => void;
  onStatusChange: (id: number, status: 'active' | 'inactive') => void;
  onDelete: (id: number) => void;
}

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

// ----------------------------------------------------------------------

export function ApiKeyDetailDialog({
  open,
  apiKey,
  onClose,
  onStatusChange,
  onDelete
}: ApiKeyDetailDialogProps) {
  const { t } = useTranslation();
  const [tabValue, setTabValue] = useState(0);
  const [loading, setLoading] = useState(false);
  const [usageLogs, setUsageLogs] = useState<UsageLog[]>([]);
  const [billingRecords, setBillingRecords] = useState<BillingRecord[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [totalTokens, setTotalTokens] = useState(0);

  // 分页状态
  const [usageLogsPage, setUsageLogsPage] = useState(0);
  const [usageLogsRowsPerPage, setUsageLogsRowsPerPage] = useState(10);
  const [usageLogsTotal, setUsageLogsTotal] = useState(0);
  const [billingRecordsPage, setBillingRecordsPage] = useState(0);
  const [billingRecordsRowsPerPage, setBillingRecordsRowsPerPage] = useState(10);
  const [billingRecordsTotal, setBillingRecordsTotal] = useState(0);

  // 日期过滤状态
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  // 获取使用日志
  const fetchUsageLogs = async (page = 0, pageSize = 10) => {
    if (!apiKey) return;

    try {
      setLoading(true);
      const token = localStorage.getItem('access_token');
      if (!token) return;

      const params = new URLSearchParams({
        page: (page + 1).toString(),
        page_size: pageSize.toString(),
      });

      if (startDate) {
        // 将日期转换为用户时区的ISO字符串
        const startDateTime = new Date(startDate);
        startDateTime.setHours(0, 0, 0, 0);
        const startDateISO = startDateTime.toISOString();
        params.append('start_date', startDateISO);
        console.log('Start date:', startDate, '-> ISO:', startDateISO);
      }
      if (endDate) {
        // 将日期转换为用户时区的ISO字符串，设置为当天结束时间
        const endDateTime = new Date(endDate);
        endDateTime.setHours(23, 59, 59, 999);
        const endDateISO = endDateTime.toISOString();
        params.append('end_date', endDateISO);
        console.log('End date:', endDate, '-> ISO:', endDateISO);
      }

      const response = await api.get(`/admin/api-keys/${apiKey.id}/usage-logs?${params}`);

      if (response.success && response.data) {
        const data = response.data;
        setUsageLogs(data.data || []);
        setUsageLogsTotal(data.total || 0);

        // 计算总计
        const logs = data.data || [];
        const totalCostCalc = logs.reduce((sum: number, log: UsageLog) => sum + log.cost, 0);
        const totalTokensCalc = logs.reduce((sum: number, log: UsageLog) => sum + log.tokens_used, 0);
        setTotalCost(totalCostCalc);
        setTotalTokens(totalTokensCalc);
      }
    } catch (error) {
      console.error('Error fetching usage logs:', error);
    } finally {
      setLoading(false);
    }
  };

  // 获取扣费记录
  const fetchBillingRecords = async (page = 0, pageSize = 10) => {
    if (!apiKey) return;

    try {
      setLoading(true);
      const token = localStorage.getItem('access_token');
      if (!token) return;

      const params = new URLSearchParams({
        page: (page + 1).toString(),
        page_size: pageSize.toString(),
      });

      if (startDate) {
        // 将日期转换为用户时区的ISO字符串
        const startDateTime = new Date(startDate);
        startDateTime.setHours(0, 0, 0, 0);
        params.append('start_date', startDateTime.toISOString());
      }
      if (endDate) {
        // 将日期转换为用户时区的ISO字符串，设置为当天结束时间
        const endDateTime = new Date(endDate);
        endDateTime.setHours(23, 59, 59, 999);
        params.append('end_date', endDateTime.toISOString());
      }

      const response = await api.get(`/admin/api-keys/${apiKey.id}/billing-records?${params}`);

      if (response.success && response.data) {
        const data = response.data;
        setBillingRecords(data.data || []);
        setBillingRecordsTotal(data.total || 0);
      }
    } catch (error) {
      console.error('Error fetching billing records:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && apiKey) {
      setUsageLogsPage(0);
      setBillingRecordsPage(0);
      fetchUsageLogs(0, usageLogsRowsPerPage);
      fetchBillingRecords(0, billingRecordsRowsPerPage);
    }
  }, [open, apiKey, startDate, endDate]);

  // 处理使用日志分页
  const handleUsageLogsPageChange = (event: unknown, newPage: number) => {
    setUsageLogsPage(newPage);
    fetchUsageLogs(newPage, usageLogsRowsPerPage);
  };

  const handleUsageLogsRowsPerPageChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newRowsPerPage = parseInt(event.target.value, 10);
    setUsageLogsRowsPerPage(newRowsPerPage);
    setUsageLogsPage(0);
    fetchUsageLogs(0, newRowsPerPage);
  };

  // 处理扣费记录分页
  const handleBillingRecordsPageChange = (event: unknown, newPage: number) => {
    setBillingRecordsPage(newPage);
    fetchBillingRecords(newPage, billingRecordsRowsPerPage);
  };

  const handleBillingRecordsRowsPerPageChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const newRowsPerPage = parseInt(event.target.value, 10);
    setBillingRecordsRowsPerPage(newRowsPerPage);
    setBillingRecordsPage(0);
    fetchBillingRecords(0, newRowsPerPage);
  };

  // 清除日期过滤器
  const handleClearDateRange = () => {
    setStartDate('');
    setEndDate('');
  };

  const handleStatusToggle = () => {
    if (!apiKey) return;
    const newStatus = apiKey.status === 'active' ? 'inactive' : 'active';
    onStatusChange(apiKey.id, newStatus);
  };

  const handleDelete = () => {
    if (!apiKey) return;
    if (window.confirm(t('api_keys.delete_confirm'))) {
      onDelete(apiKey.id);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'inactive':
        return 'warning';
      case 'revoked':
        return 'error';
      default:
        return 'default';
    }
  };

  const formatDate = (dateString: string) => new Date(dateString).toLocaleString();

  const formatCurrency = (amount: number) => `$${amount.toFixed(4)}`;



  if (!apiKey) return null;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Box>
            <Typography variant="h6">{apiKey.name}</Typography>
            <Typography variant="body2" color="text.secondary">
              {apiKey.key_prefix}...
            </Typography>
          </Box>
          <Chip
            label={apiKey.status}
            color={getStatusColor(apiKey.status) as any}
            variant="outlined"
          />
        </Box>
      </DialogTitle>

      <DialogContent>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <Tabs value={tabValue} onChange={handleTabChange}>
            <Tab label={t('dashboard.overview')} />
            <Tab label={t('api_keys.usage_logs')} />
            <Tab label={t('api_keys.billing_records')} />
          </Tabs>
        </Box>

        <TabPanel value={tabValue} index={0}>
          <Box sx={{ display: 'grid', gap: 3 }}>
            {/* 基本信息 */}
            <Card>
              <CardContent>
                <Typography variant="h6" sx={{ mb: 2 }}>
                  Basic Information
                </Typography>
                <Box sx={{ display: 'grid', gap: 2 }}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography color="text.secondary">Name:</Typography>
                    <Typography>{apiKey.name}</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography color="text.secondary">Key:</Typography>
                    <Typography sx={{ fontFamily: 'monospace' }}>{apiKey.key_prefix}••••••••••••</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography color="text.secondary">Status:</Typography>
                    <Chip
                      label={apiKey.status}
                      color={getStatusColor(apiKey.status) as any}
                      size="small"
                    />
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography color="text.secondary">Created:</Typography>
                    <Typography>{formatDate(apiKey.created_at)}</Typography>
                  </Box>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                    <Typography color="text.secondary">Last Used:</Typography>
                    <Typography>
                      {apiKey.last_used_at ? formatDate(apiKey.last_used_at) : 'Never'}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>

            {/* 使用统计 */}
            <Card>
              <CardContent>
                <Typography variant="h6" sx={{ mb: 2 }}>
                  Usage Statistics
                </Typography>
                <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 2 }}>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.100', borderRadius: 1 }}>
                    <Typography variant="h4" color="primary">
                      {totalTokens.toLocaleString()}
                    </Typography>
                    <Typography color="text.secondary">Total Tokens</Typography>
                  </Box>
                  <Box sx={{ textAlign: 'center', p: 2, bgcolor: 'grey.100', borderRadius: 1 }}>
                    <Typography variant="h4" color="error">
                      {formatCurrency(totalCost)}
                    </Typography>
                    <Typography color="text.secondary">Total Cost</Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          {/* 日期过滤器 */}
          <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <DateRangePicker
              startDate={startDate}
              endDate={endDate}
              onStartDateChange={setStartDate}
              onEndDateChange={setEndDate}
              onClear={handleClearDateRange}
            />
          </Box>

          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (
            <>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>{t('usage_logs.time')}</TableCell>
                      <TableCell>{t('usage_logs.model')}</TableCell>
                      <TableCell>{t('usage_logs.type')}</TableCell>
                      <TableCell align="right">{t('usage_logs.tokens')}</TableCell>
                      <TableCell align="right">{t('usage_logs.cost')}</TableCell>
                      <TableCell>{t('common.status')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {usageLogs.map((log) => (
                      <TableRow key={log.id}>
                        <TableCell>{formatDate(log.timestamp)}</TableCell>
                        <TableCell>{log.model}</TableCell>
                        <TableCell>{log.request_type}</TableCell>
                        <TableCell align="right">{log.tokens_used.toLocaleString()}</TableCell>
                        <TableCell align="right">{formatCurrency(log.cost)}</TableCell>
                        <TableCell>
                          <Chip
                            label={log.status}
                            color={log.status === 'success' ? 'success' : 'error'}
                            size="small"
                          />
                        </TableCell>
                      </TableRow>
                    ))}
                    {usageLogs.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={6} align="center">
                          {t('usage_logs.no_logs')}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>

              <TablePagination
                component="div"
                count={usageLogsTotal}
                page={usageLogsPage}
                onPageChange={handleUsageLogsPageChange}
                rowsPerPage={usageLogsRowsPerPage}
                onRowsPerPageChange={handleUsageLogsRowsPerPageChange}
                rowsPerPageOptions={[5, 10, 25, 50]}
              />
            </>
          )}
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          {/* 日期过滤器 */}
          <Box sx={{ mb: 3, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <DateRangePicker
              startDate={startDate}
              endDate={endDate}
              onStartDateChange={setStartDate}
              onEndDateChange={setEndDate}
              onClear={handleClearDateRange}
            />
          </Box>

          {loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (
            <>
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>{t('billing.timestamp')}</TableCell>
                      <TableCell>{t('billing.transaction_type')}</TableCell>
                      <TableCell>{t('common.description')}</TableCell>
                      <TableCell align="right">{t('billing.amount')}</TableCell>
                      <TableCell align="right">{t('billing.balance_before')}</TableCell>
                      <TableCell align="right">{t('billing.balance_after')}</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {billingRecords.map((record) => (
                      <TableRow key={record.id}>
                        <TableCell>{formatDate(record.timestamp)}</TableCell>
                        <TableCell>{record.transaction_type}</TableCell>
                        <TableCell>{record.description}</TableCell>
                        <TableCell align="right" sx={{ color: record.amount < 0 ? 'error.main' : 'success.main' }}>
                          {formatCurrency(record.amount)}
                        </TableCell>
                        <TableCell align="right">{formatCurrency(record.balance_before)}</TableCell>
                        <TableCell align="right">{formatCurrency(record.balance_after)}</TableCell>
                      </TableRow>
                    ))}
                    {billingRecords.length === 0 && (
                      <TableRow>
                        <TableCell colSpan={6} align="center">
                          {t('billing.no_records')}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>

              <TablePagination
                component="div"
                count={billingRecordsTotal}
                page={billingRecordsPage}
                onPageChange={handleBillingRecordsPageChange}
                rowsPerPage={billingRecordsRowsPerPage}
                onRowsPerPageChange={handleBillingRecordsRowsPerPageChange}
                rowsPerPageOptions={[5, 10, 25, 50]}
              />
            </>
          )}
        </TabPanel>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>Close</Button>
        <Button
          onClick={handleStatusToggle}
          color={apiKey.status === 'active' ? 'warning' : 'success'}
          startIcon={<Iconify icon={apiKey.status === 'active' ? 'solar:pause-bold' : 'solar:play-bold'} />}
        >
          {apiKey.status === 'active' ? 'Disable' : 'Enable'}
        </Button>
        <Button
          onClick={handleDelete}
          color="error"
          startIcon={<Iconify icon="solar:trash-bin-trash-bold" />}
        >
          Delete
        </Button>
      </DialogActions>
    </Dialog>
  );
}
