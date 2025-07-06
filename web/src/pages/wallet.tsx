import { CONFIG } from 'src/config-global';

import { WalletView } from 'src/sections/wallet/view';

// ----------------------------------------------------------------------

export default function WalletPage() {
  return (
    <>
      <title>{`Wallet - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="Manage your wallet balance and recharge your account"
      />
      <meta name="keywords" content="wallet,balance,recharge,payment,money" />

      <WalletView />
    </>
  );
}
