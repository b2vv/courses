import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginTailwindcss } from '@rsbuild/plugin-tailwindcss';
import { tanstackStart } from '@tanstack/react-start/plugin/rsbuild';

// Storybook loads this same config (by design) but has its own entry and must
// not run the TanStack Start plugin — Start emits a server entry that breaks the
// Storybook preview build. main.ts sets STORYBOOK=true before this is loaded.
const underStorybook = process.env.STORYBOOK === 'true';

export default defineConfig({
  plugins: [pluginReact(), pluginTailwindcss(), ...(underStorybook ? [] : [tanstackStart()])],
  source: {
    define: {
      // Empty in normal builds; the e2e Compose build passes localhost for its HTTP stub.
      'process.env.PUBLIC_WAYFORPAY_PAY_HOST': JSON.stringify(
        process.env.PUBLIC_WAYFORPAY_PAY_HOST ?? '',
      ),
    },
  },
  server: { port: 3000 },
});
