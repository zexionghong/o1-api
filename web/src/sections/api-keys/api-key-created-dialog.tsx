import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import InputAdornment from '@mui/material/InputAdornment';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

interface ApiKeyCreatedDialogProps {
  open: boolean;
  apiKey: string;
  apiKeyName: string;
  onClose: () => void;
}

// ----------------------------------------------------------------------

export function ApiKeyCreatedDialog({ open, apiKey, apiKeyName, onClose }: ApiKeyCreatedDialogProps) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000); // 2秒后重置状态
    } catch (error) {
      console.error('Failed to copy API key:', error);
      // 降级方案：选择文本
      const textField = document.getElementById('api-key-field') as HTMLInputElement;
      if (textField) {
        textField.select();
        textField.setSelectionRange(0, 99999); // 移动端兼容
        try {
          document.execCommand('copy');
          setCopied(true);
          setTimeout(() => setCopied(false), 2000);
        } catch (err) {
          console.error('Fallback copy failed:', err);
        }
      }
    }
  };

  const handleClose = () => {
    setCopied(false);
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Iconify icon="solar:key-bold" sx={{ color: 'success.main' }} />
          <Typography variant="h6">{t('api_keys.created_successfully')}</Typography>
        </Box>
      </DialogTitle>

      <DialogContent>
        <Box sx={{ mb: 3 }}>
          <Alert severity="warning" sx={{ mb: 3 }}>
            <Typography variant="body2">
              <strong>{t('api_keys.important')}:</strong> {t('api_keys.security_warning')}
            </Typography>
          </Alert>

          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
            {t('api_keys.api_key_name')}:
          </Typography>
          <Typography variant="body1" sx={{ mb: 3, color: 'text.secondary' }}>
            {apiKeyName}
          </Typography>

          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
            {t('api_keys.your_api_key')}:
          </Typography>
          
          <TextField
            id="api-key-field"
            fullWidth
            value={apiKey}
            variant="outlined"
            InputProps={{
              readOnly: true,
              sx: { 
                fontFamily: 'monospace',
                fontSize: '0.875rem',
                bgcolor: 'grey.50',
              },
              endAdornment: (
                <InputAdornment position="end">
                  <IconButton
                    onClick={handleCopy}
                    edge="end"
                    color={copied ? 'success' : 'default'}
                    sx={{ mr: 1 }}
                  >
                    <Iconify 
                      icon={copied ? 'solar:check-circle-bold' : 'solar:copy-bold'} 
                    />
                  </IconButton>
                </InputAdornment>
              ),
            }}
            sx={{ mb: 2 }}
          />

          {copied && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {t('api_keys.key_copied')}
            </Alert>
          )}

          <Box sx={{ p: 2, bgcolor: 'info.lighter', borderRadius: 1 }}>
            <Typography variant="body2" color="info.dark">
              <strong>{t('api_keys.usage_instructions')}:</strong>
              <br />
              • {t('api_keys.usage_instruction_1')}: <code>Authorization: Bearer {apiKey.substring(0, 20)}...</code>
              <br />
              • {t('api_keys.usage_instruction_2')}
              <br />
              • {t('api_keys.usage_instruction_3')}
            </Typography>
          </Box>
        </Box>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} variant="contained" color="primary">
          {t('api_keys.saved_api_key')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
