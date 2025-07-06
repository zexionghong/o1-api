import { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Grid from '@mui/system/Grid';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

// 工具数据类型
interface Tool {
  id: string;
  name: string;
  description: string;
  category: string;
  icon: string;
  status: 'available' | 'coming_soon' | 'beta';
  features: string[];
  color: string;
}

// 硬编码的工具数据
const TOOLS_DATA: Tool[] = [
  {
    id: 'n8n-workflow',
    name: 'Workflow Automation',
    description: 'Create powerful automation workflows with visual node-based editor, similar to n8n',
    category: 'Automation',
    icon: 'solar:settings-bold-duotone',
    status: 'coming_soon',
    features: ['Visual Editor', 'API Integrations', 'Scheduled Tasks', 'Webhooks'],
    color: '#FF6B6B'
  },
  {
    id: 'image-generator',
    name: 'AI Image Generator',
    description: 'Generate stunning images from text descriptions using advanced AI models',
    category: 'Creative',
    icon: 'solar:gallery-bold-duotone',
    status: 'beta',
    features: ['Text to Image', 'Style Transfer', 'Image Editing', 'Batch Generation'],
    color: '#4ECDC4'
  },
  {
    id: 'chatbot',
    name: 'AI Chatbot',
    description: 'Build and deploy intelligent chatbots for customer service and support',
    category: 'Communication',
    icon: 'solar:chat-round-bold-duotone',
    status: 'available',
    features: ['Natural Language', 'Multi-language', 'Custom Training', 'API Integration'],
    color: '#45B7D1'
  },
  {
    id: 'video-generator',
    name: 'AI Video Generator',
    description: 'Create professional videos from text scripts with AI-powered generation',
    category: 'Creative',
    icon: 'solar:videocamera-bold-duotone',
    status: 'coming_soon',
    features: ['Text to Video', 'Voice Synthesis', 'Scene Generation', 'Auto Editing'],
    color: '#96CEB4'
  },
  {
    id: 'code-assistant',
    name: 'Code Assistant',
    description: 'AI-powered coding assistant for code generation, review, and optimization',
    category: 'Development',
    icon: 'solar:code-bold-duotone',
    status: 'beta',
    features: ['Code Generation', 'Bug Detection', 'Code Review', 'Documentation'],
    color: '#FFEAA7'
  },
  {
    id: 'data-analyzer',
    name: 'Data Analyzer',
    description: 'Analyze and visualize data with AI-powered insights and recommendations',
    category: 'Analytics',
    icon: 'solar:chart-bold-duotone',
    status: 'coming_soon',
    features: ['Data Visualization', 'Pattern Recognition', 'Predictive Analytics', 'Reports'],
    color: '#DDA0DD'
  }
];

const CATEGORIES = ['All', 'Automation', 'Creative', 'Communication', 'Development', 'Analytics'];

// ----------------------------------------------------------------------

export function ToolsView() {
  const { t } = useTranslation();
  const [selectedCategory, setSelectedCategory] = useState('All');

  const handleCategoryChange = useCallback((category: string) => {
    setSelectedCategory(category);
  }, []);

  const filteredTools = selectedCategory === 'All' 
    ? TOOLS_DATA 
    : TOOLS_DATA.filter(tool => tool.category === selectedCategory);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'available':
        return 'success';
      case 'beta':
        return 'warning';
      case 'coming_soon':
        return 'default';
      default:
        return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'available':
        return t('tools.available');
      case 'beta':
        return t('tools.beta');
      case 'coming_soon':
        return t('tools.coming_soon');
      default:
        return status;
    }
  };

  const handleToolAction = useCallback((tool: Tool) => {
    if (tool.status === 'available') {
      // TODO: 导航到工具页面
      console.log('Launch tool:', tool.id);
    } else if (tool.status === 'beta') {
      // TODO: 显示beta访问对话框
      console.log('Request beta access:', tool.id);
    } else {
      // TODO: 显示即将推出提示
      console.log('Tool coming soon:', tool.id);
    }
  }, []);

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ mb: 1 }}>
          {t('tools.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('tools.description')}
        </Typography>
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
      <Grid container spacing={3}>
        {filteredTools.map((tool) => (
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
                      bgcolor: tool.color,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mr: 2,
                    }}
                  >
                    <Iconify 
                      icon={tool.icon} 
                      sx={{ width: 24, height: 24, color: 'white' }} 
                    />
                  </Box>
                  <Box sx={{ flexGrow: 1 }}>
                    <Typography variant="h6" sx={{ mb: 0.5 }}>
                      {tool.name}
                    </Typography>
                    <Chip
                      label={getStatusText(tool.status)}
                      size="small"
                      color={getStatusColor(tool.status) as any}
                      variant="outlined"
                    />
                  </Box>
                </Box>

                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  {tool.description}
                </Typography>

                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                    {t('tools.features')}:
                  </Typography>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                    {tool.features.map((feature, index) => (
                      <Chip
                        key={index}
                        label={feature}
                        size="small"
                        variant="outlined"
                        sx={{ fontSize: '0.75rem' }}
                      />
                    ))}
                  </Box>
                </Box>
              </CardContent>

              <CardActions sx={{ p: 2, pt: 0 }}>
                <Button
                  fullWidth
                  variant={tool.status === 'available' ? 'contained' : 'outlined'}
                  onClick={() => handleToolAction(tool)}
                  disabled={tool.status === 'coming_soon'}
                  startIcon={
                    tool.status === 'available' ? (
                      <Iconify icon="solar:play-bold" />
                    ) : tool.status === 'beta' ? (
                      <Iconify icon="solar:test-tube-bold" />
                    ) : (
                      <Iconify icon="solar:clock-circle-bold" />
                    )
                  }
                >
                  {tool.status === 'available' && t('tools.launch')}
                  {tool.status === 'beta' && t('tools.try_beta')}
                  {tool.status === 'coming_soon' && t('tools.coming_soon')}
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* 空状态 */}
      {filteredTools.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Iconify 
            icon="solar:box-bold-duotone" 
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
    </Box>
  );
}
