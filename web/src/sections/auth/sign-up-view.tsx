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

export function SignUpView() {
  const { t } = useTranslation();
  const router = useRouter();
  const { state, register, clearError } = useAuth();

  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    full_name: '',
  });
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [registrationSuccess, setRegistrationSuccess] = useState(false);

  const handleInputChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = event.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
    
    // 清除对应字段的错误
    if (formErrors[name]) {
      setFormErrors(prev => ({
        ...prev,
        [name]: '',
      }));
    }
    
    // 清除全局错误信息
    if (state.error) {
      clearError();
    }
  }, [formErrors, state.error, clearError]);

  const validateForm = useCallback(() => {
    const errors: Record<string, string> = {};

    if (!formData.username.trim()) {
      errors.username = t('auth.username_required');
    } else if (formData.username.length < 3) {
      errors.username = t('auth.username_min_length');
    }

    if (!formData.email.trim()) {
      errors.email = t('auth.email_required');
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = t('auth.email_invalid');
    }

    if (!formData.password) {
      errors.password = t('auth.password_required');
    } else if (formData.password.length < 6) {
      errors.password = t('auth.password_min_length');
    }

    if (!formData.confirmPassword) {
      errors.confirmPassword = t('auth.confirm_password_required');
    } else if (formData.password !== formData.confirmPassword) {
      errors.confirmPassword = t('auth.passwords_not_match');
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  }, [formData, t]);

  const handleSignUp = useCallback(async (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      await register({
        username: formData.username,
        email: formData.email,
        password: formData.password,
        full_name: formData.full_name || undefined,
      });
      
      // 注册成功
      setRegistrationSuccess(true);
    } catch (error) {
      // 错误已经在context中处理了
      console.error('Registration failed:', error);
    }
  }, [formData, register, validateForm]);

  const handleGoToSignIn = useCallback(() => {
    router.push('/sign-in');
  }, [router]);

  if (registrationSuccess) {
    return (
      <Box sx={{ textAlign: 'center', p: 3 }}>
        <Typography variant="h4" sx={{ mb: 2 }}>
          {t('auth.registration_successful')}
        </Typography>
        <Typography variant="body1" sx={{ mb: 3, color: 'text.secondary' }}>
          {t('auth.account_created')}
        </Typography>
        <Button
          variant="contained"
          size="large"
          onClick={handleGoToSignIn}
        >
          Go to Sign In
        </Button>
      </Box>
    );
  }

  const renderForm = (
    <Box
      component="form"
      onSubmit={handleSignUp}
      sx={{
        display: 'flex',
        flexDirection: 'column',
        gap: 3,
      }}
    >
      {state.error && (
        <Alert severity="error">
          {state.error}
        </Alert>
      )}

      <TextField
        fullWidth
        name="username"
        label={t('auth.username')}
        value={formData.username}
        onChange={handleInputChange}
        disabled={state.isLoading}
        required
        error={!!formErrors.username}
        helperText={formErrors.username}
        slotProps={{
          inputLabel: { shrink: true },
        }}
      />

      <TextField
        fullWidth
        name="email"
        label={t('auth.email')}
        type="email"
        value={formData.email}
        onChange={handleInputChange}
        disabled={state.isLoading}
        required
        error={!!formErrors.email}
        helperText={formErrors.email}
        slotProps={{
          inputLabel: { shrink: true },
        }}
      />

      <TextField
        fullWidth
        name="full_name"
        label={t('users.full_name')}
        value={formData.full_name}
        onChange={handleInputChange}
        disabled={state.isLoading}
        slotProps={{
          inputLabel: { shrink: true },
        }}
      />

      <TextField
        fullWidth
        name="password"
        label={t('auth.password')}
        value={formData.password}
        onChange={handleInputChange}
        disabled={state.isLoading}
        required
        error={!!formErrors.password}
        helperText={formErrors.password}
        type={showPassword ? 'text' : 'password'}
        slotProps={{
          inputLabel: { shrink: true },
          input: {
            endAdornment: (
              <InputAdornment position="end">
                <IconButton 
                  onClick={() => setShowPassword(!showPassword)} 
                  edge="end"
                  disabled={state.isLoading}
                >
                  <Iconify icon={showPassword ? 'solar:eye-bold' : 'solar:eye-closed-bold'} />
                </IconButton>
              </InputAdornment>
            ),
          },
        }}
      />

      <TextField
        fullWidth
        name="confirmPassword"
        label={t('auth.confirm_password')}
        value={formData.confirmPassword}
        onChange={handleInputChange}
        disabled={state.isLoading}
        required
        error={!!formErrors.confirmPassword}
        helperText={formErrors.confirmPassword}
        type={showConfirmPassword ? 'text' : 'password'}
        slotProps={{
          inputLabel: { shrink: true },
          input: {
            endAdornment: (
              <InputAdornment position="end">
                <IconButton 
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)} 
                  edge="end"
                  disabled={state.isLoading}
                >
                  <Iconify icon={showConfirmPassword ? 'solar:eye-bold' : 'solar:eye-closed-bold'} />
                </IconButton>
              </InputAdornment>
            ),
          },
        }}
      />

      <Button
        fullWidth
        size="large"
        type="submit"
        color="inherit"
        variant="contained"
        disabled={state.isLoading}
        startIcon={state.isLoading ? <CircularProgress size={20} /> : null}
      >
        {state.isLoading ? t('common.loading') : t('auth.register')}
      </Button>
    </Box>
  );

  return (
    <Box sx={{ p: 3 }}>
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

      <Typography variant="h3" sx={{ mb: 2 }}>
        {t('auth.register')}
      </Typography>

      <Typography variant="body2" sx={{ color: 'text.secondary', mb: 5 }}>
        {t('auth.have_account')}{' '}
        <Link
          variant="subtitle2"
          sx={{ cursor: 'pointer' }}
          onClick={handleGoToSignIn}
        >
          {t('auth.login')}
        </Link>
      </Typography>

      {renderForm}

      <Divider sx={{ my: 3 }}>
        <Typography variant="body2" sx={{ color: 'text.secondary' }}>
          {t('auth.or')}
        </Typography>
      </Divider>

      <Button
        fullWidth
        size="large"
        color="inherit"
        variant="outlined"
        onClick={handleGoToSignIn}
        startIcon={<Iconify icon="eva:arrow-ios-upward-fill" />}
      >
        {t('auth.back_to_signin')}
      </Button>
    </Box>
  );
}
