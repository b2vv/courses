import { defineConfig } from '@rstest/core';

const serverAlias = {
  '@tanstack/react-start/server$': '@tanstack/start-server-core/request-response',
};

const sharedTools = {
  rspack: {
    resolve: {
      // The public server barrel pulls virtual production entries that do not
      // exist in Rstest. Tests need only its request/response implementation.
      // ponytail: shared by both projects, so a test wanting anything the
      // barrel re-exports beyond request/response gets a module without it and
      // fails confusingly. Split it when a test needs another barrel export.
      alias: serverAlias,
    },
  },
  swc: { jsc: { transform: { react: { runtime: 'automatic' as const } } } },
};

export default defineConfig({
  projects: [
    {
      name: 'adapter-node',
      testEnvironment: 'node',
      include: ['src/adapters/api/**/*.test.{ts,tsx}'],
      setupFiles: ['./src/test-setup.ts'],
      env: { API_BASE_URL: 'http://api.test' },
      tools: {
        ...sharedTools,
        rspack: {
          resolve: {
            alias: {
              ...serverAlias,
              '@opentelemetry/api$': '@opentelemetry/api/build/src/index.js',
            },
          },
        },
      },
    },
    {
      name: 'dom',
      testEnvironment: 'happy-dom',
      include: ['src/**/*.test.{ts,tsx}'],
      exclude: ['src/adapters/api/**'],
      setupFiles: ['./src/test-setup.ts'],
      tools: sharedTools,
    },
  ],
});
