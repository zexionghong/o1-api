import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import TextField from '@mui/material/TextField';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import InputAdornment from '@mui/material/InputAdornment';
import CircularProgress from '@mui/material/CircularProgress';

import { useRouter } from 'src/routes/hooks';

import { useAuth } from 'src/contexts/auth-context';

import { Iconify } from 'src/components/iconify';
import { AuthLanguageSwitcher } from 'src/components/language-switcher';

// ----------------------------------------------------------------------

export function SignInView() {
  const { t } = useTranslation();
  const router = useRouter();
  const { login, state, clearError } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });

  const handleInputChange = useCallback((field: string) => (event: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({
      ...prev,
      [field]: event.target.value,
    }));
    if (state.error) {
      clearError();
    }
  }, [state.error, clearError]);

  const handleSubmit = useCallback(async (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!formData.username.trim() || !formData.password.trim()) {
      return;
    }

    try {
      await login({
        username: formData.username,
        password: formData.password,
      });
      router.push('/');
    } catch {
      // Error is handled by the auth context
    }
  }, [formData, login, router]);

  const renderForm = (
    <Box component="form" onSubmit={handleSubmit} sx={{ mt: 3 }}>
      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
        <TextField
          fullWidth
          name="username"
          label={t('auth.username')}
          value={formData.username}
          onChange={handleInputChange('username')}
          slotProps={{
            inputLabel: { shrink: true },
          }}
        />

        <Link variant="body2" color="inherit" sx={{ alignSelf: 'flex-end' }}>
          {t('auth.forgot_password')}
        </Link>

        <TextField
          fullWidth
          name="password"
          label={t('auth.password')}
          type={showPassword ? 'text' : 'password'}
          value={formData.password}
          onChange={handleInputChange('password')}
          slotProps={{
            inputLabel: { shrink: true },
            input: {
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton onClick={() => setShowPassword(!showPassword)} edge="end">
                    <Iconify icon={showPassword ? 'solar:eye-bold' : 'solar:eye-closed-bold'} />
                  </IconButton>
                </InputAdornment>
              ),
            },
          }}
        />
      </Box>

      {state.error && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {state.error}
        </Alert>
      )}

      <Button
        fullWidth
        size="large"
        type="submit"
        color="inherit"
        variant="contained"
        disabled={state.isLoading}
        startIcon={state.isLoading ? <CircularProgress size={20} /> : null}
        sx={{ mt: 3 }}
      >
        {state.isLoading ? t('common.loading') : t('auth.login')}
      </Button>
    </Box>
  );

  return (
    <>
      {/* 语言切换器 */}
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'flex-end',
          mb: 3,
        }}
      >
        <AuthLanguageSwitcher variant="icon" />
      </Box>

      <Box
        sx={{
          gap: 1.5,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          mb: 5,
        }}
      >
        <Typography variant="h5">{t('auth.login')}</Typography>
        <Typography
          variant="body2"
          sx={{
            color: 'text.secondary',
          }}
        >
          {t('auth.no_account')}{' '}
          <Link
            variant="subtitle2"
            sx={{ ml: 0.5, cursor: 'pointer' }}
            onClick={() => router.push('/sign-up')}
          >
            {t('auth.get_started')}
          </Link>
        </Typography>
      </Box>
      {renderForm}
      <Divider sx={{ my: 3, '&::before, &::after': { borderTopStyle: 'dashed' } }}>
        <Typography
          variant="overline"
          sx={{ color: 'text.disabled', fontWeight: 'fontWeightMedium' }}
        >
          {t('auth.or')}
        </Typography>
      </Divider>
    </>
  );
}
