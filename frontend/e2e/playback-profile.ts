import { execFileSync } from 'node:child_process';

// Не тест-файл: Playwright забороняє тестам імпортувати одне одного, тож спільна
// операція живе тут, як і `auth.ts`.
const repositoryRoot = new URL('../..', import.meta.url).pathname;

export function recreateApi(heartbeatSeconds: string, ownershipTtlSeconds: string): void {
  execFileSync('docker', ['compose', 'up', '-d', '--force-recreate', '--wait', 'api'], {
    cwd: repositoryRoot,
    env: {
      ...process.env,
      PLAYBACK_HEARTBEAT_SECONDS: heartbeatSeconds,
      PLAYBACK_OWNERSHIP_TTL_SECONDS: ownershipTtlSeconds,
    },
    stdio: 'pipe',
  });
}
