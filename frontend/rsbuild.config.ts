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
      // Release identifier baked into the client bundle at build time (git short
      // SHA passed as APP_RELEASE; "dev" in dev/gate builds). The SSR process
      // (server.mjs) reads the SAME process.env.APP_RELEASE at runtime, so client
      // and SSR log lines carry one matching release. Using the process.env key
      // keeps it typed by @types/node — no ambient declaration needed.
      'process.env.APP_RELEASE': JSON.stringify(process.env.APP_RELEASE ?? 'dev'),
    },
  },
  server: { port: 3000 },
});
