import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';

import { useRouter } from 'src/routes/hooks';

import { useAuth } from 'src/contexts/auth-context';
import { DashboardContent } from 'src/layouts/dashboard';

import { AnalyticsWidgetSummary } from '../analytics-widget-summary';

// ----------------------------------------------------------------------

// 数据类型定义
interface UserStats {
  balance: number;
  total_requests: number;
  total_tokens: number;
  total_cost: number;
}

interface ApiKey {
  id: number;
  name: string;
  key_prefix: string;
  status: string;
  last_used_at: string | null;
  created_at: string;
}

interface UsageRecord {
  id: number;
  request_id: string;
  method: string;
  endpoint: string;
  total_tokens: number;
  cost: number;
  status_code: number;
  created_at: string;
}

// ----------------------------------------------------------------------

export function RealDashboardView() {
  const { t } = useTranslation();
  const { state } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [userStats, setUserStats] = useState<UserStats | null>(null);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);

  // 获取用户统计数据
  const fetchUserStats = async () => {
    try {
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('No access token found');
      }

      // 获取用户资料（包含余额）
      const profileResponse = await fetch('http://localhost:8080/auth/profile', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!profileResponse.ok) {
        throw new Error('Failed to fetch user profile');
      }

      const profileData = await profileResponse.json();
      const userProfile = profileData.data;

      // 获取用户的API密钥
      const apiKeysResponse = await fetch(`http://localhost:8080/admin/users/${userProfile.id}/api-keys`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      let apiKeysData = [];
      if (apiKeysResponse.ok) {
        const apiKeysResult = await apiKeysResponse.json();
        apiKeysData = apiKeysResult.data || [];
      }

      // 获取使用情况统计
      const usageResponse = await fetch('http://localhost:8080/v1/usage', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      let usageData = { total_requests: 0, total_tokens: 0, total_cost: 0 };
      if (usageResponse.ok) {
        const usageResult = await usageResponse.json();
        usageData = usageResult.data || usageData;
      }

      // 组合统计数据
      const stats: UserStats = {
        balance: userProfile.balance || 0,
        total_requests: usageData.total_requests || 0,
        total_tokens: usageData.total_tokens || 0,
        total_cost: usageData.total_cost || 0,
      };

      setUserStats(stats);
      setApiKeys(apiKeysData.slice(0, 5)); // 只显示前5个API密钥

    } catch (err) {
      console.error('Error fetching user stats:', err);
      setError(err instanceof Error ? err.message : 'Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (state.user?.id) {
      fetchUserStats();
    }
  }, [state.user?.id]);

  if (loading) {
    return (
      <DashboardContent maxWidth="xl">
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
          <CircularProgress />
        </div>
      </DashboardContent>
    );
  }

  if (error) {
    return (
      <DashboardContent maxWidth="xl">
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      </DashboardContent>
    );
  }

  return (
    <DashboardContent maxWidth="xl">
      <Typography variant="h4" sx={{ mb: { xs: 3, md: 5 } }}>
        {t('dashboard.welcome')}, {state.user?.username} 👋
      </Typography>

      <Grid container spacing={3}>
        {/* 账户余额 */}
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <AnalyticsWidgetSummary
            title={t('dashboard.balance')}
            total={userStats?.balance || 0}
            percent={0}
            icon={<img alt="Account Balance" src="/assets/icons/glass/ic-glass-bag.svg" />}
            chart={{
              categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug'],
              series: [22, 8, 35, 50, 82, 84, 77, 12],
            }}
          />
        </Grid>

        {/* API密钥数量 */}
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <AnalyticsWidgetSummary
            title={t('api_keys.title')}
            total={apiKeys.length}
            percent={0}
            color="secondary"
            icon={<img alt="API Keys" src="/assets/icons/glass/ic-glass-users.svg" />}
            chart={{
              categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug'],
              series: [56, 47, 40, 62, 73, 30, 23, 54],
            }}
          />
        </Grid>

        {/* 总请求数 */}
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <AnalyticsWidgetSummary
            title={t('dashboard.total_requests')}
            total={userStats?.total_requests || 0}
            percent={0}
            color="warning"
            icon={<img alt="Total Requests" src="/assets/icons/glass/ic-glass-buy.svg" />}
            chart={{
              categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug'],
              series: [40, 70, 75, 70, 50, 28, 7, 64],
            }}
          />
        </Grid>

        {/* 总花费 */}
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <AnalyticsWidgetSummary
            title={t('dashboard.total_cost')}
            total={userStats?.total_cost || 0}
            percent={0}
            color="error"
            icon={<img alt="Total Cost" src="/assets/icons/glass/ic-glass-message.svg" />}
            chart={{
              categories: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug'],
              series: [56, 30, 23, 54, 47, 40, 62, 73],
            }}
          />
        </Grid>

        {/* API密钥列表 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6">
                  {t('api_keys.title')}
                </Typography>
                <Button
                  variant="outlined"
                  size="small"
                  onClick={() => router.push('/api-keys')}
                >
                  {t('api_keys.manage_keys')}
                </Button>
              </Box>
              {apiKeys.length > 0 ? (
                <TableContainer component={Paper} variant="outlined">
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>{t('common.name')}</TableCell>
                        <TableCell>{t('api_keys.key_prefix')}</TableCell>
                        <TableCell>{t('common.status')}</TableCell>
                        <TableCell>{t('common.created_at')}</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {apiKeys.map((apiKey) => (
                        <TableRow key={apiKey.id}>
                          <TableCell>{apiKey.name}</TableCell>
                          <TableCell>
                            <code>{apiKey.key_prefix}...</code>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={apiKey.status}
                              color={apiKey.status === 'active' ? 'success' : 'default'}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            {new Date(apiKey.created_at).toLocaleDateString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              ) : (
                <Box sx={{ textAlign: 'center', py: 3 }}>
                  <Typography color="text.secondary" sx={{ mb: 2 }}>
                    {t('api_keys.no_keys')}
                  </Typography>
                  <Button
                    variant="contained"
                    onClick={() => router.push('/api-keys')}
                  >
                    {t('api_keys.create_key')}
                  </Button>
                </Box>
              )}
            </CardContent>
          </Card>
        </Grid>

        {/* 使用统计 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>
                {t('dashboard.statistics')}
              </Typography>
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography color="text.secondary">{t('dashboard.total_requests')}:</Typography>
                  <Typography variant="h6">{userStats?.total_requests || 0}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography color="text.secondary">{t('dashboard.total_tokens')}:</Typography>
                  <Typography variant="h6">{userStats?.total_tokens || 0}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography color="text.secondary">{t('dashboard.total_cost')}:</Typography>
                  <Typography variant="h6">${userStats?.total_cost?.toFixed(4) || '0.0000'}</Typography>
                </Box>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography color="text.secondary">{t('dashboard.balance')}:</Typography>
                  <Typography variant="h6" color={userStats?.balance && userStats.balance > 0 ? 'success.main' : 'error.main'}>
                    ${userStats?.balance?.toFixed(2) || '0.00'}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </DashboardContent>
  );
}
