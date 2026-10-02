import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginTailwindcss } from '@rsbuild/plugin-tailwindcss';
import { tanstackStart } from '@tanstack/react-start/plugin/rsbuild';

export default defineConfig({
  plugins: [pluginReact(), pluginTailwindcss(), tanstackStart()],
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
