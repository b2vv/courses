// Mark the process as Storybook BEFORE the Rsbuild builder auto-loads the app's
// rsbuild.config.ts, so that config drops the TanStack Start plugin (its server
// entry breaks the Storybook preview build). Tailwind + React come from the app
// config, which the builder loads by design.
process.env.STORYBOOK = 'true';

import type { StorybookConfig } from 'storybook-react-rsbuild';

/** Storybook on the Rsbuild builder (no Vite). Admin design system only, for now. */
const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(ts|tsx)'],
  addons: [],
  framework: { name: 'storybook-react-rsbuild', options: {} },
};

export default config;
