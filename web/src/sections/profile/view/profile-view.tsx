import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Grid from '@mui/system/Grid';
import Button from '@mui/material/Button';
import Avatar from '@mui/material/Avatar';
import Typography from '@mui/material/Typography';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';

import { useAuthContext } from 'src/contexts/auth-context';

import AuthService from 'src/services/auth';
import type { UserProfile } from 'src/services/auth';

import { Iconify } from 'src/components/iconify';

import { ProfilePasswordForm } from '../profile-password-form';

// ----------------------------------------------------------------------

export function ProfileView() {
  const { t } = useTranslation();
  const { state } = useAuthContext();
  const [showPasswordForm, setShowPasswordForm] = useState(false);
  const [userProfile, setUserProfile] = useState<UserProfile | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);

  const handlePasswordFormToggle = useCallback(() => {
    setShowPasswordForm((prev) => !prev);
  }, []);

  // 获取最新的用户资料
  const fetchUserProfile = useCallback(async () => {
    if (!state.isAuthenticated) return;

    setProfileLoading(true);
    try {
      const profile = await AuthService.getProfile();
      setUserProfile(profile);
    } catch (error) {
      console.error('Failed to fetch user profile:', error);
      // 如果获取失败，使用认证上下文中的用户信息作为备选
    } finally {
      setProfileLoading(false);
    }
  }, [state.isAuthenticated]);

  // 组件加载时获取用户资料
  useEffect(() => {
    fetchUserProfile();
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
          {t('profile.user_not_found')}
        </Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
          {t('profile.please_login_again')}
        </Typography>
      </Box>
    );
  }

  // 使用最新的用户资料数据，如果没有则使用认证上下文中的数据作为备选
  const currentUser = userProfile || state.user;

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        {t('profile.title')}
      </Typography>

      <Grid container spacing={3}>
        {/* 用户信息卡片 */}
        <Grid size={{ xs: 12, md: 4 }}>
          <Card>
            <CardContent sx={{ textAlign: 'center', py: 4 }}>
              <Avatar
                sx={{
                  width: 120,
                  height: 120,
                  mx: 'auto',
                  mb: 2,
                  fontSize: '3rem',
                  bgcolor: 'primary.main',
                }}
              >
                {currentUser.username?.charAt(0).toUpperCase()}
              </Avatar>

              <Typography variant="h6" sx={{ mb: 1 }}>
                {currentUser.full_name || currentUser.username}
              </Typography>

              <Typography variant="body2" color="text.secondary">
                {currentUser.email}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        {/* 账户设置 */}
        <Grid size={{ xs: 12, md: 8 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 3 }}>
                {t('profile.account_settings')}
              </Typography>

              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                {/* 基本信息 */}
                <Box
                  sx={{
                    p: 2,
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: 1,
                  }}
                >
                  <Typography variant="subtitle1" sx={{ mb: 1 }}>
                    {t('profile.basic_info')}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    {t('profile.basic_info_desc')}
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="caption" color="text.secondary">
                        {t('profile.username')}
                      </Typography>
                      <Typography variant="body2">{currentUser.username}</Typography>
                    </Box>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="caption" color="text.secondary">
                        {t('profile.email')}
                      </Typography>
                      <Typography variant="body2">{currentUser.email}</Typography>
                    </Box>
                  </Box>
                </Box>

                {/* 密码设置 */}
                <Box
                  sx={{
                    p: 2,
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: 1,
                  }}
                >
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                    <Typography variant="subtitle1">
                      {t('profile.password')}
                    </Typography>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<Iconify icon="solar:key-bold" />}
                      onClick={handlePasswordFormToggle}
                    >
                      {t('profile.change_password')}
                    </Button>
                  </Box>
                  <Typography variant="body2" color="text.secondary">
                    {t('profile.password_desc')}
                  </Typography>
                </Box>

                {/* 账户状态 */}
                <Box
                  sx={{
                    p: 2,
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: 1,
                  }}
                >
                  <Typography variant="subtitle1" sx={{ mb: 1 }}>
                    {t('profile.account_status')}
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="caption" color="text.secondary">
                        {t('profile.status')}
                      </Typography>
                      <Typography variant="body2" color="success.main">
                        {t('profile.active')}
                      </Typography>
                    </Box>
                    <Box sx={{ flex: 1 }}>
                      <Typography variant="caption" color="text.secondary">
                        {t('profile.balance')}
                      </Typography>
                      <Typography variant="body2" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        ${currentUser.balance?.toFixed(6) || '0.000000'}
                        {profileLoading && (
                          <CircularProgress size={12} />
                        )}
                      </Typography>
                    </Box>
                  </Box>
                </Box>
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* 修改密码对话框 */}
      <ProfilePasswordForm
        open={showPasswordForm}
        onClose={handlePasswordFormToggle}
      />
    </Box>
  );
}
