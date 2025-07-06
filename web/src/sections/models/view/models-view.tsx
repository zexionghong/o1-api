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

// 模型数据类型
interface Model {
  id: string;
  name: string;
  provider: string;
  description: string;
  category: string;
  type: 'text' | 'image' | 'audio' | 'video' | 'multimodal';
  pricing: {
    input: number;  // per 1K tokens
    output: number; // per 1K tokens
    unit: string;
  };
  capabilities: string[];
  maxTokens: number;
  status: 'available' | 'beta' | 'deprecated';
  icon: string;
  color: string;
}

// 硬编码的模型数据
const MODELS_DATA: Model[] = [
  {
    id: 'gpt-4o',
    name: 'GPT-4o',
    provider: 'OpenAI',
    description: 'Most advanced multimodal model with vision, audio, and text capabilities',
    category: 'Premium',
    type: 'multimodal',
    pricing: { input: 0.005, output: 0.015, unit: '1K tokens' },
    capabilities: ['Text Generation', 'Vision', 'Audio', 'Code', 'Reasoning'],
    maxTokens: 128000,
    status: 'available',
    icon: 'solar:cpu-bolt-bold-duotone',
    color: '#10B981'
  },
  {
    id: 'gpt-4-turbo',
    name: 'GPT-4 Turbo',
    provider: 'OpenAI',
    description: 'High-performance model optimized for speed and efficiency',
    category: 'Premium',
    type: 'text',
    pricing: { input: 0.01, output: 0.03, unit: '1K tokens' },
    capabilities: ['Text Generation', 'Code', 'Analysis', 'Reasoning'],
    maxTokens: 128000,
    status: 'available',
    icon: 'solar:cpu-bolt-bold-duotone',
    color: '#3B82F6'
  },
  {
    id: 'claude-3-5-sonnet',
    name: 'Claude 3.5 Sonnet',
    provider: 'Anthropic',
    description: 'Advanced reasoning and analysis with excellent safety features',
    category: 'Premium',
    type: 'text',
    pricing: { input: 0.003, output: 0.015, unit: '1K tokens' },
    capabilities: ['Text Generation', 'Analysis', 'Code', 'Safety'],
    maxTokens: 200000,
    status: 'available',
    icon: 'solar:brain-bold-duotone',
    color: '#8B5CF6'
  },
  {
    id: 'dall-e-3',
    name: 'DALL-E 3',
    provider: 'OpenAI',
    description: 'State-of-the-art image generation from text descriptions',
    category: 'Creative',
    type: 'image',
    pricing: { input: 0.04, output: 0, unit: 'image' },
    capabilities: ['Text to Image', 'High Quality', 'Style Control'],
    maxTokens: 4000,
    status: 'available',
    icon: 'solar:gallery-bold-duotone',
    color: '#F59E0B'
  },
  {
    id: 'stable-diffusion-xl',
    name: 'Stable Diffusion XL',
    provider: 'Stability AI',
    description: 'Open-source image generation with fine-tuning capabilities',
    category: 'Creative',
    type: 'image',
    pricing: { input: 0.02, output: 0, unit: 'image' },
    capabilities: ['Text to Image', 'Style Transfer', 'Fine-tuning'],
    maxTokens: 2000,
    status: 'available',
    icon: 'solar:palette-bold-duotone',
    color: '#EC4899'
  },
  {
    id: 'whisper-1',
    name: 'Whisper',
    provider: 'OpenAI',
    description: 'Automatic speech recognition and transcription',
    category: 'Audio',
    type: 'audio',
    pricing: { input: 0.006, output: 0, unit: 'minute' },
    capabilities: ['Speech to Text', 'Multi-language', 'Noise Robust'],
    maxTokens: 0,
    status: 'available',
    icon: 'solar:microphone-bold-duotone',
    color: '#06B6D4'
  }
];

const CATEGORIES = ['All', 'Premium', 'Creative', 'Audio', 'Experimental'];
const TYPES = ['All', 'text', 'image', 'audio', 'video', 'multimodal'];

// ----------------------------------------------------------------------

export function ModelsView() {
  const { t } = useTranslation();
  const [selectedCategory, setSelectedCategory] = useState('All');
  const [selectedType, setSelectedType] = useState('All');

  const handleCategoryChange = useCallback((category: string) => {
    setSelectedCategory(category);
  }, []);

  const handleTypeChange = useCallback((type: string) => {
    setSelectedType(type);
  }, []);

  const filteredModels = MODELS_DATA.filter(model => {
    const categoryMatch = selectedCategory === 'All' || model.category === selectedCategory;
    const typeMatch = selectedType === 'All' || model.type === selectedType;
    return categoryMatch && typeMatch;
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'available':
        return 'success';
      case 'beta':
        return 'warning';
      case 'deprecated':
        return 'error';
      default:
        return 'default';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'text':
        return 'solar:text-bold-duotone';
      case 'image':
        return 'solar:gallery-bold-duotone';
      case 'audio':
        return 'solar:microphone-bold-duotone';
      case 'video':
        return 'solar:videocamera-bold-duotone';
      case 'multimodal':
        return 'solar:layers-bold-duotone';
      default:
        return 'solar:cpu-bolt-bold-duotone';
    }
  };

  const formatPricing = (model: Model) => {
    if (model.type === 'text' || model.type === 'multimodal') {
      return `$${model.pricing.input}/$${model.pricing.output} per ${model.pricing.unit}`;
    }
    return `$${model.pricing.input} per ${model.pricing.unit}`;
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ mb: 1 }}>
          {t('models.title')}
        </Typography>
        <Typography variant="body1" color="text.secondary">
          {t('models.description')}
        </Typography>
      </Box>

      {/* 筛选器 */}
      <Box sx={{ mb: 4 }}>
        <Box sx={{ mb: 3 }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            {t('models.categories')}
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
            {CATEGORIES.map((category) => (
              <Chip
                key={category}
                label={t(`models.category_${category.toLowerCase()}`)}
                onClick={() => handleCategoryChange(category)}
                variant={selectedCategory === category ? 'filled' : 'outlined'}
                color={selectedCategory === category ? 'primary' : 'default'}
                sx={{ cursor: 'pointer' }}
              />
            ))}
          </Box>
        </Box>

        <Box>
          <Typography variant="h6" sx={{ mb: 2 }}>
            {t('models.types')}
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
            {TYPES.map((type) => (
              <Chip
                key={type}
                label={t(`models.type_${type.toLowerCase()}`)}
                onClick={() => handleTypeChange(type)}
                variant={selectedType === type ? 'filled' : 'outlined'}
                color={selectedType === type ? 'secondary' : 'default'}
                sx={{ cursor: 'pointer' }}
                icon={<Iconify icon={getTypeIcon(type)} />}
              />
            ))}
          </Box>
        </Box>
      </Box>

      {/* 模型网格 */}
      <Grid container spacing={3}>
        {filteredModels.map((model) => (
          <Grid key={model.id} size={{ xs: 12, md: 6, lg: 4 }}>
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
                      bgcolor: model.color,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mr: 2,
                    }}
                  >
                    <Iconify 
                      icon={model.icon} 
                      sx={{ width: 24, height: 24, color: 'white' }} 
                    />
                  </Box>
                  <Box sx={{ flexGrow: 1 }}>
                    <Typography variant="h6" sx={{ mb: 0.5 }}>
                      {model.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {model.provider}
                    </Typography>
                  </Box>
                  <Chip
                    label={model.status}
                    size="small"
                    color={getStatusColor(model.status) as any}
                    variant="outlined"
                  />
                </Box>

                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  {model.description}
                </Typography>

                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                    {t('models.pricing')}:
                  </Typography>
                  <Typography variant="body2" sx={{ fontWeight: 600, color: 'primary.main' }}>
                    {formatPricing(model)}
                  </Typography>
                </Box>

                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary" sx={{ mb: 1, display: 'block' }}>
                    {t('models.capabilities')}:
                  </Typography>
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                    {model.capabilities.map((capability, index) => (
                      <Chip
                        key={index}
                        label={capability}
                        size="small"
                        variant="outlined"
                        sx={{ fontSize: '0.75rem' }}
                      />
                    ))}
                  </Box>
                </Box>

                {model.maxTokens > 0 && (
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      {t('models.max_tokens')}: {model.maxTokens.toLocaleString()}
                    </Typography>
                  </Box>
                )}
              </CardContent>

              <CardActions sx={{ p: 2, pt: 0 }}>
                <Button
                  fullWidth
                  variant="outlined"
                  startIcon={<Iconify icon="solar:info-circle-bold" />}
                >
                  {t('models.view_details')}
                </Button>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      {/* 空状态 */}
      {filteredModels.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Iconify 
            icon="solar:box-bold-duotone" 
            sx={{ width: 64, height: 64, color: 'text.disabled', mb: 2 }} 
          />
          <Typography variant="h6" color="text.secondary">
            {t('models.no_models_found')}
          </Typography>
          <Typography variant="body2" color="text.disabled">
            {t('models.try_different_filter')}
          </Typography>
        </Box>
      )}
    </Box>
  );
}
