import { CONFIG } from 'src/config-global';

import { ProfileView } from 'src/sections/profile/view';

// ----------------------------------------------------------------------

export default function ProfilePage() {
  return (
    <>
      <title>{`Profile - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="Manage your profile settings, change password and avatar"
      />
      <meta name="keywords" content="profile,settings,password,avatar,account" />

      <ProfileView />
    </>
  );
}
