import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Typography from '@mui/material/Typography';
import CardContent from '@mui/material/CardContent';
import { alpha, useTheme } from '@mui/material/styles';

// ----------------------------------------------------------------------

type Props = {
  title: string;
  total: number;
  percent?: number;
  color?: 'primary' | 'secondary' | 'info' | 'success' | 'warning' | 'error';
  icon?: React.ReactNode;
  chart?: {
    categories: string[];
    series: number[];
  };
};

export function AnalyticsWidgetSummary({
  title,
  total,
  percent = 0,
  color = 'primary',
  icon,
  chart,
}: Props) {
  const theme = useTheme();

  return (
    <Card
      sx={{
        py: 3,
        px: 3,
        borderRadius: 2,
        bgcolor: alpha(theme.palette[color].main, 0.08),
        border: `1px solid ${alpha(theme.palette[color].main, 0.24)}`,
      }}
    >
      <CardContent sx={{ p: 0 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          {icon && (
            <Box
              sx={{
                width: 48,
                height: 48,
                borderRadius: 1.5,
                bgcolor: alpha(theme.palette[color].main, 0.16),
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                mr: 2,
              }}
            >
              {icon}
            </Box>
          )}
          <Box sx={{ flexGrow: 1 }}>
            <Typography variant="h3" sx={{ color: theme.palette[color].main }}>
              {typeof total === 'number' ? total.toLocaleString() : total}
            </Typography>
            <Typography variant="subtitle2" sx={{ color: 'text.secondary' }}>
              {title}
            </Typography>
          </Box>
        </Box>

        {percent !== 0 && (
          <Typography
            variant="body2"
            sx={{
              color: percent > 0 ? 'success.main' : 'error.main',
            }}
          >
            {percent > 0 ? '+' : ''}
            {percent}%
          </Typography>
        )}
      </CardContent>
    </Card>
  );
}
