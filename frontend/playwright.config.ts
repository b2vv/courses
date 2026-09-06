import { defineConfig, devices } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  retries: 0,
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:8080',
    locale: 'uk-UA',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'setup',
      testMatch: ['**/auth.setup.ts', '**/media.setup.ts'],
      teardown: 'login-budget',
    },
    {
      name: 'chromium',
      testIgnore: ['**/recovery.spec.ts', '**/media.spec.ts'],
      dependencies: ['setup'],
      use: { ...devices['Desktop Chrome'] },
    },
    {
      // The media journey restarts the worker and saturates its two CPUs with a five-minute
      // transcode, which starves anything running beside it. Same reason recovery is isolated.
      name: 'media',
      testMatch: '**/media.spec.ts',
      dependencies: ['setup', 'chromium'],
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'recovery',
      testMatch: '**/recovery.spec.ts',
      dependencies: ['setup', 'chromium', 'media'],
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'login-budget',
      testMatch: '**/auth.teardown.ts',
    },
  ],
});
