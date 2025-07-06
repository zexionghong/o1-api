import { useTranslation } from 'react-i18next';

import { SvgColor } from 'src/components/svg-color';

// ----------------------------------------------------------------------

const icon = (name: string) => <SvgColor src={`/assets/icons/navbar/${name}.svg`} />;

export type NavItem = {
  title: string;
  path: string;
  icon: React.ReactNode;
  info?: React.ReactNode;
};

export function useNavData() {
  const { t } = useTranslation();

  return [
    {
      title: t('navigation.dashboard'),
      path: '/',
      icon: icon('ic-analytics'),
    },
    {
      title: t('navigation.api_keys'),
      path: '/api-keys',
      icon: icon('ic-lock'),
    },
    {
      title: t('navigation.profile'),
      path: '/profile',
      icon: icon('ic-user'),
    },
  ];
}

// 保持向后兼容
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
  {
    title: 'Profile',
    path: '/profile',
    icon: icon('ic-user'),
  },
];
