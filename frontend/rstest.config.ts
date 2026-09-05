import { defineConfig } from '@rstest/core';

export default defineConfig({
  testEnvironment: 'happy-dom',
  tools: {
    rspack: {
      resolve: {
        // The public server barrel pulls virtual production entries that do not
        // exist in Rstest. Tests need only its request/response implementation.
        // ponytail: global alias for every test, so a test wanting anything the
        // barrel re-exports beyond request/response gets a module without it and
        // fails confusingly. Scope it per test project when a second test needs
        // another export from this barrel.
        alias: {
          '@tanstack/react-start/server$': '@tanstack/start-server-core/request-response',
        },
      },
    },
    swc: { jsc: { transform: { react: { runtime: 'automatic' } } } },
  },
  include: ['src/**/*.test.{ts,tsx}'],
  setupFiles: ['./src/test-setup.ts'],
});
