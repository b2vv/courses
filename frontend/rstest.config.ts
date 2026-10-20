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
  // B-4: the 5s default timed out CourseEditorView tests whose measured cause was external host
  // load (load average ~47 from unrelated processes), not the tests — one test reached 162
  // assertions before the deadline. A single run at 30s gave 12/12 and 364/364. The ceiling is
  // generous on purpose: it only bites a genuinely hung test, never a slow-but-working one.
  testTimeout: 30000,
  projects: [
    {
      name: 'adapter-node',
      testEnvironment: 'node',
      // Adapter tests run under node for real fetch/Headers wire semantics. Feature modules keep
      // their adapters under src/modules/<m>/adapters, so match those too — not only the flat
      // src/adapters/api left from the pre-module layout.
      include: ['src/adapters/api/**/*.test.{ts,tsx}', 'src/modules/*/adapters/**/*.test.{ts,tsx}'],
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
      exclude: ['src/adapters/api/**', 'src/modules/*/adapters/**'],
      setupFiles: ['./src/test-setup.ts'],
      tools: sharedTools,
    },
  ],
});
