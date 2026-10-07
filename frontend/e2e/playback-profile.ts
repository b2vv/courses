import { execFileSync } from 'node:child_process';

// Не тест-файл: Playwright забороняє тестам імпортувати одне одного, тож спільна
// операція живе тут, як і `auth.ts`.
const repositoryRoot = new URL('../..', import.meta.url).pathname;

export function recreateApi(heartbeatSeconds: string, ownershipTtlSeconds: string): void {
  // --no-deps: this recreate is api-only by contract. Without it, --force-recreate's reach over
  // api's depends_on (redis, postgres, mailpit, and wayforpay-stub once COMPOSE_FILE resolves the
  // overlay) is compose-version cascade semantics — the one thing a playback fixture must not
  // leave to chance, since redis holds the ownership keys. --wait then waits for api alone; the
  // deps are already healthy from `make e2e`.
  execFileSync(
    'docker',
    ['compose', 'up', '-d', '--no-deps', '--force-recreate', '--wait', 'api'],
    {
      cwd: repositoryRoot,
      env: {
        ...process.env,
        PLAYBACK_HEARTBEAT_SECONDS: heartbeatSeconds,
        PLAYBACK_OWNERSHIP_TTL_SECONDS: ownershipTtlSeconds,
      },
      stdio: 'pipe',
    },
  );
}
