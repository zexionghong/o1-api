import { CONFIG } from 'src/config-global';

import { RealDashboardView as DashboardView } from 'src/sections/overview/view';

// ----------------------------------------------------------------------

export default function Page() {
  return (
    <>
      <title>{`Real Dashboard - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="Real-time dashboard with data from backend API"
      />
      <meta name="keywords" content="react,material,kit,application,dashboard,admin,template,real-time,api" />

      <DashboardView />
    </>
  );
}
