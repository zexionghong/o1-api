import { useState } from 'react';
import { useTranslation } from 'react-i18next';

import Box from '@mui/material/Box';
import Popover from '@mui/material/Popover';
import TableRow from '@mui/material/TableRow';
import Checkbox from '@mui/material/Checkbox';
import MenuList from '@mui/material/MenuList';
import TableCell from '@mui/material/TableCell';
import IconButton from '@mui/material/IconButton';
import MenuItem, { menuItemClasses } from '@mui/material/MenuItem';
import Chip from '@mui/material/Chip';
import Typography from '@mui/material/Typography';
import Tooltip from '@mui/material/Tooltip';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

interface ApiKey {
  id: number;
  name: string;
  key?: string; // 完整的API密钥
  key_prefix: string;
  status: 'active' | 'inactive' | 'revoked';
  permissions?: {
    allowed_providers?: string[];
    allowed_models?: string[];
  };
  expires_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

interface ApiKeyTableRowProps {
  row: ApiKey;
  selected: boolean;
  onSelectRow: (event: React.MouseEvent<unknown>) => void;
  onViewDetails: () => void;
  onStatusChange: (status: 'active' | 'inactive') => void;
  onDeleteRow: () => void;
}

// ----------------------------------------------------------------------

export function ApiKeyTableRow({ row, selected, onSelectRow, onViewDetails, onStatusChange, onDeleteRow }: ApiKeyTableRowProps) {
  const { t } = useTranslation();
  const [openPopover, setOpenPopover] = useState<HTMLButtonElement | null>(null);
  const [copied, setCopied] = useState(false);

  const handleOpenPopover = (event: React.MouseEvent<HTMLButtonElement>) => {
    setOpenPopover(event.currentTarget);
  };

  const handleClosePopover = () => {
    setOpenPopover(null);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'inactive':
        return 'warning';
      case 'revoked':
        return 'error';
      default:
        return 'default';
    }
  };

  const formatDate = (dateString: string | null | undefined) => {
    if (!dateString) return t('api_keys.never');
    return new Date(dateString).toLocaleDateString();
  };

  const handleCopyApiKey = async () => {
    if (!row.key) {
      console.error('API key not available for copying');
      return;
    }

    try {
      await navigator.clipboard.writeText(row.key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error('Failed to copy API key:', error);
      // 降级方案
      try {
        const textArea = document.createElement('textarea');
        textArea.value = row.key;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (fallbackError) {
        console.error('Fallback copy failed:', fallbackError);
      }
    }
  };

  return (
    <>
      <TableRow hover tabIndex={-1} role="checkbox" selected={selected}>
        <TableCell padding="checkbox">
          <Checkbox disableRipple checked={selected} onChange={onSelectRow} />
        </TableCell>

        <TableCell component="th" scope="row">
          <Typography variant="subtitle2" noWrap>
            {row.name}
          </Typography>
        </TableCell>

        <TableCell>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Box sx={{ fontFamily: 'monospace', fontSize: '0.875rem' }}>
              {row.key_prefix}••••••••••••
            </Box>
            <Tooltip title={copied ? t('api_keys.key_copied') : t('api_keys.copy_key')}>
              <IconButton
                size="small"
                onClick={handleCopyApiKey}
                color={copied ? 'success' : 'default'}
                sx={{ opacity: 0.7, '&:hover': { opacity: 1 } }}
                disabled={!row.key}
              >
                <Iconify
                  icon={copied ? 'solar:check-circle-bold' : 'solar:copy-bold'}
                  width={14}
                />
              </IconButton>
            </Tooltip>
          </Box>
        </TableCell>

        <TableCell>
          <Chip
            label={row.status}
            color={getStatusColor(row.status) as any}
            size="small"
            variant="outlined"
          />
        </TableCell>

        <TableCell>{formatDate(row.last_used_at)}</TableCell>

        <TableCell>{formatDate(row.created_at)}</TableCell>

        <TableCell align="right">
          <IconButton onClick={handleOpenPopover}>
            <Iconify icon="eva:more-vertical-fill" />
          </IconButton>
        </TableCell>
      </TableRow>

      <Popover
        open={!!openPopover}
        anchorEl={openPopover}
        onClose={handleClosePopover}
        anchorOrigin={{ vertical: 'top', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'right' }}
      >
        <MenuList
          disablePadding
          sx={{
            p: 0.5,
            gap: 0.5,
            width: 160,
            display: 'flex',
            flexDirection: 'column',
            [`& .${menuItemClasses.root}`]: {
              px: 1,
              gap: 2,
              borderRadius: 0.75,
              [`&.${menuItemClasses.selected}`]: { bgcolor: 'action.selected' },
            },
          }}
        >
          <MenuItem
            onClick={() => {
              handleClosePopover();
              onViewDetails();
            }}
          >
            <Iconify icon="solar:eye-bold" />
            {t('api_keys.view_details')}
          </MenuItem>

          <MenuItem
            onClick={() => {
              handleClosePopover();
              const newStatus = row.status === 'active' ? 'inactive' : 'active';
              onStatusChange(newStatus);
            }}
            sx={{ color: row.status === 'active' ? 'warning.main' : 'success.main' }}
          >
            <Iconify icon={row.status === 'active' ? 'solar:pause-bold' : 'solar:play-bold'} />
            {row.status === 'active' ? 'Disable' : 'Enable'}
          </MenuItem>

          <MenuItem
            onClick={() => {
              handleClosePopover();
              onDeleteRow();
            }}
            sx={{ color: 'error.main' }}
          >
            <Iconify icon="solar:trash-bin-trash-bold" />
            {t('common.delete')}
          </MenuItem>
        </MenuList>
      </Popover>
    </>
  );
}
