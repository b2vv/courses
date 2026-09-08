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
      testIgnore: [
        '**/recovery.spec.ts',
        '**/media.spec.ts',
        '**/playback.spec.ts',
        '**/playback.setup.ts',
        '**/playback.teardown.ts',
      ],
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
      // Playback owns another long transcode plus 30-second/expiry measurements. It reuses the
      // media setup bytes but runs alone so its network counts cannot race the Phase 5 journey.
      name: 'playback',
      testMatch: '**/playback.spec.ts',
      // P08 must observe several ownership renewals inside one view, which the shipped
      // 30/180 cadence cannot show without two minutes of playback. The profile project
      // recreates api at 2/6 and the teardown puts the documented defaults back, so
      // docker-compose.yml keeps 30/180 for development and for every other project.
      dependencies: ['media', 'playback-profile'],
      teardown: 'playback-profile-teardown',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'playback-profile',
      testMatch: '**/playback.setup.ts',
    },
    {
      name: 'playback-profile-teardown',
      testMatch: '**/playback.teardown.ts',
    },
    {
      name: 'recovery',
      testMatch: '**/recovery.spec.ts',
      dependencies: ['playback'],
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'login-budget',
      testMatch: '**/auth.teardown.ts',
    },
  ],
});
