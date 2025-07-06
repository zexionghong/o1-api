import { SvgColor } from 'src/components/svg-color';

// ----------------------------------------------------------------------

const icon = (name: string) => <SvgColor src={`/assets/icons/navbar/${name}.svg`} />;

export type NavItem = {
  title: string;
  path: string;
  icon: React.ReactNode;
  info?: React.ReactNode;
};

export const navData = [
  {
    title: 'Real Dashboard',
    path: '/',
    icon: icon('ic-analytics'),
  },
  {
    title: 'API Keys',
    path: '/api-keys',
    icon: icon('ic-lock'),
  },
];
