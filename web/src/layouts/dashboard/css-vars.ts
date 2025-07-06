import type { Theme } from '@mui/material/styles';

// ----------------------------------------------------------------------

export function dashboardLayoutVars(theme: Theme, navOpen = true) {
  return {
    '--layout-transition-easing': 'linear',
    '--layout-transition-duration': '120ms',
    '--layout-nav-vertical-width': navOpen ? '300px' : '80px',
    '--layout-nav-vertical-width-collapsed': '80px',
    '--layout-dashboard-content-pt': theme.spacing(1),
    '--layout-dashboard-content-pb': theme.spacing(8),
    '--layout-dashboard-content-px': theme.spacing(5),
  };
}
