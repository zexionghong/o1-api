import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import {
  Box,
  Menu,
  Button,
  MenuItem,
  Typography,
  IconButton,
} from '@mui/material';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

const LANGUAGES = [
  {
    value: 'zh',
    label: '中文',
    icon: '🇨🇳',
  },
  {
    value: 'en',
    label: 'English',
    icon: '🇺🇸',
  },
  {
    value: 'ja',
    label: '日本語',
    icon: '🇯🇵',
  },
];

// ----------------------------------------------------------------------

type Props = {
  variant?: 'button' | 'icon';
  size?: 'small' | 'medium' | 'large';
};

export function AuthLanguageSwitcher({ variant = 'icon', size = 'medium' }: Props) {
  const { i18n } = useTranslation();
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

  const currentLanguage = LANGUAGES.find((lang) => lang.value === i18n.language) || LANGUAGES[0];

  const handleOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLanguageChange = (languageValue: string) => {
    i18n.changeLanguage(languageValue);
    handleClose();
  };

  const open = Boolean(anchorEl);

  if (variant === 'button') {
    return (
      <>
        <Button
          onClick={handleOpen}
          startIcon={
            <Typography component="span" sx={{ fontSize: '1.2rem' }}>
              {currentLanguage.icon}
            </Typography>
          }
          endIcon={<Iconify icon="eva:chevron-down-fill" />}
          sx={{ 
            color: 'text.primary',
            minWidth: 120,
            justifyContent: 'flex-start',
          }}
          size={size}
        >
          {currentLanguage.label}
        </Button>

        <Menu
          anchorEl={anchorEl}
          open={open}
          onClose={handleClose}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'left',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'left',
          }}
        >
          {LANGUAGES.map((language) => (
            <MenuItem
              key={language.value}
              selected={language.value === i18n.language}
              onClick={() => handleLanguageChange(language.value)}
              sx={{ minWidth: 150 }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Typography component="span" sx={{ fontSize: '1.2rem' }}>
                  {language.icon}
                </Typography>
                <Typography variant="body2">
                  {language.label}
                </Typography>
              </Box>
            </MenuItem>
          ))}
        </Menu>
      </>
    );
  }

  return (
    <>
      <IconButton 
        onClick={handleOpen} 
        sx={{ 
          p: 1,
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 1,
          '&:hover': {
            borderColor: 'primary.main',
          }
        }}
        size={size}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
          <Typography component="span" sx={{ fontSize: '1.2rem' }}>
            {currentLanguage.icon}
          </Typography>
          <Iconify icon="eva:chevron-down-fill" width={16} />
        </Box>
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'right',
        }}
      >
        {LANGUAGES.map((language) => (
          <MenuItem
            key={language.value}
            selected={language.value === i18n.language}
            onClick={() => handleLanguageChange(language.value)}
            sx={{ minWidth: 150 }}
          >
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Typography component="span" sx={{ fontSize: '1.2rem' }}>
                {language.icon}
              </Typography>
              <Typography variant="body2">
                {language.label}
              </Typography>
            </Box>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
