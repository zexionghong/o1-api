import { CONFIG } from 'src/config-global';

import { ModelsView } from 'src/sections/models/view';

// ----------------------------------------------------------------------

export default function ModelsPage() {
  return (
    <>
      <title>{`Models - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="Browse and explore available AI models for text generation, image creation, and more"
      />
      <meta name="keywords" content="ai models,gpt,claude,dall-e,stable diffusion,text generation,image generation" />

      <ModelsView />
    </>
  );
}
