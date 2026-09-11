import { execFileSync } from 'node:child_process';
import { test as teardown } from '@playwright/test';
import { recreateApi } from './playback-profile';

const repositoryRoot = new URL('../..', import.meta.url).pathname;

// The block that made the mess clears it. Playback's uploads leave transcode jobs in Redis, and
// the queue outlives the stack: `recovery` restarts everything a few minutes later, the worker
// recovers those jobs immediately, ffmpeg takes its whole two-CPU budget, and `/readyz` stops
// answering inside the healthcheck's three seconds. Compose then calls the worker unhealthy and
// the restart fails - a red with nothing wrong in the product.
//
// This is the half of B-5 that the setup drain cannot reach: that one runs before the fixtures,
// and these jobs are created afterwards.
function drainTranscodeQueue(): void {
  execFileSync(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'redis',
      'sh',
      '-c',
      'export REDISCLI_AUTH="$REDIS_PASSWORD"; ' +
        'redis-cli --scan --pattern "asynq:{media}:*" | xargs -r -n 100 redis-cli DEL > /dev/null',
    ],
    { cwd: repositoryRoot, stdio: 'pipe' },
  );
}

// Повертає api до задокументованих дефолтів, щоб наступні проєкти й розробка не
// успадкували прискорений профіль P08.
teardown('restore the documented playback defaults', () => {
  teardown.setTimeout(180_000);
  drainTranscodeQueue();
  recreateApi('30', '180');
});
