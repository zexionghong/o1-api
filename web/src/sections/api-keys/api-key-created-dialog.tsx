import { useState } from 'react';

import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import InputAdornment from '@mui/material/InputAdornment';
import IconButton from '@mui/material/IconButton';

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
          <Typography variant="h6">API Key Created Successfully!</Typography>
        </Box>
      </DialogTitle>

      <DialogContent>
        <Box sx={{ mb: 3 }}>
          <Alert severity="warning" sx={{ mb: 3 }}>
            <Typography variant="body2">
              <strong>Important:</strong> This is the only time you will see the complete API key. 
              Please copy it now and store it securely. You won't be able to see it again.
            </Typography>
          </Alert>

          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
            API Key Name:
          </Typography>
          <Typography variant="body1" sx={{ mb: 3, color: 'text.secondary' }}>
            {apiKeyName}
          </Typography>

          <Typography variant="subtitle1" sx={{ mb: 1, fontWeight: 600 }}>
            Your API Key:
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
              API Key copied to clipboard!
            </Alert>
          )}

          <Box sx={{ p: 2, bgcolor: 'info.lighter', borderRadius: 1 }}>
            <Typography variant="body2" color="info.dark">
              <strong>Usage Instructions:</strong>
              <br />
              • Include this API key in your request headers: <code>Authorization: Bearer {apiKey.substring(0, 20)}...</code>
              <br />
              • Keep your API key secure and never share it publicly
              <br />
              • You can manage this key from the API Keys page (enable/disable/delete)
            </Typography>
          </Box>
        </Box>
      </DialogContent>

      <DialogActions>
        <Button onClick={handleClose} variant="contained" color="primary">
          I've Saved My API Key
        </Button>
      </DialogActions>
    </Dialog>
  );
}
