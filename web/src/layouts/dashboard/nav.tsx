import type { Theme, SxProps, Breakpoint } from '@mui/material/styles';

import { useEffect } from 'react';
import { varAlpha } from 'minimal-shared/utils';

import Box from '@mui/material/Box';
import ListItem from '@mui/material/ListItem';
import { useTheme } from '@mui/material/styles';
import IconButton from '@mui/material/IconButton';
import ListItemButton from '@mui/material/ListItemButton';
import Drawer, { drawerClasses } from '@mui/material/Drawer';

import { usePathname } from 'src/routes/hooks';
import { RouterLink } from 'src/routes/components';

import { Logo } from 'src/components/logo';
import { Iconify } from 'src/components/iconify';
import { Scrollbar } from 'src/components/scrollbar';



import type { NavItem } from '../nav-config-dashboard';

// ----------------------------------------------------------------------

export type NavContentProps = {
  data: NavItem[];
  slots?: {
    topArea?: React.ReactNode;
    bottomArea?: React.ReactNode;
  };
  sx?: SxProps<Theme>;
  collapsed?: boolean;
  onToggle?: () => void;
};

export function NavDesktop({
  sx,
  data,
  slots,
  layoutQuery,
  open = true,
  onToggle,
}: NavContentProps & {
  layoutQuery: Breakpoint;
  open?: boolean;
  onToggle?: () => void;
}) {
  const theme = useTheme();

  return (
    <Box
      sx={{
        pt: 2.5,
        px: 2.5,
        top: 0,
        left: 0,
        height: 1,
        display: 'none',
        position: 'fixed',
        flexDirection: 'column',
        zIndex: 'var(--layout-nav-zIndex)',
        width: open ? 'var(--layout-nav-vertical-width)' : '80px',
        borderRight: `1px solid ${varAlpha(theme.vars.palette.grey['500Channel'], 0.12)}`,
        transition: theme.transitions.create(['width'], {
          duration: theme.transitions.duration.shorter,
        }),
        [theme.breakpoints.up(layoutQuery)]: {
          display: 'flex',
        },
        ...sx,
      }}
    >
      <NavContent data={data} slots={slots} collapsed={!open} onToggle={onToggle} />
    </Box>
  );
}

// ----------------------------------------------------------------------

export function NavMobile({
  sx,
  data,
  open,
  slots,
  onClose,
  collapsed = false,
  onToggle,
}: NavContentProps & { open: boolean; onClose: () => void; collapsed?: boolean; onToggle?: () => void }) {
  const pathname = usePathname();

  useEffect(() => {
    if (open) {
      onClose();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname]);

  return (
    <Drawer
      open={open}
      onClose={onClose}
      sx={{
        [`& .${drawerClasses.paper}`]: {
          pt: 2.5,
          px: 2.5,
          overflow: 'unset',
          width: collapsed ? '80px' : 'var(--layout-nav-mobile-width)',
          transition: (theme) => theme.transitions.create(['width'], {
            duration: theme.transitions.duration.shorter,
          }),
          ...sx,
        },
      }}
    >
      <NavContent data={data} slots={slots} collapsed={collapsed} onToggle={onToggle} />
    </Drawer>
  );
}

// ----------------------------------------------------------------------

export function NavContent({ data, slots, sx, collapsed = false, onToggle }: NavContentProps) {
  const pathname = usePathname();

  return (
    <>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: collapsed ? 'center' : 'space-between', mb: 2 }}>
        {!collapsed && <Logo />}
        {onToggle && (
          <IconButton
            onClick={onToggle}
            sx={{
              width: 36,
              height: 36,
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 1,
              '&:hover': {
                borderColor: 'primary.main',
                bgcolor: 'action.hover',
              },
              ...(collapsed && { mx: 'auto' })
            }}
          >
            <Iconify
              icon={collapsed ? "eva:arrow-ios-forward-outline" : "eva:arrow-ios-back-outline"}
              width={20}
            />
          </IconButton>
        )}
      </Box>

      <Scrollbar fillContent>
        <Box
          component="nav"
          sx={[
            {
              display: 'flex',
              flex: '1 1 auto',
              flexDirection: 'column',
            },
            ...(Array.isArray(sx) ? sx : [sx]),
          ]}
        >
          <Box
            component="ul"
            sx={{
              gap: 0.5,
              display: 'flex',
              flexDirection: 'column',
            }}
          >
            {data.map((item) => {
              const isActived = item.path === pathname;

              return (
                <ListItem disableGutters disablePadding key={item.title}>
                  <ListItemButton
                    disableGutters
                    component={RouterLink}
                    href={item.path}
                    sx={[
                      (theme) => ({
                        pl: collapsed ? 1.5 : 2,
                        py: 1,
                        gap: collapsed ? 0 : 2,
                        pr: 1.5,
                        borderRadius: 0.75,
                        typography: 'body2',
                        fontWeight: 'fontWeightMedium',
                        color: theme.vars.palette.text.secondary,
                        minHeight: 44,
                        justifyContent: collapsed ? 'center' : 'flex-start',
                        ...(isActived && {
                          fontWeight: 'fontWeightSemiBold',
                          color: theme.vars.palette.primary.main,
                          bgcolor: varAlpha(theme.vars.palette.primary.mainChannel, 0.08),
                          '&:hover': {
                            bgcolor: varAlpha(theme.vars.palette.primary.mainChannel, 0.16),
                          },
                        }),
                      }),
                    ]}
                    title={collapsed ? item.title : undefined}
                  >
                    <Box component="span" sx={{ width: 24, height: 24 }}>
                      {item.icon}
                    </Box>

                    {!collapsed && (
                      <Box component="span" sx={{ flexGrow: 1 }}>
                        {item.title}
                      </Box>
                    )}

                    {!collapsed && item.info && item.info}
                  </ListItemButton>
                </ListItem>
              );
            })}
          </Box>
        </Box>
      </Scrollbar>
    </>
  );
}
