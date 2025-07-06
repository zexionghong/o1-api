import { useState } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Menu from '@mui/material/Menu';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { Iconify } from 'src/components/iconify';

// ----------------------------------------------------------------------

interface DateRangePickerProps {
  startDate: string;
  endDate: string;
  onStartDateChange: (date: string) => void;
  onEndDateChange: (date: string) => void;
  onClear: () => void;
}

// ----------------------------------------------------------------------

export function DateRangePicker({
  startDate,
  endDate,
  onStartDateChange,
  onEndDateChange,
  onClear,
}: DateRangePickerProps) {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleQuickSelect = (days: number) => {
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - days);

    onStartDateChange(start.toISOString().split('T')[0]);
    onEndDateChange(end.toISOString().split('T')[0]);
    handleClose();
  };

  const handleThisMonth = () => {
    const now = new Date();
    const start = new Date(now.getFullYear(), now.getMonth(), 1);
    const end = new Date(now.getFullYear(), now.getMonth() + 1, 0);

    onStartDateChange(start.toISOString().split('T')[0]);
    onEndDateChange(end.toISOString().split('T')[0]);
    handleClose();
  };

  const handleLastMonth = () => {
    const now = new Date();
    const start = new Date(now.getFullYear(), now.getMonth() - 1, 1);
    const end = new Date(now.getFullYear(), now.getMonth(), 0);

    onStartDateChange(start.toISOString().split('T')[0]);
    onEndDateChange(end.toISOString().split('T')[0]);
    handleClose();
  };

  const formatDateRange = () => {
    if (!startDate && !endDate) {
      return 'Select date range';
    }
    if (startDate && endDate) {
      return `${startDate} to ${endDate}`;
    }
    if (startDate) {
      return `From ${startDate}`;
    }
    if (endDate) {
      return `Until ${endDate}`;
    }
    return 'Select date range';
  };

  const hasDateRange = startDate || endDate;

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
      <Button
        variant="outlined"
        onClick={handleClick}
        startIcon={<Iconify icon="solar:calendar-bold" />}
        endIcon={<Iconify icon="eva:arrow-down-fill" />}
        sx={{
          color: hasDateRange ? 'primary.main' : 'text.secondary',
          borderColor: hasDateRange ? 'primary.main' : 'divider',
        }}
      >
        {formatDateRange()}
      </Button>

      {hasDateRange && (
        <Chip
          label="Clear"
          size="small"
          onDelete={onClear}
          deleteIcon={<Iconify icon="solar:close-circle-bold" />}
          sx={{ ml: 1 }}
        />
      )}

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
        PaperProps={{
          sx: { width: 320, p: 2 },
        }}
      >
        <Typography variant="subtitle2" sx={{ mb: 2 }}>
          Quick Select
        </Typography>
        
        <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 1, mb: 2 }}>
          <MenuItem onClick={() => handleQuickSelect(7)} sx={{ borderRadius: 1 }}>
            Last 7 days
          </MenuItem>
          <MenuItem onClick={() => handleQuickSelect(30)} sx={{ borderRadius: 1 }}>
            Last 30 days
          </MenuItem>
          <MenuItem onClick={() => handleQuickSelect(90)} sx={{ borderRadius: 1 }}>
            Last 90 days
          </MenuItem>
          <MenuItem onClick={() => handleQuickSelect(365)} sx={{ borderRadius: 1 }}>
            Last year
          </MenuItem>
          <MenuItem onClick={handleThisMonth} sx={{ borderRadius: 1 }}>
            This month
          </MenuItem>
          <MenuItem onClick={handleLastMonth} sx={{ borderRadius: 1 }}>
            Last month
          </MenuItem>
        </Box>

        <Typography variant="subtitle2" sx={{ mb: 2 }}>
          Custom Range
        </Typography>

        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <TextField
            fullWidth
            label="Start Date"
            type="date"
            value={startDate}
            onChange={(e) => onStartDateChange(e.target.value)}
            InputLabelProps={{ shrink: true }}
            size="small"
          />
          <TextField
            fullWidth
            label="End Date"
            type="date"
            value={endDate}
            onChange={(e) => onEndDateChange(e.target.value)}
            InputLabelProps={{ shrink: true }}
            size="small"
          />
        </Box>

        <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 2 }}>
          <Button size="small" onClick={onClear}>
            Clear
          </Button>
          <Button size="small" variant="contained" onClick={handleClose}>
            Apply
          </Button>
        </Box>
      </Menu>
    </Box>
  );
}
