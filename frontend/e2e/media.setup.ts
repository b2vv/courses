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
    timeout: 300_000,
    maxBuffer: 1024 * 1024,
  });
}

// Redis працює з `appendonly yes` на постійному томі `redis-data`, а чергу asynq між
// прогонами не чистить ніхто. Прогін, обірваний посеред транскодування, лишає задачу; на
// наступному `make up` воркер її відновлює й зʼїдає свій бюджет 2 CPU / 2 GB, поки транскод
// цього прогону стоїть за нею в черзі — і фікстура не встигає у власний таймаут. Червоне
// виходить із причини, не повʼязаної з кодом.
//
// ПЕРЕДУМОВА, за якої дренаж безпечний: важкі гейти цього репо запускає рівно одна сесія за
// раз (домовленість між сесіями, 2026-09-08). Конкурентного прогону, чию легітимну задачу
// можна знести, немає. **Якщо це зміниться — цей дренаж прибрати першим.**
//
// `export` тут несучий, а не стиль. Префікс `VAR=cmd` ставить змінну лише для першої команди,
// тож другий `redis-cli` у пайпі її не успадкував би — і пароль довелося б передати через `-a`,
// тобто в argv, де його видно з `ps` і `docker top`. Саме це забороняє `docker-compose.yml:29-30`
// для самого сервера. Через оточення його підхоплюють обидва виклики, і глушник
// `--no-auth-warning` стає непотрібним: попереджати нема про що.
//
// Патерн б'є рівно по чергах воркера (`Queues: {"media": 1}`, `cmd/media-worker/main.go:70`).
// Ширший `asynq:*` зніс би ще й `asynq:servers:*` та `asynq:workers` — реєстрацію для
// моніторингу, а не задачі.
async function drainTranscodeQueue() {
  await compose(
    'exec',
    '-T',
    'redis',
    'sh',
    '-c',
    'export REDISCLI_AUTH="$REDIS_PASSWORD"; ' +
      'redis-cli --scan --pattern "asynq:{media}:*" | xargs -r -n 100 redis-cli DEL > /dev/null',
  );
}

test('generate the nine media fixtures with the worker ffmpeg', async () => {
  // Encoding the five-minute seek fixture alone runs past two minutes on a busy machine,
  // so the budget covers all five encodes plus streaming the bytes out.
  test.setTimeout(720_000);
  await drainTranscodeQueue();
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
      // Multi-audio S5-L2b: a video-only source (no audio stream → awaiting_audio) and a short
      // AAC tone in an M4A container to upload as its audio track.
      'ffmpeg -nostdin -loglevel error -f lavfi -i testsrc2=size=640x360:rate=15 -t 2 -an -c:v libx264 -preset veryfast -y /tmp/e2e-640x360-silent.mp4',
      'ffmpeg -nostdin -loglevel error -f lavfi -i sine=frequency=660 -t 2 -c:a aac -y /tmp/e2e-tone.m4a',
      // Multi-audio S6: an uploaded track with speech, so recognition of the default track is observable.
      'ffmpeg -nostdin -loglevel error -f lavfi -i "flite=text=\'This is the dubbed audio track of the lesson about multiple languages.\':voice=slt" -c:a aac -y /tmp/e2e-speech-dub.m4a',
      'ffmpeg -nostdin -loglevel error -f lavfi -i testsrc2=size=640x360:rate=15 -f lavfi -i "flite=text=\'Welcome to the course. Today we will learn how video transcripts work.\':voice=slt" -shortest -c:v libx264 -preset veryfast -c:a aac -y /tmp/e2e-speech-640x360.mp4',
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
    '100',
    '-c:v',
    'libx264',
    '-preset',
    'veryfast',
    '-crf',
    '32',
    '-c:a',
    'aac',
    '-y',
    '/tmp/e2e-1280x720-100s.mp4',
  );

  // `docker compose cp` cannot read from the worker's `/tmp`, which is a tmpfs mount rather
  // than part of the container filesystem, so the bytes are streamed out through exec instead.
  const names = [
    ...shortSizes.map((size) => `${size}.mp4`),
    'speech-640x360.mp4',
    '1280x720-5min.mp4',
    '1280x720-100s.mp4',
    '640x360-silent.mp4',
    'tone.m4a',
    'speech-dub.m4a',
  ];
  for (const name of names) {
    const destination = path.join(fixtureDirectory, `e2e-${name}`);
    const { stdout } = await execute(
      'docker',
      ['compose', 'exec', '-T', 'media-worker', 'cat', `/tmp/e2e-${name}`],
      { cwd: repositoryRoot, timeout: 120_000, maxBuffer: 128 * 1024 * 1024, encoding: 'buffer' },
    );
    await writeFile(destination, stdout);
    expect((await stat(destination)).size).toBeGreaterThan(0);
  }
});
