import { useState, useCallback, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
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

interface ToolType {
  id: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  supported_models: string[];
}

interface Model {
  id: number;
  name: string;
  display_name: string;
  model_type: string;
  status: string;
}

interface ApiKey {
  id: number;
  name: string;
  key_prefix: string;
  status: string;
}

type Props = {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
};

// 硬编码的工具类型（只在创建时使用）
const TOOL_TYPES: ToolType[] = [
  {
    id: 'chatbot',
    name: 'AI Chatbot',
    description: 'Create intelligent conversational AI',
    icon: 'solar:chat-round-bold-duotone',
    color: '#45B7D1',
    supported_models: ['gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet']
  },
  {
    id: 'image_generator',
    name: 'Image Generator',
    description: 'Generate images from text descriptions',
    icon: 'solar:gallery-bold-duotone',
    color: '#4ECDC4',
    supported_models: ['dall-e-3', 'stable-diffusion-xl']
  },
  {
    id: 'text_generator',
    name: 'Text Generator',
    description: 'Generate and edit text content',
    icon: 'solar:text-bold-duotone',
    color: '#FF6B6B',
    supported_models: ['gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet']
  },
  {
    id: 'code_assistant',
    name: 'Code Assistant',
    description: 'AI-powered coding helper',
    icon: 'solar:code-bold-duotone',
    color: '#FFEAA7',
    supported_models: ['gpt-4o', 'claude-3-5-sonnet']
  }
];

// 硬编码的模型数据
const AVAILABLE_MODELS: Model[] = [
  { id: 'gpt-4o', name: 'GPT-4o', provider: 'OpenAI' },
  { id: 'gpt-4-turbo', name: 'GPT-4 Turbo', provider: 'OpenAI' },
  { id: 'claude-3-5-sonnet', name: 'Claude 3.5 Sonnet', provider: 'Anthropic' },
  { id: 'dall-e-3', name: 'DALL-E 3', provider: 'OpenAI' },
  { id: 'stable-diffusion-xl', name: 'Stable Diffusion XL', provider: 'Stability AI' }
];

// ----------------------------------------------------------------------

export function ToolCreateDialog({ open, onClose, onSuccess }: Props) {
  const { t } = useTranslation();
  const { state } = useAuthContext();
  
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState<Model[]>([]);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [toolTypes, setToolTypes] = useState<any[]>([]);
  const [selectedToolType, setSelectedToolType] = useState('');
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    model_id: 0,
    api_key_id: 0,
    is_public: false
  });

  // 获取用户的API Keys
  const fetchApiKeys = useCallback(async () => {
    if (!state.isAuthenticated) return;

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
  }, [state.isAuthenticated]);

  // 获取工具类型列表
  const fetchToolTypes = useCallback(async () => {
    try {
      const response = await fetch('http://localhost:8080/tools/types');
      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setToolTypes(result.data);
        }
      }
    } catch (error) {
      console.error('Failed to fetch tool types:', error);
    }
  }, []);

  // 获取模型列表
  const fetchModels = useCallback(async () => {
    try {
      const response = await fetch('http://localhost:8080/tools/models');
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

  useEffect(() => {
    if (open) {
      fetchToolTypes();
      fetchApiKeys();
      fetchModels();
    }
  }, [open, fetchToolTypes, fetchApiKeys, fetchModels]);

  // 根据选择的工具类型获取支持的模型
  const getSupportedModels = useCallback(() => {
    // 优先使用从API获取的工具类型数据
    if (toolTypes.length > 0) {
      const toolType = toolTypes.find(t => t.id === selectedToolType);
      if (toolType && toolType.supported_models) {
        return toolType.supported_models;
      }
    }

    // 如果API数据不可用，回退到硬编码数据
    const toolType = TOOL_TYPES.find(t => t.id === selectedToolType);
    if (!toolType) return [];

    // 使用从API获取的模型数据，如果没有则使用硬编码数据
    const availableModels = models.length > 0 ? models : AVAILABLE_MODELS;

    return availableModels.filter(model => {
      // 对于API数据，使用name字段匹配；对于硬编码数据，使用id字段匹配
      const modelIdentifier = models.length > 0 ? model.name : model.id;
      return toolType.supported_models.includes(modelIdentifier);
    });
  }, [selectedToolType, models, toolTypes]);

  const supportedModels = getSupportedModels();

  const handleToolTypeChange = useCallback((toolTypeId: string) => {
    setSelectedToolType(toolTypeId);
    setFormData(prev => ({ ...prev, model_id: 0 })); // 重置模型选择
  }, []);

  const handleSubmit = useCallback(async () => {
    if (!selectedToolType || !formData.name || !formData.model_id || !formData.api_key_id) {
      return;
    }

    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('http://localhost:8080/admin/tools/', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: formData.name,
          description: formData.description,
          tool_id: selectedToolType,
          model_id: formData.model_id,
          api_key_id: formData.api_key_id,
          is_public: formData.is_public,
          config: {}
        }),
      });

      if (response.ok) {
        onSuccess();
        handleClose();
      } else {
        const error = await response.json();
        console.error('Failed to create tool:', error);
      }
    } catch (error) {
      console.error('Failed to create tool:', error);
    } finally {
      setLoading(false);
    }
  }, [selectedToolType, formData, onSuccess]);

  const handleClose = useCallback(() => {
    setSelectedToolType('');
    setFormData({
      name: '',
      description: '',
      model_id: 0,
      api_key_id: 0,
      is_public: false
    });
    onClose();
  }, [onClose]);

  const selectedToolTypeData = TOOL_TYPES.find(t => t.id === selectedToolType);

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Iconify icon="solar:add-circle-bold" sx={{ color: 'primary.main' }} />
          <Typography variant="h6">{t('tools.create_tool')}</Typography>
        </Box>
      </DialogTitle>

      <DialogContent>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 3, pt: 1 }}>
          {/* 工具类型选择 */}
          <Box>
            <Typography variant="subtitle1" sx={{ mb: 2 }}>
              {t('tools.select_tool_type')}
            </Typography>
            <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 2 }}>
              {(toolTypes.length > 0 ? toolTypes : TOOL_TYPES).map((toolType) => (
                <Card
                  key={toolType.id}
                  sx={{
                    p: 2,
                    cursor: 'pointer',
                    border: selectedToolType === toolType.id ? 2 : 1,
                    borderColor: selectedToolType === toolType.id ? 'primary.main' : 'divider',
                    '&:hover': {
                      borderColor: 'primary.main',
                    }
                  }}
                  onClick={() => handleToolTypeChange(toolType.id)}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                    <Box
                      sx={{
                        width: 32,
                        height: 32,
                        borderRadius: 1,
                        bgcolor: toolType.color,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        mr: 1,
                      }}
                    >
                      <Iconify 
                        icon={toolType.icon} 
                        sx={{ width: 16, height: 16, color: 'white' }} 
                      />
                    </Box>
                    <Typography variant="subtitle2">{toolType.name}</Typography>
                  </Box>
                  <Typography variant="caption" color="text.secondary">
                    {toolType.description}
                  </Typography>
                </Card>
              ))}
            </Box>
          </Box>

          {/* 工具配置 */}
          {selectedToolType && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <Typography variant="subtitle1">
                {t('tools.tool_configuration')}
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

              {/* 模型选择 */}
              <FormControl fullWidth>
                <InputLabel>{t('tools.select_model')}</InputLabel>
                <Select
                  value={formData.model_id}
                  onChange={(e) => setFormData(prev => ({ ...prev, model_id: e.target.value }))}
                  label={t('tools.select_model')}
                >
                  {supportedModels.map((model) => (
                    <MenuItem key={model.id} value={model.id}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                        <Typography>{model.display_name || model.name}</Typography>
                        <Typography variant="caption" color="text.secondary">
                          {model.model_type || model.provider}
                        </Typography>
                      </Box>
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              {/* API Key选择 */}
              <FormControl fullWidth>
                <InputLabel>{t('tools.select_api_key')}</InputLabel>
                <Select
                  value={formData.api_key_id}
                  onChange={(e) => setFormData(prev => ({ ...prev, api_key_id: e.target.value }))}
                  label={t('tools.select_api_key')}
                >
                  {apiKeys.map((apiKey) => (
                    <MenuItem key={apiKey.id} value={apiKey.id}>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography>{apiKey.name}</Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                          {apiKey.key_prefix}
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

              {/* 工具预览 */}
              {selectedToolTypeData && (
                <Card variant="outlined" sx={{ p: 2 }}>
                  <Typography variant="subtitle2" sx={{ mb: 1 }}>
                    {t('tools.tool_preview')}
                  </Typography>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Box
                      sx={{
                        width: 40,
                        height: 40,
                        borderRadius: 2,
                        bgcolor: selectedToolTypeData.color,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <Iconify 
                        icon={selectedToolTypeData.icon} 
                        sx={{ width: 20, height: 20, color: 'white' }} 
                      />
                    </Box>
                    <Box>
                      <Typography variant="subtitle2">
                        {formData.name || t('tools.untitled_tool')}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {formData.description || selectedToolTypeData.description}
                      </Typography>
                    </Box>
                  </Box>
                </Card>
              )}
            </Box>
          )}

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
          disabled={!selectedToolType || !formData.name || !formData.model_id || !formData.api_key_id || loading}
          startIcon={loading ? <CircularProgress size={16} /> : <Iconify icon="solar:add-circle-bold" />}
        >
          {loading ? t('common.creating') : t('tools.create_tool')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
