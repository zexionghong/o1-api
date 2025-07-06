import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Grid from '@mui/system/Grid';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';

import { useAuthContext } from 'src/contexts/auth-context';

import AuthService from 'src/services/auth';
import type { UserProfile } from 'src/services/auth';

import { Iconify } from 'src/components/iconify';

import { RechargeDialog } from '../recharge-dialog';

// ----------------------------------------------------------------------

export function WalletView() {
  const { t } = useTranslation();
  const { state } = useAuthContext();
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  const [showRechargeDialog, setShowRechargeDialog] = useState(false);

  // 获取最新的用户资料
  const fetchUserProfile = useCallback(async () => {
    if (!state.isAuthenticated) return;
    
    setProfileLoading(true);
    try {
      const profile = await AuthService.getProfile();
      setUserProfile(profile);
    } catch (error) {
      console.error('Failed to fetch user profile:', error);
    } finally {
      setProfileLoading(false);
    }
  }, [state.isAuthenticated]);

  // 组件加载时获取用户资料
  useEffect(() => {
    fetchUserProfile();
  }, [fetchUserProfile]);

  const handleRechargeDialogToggle = useCallback(() => {
    setShowRechargeDialog((prev) => !prev);
  }, []);

  const handleRechargeSuccess = useCallback(() => {
    // 充值成功后刷新用户资料
    fetchUserProfile();
    setShowRechargeDialog(false);
  }, [fetchUserProfile]);

  // 如果正在加载，显示加载状态
  if (state.isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <CircularProgress />
      </Box>
    );
  }

  // 如果没有用户信息，显示错误状态
  if (!state.user) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="h6" color="error">
          {t('wallet.user_not_found')}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          {t('wallet.please_login_again')}
        </Typography>
      </Box>
    );
  }

  // 使用最新的用户资料数据，如果没有则使用认证上下文中的数据作为备选
  const currentUser = userProfile || state.user;

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        {t('wallet.title')}
      </Typography>

      <Grid container spacing={3}>
        {/* 余额卡片 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 4 }}>
              <Box sx={{ mb: 3 }}>
                <Iconify 
                  icon="solar:wallet-bold" 
                  sx={{ 
                    width: 80, 
                    height: 80, 
                    color: 'primary.main',
                    mx: 'auto',
                    mb: 2
                  }} 
                />
                <Typography variant="h6" sx={{ mb: 1 }}>
                  {t('wallet.current_balance')}
                </Typography>
                <Typography 
                  variant="h3" 
                  sx={{ 
                    color: 'primary.main',
                    fontWeight: 'bold',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    gap: 1
                  }}
                >
                  ${currentUser.balance?.toFixed(6) || '0.000000'}
                  {profileLoading && (
                    <CircularProgress size={20} />
                  )}
                </Typography>
              </Box>

              <Button
                variant="contained"
                size="large"
                startIcon={<Iconify icon="solar:card-bold" />}
                onClick={handleRechargeDialogToggle}
                fullWidth
                sx={{ py: 1.5 }}
              >
                {t('wallet.recharge')}
              </Button>
            </CardContent>
          </Card>
        </Grid>

        {/* 账户信息卡片 */}
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 3 }}>
                {t('wallet.account_info')}
              </Typography>

              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2" color="text.secondary">
                    {t('wallet.username')}
                  </Typography>
                  <Typography variant="body2">
                    {currentUser.username}
                  </Typography>
                </Box>

                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2" color="text.secondary">
                    {t('wallet.email')}
                  </Typography>
                  <Typography variant="body2">
                    {currentUser.email}
                  </Typography>
                </Box>

                <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                  <Typography variant="body2" color="text.secondary">
                    {t('wallet.account_status')}
                  </Typography>
                  <Typography variant="body2" color="success.main">
                    {t('wallet.active')}
                  </Typography>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* 使用说明卡片 */}
        <Grid size={{ xs: 12 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>
                {t('wallet.usage_instructions')}
              </Typography>
              
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                <Typography variant="body2" color="text.secondary">
                  • {t('wallet.instruction_1')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  • {t('wallet.instruction_2')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  • {t('wallet.instruction_3')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  • {t('wallet.instruction_4')}
                </Typography>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* 充值对话框 */}
      <RechargeDialog
        open={showRechargeDialog}
        onClose={handleRechargeDialogToggle}
        onSuccess={handleRechargeSuccess}
        currentBalance={currentUser.balance || 0}
      />
    </Box>
  );
}
