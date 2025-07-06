import { CONFIG } from 'src/config-global';

import { ToolsView } from 'src/sections/tools/view';

// ----------------------------------------------------------------------

export default function ToolsPage() {
  return (
    <>
      <title>{`Tools - ${CONFIG.appName}`}</title>
      <meta
        name="description"
        content="AI tools collection including workflow automation, image generation, chatbots, and more"
      />
      <meta name="keywords" content="ai tools,automation,image generation,chatbot,video generation,n8n" />

      <ToolsView />
    </>
  );
}
