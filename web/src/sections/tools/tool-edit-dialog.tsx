import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import FormControl from '@mui/material/FormControl';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import InputLabel from '@mui/material/InputLabel';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import CircularProgress from '@mui/material/CircularProgress';

import { useAuthContext } from 'src/contexts/auth-context';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

interface Model {
  id: number;
  name: string;
  display_name?: string;
  provider: string;
}

interface ApiKey {
  id: number;
  name: string;
  key_prefix: string;
  status: string;
}

interface UserTool {
  id: string;
  name: string;
  description: string;
  type: string;
  model_id: string;
  model_name: string;
  api_key_id: string;
  config: Record<string, any>;
  is_public: boolean;
  share_url?: string;
  created_at: string;
  updated_at: string;
  usage_count: number;
  tool: {
    name: string;
  };
  creator: {
    username: string;
    avatar?: string;
  };
}

interface Props {
  open: boolean;
  tool: UserTool | null;
  onClose: () => void;
  onSuccess: () => void;
}

export function ToolEditDialog({ open, tool, onClose, onSuccess }: Props) {
  const { t } = useTranslation();
  const { state } = useAuthContext();
  
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState<Model[]>([]);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    model_id: 0,
    api_key_id: 0,
    is_public: false
  });

  // 初始化表单数据
  useEffect(() => {
    if (tool && open) {
      setFormData({
        name: tool.name,
        description: tool.description,
        model_id: parseInt(tool.model_id, 10),
        api_key_id: parseInt(tool.api_key_id, 10),
        is_public: tool.is_public
      });
    }
  }, [tool, open]);

  // 获取模型列表
  const fetchModels = useCallback(async () => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('http://localhost:8080/tools/models', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setModels(result.data);
        }
      }
    } catch (error) {
      console.error('Failed to fetch models:', error);
    }
  }, []);

  // 获取API密钥列表
  const fetchApiKeys = useCallback(async () => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('http://localhost:8080/admin/tools/api-keys', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setApiKeys(result.data);
        }
      }
    } catch (error) {
      console.error('Failed to fetch API keys:', error);
    }
  }, []);

  useEffect(() => {
    if (open) {
      fetchModels();
      fetchApiKeys();
    }
  }, [open, fetchModels, fetchApiKeys]);

  const handleSubmit = useCallback(async () => {
    if (!tool) return;

    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`http://localhost:8080/admin/tools/${tool.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          model_id: formData.model_id,
          api_key_id: formData.api_key_id,
          is_public: formData.is_public,
        }),
      });

      if (response.ok) {
        onSuccess();
        handleClose();
      } else {
        const error = await response.json();
        console.error('Failed to update tool:', error);
      }
    } catch (error) {
      console.error('Failed to update tool:', error);
    } finally {
      setLoading(false);
    }
  }, [tool, formData, onSuccess]);

  const handleClose = useCallback(() => {
    setFormData({
      name: '',
      description: '',
      model_id: 0,
      api_key_id: 0,
      is_public: false
    });
    onClose();
  }, [onClose]);

  if (!tool) return null;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Iconify icon="solar:pen-bold" sx={{ color: 'primary.main' }} />
          <Typography variant="h6">{t('tools.edit')}: {tool.name}</Typography>
        </Box>
      </DialogTitle>

      <DialogContent sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
          {/* 基本信息 */}
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Typography variant="subtitle1">
              {t('tools.basic_info')}
            </Typography>

            <TextField
              fullWidth
              label={t('tools.tool_name')}
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              required
            />

            <TextField
              fullWidth
              label={t('tools.tool_description')}
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              multiline
              rows={3}
            />
          </Box>

          {/* 配置信息 */}
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <Typography variant="subtitle1">
              {t('tools.configuration')}
            </Typography>

            {/* 模型选择 */}
            <FormControl fullWidth required>
              <InputLabel>{t('tools.select_model')}</InputLabel>
              <Select
                value={formData.model_id}
                onChange={(e) => setFormData(prev => ({ ...prev, model_id: e.target.value as number }))}
                label={t('tools.select_model')}
              >
                {models.map((model) => (
                  <MenuItem key={model.id} value={model.id}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                      <Typography>{model.display_name || model.name}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {model.provider}
                      </Typography>
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* API密钥选择 */}
            <FormControl fullWidth required>
              <InputLabel>{t('tools.select_api_key')}</InputLabel>
              <Select
                value={formData.api_key_id}
                onChange={(e) => setFormData(prev => ({ ...prev, api_key_id: e.target.value as number }))}
                label={t('tools.select_api_key')}
              >
                {apiKeys.map((apiKey) => (
                  <MenuItem key={apiKey.id} value={apiKey.id}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                      <Typography>{apiKey.name}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {apiKey.key_prefix}••••
                      </Typography>
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>

            {/* 公开设置 */}
            <FormControlLabel
              control={
                <Switch
                  checked={formData.is_public}
                  onChange={(e) => setFormData(prev => ({ ...prev, is_public: e.target.checked }))}
                />
              }
              label={t('tools.make_public')}
            />
          </Box>

          {/* 错误提示 */}
          {apiKeys.length === 0 && (
            <Alert severity="warning">
              {t('tools.no_api_keys_available')}
              <Button 
                size="small" 
                sx={{ ml: 1 }}
                onClick={() => window.open('/api-keys', '_blank')}
              >
                {t('tools.create_api_key')}
              </Button>
            </Alert>
          )}
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 3 }}>
        <Button onClick={handleClose}>
          {t('common.cancel')}
        </Button>
        <Button
          variant="contained"
          onClick={handleSubmit}
          disabled={!formData.name || !formData.model_id || !formData.api_key_id || loading}
          startIcon={loading ? <CircularProgress size={16} /> : <Iconify icon="solar:pen-bold" />}
        >
          {loading ? t('common.updating') : t('tools.update_tool')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
