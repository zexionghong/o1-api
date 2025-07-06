import type { IconButtonProps } from '@mui/material/IconButton';

import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { usePopover } from 'minimal-shared/hooks';

import Box from '@mui/material/Box';
import Popover from '@mui/material/Popover';
import MenuList from '@mui/material/MenuList';
import IconButton from '@mui/material/IconButton';
import MenuItem, { menuItemClasses } from '@mui/material/MenuItem';

// ----------------------------------------------------------------------

const LANGUAGES = [
  {
    value: 'en',
    label: 'English',
    icon: '/assets/icons/flags/ic-flag-en.svg',
  },
  {
    value: 'zh',
    label: '中文',
    icon: '/assets/icons/flags/ic-flag-cn.svg',
  },
  {
    value: 'ja',
    label: '日本語',
    icon: '/assets/icons/flags/ic-flag-jp.svg',
  },
];

export type LanguagePopoverProps = IconButtonProps;

export function LanguagePopover({ sx, ...other }: LanguagePopoverProps) {
  const { i18n } = useTranslation();
  const { open, anchorEl, onClose, onOpen } = usePopover();

  const handleChangeLang = useCallback(
    (newLang: string) => {
      i18n.changeLanguage(newLang);
      onClose();
    },
    [i18n, onClose]
  );

  const currentLang = LANGUAGES.find((lang) => lang.value === i18n.language) || LANGUAGES[0];

  const renderFlag = (label?: string, icon?: string) => (
    <Box
      component="img"
      alt={label}
      src={icon}
      sx={{ width: 26, height: 20, borderRadius: 0.5, objectFit: 'cover' }}
    />
  );

  const renderMenuList = () => (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={onClose}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      transformOrigin={{ vertical: 'top', horizontal: 'right' }}
    >
      <MenuList
        sx={{
          p: 0.5,
          gap: 0.5,
          width: 160,
          minHeight: 72,
          display: 'flex',
          flexDirection: 'column',
          [`& .${menuItemClasses.root}`]: {
            px: 1,
            gap: 2,
            borderRadius: 0.75,
            [`&.${menuItemClasses.selected}`]: {
              bgcolor: 'action.selected',
              fontWeight: 'fontWeightSemiBold',
            },
          },
        }}
      >
        {LANGUAGES.map((option) => (
          <MenuItem
            key={option.value}
            selected={option.value === currentLang?.value}
            onClick={() => handleChangeLang(option.value)}
          >
            {renderFlag(option.label, option.icon)}
            {option.label}
          </MenuItem>
        ))}
      </MenuList>
    </Popover>
  );

  return (
    <>
      <IconButton
        aria-label="Languages button"
        onClick={onOpen}
        sx={[
          (theme) => ({
            p: 0,
            width: 40,
            height: 40,
            ...(open && { bgcolor: theme.vars.palette.action.selected }),
          }),
          ...(Array.isArray(sx) ? sx : [sx]),
        ]}
        {...other}
      >
        {renderFlag(currentLang?.label, currentLang?.icon)}
      </IconButton>

      {renderMenuList()}
    </>
  );
}
