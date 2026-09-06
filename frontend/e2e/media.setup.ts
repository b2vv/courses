import { execFile } from 'node:child_process';
import { mkdir, stat, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { expect, test } from '@playwright/test';
import { promisify } from 'node:util';

const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const shortSizes = ['640x360', '1280x720', '320x240'] as const;

async function compose(...args: string[]) {
  return execute('docker', ['compose', ...args], {
    cwd: repositoryRoot,
    timeout: 120_000,
    maxBuffer: 1024 * 1024,
  });
}

test('generate the four media fixtures with the worker ffmpeg', async () => {
  test.setTimeout(120_000);
  await mkdir(fixtureDirectory, { recursive: true });
  await compose(
    'exec',
    '-T',
    'media-worker',
    'sh',
    '-eu',
    '-c',
    [
      'for size in 640x360 1280x720 320x240; do',
      '  ffmpeg -nostdin -loglevel error -f lavfi -i testsrc2=size=$size:rate=15 -f lavfi -i sine=frequency=440 -t 2 -c:v libx264 -preset veryfast -c:a aac -y /tmp/e2e-$size.mp4',
      'done',
    ].join('\n'),
  );
  await compose(
    'exec',
    '-T',
    'media-worker',
    'ffmpeg',
    '-nostdin',
    '-loglevel',
    'error',
    '-f',
    'lavfi',
    '-i',
    'testsrc2=size=1280x720:rate=15',
    '-f',
    'lavfi',
    '-i',
    'sine=frequency=440',
    '-t',
    '300',
    '-c:v',
    'libx264',
    '-preset',
    'veryfast',
    '-crf',
    '32',
    '-c:a',
    'aac',
    '-y',
    '/tmp/e2e-1280x720-5min.mp4',
  );

  // `docker compose cp` cannot read from the worker's `/tmp`, which is a tmpfs mount rather
  // than part of the container filesystem, so the bytes are streamed out through exec instead.
  for (const name of [...shortSizes, '1280x720-5min'] as const) {
    const destination = path.join(fixtureDirectory, `e2e-${name}.mp4`);
    const { stdout } = await execute(
      'docker',
      ['compose', 'exec', '-T', 'media-worker', 'cat', `/tmp/e2e-${name}.mp4`],
      { cwd: repositoryRoot, timeout: 120_000, maxBuffer: 128 * 1024 * 1024, encoding: 'buffer' },
    );
    await writeFile(destination, stdout);
    expect((await stat(destination)).size).toBeGreaterThan(0);
  }
});
