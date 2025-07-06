import { CONFIG } from 'src/config-global';

import { ApiKeysView } from 'src/sections/api-keys/view';

// ----------------------------------------------------------------------

export default function Page() {
  return (
    <>
      <title>{`API Keys - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="Manage your API keys for accessing AI services"
      />
      <meta name="keywords" content="api,keys,management,ai,services" />

      <ApiKeysView />
    </>
  );
}
