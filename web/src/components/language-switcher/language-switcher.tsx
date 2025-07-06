import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import {
  Button,
  Popover,
  MenuItem,
  MenuList,
  Typography,
  IconButton,
} from '@mui/material';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

const LANGUAGES = [
  {
    value: 'en',
    label: 'English',
    icon: '🇺🇸',
  },
  {
    value: 'zh',
    label: '中文',
    icon: '🇨🇳',
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
};

export function LanguageSwitcher({ variant = 'button' }: Props) {
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

  if (variant === 'icon') {
    return (
      <>
        <IconButton onClick={handleOpen} sx={{ p: 0 }}>
          <Typography fontSize="1.5rem">{currentLanguage.icon}</Typography>
        </IconButton>

        <Popover
          open={open}
          anchorEl={anchorEl}
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
          <MenuList sx={{ py: 1, minWidth: 180 }}>
            {LANGUAGES.map((language) => (
              <MenuItem
                key={language.value}
                selected={language.value === i18n.language}
                onClick={() => handleLanguageChange(language.value)}
                sx={{ typography: 'body2', py: 1 }}
              >
                <Typography component="span" sx={{ mr: 2, fontSize: '1.2rem' }}>
                  {language.icon}
                </Typography>
                {language.label}
              </MenuItem>
            ))}
          </MenuList>
        </Popover>
      </>
    );
  }

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
        sx={{ color: 'text.secondary' }}
      >
        {currentLanguage.label}
      </Button>

      <Popover
        open={open}
        anchorEl={anchorEl}
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
        <MenuList sx={{ py: 1, minWidth: 180 }}>
          {LANGUAGES.map((language) => (
            <MenuItem
              key={language.value}
              selected={language.value === i18n.language}
              onClick={() => handleLanguageChange(language.value)}
              sx={{ typography: 'body2', py: 1 }}
            >
              <Typography component="span" sx={{ mr: 2, fontSize: '1.2rem' }}>
                {language.icon}
              </Typography>
              {language.label}
            </MenuItem>
          ))}
        </MenuList>
      </Popover>
    </>
  );
}
