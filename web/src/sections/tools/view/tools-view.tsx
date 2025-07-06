import { useTranslation } from 'react-i18next';
import { useCallback, useEffect, useMemo, useState } from 'react';

import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Snackbar from '@mui/material/Snackbar';
import Typography from '@mui/material/Typography';
import Grid from '@mui/system/Grid';

import { Iconify } from 'src/components/iconify';

import { ToolEditDialog } from '../tool-edit-dialog';
import { ToolCreateDialog } from '../tool-create-dialog';

// ----------------------------------------------------------------------

// 用户创建的工具数据类型
interface UserTool {
  id: string;
  name: string;
  description: string;
  type: 'chatbot' | 'image_generator' | 'text_generator' | 'code_assistant' | 'data_analyzer';
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

// 工具类型配置
const TOOL_TYPES = [
  {
    id: 'chatbot',
    name: 'AI Chatbot',
    description: 'Create intelligent conversational AI',
    icon: 'solar:chat-round-bold-duotone',
    color: '#45B7D1',
    supportedModels: ['gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet']
  },
  {
    id: 'image_generator',
    name: 'Image Generator',
    description: 'Generate images from text descriptions',
    icon: 'solar:gallery-bold-duotone',
    color: '#4ECDC4',
    supportedModels: ['dall-e-3', 'stable-diffusion-xl']
  },
  {
    id: 'text_generator',
    name: 'Text Generator',
    description: 'Generate and edit text content',
    icon: 'solar:text-bold-duotone',
    color: '#FF6B6B',
    supportedModels: ['gpt-4o', 'gpt-4-turbo', 'claude-3-5-sonnet']
  },
  {
    id: 'code_assistant',
    name: 'Code Assistant',
    description: 'AI-powered coding helper',
    icon: 'solar:code-bold-duotone',
    color: '#FFEAA7',
    supportedModels: ['gpt-4o', 'claude-3-5-sonnet']
  }
];

const CATEGORIES = ['All', 'My Tools', 'Public', 'Shared'];

// ----------------------------------------------------------------------

export function ToolsView() {
  const { t } = useTranslation();
  const [selectedCategory, setSelectedCategory] = useState('All');
  const [userTools, setUserTools] = useState<UserTool[]>([]);
  const [loading, setLoading] = useState(false);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [editingTool, setEditingTool] = useState<UserTool | null>(null);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' as 'success' | 'error' });

  const handleCategoryChange = useCallback((category: string) => {
    setSelectedCategory(category);
  }, []);

  // 获取用户工具列表
  const fetchUserTools = useCallback(async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`http://localhost:8080/admin/tools/?category=${selectedCategory.toLowerCase().replace(' ', '_')}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data) {
          setUserTools(result.data);
        }
      } else {
        // 如果API失败，使用模拟数据
        const mockTools: UserTool[] = [
          {
            id: '1',
            name: 'My AI Assistant',
            description: 'A helpful AI assistant for daily tasks',
            type: 'chatbot',
            model_id: 'gpt-4o',
            model_name: 'GPT-4o',
            api_key_id: 'key1',
            config: { temperature: 0.7, max_tokens: 2000 },
            is_public: true,
            share_url: 'https://tools.example.com/share/abc123',
            created_at: '2024-01-15T10:00:00Z',
            updated_at: '2024-01-15T10:00:00Z',
            usage_count: 156,
            tool: { name: 'AI Chatbot' },
            creator: { username: 'john_doe' }
          },
          {
            id: '2',
            name: 'Logo Generator',
            description: 'Generate professional logos for brands',
            type: 'image_generator',
            model_id: 'dall-e-3',
            model_name: 'DALL-E 3',
            api_key_id: 'key2',
            config: { style: 'professional', size: '1024x1024' },
            is_public: false,
            created_at: '2024-01-10T15:30:00Z',
            updated_at: '2024-01-12T09:15:00Z',
            usage_count: 23,
            tool: { name: 'Image Generator' },
            creator: { username: 'john_doe' }
          }
        ];
        setUserTools(mockTools);
      }
    } catch (error) {
      console.error('Failed to fetch user tools:', error);
      setUserTools([]);
    } finally {
      setLoading(false);
    }
  }, [selectedCategory]);

  useEffect(() => {
    fetchUserTools();
  }, [fetchUserTools]);

  const filteredTools = useMemo(() => {
    switch (selectedCategory) {
      case 'My Tools':
        return userTools;
      case 'Public':
        return userTools.filter(tool => tool.is_public);
      case 'Shared':
        return userTools.filter(tool => tool.share_url);
      default:
        return userTools;
    }
  }, [selectedCategory, userTools]);

  const getToolTypeConfig = (type: string) => TOOL_TYPES.find(toolType => toolType.id === type) || TOOL_TYPES[0];

  const handleCreateTool = useCallback(() => {
    setShowCreateDialog(true);
  }, []);

  const handleCreateSuccess = useCallback(() => {
    fetchUserTools(); // 重新获取工具列表
  }, [fetchUserTools]);

  const handleEditSuccess = useCallback(() => {
    fetchUserTools(); // 重新获取工具列表
    setShowEditDialog(false);
    setEditingTool(null);
  }, [fetchUserTools]);

  const handleToolAction = useCallback((tool: UserTool, action: 'edit' | 'share' | 'delete' | 'launch') => {
    switch (action) {
      case 'launch':
        // 启动工具
        window.open(`/tools/${tool.id}`, '_blank');
        break;
      case 'edit':
        // 编辑工具
        setEditingTool(tool);
        setShowEditDialog(true);
        break;
      case 'share': {
        // 分享工具
        const shareUrl = tool.share_url || `${window.location.origin}/tools/${tool.id}`;
        navigator.clipboard.writeText(shareUrl)
          .then(() => {
            setSnackbar({ open: true, message: t('tools.share_success'), severity: 'success' });
          })
          .catch(() => {
            setSnackbar({ open: true, message: t('tools.share_failed'), severity: 'error' });
          });
        break;
      }
      case 'delete':
        // 删除工具
        console.log('Delete tool:', tool.id);
        break;
      default:
        break;
    }
  }, [t]);

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <Box>
          <Typography variant="h4" sx={{ mb: 1 }}>
            {t('tools.title')}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {t('tools.description_new')}
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<Iconify icon="solar:pen-bold" />}
          onClick={handleCreateTool}
          sx={{ flexShrink: 0 }}
        >
          {t('tools.create_tool')}
        </Button>
      </Box>

      {/* 分类筛选 */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          {t('tools.categories')}
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
          {CATEGORIES.map((category) => (
            <Chip
              key={category}
              label={t(`tools.category_${category.toLowerCase()}`)}
              onClick={() => handleCategoryChange(category)}
              variant={selectedCategory === category ? 'filled' : 'outlined'}
              color={selectedCategory === category ? 'primary' : 'default'}
              sx={{ cursor: 'pointer' }}
            />
          ))}
        </Box>
      </Box>

      {/* 工具网格 */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
          <CircularProgress />
        </Box>
      ) : (
        <Grid container spacing={3}>
          {filteredTools.map((tool) => {
            const typeConfig = getToolTypeConfig(tool.type);
            return (
              <Grid key={tool.id} size={{ xs: 12, sm: 6, md: 4 }}>
                <Card
                  sx={{
                    height: '100%',
                    display: 'flex',
                    flexDirection: 'column',
                    transition: 'all 0.3s ease',
                    '&:hover': {
                      transform: 'translateY(-4px)',
                      boxShadow: (theme) => theme.shadows[8],
                    }
                  }}
                >
                  <CardContent sx={{ flexGrow: 1 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                      <Box
                        sx={{
                          width: 48,
                          height: 48,
                          borderRadius: 2,
                          bgcolor: typeConfig.color,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          mr: 2,
                        }}
                      >
                        <Iconify
                          icon="solar:pen-bold"
                          sx={{ width: 24, height: 24, color: 'white' }}
                        />
                      </Box>
                      <Box sx={{ flexGrow: 1 }}>
                        <Typography variant="h6" sx={{ mb: 0.5 }}>
                          {tool.name}
                        </Typography>
                        <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                          {tool.tool.name}
                        </Typography>
                        <Box sx={{ display: 'flex', gap: 0.5 }}>
                          <Chip
                            label={tool.model_name}
                            size="small"
                            variant="outlined"
                            sx={{ fontSize: '0.75rem' }}
                          />
                          {tool.is_public && (
                            <Chip
                              label={t('tools.public')}
                              size="small"
                              color="success"
                              variant="outlined"
                              sx={{ fontSize: '0.75rem' }}
                            />
                          )}
                        </Box>
                      </Box>
                    </Box>

                    <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                      {tool.description}
                    </Typography>

                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                      <Box>
                        <Typography variant="caption" color="text.secondary">
                          {t('tools.creator')}:
                        </Typography>
                        <Typography variant="body2">
                          {tool.creator.username}
                        </Typography>
                      </Box>
                      <Box sx={{ textAlign: 'right' }}>
                        <Typography variant="caption" color="text.secondary">
                          {t('tools.usage_count')}:
                        </Typography>
                        <Typography variant="body2">
                          {tool.usage_count}
                        </Typography>
                      </Box>
                    </Box>

                    <Typography variant="caption" color="text.secondary">
                      {t('tools.created')}: {new Date(tool.created_at).toLocaleDateString()}
                    </Typography>
                  </CardContent>

                  <CardActions sx={{ p: 2, pt: 0, gap: 1 }}>
                    <Button
                      variant="contained"
                      size="small"
                      onClick={() => handleToolAction(tool, 'launch')}
                      startIcon={<Iconify icon="solar:play-bold" />}
                      sx={{ flex: 1 }}
                    >
                      {t('tools.launch')}
                    </Button>
                    <Button
                      variant="outlined"
                      size="small"
                      onClick={() => handleToolAction(tool, 'edit')}
                      startIcon={<Iconify icon="solar:pen-bold" />}
                    >
                      {t('tools.edit')}
                    </Button>
                    <Button
                      variant="outlined"
                      size="small"
                      onClick={() => handleToolAction(tool, 'share')}
                      startIcon={<Iconify icon="solar:share-bold" />}
                    >
                      {t('tools.share')}
                    </Button>
                  </CardActions>
                </Card>
              </Grid>
            );
          })}
        </Grid>
      )}

      {/* 空状态 */}
      {filteredTools.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Iconify
            icon="solar:pen-bold"
            sx={{ width: 64, height: 64, color: 'text.disabled', mb: 2 }}
          />
          <Typography variant="h6" color="text.secondary">
            {t('tools.no_tools_found')}
          </Typography>
          <Typography variant="body2" color="text.disabled">
            {t('tools.try_different_category')}
          </Typography>
        </Box>
      )}

      {/* 创建工具对话框 */}
      <ToolCreateDialog
        open={showCreateDialog}
        onClose={() => setShowCreateDialog(false)}
        onSuccess={handleCreateSuccess}
      />

      {/* 编辑工具对话框 */}
      <ToolEditDialog
        open={showEditDialog}
        tool={editingTool}
        onClose={() => {
          setShowEditDialog(false);
          setEditingTool(null);
        }}
        onSuccess={handleEditSuccess}
      />

      {/* 分享成功/失败提示 */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      >
        <Alert
          onClose={() => setSnackbar({ ...snackbar, open: false })}
          severity={snackbar.severity}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
