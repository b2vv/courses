import { defineConfig, devices } from '@playwright/test';
export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  retries: 0,
  // Pinned, not tuned. Left unset, Playwright takes half the machine's cores, which makes the
  // suite's parallelism a property of the hardware: the same commit passes on one host and
  // fails on another, and each failure reads as a flake in the newest change. Measured here on
  // 2026-09-08 - five workers under a host at load 22 produced five failures across four spec
  // files, while the server answered /courses in 0.108s and the API in 0.010s, so the
  // contention was in the browsers, not the application.
  //
  // The cost is small and one-sided: only `chromium` holds several files, and the single-file
  // projects (media, playback, recovery) are unaffected by this number at all.
  //
  // The trade is named so it can be revisited deliberately: a large CI runner will leave
  // capacity unused. That is the price of a suite that behaves the same everywhere.
  // One worker, not two, and the reason is isolation rather than speed. `fullyParallel: false`
  // orders tests inside a file; separate files still run side by side, and the public catalogue
  // is global state that several of them share. `admin.spec` publishes a course mid-scenario
  // with published_at = now, which sorts ahead of every seeded course, so a catalogue test
  // reading offset=1 concurrently gets the wrong row - measured 2026-09-09, twice, and the
  // interleaved completion order shows the overlap. Serialising the project costs a couple of
  // minutes only in `chromium`; media, playback and recovery hold one file each and are
  // unaffected by this number at all.
  //
  // The trade stays named: a large CI runner will leave capacity unused. That is the price of a
  // suite whose result depends on the code rather than on which tests happened to overlap.
  workers: 1,
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
      dependencies: ['playback-profile'],
      teardown: 'playback-profile-teardown',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      // Depends on `media` so the whole chain is serial: setup, chromium, media, then this.
      // Without an edge here the project ran beside `setup`, and recreating api mid-login gave
      // nginx no upstream: the student captured its state in 3s, the instructor got 502 in 68ms.
      // The dependency is what orders the restart, not the absence of `fullyParallel`.
      name: 'playback-profile',
      testMatch: '**/playback.setup.ts',
      dependencies: ['media'],
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
