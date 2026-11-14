import { execFile, spawn } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  expect,
  test,
  type Browser,
  type BrowserContext,
  type Page,
  type Request,
  type Response,
} from '@playwright/test';
import { promisify } from 'node:util';
import type {
  AdminCourseTree,
  ErrorResponse,
  ErrorResponseError,
  LearningCourse as LearningCourseContract,
  LearningLesson,
  LearningModule,
  MediaUploadTicket,
  PlaybackHeartbeatRequest,
  PlaybackSession as PlaybackSessionContract,
} from '../src/adapters/api/generated/models';
import { authStatePath, trackPageRequests } from './auth';
import { contractField } from './contract';

const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const namespace = `e2e-playback-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const crossCourseNamespace = `${namespace}-cross-course`;
const adminID = '40000000-0000-4000-8000-000000000003';
// The seeded student fixture (student@courseforge.local); after the enrollment phase, playing a
// lesson needs membership, so the milestone fixtures admin-grant this student on their courses.
const studentID = '40000000-0000-4000-8000-000000000001';
const instructorID = '40000000-0000-4000-8000-000000000002';
const moduleID = randomUUID();
const lessonID = randomUUID();
const secondLessonID = randomUUID();
const secondModuleID = randomUUID();
const thirdLessonID = randomUUID();
const fourthLessonID = randomUUID();
const crossCourseModuleID = randomUUID();
const crossCourseLessonID = randomUUID();

const primaryCourseDraft = {
  slug: namespace,
  title: `T7 playback ${namespace}`,
  summary: 'A real HLS playback fixture.',
  description: 'Run-scoped course for the T7 browser recovery table.',
  instructorId: adminID,
};

const primaryCourseModules = [
  {
    id: moduleID,
    title: `Module one ${namespace}`,
    lessons: [
      {
        id: lessonID,
        title: `First lesson ${namespace}`,
      },
      {
        id: secondLessonID,
        title: `Second lesson ${namespace}`,
      },
    ],
  },
  {
    id: secondModuleID,
    title: `Module two ${namespace}`,
    lessons: [
      {
        id: thirdLessonID,
        title: `Third lesson ${namespace}`,
      },
      {
        id: fourthLessonID,
        title: `Fourth lesson ${namespace}`,
      },
    ],
  },
];

type CourseToken = { readonly id: string; readonly updatedAt: string };
type UploadTarget = {
  readonly lessonID: string;
  readonly moduleNumber: number;
  readonly lessonNumber: number;
};
type PlaybackSession = {
  readonly playbackId: string;
  readonly mediaPath: string;
  readonly clientSeq: number;
  readonly heartbeatSeconds: number;
  readonly ownershipTtlSeconds: number;
};
type HeartbeatPayload = {
  readonly positionSeconds: number;
  readonly clientSeq: number;
  readonly reason: 'tick' | 'pause' | 'seek' | 'hide' | 'nav' | 'ended';
};
type AuthorizeCounts = {
  readonly observed: boolean;
  readonly successful: number;
  readonly rejected: number;
  readonly total: number;
};
type ContainerNetworkSnapshot = {
  readonly id: string;
  readonly txBytes: bigint;
};

function isHeartbeatReason(value: unknown): value is HeartbeatPayload['reason'] {
  return (
    value === 'tick' ||
    value === 'pause' ||
    value === 'seek' ||
    value === 'hide' ||
    value === 'nav' ||
    value === 'ended'
  );
}

let course: CourseToken | undefined;
let crossCourse: CourseToken | undefined;
let fourthLessonVideoID: string | undefined;
let crossCourseVideoID: string | undefined;
let finishRequestTracking = () => Promise.resolve();

function object(payload: unknown, name: string): object {
  if (typeof payload !== 'object' || payload === null) throw new Error(`${name} must be an object`);
  return payload;
}

function courseToken(payload: unknown): CourseToken {
  const record = object(payload, 'Admin course response');
  const id = contractField<AdminCourseTree>(record, 'id');
  const updatedAt = contractField<AdminCourseTree>(record, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('Admin course response must carry id and updatedAt');
  return { id, updatedAt };
}

function playbackSession(payload: unknown): PlaybackSession {
  const record = object(payload, 'Playback session');
  const playbackId = contractField<PlaybackSessionContract>(record, 'playbackId');
  const mediaPath = contractField<PlaybackSessionContract>(record, 'mediaPath');
  const clientSeq = contractField<PlaybackSessionContract>(record, 'clientSeq');
  const heartbeatSeconds = contractField<PlaybackSessionContract>(record, 'heartbeatSeconds');
  const ownershipTtlSeconds = contractField<PlaybackSessionContract>(record, 'ownershipTtlSeconds');
  if (
    typeof playbackId !== 'string' ||
    typeof mediaPath !== 'string' ||
    typeof clientSeq !== 'number' ||
    typeof heartbeatSeconds !== 'number' ||
    typeof ownershipTtlSeconds !== 'number'
  )
    throw new Error('Playback session response is incomplete');
  return {
    playbackId,
    mediaPath,
    clientSeq,
    heartbeatSeconds,
    ownershipTtlSeconds,
  };
}

function videoID(payload: unknown): string {
  const id = contractField<MediaUploadTicket>(object(payload, 'Media ticket'), 'videoId');
  if (typeof id !== 'string') throw new Error('Media ticket response must carry videoId');
  return id;
}

async function adminContext(browser: Browser, baseURL: string): Promise<BrowserContext> {
  return browser.newContext({ baseURL, storageState: authStatePath('admin') });
}

function heartbeatPayload(response: Response, expectedLessonID: string): HeartbeatPayload | null {
  if (
    response.status() !== 204 ||
    new URL(response.url()).pathname !== '/api/v1/playback/heartbeat'
  )
    return null;
  const body: unknown = response.request().postDataJSON();
  if (typeof body !== 'object' || body === null) return null;
  const position = contractField<PlaybackHeartbeatRequest>(body, 'positionSeconds');
  const clientSeq = contractField<PlaybackHeartbeatRequest>(body, 'clientSeq');
  const reason = contractField<PlaybackHeartbeatRequest>(body, 'reason');
  if (
    contractField<PlaybackHeartbeatRequest>(body, 'lessonId') !== expectedLessonID ||
    typeof position !== 'number' ||
    typeof clientSeq !== 'number' ||
    !isHeartbeatReason(reason)
  )
    return null;
  return {
    positionSeconds: position,
    clientSeq,
    reason,
  };
}

function describeResponseError(
  payload: unknown,
): { readonly code: string; readonly requestId?: string } | null {
  if (typeof payload !== 'object' || payload === null) return null;
  const error = contractField<ErrorResponse>(payload, 'error');
  if (typeof error !== 'object' || error === null) return null;
  const code = contractField<ErrorResponseError>(error, 'code');
  const requestId = contractField<ErrorResponseError>(error, 'requestId');
  if (typeof code !== 'string') return null;
  return typeof requestId === 'string' ? { code, requestId } : { code };
}

// K4 (R23 heartbeat 403, O05 — mechanism never established): a refused heartbeat used to be
// invisible. `heartbeatPayload` ignores every non-204 answer, so the `ended` wait merely timed
// out and the code behind the 403 was lost with it. This rejects the moment a heartbeat is
// refused, naming the body's error code — `playback_forbidden`/`course_purchase_required`
// (the access rule) versus `origin_forbidden` (same-origin) — and its requestId, which joins
// the refusal to the api's own "http request" line in the log the gate now streams for the whole
// run (frontend/test-logs/api-nginx-<run stamp>.log locally, api-nginx.log in CI). The api mints
// that id itself (httpserver/server.go) and
// does not adopt nginx's X-Request-ID, so nginx's `safe` access lines carry a different id.
function heartbeatDenialFailure(page: Page): Promise<never> {
  return page
    .waitForResponse(
      (response) =>
        new URL(response.url()).pathname === '/api/v1/playback/heartbeat' &&
        response.status() >= 400,
      { timeout: 0 },
    )
    .then(
      async (response) => {
        const payload: unknown = await response.json().catch(() => null);
        const error = describeResponseError(payload);
        throw new Error(
          `heartbeat refused: HTTP ${response.status()} ${error?.code ?? '<no error code in body>'}` +
            (error?.requestId === undefined ? '' : ` requestId=${error.requestId}`),
        );
      },
      // The page closing after a passing test rejects the still-pending wait; that is not a
      // refusal, and nothing awaits this branch any more, so it simply never settles.
      () => new Promise<never>(() => {}),
    );
}

function heartbeatPosition(response: Response, expectedLessonID: string): number | null {
  return heartbeatPayload(response, expectedLessonID)?.positionSeconds ?? null;
}

function learningLessonPosition(payload: unknown, expectedLessonID: string): number {
  const modules = contractField<LearningCourseContract>(
    object(payload, 'Learning course'),
    'modules',
  );
  if (!Array.isArray(modules)) throw new Error('Learning course modules must be an array');
  for (const module of modules) {
    if (typeof module !== 'object' || module === null) continue;
    const lessons = contractField<LearningModule>(module, 'lessons');
    if (!Array.isArray(lessons)) continue;
    for (const lesson of lessons) {
      if (typeof lesson !== 'object' || lesson === null) continue;
      if (contractField<LearningLesson>(lesson, 'id') !== expectedLessonID) continue;
      const position = contractField<LearningLesson>(lesson, 'position');
      if (typeof position !== 'number')
        throw new Error('Learning lesson must carry numeric position');
      return position;
    }
  }
  throw new Error(`Learning course does not contain lesson ${expectedLessonID}`);
}

async function uploadReady(
  page: Page,
  file: string,
  target: UploadTarget & { readonly courseID?: string } = {
    lessonID,
    moduleNumber: 1,
    lessonNumber: 1,
  },
): Promise<string> {
  const targetCourseID = target.courseID ?? course?.id ?? 'missing';
  await page.goto(`/admin/courses/${targetCourseID}`);
  // Modules/lessons sit on the Lessons tab of the course editor (view-3 layout).
  await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
  const module = page.getByRole('group', {
    name: `Модуль ${target.moduleNumber}`,
  });
  const lesson = module.getByRole('group', {
    name: `Урок ${target.lessonNumber}`,
  });
  const minted = page.waitForResponse(
    (response) =>
      new URL(response.url()).pathname === `/api/v1/admin/lessons/${target.lessonID}/video` &&
      response.request().method() === 'POST',
  );
  await lesson.locator('input[type="file"]:not(.admin-audio-field input)').setInputFiles(file);
  const response = await minted;
  expect(response.status()).toBe(201);
  const id = videoID(await response.json());
  // Чекати на data-video-state уроку не можна: якщо в уроку вже була готова генерація,
  // атрибут показує 'ready' МИТТЄВО, і хелпер повертається до того, як нове відео
  // дотранскодувалось. Питаємо стан САМЕ цього asset id.
  await expect
    .poll(
      async () => {
        const listed = await page.request.get(`/api/v1/admin/courses/${targetCourseID}/videos`);
        if (listed.status() !== 200) return `http ${listed.status()}`;
        const body: unknown = await listed.json();
        if (typeof body !== 'object' || body === null || !('items' in body)) return 'no items';
        const items = (body as { items: unknown }).items;
        if (!Array.isArray(items)) return 'items not an array';
        for (const item of items) {
          if (typeof item !== 'object' || item === null) continue;
          const record = item as { videoId?: unknown; state?: unknown };
          if (record.videoId === id && typeof record.state === 'string') return record.state;
        }
        return 'absent';
      },
      { timeout: 240_000, intervals: [1_000] },
    )
    .toBe('ready');
  return id;
}

async function openAndPlay(
  page: Page,
  targetLessonID = lessonID,
  targetNamespace = namespace,
): Promise<PlaybackSession> {
  if (course === undefined) throw new Error('Playback course was not created');
  await page.goto(`/learn/${targetNamespace}/${targetLessonID}`);
  const minted = page.waitForResponse(
    (response) =>
      new URL(response.url()).pathname === '/api/v1/playback/sessions' &&
      response.request().method() === 'POST',
  );
  await page.getByRole('button', { name: 'Відтворити' }).click();
  const response = await minted;
  expect(response.status()).toBe(201);
  // `videoTime > 0` alone does not mean playback started: the player restores a saved position
  // on ready, so a resumed lesson satisfies that guard before `play()` has resolved. The pause
  // The label appears only after the video element emits `playing`, which is the observable that
  // playback actually began - without it a helper returns success and the caller fails elsewhere.
  await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
  await expect.poll(() => videoTime(page)).toBeGreaterThan(0);
  return playbackSession(await response.json());
}

async function playFromAPIContainer(
  baseURL: string,
  cookieHeader: string,
  targetLessonID: string,
): Promise<PlaybackSession> {
  const stdout = await new Promise<string>((resolve, reject) => {
    const child = spawn(
      'docker',
      [
        'compose',
        'exec',
        '-T',
        'api',
        'curl',
        '-sS',
        '-w',
        '\n%{http_code}',
        '-X',
        'POST',
        'http://nginx:8080/api/v1/playback/sessions',
        '-H',
        `Origin: ${new URL(baseURL).origin}`,
        '-H',
        'Content-Type: application/json',
        '-H',
        '@-',
        '--data',
        JSON.stringify({ lessonId: targetLessonID }),
      ],
      { cwd: repositoryRoot, stdio: ['pipe', 'pipe', 'pipe'], timeout: 30_000 },
    );
    let output = '';
    let errorOutput = '';
    child.stdout.setEncoding('utf8');
    child.stderr.setEncoding('utf8');
    child.stdout.on('data', (chunk: string) => {
      output += chunk;
    });
    child.stderr.on('data', (chunk: string) => {
      errorOutput += chunk;
    });
    child.once('error', reject);
    child.once('close', (code) => {
      if (code === 0) resolve(output);
      else reject(new Error(`Container playback request failed (${code}): ${errorOutput.trim()}`));
    });
    child.stdin.end(`Cookie: ${cookieHeader}\n`);
  });
  const lines = stdout.trimEnd().split('\n');
  const status = lines.pop();
  expect(status).toBe('201');
  const body: unknown = JSON.parse(lines.join('\n'));
  return playbackSession(body);
}

async function composeOutput(args: string[], timeout = 30_000): Promise<string> {
  const { stdout } = await execute('docker', ['compose', ...args], {
    cwd: repositoryRoot,
    encoding: 'utf8',
    timeout,
  });
  return stdout.trim();
}

async function authorizeCounts(): Promise<AuthorizeCounts> {
  const metrics = await composeOutput([
    'exec',
    '-T',
    'api',
    'curl',
    '-sS',
    'http://127.0.0.1:8080/metrics',
  ]);
  let observed = false;
  let successful = 0;
  let rejected = 0;
  for (const match of metrics.matchAll(/^courses_http_requests_total\{([^}]*)\}\s+([^\s]+)$/gm)) {
    const labels = match[1];
    const rawValue = match[2];
    if (
      labels === undefined ||
      rawValue === undefined ||
      !labels.includes('method="GET"') ||
      !labels.includes('route="GET /internal/playback/authorize"')
    )
      continue;
    const value = Number(rawValue);
    if (!Number.isFinite(value)) throw new Error(`Invalid authorize metric value: ${rawValue}`);
    observed = true;
    if (labels.includes('status="204"')) successful += value;
    else rejected += value;
  }
  return { observed, successful, rejected, total: successful + rejected };
}

async function waitForAuthorizeSilence(page: Page): Promise<AuthorizeCounts> {
  let previous = await authorizeCounts();
  let stableIntervals = 0;
  for (let attempt = 0; attempt < 30; attempt += 1) {
    await page.waitForTimeout(500);
    const current = await authorizeCounts();
    if (current.total === previous.total) {
      stableIntervals += 1;
      if (stableIntervals === 3) return current;
    } else {
      stableIntervals = 0;
    }
    previous = current;
  }
  throw new Error(
    'B21 could not observe three consecutive quiet 500ms authorize intervals within 15 seconds',
  );
}

async function containerNetworkSnapshot(service: 'api' | 'frontend') {
  const id = await composeOutput(['ps', '-q', service]);
  if (id === '') throw new Error(`Container ${service} is not running`);
  const rawTxBytes = await composeOutput([
    'exec',
    '-T',
    service,
    'cat',
    '/sys/class/net/eth0/statistics/tx_bytes',
  ]);
  if (!/^\d+$/.test(rawTxBytes))
    throw new Error(`Container ${service} returned invalid tx_bytes: ${rawTxBytes}`);
  return { id, txBytes: BigInt(rawTxBytes) } satisfies ContainerNetworkSnapshot;
}

function isMediaRequest(request: Request): boolean {
  return new URL(request.url()).pathname.startsWith('/media/p/');
}

async function videoTime(page: Page): Promise<number> {
  return page.locator('video').evaluate((node) => {
    if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
    return node.currentTime;
  });
}

// A position read AFTER openAndPlay measures the test process, not the player: the helper has
// already waited for the Pause label (the `playing` event) and polled videoTime, and the video keeps
// advancing meanwhile. Against a [39, 41] restore window that read measured 41.37 s on a loaded host
// (2026-10-07, local e2e beside a format/typecheck run) and would fail routinely on a 2 vCPU runner.
// The rows "resumes at the saved position" and "reopens from zero" name currentTime AT `playing`, so
// that is what is captured: an init script (the helpers navigate again), on document in the capture
// phase (`playing` does not bubble), first occurrence only. Init scripts are page-scoped, so arm the
// probe on the page that will navigate — a new context's or a new page's — before its openAndPlay.
async function armStartPositionProbe(page: Page) {
  await page.addInitScript(() => {
    document.addEventListener(
      'playing',
      (event) => {
        const probe = window as Window & { __startPosition?: number };
        if (event.target instanceof HTMLVideoElement && probe.__startPosition === undefined) {
          probe.__startPosition = event.target.currentTime;
        }
      },
      true,
    );
  });
}

async function startPosition(page: Page): Promise<number> {
  const position = await page.evaluate(
    () => (window as Window & { __startPosition?: number }).__startPosition,
  );
  if (typeof position !== 'number') {
    throw new Error('playing never fired on this page, or the probe was armed after navigation');
  }
  return position;
}

// Pauses and returns the position the pause heartbeat recorded - the value a later restore has to
// reproduce. Waiting for the heartbeat response, not for the click, is what makes it the saved one:
// heartbeatPayload reads the request body only behind a 204, so this is what the server stored, not
// a client-side estimate. A paused player schedules no further ticks (LessonPlayer's tick effect is
// gated on `playing`), so the target stops moving. Callers assert the restored start within ±1 s of
// it: that allowance is seek granularity — the player seeks HLS/MSE media to the stored value and
// `playing` reports currentTime after the seek lands on a frame — not a timing assumption.
async function pauseAndRecord(page: Page, targetLessonID = lessonID): Promise<number> {
  const paused = page.waitForResponse(
    (response) => heartbeatPayload(response, targetLessonID)?.reason === 'pause',
  );
  await page.getByRole('button', { name: 'Пауза' }).click();
  const payload = heartbeatPayload(await paused, targetLessonID);
  if (payload === null) throw new Error('the pause heartbeat was not recorded');
  return payload.positionSeconds;
}

async function videoEnded(page: Page): Promise<boolean> {
  return page.locator('video').evaluate((node) => {
    if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
    return node.ended;
  });
}

async function setPosition(page: Page, seconds: number) {
  await page.getByRole('slider', { name: 'Позиція відтворення' }).evaluate((node, value) => {
    if (!(node instanceof HTMLInputElement)) throw new Error('Playback position is not an input');
    const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
    if (setter === undefined) throw new Error('Native input value setter is unavailable');
    Reflect.apply(setter, node, [String(value)]);
    node.dispatchEvent(new Event('input', { bubbles: true }));
    node.dispatchEvent(new Event('change', { bubbles: true }));
  }, seconds);
}

async function segmentCovering(
  page: Page,
  session: PlaybackSession,
  rendition: string,
  positionSeconds: number,
): Promise<string> {
  const master = new URL(session.mediaPath, page.url());
  const playlist = new URL(`${rendition}/index.m3u8`, master);
  const response = await page.request.get(playlist.toString());
  expect(response.status()).toBe(200);
  const lines = (await response.text()).split(/\r?\n/);
  let start = 0;
  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index]?.trim() ?? '';
    if (!line.startsWith('#EXTINF:')) continue;
    const duration = Number.parseFloat(line.slice('#EXTINF:'.length).split(',')[0] ?? '');
    if (!Number.isFinite(duration) || duration <= 0)
      throw new Error(`Invalid EXTINF duration in ${playlist.pathname}`);
    const segment = lines.slice(index + 1).find((candidate) => {
      const trimmed = candidate.trim();
      return trimmed !== '' && !trimmed.startsWith('#');
    });
    if (segment === undefined) throw new Error(`EXTINF has no segment in ${playlist.pathname}`);
    if (positionSeconds >= start && positionSeconds < start + duration)
      return new URL(segment.trim(), playlist).pathname;
    start += duration;
  }
  throw new Error(`No ${rendition} segment covers ${positionSeconds}s`);
}

// Both statements live at module level on purpose, and the reason is not tidiness. A heredoc
// terminator must sit in column zero, but it lives inside a template literal, where JavaScript
// indentation is free and nothing guards it: lint, typecheck and the formatter all see mere
// characters in a string. Nested inside a describe, the terminator moves the moment anyone adds
// a level of nesting or relocates the block - which is exactly how it broke once already. The
// top level cannot be indented by test nesting, so the mine cannot be laid again.
//
// psql interpolates `:'name'` only while parsing its own input, never inside `-c`, so the
// statement is fed on stdin.
const deleteCourseStatement = `exec psql -v ON_ERROR_STOP=1 -v "course_id=$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
-- Memberships reference the course without ON DELETE CASCADE (grants preserve ownership records),
-- and grant_audit references the grant, so cleanup unwinds the chain audit → grants → course.
DELETE FROM grant_audit WHERE grant_id IN (SELECT id FROM course_grants WHERE course_id = :'course_id');
DELETE FROM course_grants WHERE course_id = :'course_id';
DELETE FROM courses WHERE id = :'course_id';
SQL`;

const resetLessonProgressStatement = `exec psql -v ON_ERROR_STOP=1 -v "course_id=$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
DELETE FROM lesson_progress WHERE lesson_id IN (
  SELECT lessons.id FROM course_lessons AS lessons
  JOIN course_modules AS modules ON modules.id = lessons.module_id
  WHERE modules.course_id = :'course_id'
);
SQL`;

test.describe('T7 playback milestone', () => {
  test.describe.configure({ mode: 'serial', timeout: 420_000 });

  test.beforeAll(async ({ browser, baseURL }) => {
    // `describe.configure` sets the timeout of the tests, not of this hook, and the hook
    // transcodes two five-minute fixtures before the first test can run.
    test.setTimeout(720_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const context = await adminContext(browser, baseURL);
    const page = await context.newPage();
    try {
      const created = await context.request.post('/api/v1/admin/courses', {
        data: primaryCourseDraft,
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(created.status()).toBe(201);
      course = courseToken(await created.json());
      const saved = await context.request.put(`/api/v1/admin/courses/${course.id}`, {
        data: {
          expectedUpdatedAt: course.updatedAt,
          course: primaryCourseDraft,
          modules: primaryCourseModules,
        },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(saved.status()).toBe(200);
      course = courseToken(await saved.json());
      await uploadReady(page, path.join(fixtureDirectory, 'e2e-1280x720-5min.mp4'));
      await uploadReady(page, path.join(fixtureDirectory, 'e2e-1280x720-5min.mp4'), {
        lessonID: thirdLessonID,
        moduleNumber: 2,
        lessonNumber: 1,
      });
      fourthLessonVideoID = await uploadReady(
        page,
        path.join(fixtureDirectory, 'e2e-1280x720-100s.mp4'),
        {
          lessonID: fourthLessonID,
          moduleNumber: 2,
          lessonNumber: 2,
        },
      );
      const published = await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { status: 'published', expectedUpdatedAt: course.updatedAt },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(published.status()).toBe(200);
      course = courseToken(await published.json());

      // Enroll the playing student: lesson content needs membership after the enrollment phase.
      const enrolled = await context.request.post(`/api/v1/admin/courses/${course.id}/grants`, {
        data: { userId: studentID, reason: 'e2e playback access for the student' },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(enrolled.status(), await enrolled.text()).toBe(201);

      const crossCourseDraft = {
        slug: crossCourseNamespace,
        title: `T10 cross-course ${namespace}`,
        summary: 'A second course for the account-wide ownership limit.',
        description: 'Run-scoped cross-course playback fixture.',
        instructorId: adminID,
      };
      const crossCourseCreated = await context.request.post('/api/v1/admin/courses', {
        data: crossCourseDraft,
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(crossCourseCreated.status()).toBe(201);
      crossCourse = courseToken(await crossCourseCreated.json());
      const crossCourseSaved = await context.request.put(
        `/api/v1/admin/courses/${crossCourse.id}`,
        {
          data: {
            expectedUpdatedAt: crossCourse.updatedAt,
            course: crossCourseDraft,
            modules: [
              {
                id: crossCourseModuleID,
                title: `Cross-course module ${namespace}`,
                lessons: [
                  {
                    id: crossCourseLessonID,
                    title: `Cross-course lesson ${namespace}`,
                  },
                ],
              },
            ],
          },
          headers: { Origin: new URL(baseURL).origin },
        },
      );
      expect(crossCourseSaved.status()).toBe(200);
      crossCourse = courseToken(await crossCourseSaved.json());
      crossCourseVideoID = await uploadReady(
        page,
        path.join(fixtureDirectory, 'e2e-1280x720-100s.mp4'),
        {
          courseID: crossCourse.id,
          lessonID: crossCourseLessonID,
          moduleNumber: 1,
          lessonNumber: 1,
        },
      );
      const crossCoursePublished = await context.request.post(
        `/api/v1/admin/courses/${crossCourse.id}/status`,
        {
          data: { status: 'published', expectedUpdatedAt: crossCourse.updatedAt },
          headers: { Origin: new URL(baseURL).origin },
        },
      );
      expect(crossCoursePublished.status()).toBe(200);
      crossCourse = courseToken(await crossCoursePublished.json());

      const crossEnrolled = await context.request.post(
        `/api/v1/admin/courses/${crossCourse.id}/grants`,
        {
          data: { userId: studentID, reason: 'e2e cross-course playback access for the student' },
          headers: { Origin: new URL(baseURL).origin },
        },
      );
      expect(crossEnrolled.status(), await crossEnrolled.text()).toBe(201);
    } finally {
      await context.close();
    }
  });

  test.beforeEach(({ page }) => {
    finishRequestTracking = trackPageRequests(page);
  });

  test.afterEach(async () => {
    await finishRequestTracking();
  });

  test.afterAll(async () => {
    for (const fixtureCourse of [course, crossCourse]) {
      if (fixtureCourse === undefined) continue;
      await execute(
        'docker',
        [
          'compose',
          'exec',
          '-T',
          'postgres',
          'sh',
          '-eu',
          '-c',
          deleteCourseStatement,
          'playback-cleanup',
          fixtureCourse.id,
        ],
        { cwd: repositoryRoot, timeout: 60_000 },
      );
    }
  });

  test.describe('as a student', () => {
    // Every test in this block starts from a declared state, not from what its neighbours
    // left. The player restores a saved position on ready, so four T7 tests were sharing one
    // lesson: 80 from P01, then P08's thirty seconds, then B39, and B40 arrived with barely two
    // minutes of runway left - which is why it went red once the admin upload it waits on took
    // longer than the remainder of the video.
    //
    // beforeEach, not beforeAll. beforeAll removes what happened BEFORE the block and would run
    // once, on state that is already clean; the coupling here is between siblings INSIDE it.
    // Resetting only before B40 would have treated the symptom and left the first three coupled.
    const resetLessonProgress = async () => {
      if (course === undefined) throw new Error('Playback course was not created');
      await execute(
        'docker',
        [
          'compose',
          'exec',
          '-T',
          'postgres',
          'sh',
          '-eu',
          '-c',
          resetLessonProgressStatement,
          'playback-progress-reset',
          course.id,
        ],
        { cwd: repositoryRoot, timeout: 60_000 },
      );
    };

    test.beforeEach(resetLessonProgress);

    test.use({ storageState: authStatePath('student') });

    test('P01 P09 P10 plays real HLS, selects a rung and seeks into unloaded media', async ({
      page,
    }) => {
      const mediaResponses: Response[] = [];
      page.on('response', (response) => {
        if (new URL(response.url()).pathname.startsWith('/media/p/')) mediaResponses.push(response);
      });
      const session = await openAndPlay(page);
      const expectedSegment = await segmentCovering(page, session, '360p', 80);
      await expect(page.getByRole('button', { name: 'Налаштування' })).toBeVisible();
      await page.getByRole('button', { name: 'Налаштування' }).click();
      // React Aria hides the radio input inside a visible label and keeps the state itself, so a
      // pointer click targets the label, not the role=radio input (which the label overlays). Same
      // convention as the ratings star group's writeReviewInUI.
      await page.locator('label.ui-radio').filter({ hasText: '360p' }).click();
      await expect(page.getByRole('radio', { name: '360p' })).toBeChecked();
      await page.keyboard.press('Escape');
      await setPosition(page, 80);
      await expect
        .poll(() =>
          mediaResponses.some(
            (response) =>
              response.status() === 200 && new URL(response.url()).pathname === expectedSegment,
          ),
        )
        .toBe(true);
      expect(mediaResponses.some((response) => response.status() === 200)).toBe(true);
      expect(await videoTime(page)).toBeGreaterThan(70);
    });

    test('R35 hls.js automatically switches from 720p to 360p on a narrowed channel', async ({
      page,
    }) => {
      const renditionRequests: string[] = [];
      const renditionRequestCount = (rendition: '360p' | '720p') =>
        renditionRequests.filter((pathname) => pathname.includes(`/${rendition}/`)).length;
      page.on('request', (request) => {
        const pathname = new URL(request.url()).pathname;
        if (pathname.endsWith('.m4s')) renditionRequests.push(pathname);
      });
      const session = await openAndPlay(page);
      // Auto is the initial rendition choice; it now lives as a checked radio inside the gear menu.
      await page.getByRole('button', { name: 'Налаштування' }).click();
      await expect(page.getByRole('radio', { name: 'Автоматично' })).toBeChecked();
      await page.keyboard.press('Escape');
      const master = await page.request.get(new URL(session.mediaPath, page.url()).toString());
      expect(master.status()).toBe(200);
      const masterLines = (await master.text())
        .split(/\r?\n/)
        .map((line) => line.trim())
        .filter((line) => line !== '');
      const streams = masterLines.flatMap((line, index) => {
        if (!line.startsWith('#EXT-X-STREAM-INF:')) return [];
        const match = /(?:^|[:,])BANDWIDTH=(\d+)(?:,|$)/.exec(line);
        const variant = masterLines[index + 1];
        if (match?.[1] === undefined || variant === undefined || variant.startsWith('#'))
          throw new Error('Each EXT-X-STREAM-INF must carry BANDWIDTH and a variant URI');
        return [{ bandwidth: Number.parseInt(match[1], 10), variant }];
      });
      expect(streams.map(({ variant }) => variant).sort()).toEqual([
        '360p/index.m3u8',
        '720p/index.m3u8',
      ]);
      const bandwidth360 = streams.find(({ variant }) => variant === '360p/index.m3u8')?.bandwidth;
      const bandwidth720 = streams.find(({ variant }) => variant === '720p/index.m3u8')?.bandwidth;
      if (bandwidth360 === undefined || bandwidth720 === undefined)
        throw new Error('R35 requires 360p and 720p BANDWIDTH values');
      const throttleBitsPerSecond = Math.sqrt(bandwidth360 * bandwidth720);
      const throttleBytesPerSecond = Math.round(throttleBitsPerSecond / 8);
      expect(throttleBitsPerSecond).toBeGreaterThan(bandwidth360);
      expect(throttleBitsPerSecond).toBeLessThan(bandwidth720);
      await expect
        .poll(() => renditionRequests.some((pathname) => pathname.includes('/720p/')), {
          timeout: 60_000,
        })
        .toBe(true);

      const highRequestsBeforeControlSeek = renditionRequestCount('720p');
      await setPosition(page, 100);
      await expect
        .poll(() => renditionRequestCount('720p'), { timeout: 30_000 })
        .toBeGreaterThan(highRequestsBeforeControlSeek);
      const lowRequestsBeforeThrottle = renditionRequestCount('360p');
      const highRequestsBeforeThrottle = renditionRequestCount('720p');
      const requestIndexBeforeThrottle = renditionRequests.length;
      const devtools = await page.context().newCDPSession(page);
      await devtools.send('Network.enable');
      try {
        // O1's 1 KB/s must stall every rung; R35 needs a master-derived midpoint where 360p is
        // viable and 720p is not, otherwise hls.js has no useful lower level to switch to.
        await devtools.send('Network.emulateNetworkConditions', {
          offline: false,
          latency: 100,
          downloadThroughput: throttleBytesPerSecond,
          uploadThroughput: 1024 * 1024,
        });
        const lowerRenditionRequest = page.waitForRequest(
          (request) => {
            const pathname = new URL(request.url()).pathname;
            return pathname.endsWith('.m4s') && pathname.includes('/360p/');
          },
          { timeout: 60_000 },
        );
        await setPosition(page, 150);
        await lowerRenditionRequest.then(
          () => true,
          () => false,
        );
        const lowRequestsAfterThrottle = renditionRequestCount('360p');
        const highRequestsAfterThrottle = renditionRequestCount('720p');
        const requestsAfterThrottle = renditionRequests.slice(requestIndexBeforeThrottle);
        console.log(
          `R35 master BANDWIDTH: 360p=${bandwidth360} bit/s, 720p=${bandwidth720} bit/s; derived throttle=${throttleBytesPerSecond} B/s (${Math.round(throttleBitsPerSecond)} bit/s)`,
        );
        console.log(
          `R35 baselines before throttle: 360p=${lowRequestsBeforeThrottle}, 720p=${highRequestsBeforeThrottle}`,
        );
        console.log(
          `R35 counts after throttle: 360p=${lowRequestsAfterThrottle}, 720p=${highRequestsAfterThrottle}`,
        );
        console.log(
          `R35 rendition requests after throttle: ${requestsAfterThrottle.length === 0 ? 'none' : requestsAfterThrottle.join(', ')}`,
        );
        // The 720p counts stay printed, never asserted. Whether another high fragment is requested
        // before the downswitch is timing, not the claim: if hls.js drops the level before the seek
        // pulls anything, the player did better than required and this row must not go red for it.
        // What R35 claims is the lower rung being reached on a narrowed channel, and that is the
        // 360p growth alone. The printed pair still separates "ABR had nothing to abandon" from
        // "ABR abandoned and stepped down" when this row fails.
        expect(lowRequestsAfterThrottle).toBeGreaterThan(lowRequestsBeforeThrottle);
        // The rung switch was hls.js's doing, not the student's: the quality choice is still Auto.
        await page.getByRole('button', { name: 'Налаштування' }).click();
        await expect(page.getByRole('radio', { name: 'Автоматично' })).toBeChecked();
        await page.keyboard.press('Escape');
      } finally {
        await devtools.send('Network.emulateNetworkConditions', {
          offline: false,
          latency: 0,
          downloadThroughput: -1,
          uploadThroughput: -1,
        });
        await devtools.detach();
      }
    });

    test('P08 renews ownership at least four times during a 30-second view', async ({ page }) => {
      let heartbeats = 0;
      let deniedSegments = 0;
      page.on('response', (response) => {
        const pathname = new URL(response.url()).pathname;
        if (pathname === '/api/v1/playback/heartbeat' && response.status() === 204) heartbeats += 1;
        if (pathname.startsWith('/media/p/') && response.status() === 403) deniedSegments += 1;
      });
      const session = await openAndPlay(page);
      expect(
        {
          heartbeatSeconds: session.heartbeatSeconds,
          ownershipTtlSeconds: session.ownershipTtlSeconds,
        },
        'Recreate api with the documented T7 2s/6s profile',
      ).toEqual({ heartbeatSeconds: 2, ownershipTtlSeconds: 6 });
      await page.waitForTimeout(30_000);
      const watched = await videoTime(page);
      // The measured values, not only the verdict, so a CI log shows how far each sits from its
      // threshold (K12: the list reporter gives durations, never an assertion's margin). At the 2 s
      // cadence a 30 s view should renew ~15 times; a count drifting toward 4 is the early warning.
      console.log(
        `P08 measurements: heartbeats=${heartbeats} (min 4), deniedSegments=${deniedSegments} (must be 0), videoTime=${watched.toFixed(1)}s (min 25)`,
      );
      expect(heartbeats).toBeGreaterThanOrEqual(4);
      expect(deniedSegments).toBe(0);
      expect(watched).toBeGreaterThan(25);
    });

    test('P04 restores the position after a reload', async ({ page }) => {
      await openAndPlay(page);
      const saved = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 40;
      });
      await setPosition(page, 40);
      await saved;

      await armStartPositionProbe(page);
      await page.reload();
      await openAndPlay(page);
      const restored = await startPosition(page);
      expect(restored).toBeGreaterThanOrEqual(39);
      expect(restored).toBeLessThanOrEqual(41);
    });

    test('P05 restores the position in another browser context', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      await openAndPlay(page);
      const saved = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 40;
      });
      await setPosition(page, 40);
      await saved;
      // The first page keeps playing and keeps recording (a tick every 2 s on the e2e profile), so
      // "the saved position" is a moving target while the second context is being opened: on the
      // 2 vCPU runner (2026-10-07) the second page started at 41.03 s, correctly restoring a position
      // the first page had meanwhile saved. Freeze it: pause, and anchor the expectation on the
      // position the pause heartbeat actually recorded instead of on the number the seek asked for.
      const pausedAt = await pauseAndRecord(page);

      const secondContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const secondPage = await secondContext.newPage();
        await armStartPositionProbe(secondPage);
        await openAndPlay(secondPage);
        const restored = await startPosition(secondPage);
        expect(restored).toBeGreaterThanOrEqual(pausedAt - 1);
        expect(restored).toBeLessThanOrEqual(pausedAt + 1);
      } finally {
        await secondContext.close();
      }
    });

    test('P06 marks the dedicated 100-second lesson at its proportional 95% threshold', async ({
      page,
    }) => {
      await openAndPlay(page, fourthLessonID);
      const duration = Number(
        await page.getByRole('slider', { name: 'Позиція відтворення' }).getAttribute('max'),
      );
      expect(duration).toBeGreaterThanOrEqual(99);
      expect(duration).toBeLessThanOrEqual(101);
      const completed = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, fourthLessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 95;
      });
      await setPosition(page, 95);
      await completed;

      await page.reload();
      const outline = page.getByRole('navigation', { name: 'Зміст курсу' });
      const lessons = outline.getByRole('link');
      await expect(lessons).toHaveCount(4);
      await expect(lessons.nth(3)).toContainText('Переглянуто');
      await expect(lessons.nth(0)).not.toContainText('Переглянуто');
    });

    test('P07 completes its own lesson and reopens it from zero', async ({ page }) => {
      await openAndPlay(page);
      const completed = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds >= 285;
      });
      await setPosition(page, 285);
      await completed;

      await armStartPositionProbe(page);
      await page.reload();
      await expect(page.getByRole('slider', { name: 'Позиція відтворення' })).toHaveValue('0');
      await openAndPlay(page);
      expect(await startPosition(page)).toBeLessThanOrEqual(1);
    });

    test('B18 keeps the last recorded position across a network break', async ({ page }) => {
      await openAndPlay(page);
      const saved = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 40;
      });
      await setPosition(page, 40);
      await saved;
      // Same anchor as P05: the row restores what the pause recorded, not the number the seek asked for.
      const pausedAt = await pauseAndRecord(page);

      // The break has to be observable, or an offline toggle that changed nothing would pass.
      // It shows up as the player's unavailable state: losing the network exhausts the media
      // retries, and `LessonPlayer.tsx:338` then replaces the whole player - controls included.
      //
      // Which is why this row does not seek while offline. That assumed the player stays
      // operable through a break; it does not, and it is right not to. The row's claim is that
      // the recorded position survives the break, not that the player keeps taking input during
      // one - a failed write moving the saved position is B11's subject, at the datastore.
      await page.context().setOffline(true);
      await expect(page.getByText('Відео зараз недоступне.')).toBeVisible({ timeout: 60_000 });
      await page.context().setOffline(false);

      const recoveredPage = await page.context().newPage();
      try {
        await armStartPositionProbe(recoveredPage);
        await openAndPlay(recoveredPage);
        const restored = await startPosition(recoveredPage);
        expect(restored).toBeGreaterThanOrEqual(pausedAt - 1);
        expect(restored).toBeLessThanOrEqual(pausedAt + 1);
      } finally {
        await recoveredPage.close();
      }
    });

    test('B39 O1 remints after 410 without takeover and prints browser measurements', async ({
      page,
    }) => {
      const active = new Set<Request>();
      let activeAtMint = -1;
      let playRequests = 0;
      let gateExpiredHeartbeat = false;
      let markHeldSegmentStarted: (() => void) | undefined;
      let allowExpiredHeartbeat: (() => void) | undefined;
      const heldSegmentStarted = new Promise<void>((resolve) => {
        markHeldSegmentStarted = resolve;
      });
      const heartbeatMayContinue = new Promise<void>((resolve) => {
        allowExpiredHeartbeat = resolve;
      });
      // The gate stays open until the segment has actually started. Clearing it on entry
      // deadlocked the test: `seekTo` queues the heartbeat right after assigning currentTime,
      // so the heartbeat reaches this handler before hls.js has reacted to the seek - the flag
      // went false, the request listener below then ignored the segment, and nothing was left
      // to resolve the promise this handler awaits. A separate latch keeps the handler
      // one-shot without closing the window the listener needs.
      let heartbeatAlreadyHeld = false;
      await page.route('**/api/v1/playback/heartbeat', async (route) => {
        if (!gateExpiredHeartbeat || heartbeatAlreadyHeld) {
          await route.continue();
          return;
        }
        heartbeatAlreadyHeld = true;
        await heldSegmentStarted;
        gateExpiredHeartbeat = false;
        await heartbeatMayContinue;
        await route.continue();
      });
      page.on('request', (request) => {
        if (isMediaRequest(request)) {
          active.add(request);
          if (gateExpiredHeartbeat && new URL(request.url()).pathname.endsWith('.m4s'))
            markHeldSegmentStarted?.();
        }
        if (new URL(request.url()).pathname === '/api/v1/playback/sessions') {
          playRequests += 1;
          if (playRequests === 2) activeAtMint = active.size;
        }
      });
      // Every denied media request with its address. When the player ends in an error state the
      // question is always which prefix the 403 carried against the one in use, and neither the
      // Playwright trace nor the server log answers it - the trace truncates under the fixture's
      // upload polling, and the server sees paths, not which session the client thought it had.
      const deniedMedia: string[] = [];
      page.on('response', (response) => {
        if (response.status() === 403 && isMediaRequest(response.request()))
          deniedMedia.push(new URL(response.url()).pathname);
      });
      const settle = (request: Request) => active.delete(request);
      page.on('requestfinished', settle);
      page.on('requestfailed', settle);
      const session = await openAndPlay(page);
      await page.getByRole('button', { name: 'Пауза' }).click();
      await page.locator('video').evaluate((video) => {
        if (!(video instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        const descriptor = Object.getOwnPropertyDescriptor(
          HTMLMediaElement.prototype,
          'currentTime',
        );
        if (descriptor?.get === undefined || descriptor.set === undefined)
          throw new Error('HTMLMediaElement.currentTime descriptor is unavailable');
        const getCurrentTime = descriptor.get;
        const setCurrentTime = descriptor.set;
        const writes: Array<{ before: number; assigned: number }> = [];
        Reflect.set(video, '__o1CurrentTimeWrites', writes);
        Object.defineProperty(video, 'currentTime', {
          configurable: true,
          get: () => Reflect.apply(getCurrentTime, video, []),
          set: (assigned: number) => {
            writes.push({
              before: Reflect.apply(getCurrentTime, video, []),
              assigned,
            });
            Reflect.apply(setCurrentTime, video, [assigned]);
          },
        });
      });
      await page.waitForTimeout((session.ownershipTtlSeconds + 1) * 1_000);
      const devtools = await page.context().newCDPSession(page);
      await devtools.send('Network.enable');
      await devtools.send('Network.emulateNetworkConditions', {
        offline: false,
        latency: 100,
        downloadThroughput: 1_024,
        uploadThroughput: 1024 * 1024,
      });
      gateExpiredHeartbeat = true;
      await setPosition(page, 150);
      await heldSegmentStarted;
      const preRemint = await videoTime(page);
      const activeImmediatelyBeforeRemint = active.size;
      allowExpiredHeartbeat?.();
      await expect.poll(() => playRequests).toBe(2);
      await devtools.send('Network.emulateNetworkConditions', {
        offline: false,
        latency: 0,
        downloadThroughput: -1,
        uploadThroughput: -1,
      });
      await devtools.detach();
      await expect(page.getByText('Відео відтворюється на іншому пристрої.')).toHaveCount(0);
      // If the player has fallen into any error state its whole subtree is replaced, `<video>`
      // included - and every locator after this point then waits out the full test timeout
      // while saying nothing about why. Seven minutes of silence for a state the page is
      // displaying in plain text. Fail here instead, carrying that text.
      if ((await page.locator('video').count()) === 0) {
        const shown = await page.locator('main, body').first().innerText();
        throw new Error(
          [
            `Player left no <video> after the remint. On screen: ${shown.replace(/\s+/g, ' ').slice(0, 300)}`,
            `First session media path: ${session.mediaPath}`,
            `Media requests answered 403: ${deniedMedia.length === 0 ? 'none' : deniedMedia.join(', ')}`,
          ].join('\n'),
        );
      }

      // Read and print the measurements BEFORE asserting on them. This test exists to produce
      // numbers; printing them after the assertion hid them on exactly the runs where they were
      // needed - the same failure shape as evidence that disappears when a container restarts.
      const readWrites = () =>
        page.locator('video').evaluate((video) => {
          const value: unknown = Reflect.get(video, '__o1CurrentTimeWrites');
          if (!Array.isArray(value)) return [];
          return value.flatMap((entry) => {
            if (typeof entry !== 'object' || entry === null) return [];
            const beforeWrite = Reflect.get(entry, 'before');
            const assigned = Reflect.get(entry, 'assigned');
            return typeof beforeWrite === 'number' && typeof assigned === 'number'
              ? [{ before: beforeWrite, assigned }]
              : [];
          });
        });
      const findRestore = (all: Array<{ before: number; assigned: number }>) =>
        all.find(({ assigned }, index) => index > 0 && Math.abs(assigned - preRemint) <= 1);
      const writes = await readWrites();
      const restoreWrite = findRestore(writes);
      const stopLoadObservation =
        activeImmediatelyBeforeRemint === 0
          ? 'inconclusive: no old request was in flight immediately before remint'
          : activeAtMint === 0
            ? 'yes'
            : `no: ${activeAtMint} old request(s) remained`;
      console.log(
        `O1 hls.js in-flight immediately before remint: ${activeImmediatelyBeforeRemint}`,
      );
      console.log(`O1 hls.js in-flight when mint request began after stopLoad: ${activeAtMint}`);
      console.log(`O1 stopLoad cleared old in-flight requests before mint: ${stopLoadObservation}`);
      console.log(
        `O1 currentTime before explicit restore: ${restoreWrite?.before ?? 'not observed'}; assigned restore: ${restoreWrite?.assigned ?? 'not observed'}; pre-remint: ${preRemint}`,
      );
      // Every write, not only the one being looked for. "Not observed" says a match was absent;
      // it does not say what was there instead, and that difference has cost three runs today.
      console.log(
        `O1 every currentTime write: ${writes.length === 0 ? 'none' : writes.map(({ before: from, assigned }) => `${from}->${assigned}`).join(', ')}`,
      );
      await expect
        .poll(
          () =>
            page.locator('video').evaluate((video, target) => {
              const value: unknown = Reflect.get(video, '__o1CurrentTimeWrites');
              if (!Array.isArray(value)) return false;
              return value.some((entry, index) => {
                if (index === 0 || typeof entry !== 'object' || entry === null) return false;
                const assigned: unknown = Reflect.get(entry, 'assigned');
                return typeof assigned === 'number' && Math.abs(assigned - target) <= 1;
              });
            }, preRemint),
          { timeout: 30_000 },
        )
        .toBe(true);
      expect(activeImmediatelyBeforeRemint).toBeGreaterThan(0);
      expect(activeAtMint).toBe(0);
      // B-23: assert on a snapshot taken after the poll. The early snapshot above is printed for diagnosis,
      // but the restore write can land between it and the poll's success — asserting on it failed with
      // `undefined` even though the poll had just proven the write.
      const settledRestore = findRestore(await readWrites());
      expect(settledRestore?.assigned).toBeGreaterThanOrEqual(preRemint - 1);
      expect(settledRestore?.assigned).toBeLessThanOrEqual(preRemint + 1);
      await expect.poll(() => videoTime(page)).toBeGreaterThanOrEqual(preRemint - 1);
    });

    test('B-7 a pause outside the button stops ownership renewal', async ({ page }) => {
      let successfulHeartbeats = 0;
      page.on('response', (response) => {
        if (
          response.status() === 204 &&
          new URL(response.url()).pathname === '/api/v1/playback/heartbeat'
        )
          successfulHeartbeats += 1;
      });
      const firstTick = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'tick';
      });
      const session = await openAndPlay(page);
      await firstTick;

      // R39. Stopping the cadence is only half of it: the pause has to record the exit as well,
      // or a student who pauses from the browser's media hub and closes the tab loses everything
      // since the last beat. Read the position before and after - the write alone proves nothing
      // without what preceded it.
      const before = await page.request.get(`/api/v1/learn/courses/${namespace}`);
      expect(before.status()).toBe(200);
      const positionBeforePause = learningLessonPosition(await before.json(), lessonID);
      // Let playback advance past the tick that was just written, so a recorded pause is
      // distinguishable from it by value and not only by count.
      await page.waitForTimeout(1_000);

      const pauseWritten = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'pause';
      });
      await page.locator('video').evaluate((node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        node.pause();
      });
      await expect(page.getByRole('button', { name: 'Відтворити' })).toBeVisible();
      await pauseWritten;

      const after = await page.request.get(`/api/v1/learn/courses/${namespace}`);
      expect(after.status()).toBe(200);
      expect(learningLessonPosition(await after.json(), lessonID)).toBeGreaterThan(
        positionBeforePause,
      );
      const heartbeatsAtPause = successfulHeartbeats;
      await page.waitForTimeout(session.heartbeatSeconds * 2_000 + 250);
      expect(successfulHeartbeats).toBe(heartbeatsAtPause);

      const resumedTick = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'tick';
      });
      await page.locator('video').evaluate(async (node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        await node.play();
      });
      await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
      await resumedTick;
      expect(successfulHeartbeats).toBeGreaterThan(heartbeatsAtPause);
    });

    test('P13 B15 B31 B32 takeover stops the first device and lets it retake the slot', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const served = page.waitForResponse(
        (response) =>
          response.status() === 200 &&
          isMediaRequest(response.request()) &&
          new URL(response.url()).pathname.endsWith('.m4s'),
      );
      const firstSession = await openAndPlay(page);
      const segmentPath = new URL((await served).url()).pathname;
      const beforeTakeover = await page.request.get(segmentPath);
      expect(beforeTakeover.status()).toBe(200);

      const secondContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const secondPage = await secondContext.newPage();
        const secondSession = await openAndPlay(secondPage);
        const secondSegmentPath = segmentPath.replace(
          firstSession.playbackId,
          secondSession.playbackId,
        );
        const beforeRetake = await secondPage.request.get(secondSegmentPath);
        expect(beforeRetake.status()).toBe(200);
        const afterTakeover = await page.request.get(segmentPath);
        expect(afterTakeover.status()).toBe(403);
        await expect(
          page.getByRole('heading', {
            name: 'Відео відтворюється на іншому пристрої.',
          }),
        ).toBeVisible();
        await expect(secondPage.getByRole('button', { name: 'Пауза' })).toBeVisible();

        const reminted = page.waitForResponse(
          (response) =>
            new URL(response.url()).pathname === '/api/v1/playback/sessions' &&
            response.request().method() === 'POST',
        );
        await page.getByRole('button', { name: 'Почати перегляд знову' }).click();
        const remintedResponse = await reminted;
        expect(remintedResponse.status()).toBe(201);
        const retakenSession = playbackSession(await remintedResponse.json());

        await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
        const retakenSegmentPath = segmentPath.replace(
          firstSession.playbackId,
          retakenSession.playbackId,
        );
        const afterRetake = await page.request.get(retakenSegmentPath);
        expect(afterRetake.status()).toBe(200);

        await expect(
          secondPage.getByRole('heading', {
            name: 'Відео відтворюється на іншому пристрої.',
          }),
        ).toBeVisible();
        const displacedSegment = await secondPage.request.get(secondSegmentPath);
        expect(displacedSegment.status()).toBe(403);
      } finally {
        await secondContext.close();
      }
    });

    test('P14 takeover is account-wide across two courses', async ({ page, browser, baseURL }) => {
      if (baseURL === undefined || crossCourse === undefined)
        throw new Error('Cross-course playback fixture is unavailable');
      const firstSession = await openAndPlay(page);
      const secondContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const secondPage = await secondContext.newPage();
        const secondSession = await openAndPlay(
          secondPage,
          crossCourseLessonID,
          crossCourseNamespace,
        );
        expect(secondSession.mediaPath).not.toBe(firstSession.mediaPath);
        await expect(
          page.getByRole('heading', {
            name: 'Відео відтворюється на іншому пристрої.',
          }),
        ).toBeVisible();
        await expect(secondPage.getByRole('button', { name: 'Пауза' })).toBeVisible();
      } finally {
        await secondContext.close();
      }
    });

    test('B33 two tabs in one browser displace each other', async ({ page }) => {
      await openAndPlay(page);
      const secondPage = await page.context().newPage();
      try {
        await openAndPlay(secondPage);
        await expect(
          page.getByRole('heading', {
            name: 'Відео відтворюється на іншому пристрої.',
          }),
        ).toBeVisible();
        await expect(secondPage.getByRole('button', { name: 'Пауза' })).toBeVisible();
      } finally {
        await secondPage.close();
      }
    });

    test('R25 two accounts behind one IP keep playing simultaneously', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const studentSession = await openAndPlay(page);
      const studentTimeBefore = await videoTime(page);
      const context = await adminContext(browser, baseURL);
      try {
        const adminPage = await context.newPage();
        await openAndPlay(adminPage);
        await page.waitForTimeout(studentSession.heartbeatSeconds * 1_000 + 250);
        await expect(
          page.getByRole('heading', {
            name: 'Відео відтворюється на іншому пристрої.',
          }),
        ).toHaveCount(0);
        await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
        await expect(adminPage.getByRole('button', { name: 'Пауза' })).toBeVisible();
        expect(await videoTime(page)).toBeGreaterThan(studentTimeBefore);
        expect(await videoTime(adminPage)).toBeGreaterThan(0);
      } finally {
        await context.close();
      }
    });

    test('R26 the same account from another network address still takes over', async ({
      page,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const firstSession = await openAndPlay(page);
      const cookies = await page.context().cookies(baseURL);
      const cookieHeader = cookies.map(({ name, value }) => `${name}=${value}`).join('; ');
      if (cookieHeader === '') throw new Error('Student session cookie is unavailable');

      const secondSession = await playFromAPIContainer(baseURL, cookieHeader, lessonID);
      expect(secondSession.playbackId).not.toBe(firstSession.playbackId);
      await expect(
        page.getByRole('heading', {
          name: 'Відео відтворюється на іншому пристрої.',
        }),
      ).toBeVisible();
    });

    test('B16 an abandoned session does not delay play on another device', async ({
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const origin = new URL(baseURL).origin;
      const abandonedContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      let abandonedSession: PlaybackSession;
      try {
        const abandoned = await abandonedContext.request.post('/api/v1/playback/sessions', {
          data: { lessonId: lessonID },
          headers: { Origin: origin },
        });
        expect(abandoned.status()).toBe(201);
        abandonedSession = playbackSession(await abandoned.json());
      } finally {
        await abandonedContext.close();
      }

      const nextContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const next = await nextContext.request.post('/api/v1/playback/sessions', {
          data: { lessonId: lessonID },
          headers: { Origin: origin },
        });
        expect(next.status()).toBe(201);
        const nextSession = playbackSession(await next.json());
        expect(nextSession.playbackId).not.toBe(abandonedSession.playbackId);
        const media = await nextContext.request.get(nextSession.mediaPath);
        expect(media.status()).toBe(200);
      } finally {
        await nextContext.close();
      }
    });

    test('B40 finishes the retained generation and sees replacement on next opening', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const first = await openAndPlay(page);
      const oldVideo = /\/hls\/([^/]+)\/master\.m3u8$/.exec(first.mediaPath)?.[1];
      if (oldVideo === undefined) throw new Error('First playback path has no video id');
      const context = await adminContext(browser, baseURL);
      try {
        const adminPage = await context.newPage();
        const replacement = await uploadReady(
          adminPage,
          path.join(fixtureDirectory, 'e2e-1280x720-5min.mp4'),
        );
        expect(replacement).not.toBe(oldVideo);
      } finally {
        await context.close();
      }
      const oldResponses: number[] = [];
      page.on('response', (response) => {
        const pathname = new URL(response.url()).pathname;
        if (pathname.includes(`/hls/${oldVideo}/`) && pathname.endsWith('.m4s'))
          oldResponses.push(response.status());
      });
      await setPosition(page, 298);
      await expect.poll(() => videoEnded(page)).toBe(true);
      expect(oldResponses).toContain(200);
      expect(oldResponses).not.toContain(404);
      expect(oldResponses).not.toContain(403);
      await page.reload();
      const next = page.waitForResponse(
        (response) => new URL(response.url()).pathname === '/api/v1/playback/sessions',
      );
      await page.getByRole('button', { name: 'Відтворити' }).click();
      const nextSession = playbackSession(await (await next).json());
      expect(nextSession.mediaPath).not.toContain(`/hls/${oldVideo}/`);
    });

    // Saved progress makes a lesson shared state between tests: the player restores
    // `lesson.position` on ready, so whoever seeks furthest decides where the next test
    // starts. Four tests seek on lesson 1 - 80, 298, 290, 120 - and the two near the end
    // are fatal, because a resumed lesson two seconds from its finish never resolves
    // `play()` and the caller waits out its whole timeout.
    //
    // Order cannot fix this: both directions poison the other. Run this block first and
    // B39 inherits 290; run it last and it inherits 298 from B40 - and then the P02
    // assertion that lesson 1 is watched would be satisfied by B40's seek rather than by
    // the chain P02 exists to prove, which is a green main assertion resting on nothing.
    //
    // So the block starts from a known state instead of a hoped-for one. Lessons 2 and 4
    // carry no video, so there is no other lesson to move onto; this runs before R04,
    // which archives the course irreversibly.
    // Grouping only: the progress reset now runs before every test in the block above, so
    // these rows need no hooks of their own.
    test.describe('progress, the outline and the clamp', () => {
      test('P02 P03 persists real progress, marks the outline, and enters at the server-selected lesson', async ({
        page,
      }) => {
        const firstHeartbeat = page.waitForResponse((response) => {
          const position = heartbeatPosition(response, lessonID);
          return position !== null && position > 0 && position < 285;
        });
        await openAndPlay(page);
        const firstHeartbeatResponse = await firstHeartbeat;
        const firstPosition = heartbeatPosition(firstHeartbeatResponse, lessonID);
        if (firstPosition === null) throw new Error('First heartbeat position was not captured');
        expect(firstPosition).toBeGreaterThan(0);
        // The exact equality below races the 2s cadence. Pausing stops that cadence and now owns
        // its required exit write, so the read is compared with the observed pause heartbeat,
        // not with the earlier tick that the pause is allowed to supersede.
        const paused = page.waitForResponse((response) => {
          const heartbeat = heartbeatPayload(response, lessonID);
          return heartbeat?.reason === 'pause';
        });
        await page.getByRole('button', { name: 'Пауза' }).click();
        const pausedHeartbeat = heartbeatPayload(await paused, lessonID);
        if (pausedHeartbeat === null) throw new Error('Pause heartbeat was not captured');
        const firstRead = await page.request.get(`/api/v1/learn/courses/${namespace}`);
        expect(firstRead.status()).toBe(200);
        expect(learningLessonPosition(await firstRead.json(), lessonID)).toBe(
          pausedHeartbeat.positionSeconds,
        );

        const completedHeartbeat = page.waitForResponse((response) => {
          const heartbeat = heartbeatPayload(response, lessonID);
          return heartbeat?.reason === 'seek' && heartbeat.positionSeconds >= 285;
        });
        // T9 makes seek itself the exit write; resuming here would add an unrelated heartbeat and
        // weaken the row's claim about which event persisted completion.
        await setPosition(page, 290);
        await completedHeartbeat;

        let learningReads = 0;
        page.on('request', (request) => {
          if (
            request.method() === 'GET' &&
            new URL(request.url()).pathname === `/api/v1/learn/courses/${namespace}`
          )
            learningReads += 1;
        });
        await page.goto(`/courses/${namespace}`);
        await page.getByRole('link', { name: 'Почати навчання' }).click();

        await expect(page).toHaveURL(`/learn/${namespace}/${secondLessonID}`);
        const outline = page.getByRole('navigation', { name: 'Зміст курсу' });
        const lessons = outline.getByRole('link');
        await expect(lessons).toHaveCount(4);
        await expect(lessons.nth(0)).toContainText('Переглянуто');
        await expect(lessons.nth(0)).not.toHaveAttribute('aria-current', 'page');
        await expect(lessons.nth(1)).toHaveAttribute('aria-current', 'page');
        await expect(lessons.nth(1)).toContainText('Поточний урок');
        // Without these three the row does not exclude an outline that marks every lesson
        // watched, which is what reading the course's `completed` instead of the lesson's would
        // produce. All five assertions above pass in that case; these are the ones that do not.
        await expect(lessons.nth(1)).not.toContainText('Переглянуто');
        await expect(lessons.nth(2)).not.toContainText('Переглянуто');
        await expect(lessons.nth(3)).not.toContainText('Переглянуто');
        expect(learningReads).toBe(1);
      });

      test('R37 clamps an oversized heartbeat to the known 300-second duration', async ({
        page,
      }) => {
        await page.goto(`/learn/${namespace}/${thirdLessonID}`);
        const origin = new URL(page.url()).origin;
        const minted = await page.request.post('/api/v1/playback/sessions', {
          data: { lessonId: thirdLessonID },
          headers: { Origin: origin },
        });
        expect(minted.status()).toBe(201);
        const session = playbackSession(await minted.json());
        const heartbeat = await page.request.post('/api/v1/playback/heartbeat', {
          data: {
            playbackId: session.playbackId,
            lessonId: thirdLessonID,
            positionSeconds: 100_000,
            clientSeq: session.clientSeq + 1,
            reason: 'seek',
          },
          headers: { Origin: origin },
        });
        expect(heartbeat.status()).toBe(204);
        const learning = await page.request.get(`/api/v1/learn/courses/${namespace}`);
        expect(learning.status()).toBe(200);
        expect(learningLessonPosition(await learning.json(), thirdLessonID)).toBe(300);
      });
    });

    test('B20 B21 B21a B22 serves independent ranged segments with one authorization each', async ({
      page,
    }) => {
      const served = page.waitForResponse(
        (response) =>
          response.status() === 200 &&
          isMediaRequest(response.request()) &&
          new URL(response.url()).pathname.endsWith('.m4s'),
      );
      const session = await openAndPlay(page, crossCourseLessonID, crossCourseNamespace);
      const segmentPath = new URL((await served).url()).pathname;

      const fullSegment = await page.request.get(segmentPath);
      expect(fullSegment.status()).toBe(200);
      const fullSegmentBytes = await fullSegment.body();
      expect(fullSegmentBytes.byteLength).toBeGreaterThan(1_024);

      const headers = fullSegment.headers();
      expect(headers['cache-control']).toBe('private, no-store');
      expect(headers['x-content-type-options']).toBe('nosniff');
      expect(headers['referrer-policy']).toBe('strict-origin-when-cross-origin');
      expect(headers['x-frame-options']).toBe('DENY');

      const ranged = await page.request.get(segmentPath, {
        headers: { Range: 'bytes=0-1023' },
      });
      expect(ranged.status()).toBe(206);
      expect((await ranged.body()).byteLength).toBe(1_024);
      expect(ranged.headers()['content-range']).toBe(`bytes 0-1023/${fullSegmentBytes.byteLength}`);

      const masterURL = new URL(session.mediaPath, page.url());
      const master = await page.request.get(masterURL.toString());
      expect(master.status()).toBe(200);
      const masterLines = (await master.text()).split(/\r?\n/).map((line) => line.trim());
      const variantURLs = masterLines.flatMap((line, index) => {
        if (!line.startsWith('#EXT-X-STREAM-INF:')) return [];
        const variant = masterLines[index + 1];
        if (variant === undefined || variant === '' || variant.startsWith('#'))
          throw new Error('Each EXT-X-STREAM-INF must be followed by a variant URI');
        return [new URL(variant, masterURL)];
      });
      expect(variantURLs).toHaveLength(2);
      const generatedSegmentPaths: string[] = [];
      for (const variantURL of variantURLs) {
        const variant = await page.request.get(variantURL.toString());
        expect(variant.status()).toBe(200);
        const playlist = await variant.text();
        expect(playlist).not.toContain('#EXT-X-BYTERANGE');
        const lines = playlist.split(/\r?\n/).map((line) => line.trim());
        const segmentURIs = lines.flatMap((line, index) => {
          if (!line.startsWith('#EXTINF:')) return [];
          const segment = lines[index + 1];
          if (segment === undefined || segment === '' || segment.startsWith('#'))
            throw new Error(`EXTINF in ${variantURL.pathname} must be followed by a segment URI`);
          return [segment];
        });
        expect(segmentURIs.length).toBeGreaterThan(0);
        expect(new Set(segmentURIs).size).toBe(segmentURIs.length);
        generatedSegmentPaths.push(
          ...segmentURIs.map((segment) => new URL(segment, variantURL).pathname),
        );
      }
      expect(new Set(generatedSegmentPaths).size).toBe(generatedSegmentPaths.length);

      const pauseWritten = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, crossCourseLessonID);
        return heartbeat?.reason === 'pause';
      });
      await page.locator('video').evaluate((node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        node.pause();
      });
      await expect(page.getByRole('button', { name: 'Відтворити' })).toBeVisible();
      await pauseWritten;
      const before = await waitForAuthorizeSilence(page);
      if (!before.observed)
        throw new Error('B21 authorize metric series is absent after serving a real segment');

      for (const range of ['bytes=0-0', 'bytes=1-1', 'bytes=2-2']) {
        const response = await page.request.get(segmentPath, { headers: { Range: range } });
        expect(response.status()).toBe(206);
      }
      const after = await authorizeCounts();
      console.log(
        `B21 authorize counters: before success=${before.successful} rejected=${before.rejected}; after success=${after.successful} rejected=${after.rejected}`,
      );
      expect(after.observed).toBe(true);
      expect(
        after.successful - before.successful,
        'B21 expected exactly three successful authorizations. If the delta exceeds 3, first check whether silence was established and hls.js buffered after pause (see the printed counters); only then investigate duplicate auth_request calls.',
      ).toBe(3);
      expect(after.rejected - before.rejected).toBe(0);
    });

    test('R23 a complete 100-second lesson sends at least ten times fewer bytes through Go and SSR', async ({
      page,
    }) => {
      if (fourthLessonVideoID === undefined)
        throw new Error('The dedicated 100-second playback asset is unavailable');
      const deliveredBodies: Array<Promise<{ readonly bytes: number; readonly pathname: string }>> =
        [];
      page.on('response', (response) => {
        const pathname = new URL(response.url()).pathname;
        if (response.status() !== 200 || !pathname.endsWith('.m4s')) return;
        deliveredBodies.push(
          response.body().then((body) => ({ bytes: body.byteLength, pathname })),
        );
      });
      const apiBefore = await containerNetworkSnapshot('api');
      const authorizeBefore = await authorizeCounts();
      // K4: the `ended` wait races a refused heartbeat, so a repeat of the O05 403 fails here
      // with its error code and requestId instead of a silent timeout (see heartbeatDenialFailure).
      const endedHeartbeat = Promise.race([
        page.waitForResponse((response) => {
          const heartbeat = heartbeatPayload(response, fourthLessonID);
          return heartbeat?.reason === 'ended';
        }),
        heartbeatDenialFailure(page),
      ]);
      const session = await openAndPlay(page, fourthLessonID);
      const startPosition = await videoTime(page);
      expect(
        startPosition,
        'R23 must start the measured 100-second lesson at or before 2 seconds',
      ).toBeLessThanOrEqual(2);
      // The frontend baseline follows the page load on purpose, unlike the api's: taken before
      // `page.goto`, its delta is the weight of the SSR page, which is constant, so the bound would
      // turn red on a shorter asset or a heavier bundle with no media passing through SSR at all.
      const frontendBefore = await containerNetworkSnapshot('frontend');
      await page.locator('video').evaluate((node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        node.playbackRate = 16;
      });
      await expect.poll(() => videoEnded(page), { timeout: 60_000 }).toBe(true);
      await endedHeartbeat;
      const terminal = await page.locator('video').evaluate((node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        return { currentTime: node.currentTime, duration: node.duration, ended: node.ended };
      });
      await page.waitForTimeout(250);
      const delivered = await Promise.all(deliveredBodies);
      const apiAfter = await containerNetworkSnapshot('api');
      const authorizeAfter = await authorizeCounts();
      const frontendAfter = await containerNetworkSnapshot('frontend');
      const apiDelta = apiAfter.txBytes - apiBefore.txBytes;
      const frontendDelta = frontendAfter.txBytes - frontendBefore.txBytes;
      const successfulAuthorizeDelta = authorizeAfter.successful - authorizeBefore.successful;
      const rejectedAuthorizeDelta = authorizeAfter.rejected - authorizeBefore.rejected;
      const deliveredBytes = delivered.reduce(
        (total, segment) => total + BigInt(segment.bytes),
        0n,
      );
      const renditionSegments = new Map<string, Set<string>>();
      for (const segment of delivered) {
        const relative = segment.pathname.split(`/hls/${fourthLessonVideoID}/`)[1];
        const [rendition] = relative?.split('/') ?? [];
        if (rendition === undefined || rendition === '')
          throw new Error(`R23 could not identify rendition for ${segment.pathname}`);
        const paths = renditionSegments.get(rendition) ?? new Set<string>();
        paths.add(segment.pathname);
        renditionSegments.set(rendition, paths);
      }
      expect(renditionSegments.size).toBeGreaterThan(0);
      const masterURL = new URL(session.mediaPath, page.url());
      const coverage: string[] = [];
      for (const [rendition, paths] of renditionSegments) {
        const playlist = await page.request.get(
          new URL(`${rendition}/index.m3u8`, masterURL).toString(),
        );
        expect(playlist.status()).toBe(200);
        const playlistSegments = (await playlist.text())
          .split(/\r?\n/)
          .map((line) => line.trim())
          .filter((line) => line !== '' && !line.startsWith('#') && line.endsWith('.m4s'));
        expect(playlistSegments.length).toBeGreaterThan(0);
        coverage.push(`${rendition}=${paths.size} of ${playlistSegments.length}`);
      }
      console.log(
        `R23 delivered segments: ${coverage.join(', ')}; bodies=${delivered.length}; bytes=${deliveredBytes}`,
      );
      console.log(
        `R23 positions: start=${startPosition}; currentTime=${terminal.currentTime}; duration=${terminal.duration}; ended=${terminal.ended}; playbackRate=16`,
      );
      console.log(`R23 raw tx bytes: api delta=${apiDelta}; frontend delta=${frontendDelta}`);
      console.log(
        `R23 authorize counters: before observed=${authorizeBefore.observed} success=${authorizeBefore.successful} rejected=${authorizeBefore.rejected}; after observed=${authorizeAfter.observed} success=${authorizeAfter.successful} rejected=${authorizeAfter.rejected}; deltas success=${successfulAuthorizeDelta} rejected=${rejectedAuthorizeDelta}`,
      );
      expect(delivered.length).toBeGreaterThan(0);
      expect(deliveredBytes).toBeGreaterThan(0n);
      expect(terminal.ended).toBe(true);
      expect(terminal.duration).toBeGreaterThan(0);
      expect(terminal.currentTime).toBeGreaterThanOrEqual(terminal.duration - 0.25);
      expect(apiAfter.id, 'R23 api container changed during the measurement').toBe(apiBefore.id);
      expect(frontendAfter.id, 'R23 frontend container changed during the measurement').toBe(
        frontendBefore.id,
      );
      expect(apiDelta, 'R23 api tx delta must prove the measurement occurred').toBeGreaterThan(0n);
      // Asymmetric on purpose: the left side is interface bytes, headers and TCP included; the right
      // side is 200 `.m4s` bodies only, without headers or init segments. Both lean towards a stricter
      // bound, so "aligning" them - adding headers or init to the denominator - weakens the proof silently.
      expect(apiDelta * 10n).toBeLessThan(deliveredBytes);
      expect(frontendDelta * 10n).toBeLessThan(deliveredBytes);
    });

    test('B36 unavailable Redis returns no bytes from a previously served segment and always recovers', async ({
      page,
    }) => {
      const served = page.waitForResponse(
        (response) =>
          response.status() === 200 &&
          isMediaRequest(response.request()) &&
          new URL(response.url()).pathname.endsWith('.m4s'),
      );
      await openAndPlay(page, crossCourseLessonID, crossCourseNamespace);
      const segmentPath = new URL((await served).url()).pathname;
      const anchor = await page.request.get(segmentPath);
      expect(anchor.status()).toBe(200);
      const mediaBytes = await anchor.body();
      const windowSize = 64;
      expect(mediaBytes.byteLength).toBeGreaterThanOrEqual(windowSize * 2);
      const middlePosition = Math.floor((mediaBytes.byteLength - windowSize) / 2);
      const referenceWindows = [
        { name: 'first', position: 0, bytes: mediaBytes.subarray(0, windowSize) },
        {
          name: 'middle',
          position: middlePosition,
          bytes: mediaBytes.subarray(middlePosition, middlePosition + windowSize),
        },
      ];
      for (const window of referenceWindows) {
        expect(window.bytes.byteLength).toBe(windowSize);
        expect(
          mediaBytes.indexOf(window.bytes),
          `B36 ${window.name} reference window must occur at its source position`,
        ).toBe(window.position);
      }

      const pauseWritten = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, crossCourseLessonID);
        return heartbeat?.reason === 'pause';
      });
      await page.locator('video').evaluate((node) => {
        if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
        node.pause();
      });
      await pauseWritten;

      let recoveredStatus: number | undefined;
      try {
        await composeOutput(['stop', 'redis'], 60_000);
        const unavailable = await page.request.get(segmentPath);
        expect(unavailable.status()).toBe(503);
        const unavailableBytes = await unavailable.body();
        console.log(
          `B36 bodies: healthy content-type=${anchor.headers()['content-type'] ?? 'missing'} bytes=${mediaBytes.byteLength}; unavailable content-type=${unavailable.headers()['content-type'] ?? 'missing'} bytes=${unavailableBytes.byteLength}`,
        );
        for (const window of referenceWindows) {
          expect(
            unavailableBytes.indexOf(window.bytes),
            `B36 503 body must not contain the ${window.name} 64-byte media window`,
          ).toBe(-1);
        }
      } finally {
        await composeOutput(['start', 'redis'], 60_000);
        await expect
          .poll(
            async () => {
              const recovered = await page.request.get(segmentPath);
              recoveredStatus = recovered.status();
              return recoveredStatus;
            },
            { timeout: 30_000, intervals: [250] },
          )
          .not.toBe(503);
        console.log(`B36 Redis recovery: same segment changed from 503 to ${recoveredStatus}`);
      }
    });

    test('B28 a ready asset with no files returns edge 404 and a named player error', async ({
      page,
    }) => {
      if (crossCourseVideoID === undefined)
        throw new Error('The destructive B28 asset is unavailable');
      if (
        !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(crossCourseVideoID)
      )
        throw new Error(`B28 refuses a non-canonical video id: ${crossCourseVideoID}`);
      const assetPath = `/media/hls/${crossCourseVideoID}`;
      await composeOutput([
        'exec',
        '-T',
        'media-worker',
        'sh',
        '-eu',
        '-c',
        'test -d "$1"; test -f "$1/master.m3u8"; [ -n "$(ls -A "$1")" ]; rm -rf -- "$1"/* "$1"/.[!.]* "$1"/..?*; test -d "$1"; [ -z "$(ls -A "$1")" ]',
        'b28-remove-ready-asset',
        assetPath,
      ]);

      await page.goto(`/learn/${crossCourseNamespace}/${crossCourseLessonID}`);
      const minted = page.waitForResponse(
        (response) =>
          new URL(response.url()).pathname === '/api/v1/playback/sessions' &&
          response.request().method() === 'POST',
      );
      const missingMaster = page.waitForResponse((response) =>
        new URL(response.url()).pathname.endsWith(`/hls/${crossCourseVideoID}/master.m3u8`),
      );
      await page.getByRole('button', { name: 'Відтворити' }).click();
      const mintedResponse = await minted;
      expect(mintedResponse.status()).toBe(201);
      const session = playbackSession(await mintedResponse.json());
      expect(session.mediaPath).toContain(`/hls/${crossCourseVideoID}/master.m3u8`);
      const edgeResponse = await missingMaster;
      expect(edgeResponse.status()).toBe(404);
      console.log(`B28 ready asset ${crossCourseVideoID}: edge status=${edgeResponse.status()}`);
      await expect(page.getByText('Відео зараз недоступне.')).toBeVisible({ timeout: 60_000 });
    });

    test('R04 archive preserves an enrolled member started view (En09 ownership)', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined || course === undefined)
        throw new Error('Playback course fixture is unavailable');
      // Remember a segment the edge served while the course was public, then ask for that exact
      // path again after the archive. En09 (ownership, not rental): archiving must NOT revoke an
      // enrolled member's access, so the same bytes are still served.
      let servedSegment: string | undefined;
      page.on('response', (response) => {
        const path = new URL(response.url()).pathname;
        if (response.status() === 200 && path.endsWith('.m4s')) servedSegment ??= path;
      });
      await openAndPlay(page);
      await page.waitForTimeout(31_000);
      if (servedSegment === undefined) throw new Error('No segment was served before the archive');
      const context = await adminContext(browser, baseURL);
      try {
        const archived = await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
          data: { status: 'archived', expectedUpdatedAt: course.updatedAt },
          headers: { Origin: new URL(baseURL).origin },
        });
        expect(archived.status()).toBe(200);
        course = courseToken(await archived.json());
      } finally {
        await context.close();
      }
      await setPosition(page, 120);
      // The same path the edge served a moment ago is still served: an enrolled member keeps
      // access through an archive (En09). The player shows no "course unavailable" stop. Cutting a
      // started view is an admin REVOKE, not an archive — covered once the revoke endpoint lands (L6).
      const stillServed = await page.request.get(servedSegment);
      expect(stillServed.status()).toBe(200);
      await expect(
        page.getByRole('heading', {
          name: 'Перегляд зупинено. Курс більше недоступний.',
        }),
      ).toBeHidden();
    });
  });

  test.describe('as the course instructor', () => {
    test.use({ storageState: authStatePath('instructor') });

    test('R34 a non-admin owner plays the archived course while another non-admin is refused', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined || course === undefined)
        throw new Error('Playback course fixture is unavailable');
      const origin = new URL(baseURL).origin;
      const context = await adminContext(browser, baseURL);
      try {
        const reassigned = await context.request.put(`/api/v1/admin/courses/${course.id}`, {
          data: {
            expectedUpdatedAt: course.updatedAt,
            course: { ...primaryCourseDraft, instructorId: instructorID },
            modules: primaryCourseModules,
          },
          headers: { Origin: origin },
        });
        expect(reassigned.status()).toBe(200);
        course = courseToken(await reassigned.json());
      } finally {
        await context.close();
      }

      const segment = page.waitForResponse(
        (response) =>
          isMediaRequest(response.request()) &&
          /\.m4s$/.test(new URL(response.url()).pathname) &&
          response.status() === 200,
      );
      await openAndPlay(page);
      expect((await segment).status()).toBe(200);

      // The refused viewer is the non-member fixture (outsider@courseforge.local): the seeded
      // student is now an enrolled member and would keep access (En09), so the "another non-admin
      // is refused" guarantee needs a genuine non-member. A fresh context logs in inline.
      const otherContext = await browser.newContext({ baseURL });
      try {
        const loggedIn = await otherContext.request.post('/api/v1/auth/login', {
          data: { email: 'outsider@courseforge.local', password: 'courseforge-student-demo' },
          headers: { Origin: origin },
        });
        expect(loggedIn.status(), await loggedIn.text()).toBe(200);
        const refused = await otherContext.request.post('/api/v1/playback/sessions', {
          data: { lessonId: lessonID },
          headers: { Origin: origin },
        });
        expect([403, 404]).toContain(refused.status());
      } finally {
        await otherContext.close();
      }
    });
  });

  test.describe('as an admin', () => {
    test.use({ storageState: authStatePath('admin') });

    test('P11 plays the archived course through the same route', async ({ page }) => {
      const segment = page.waitForResponse(
        (response) =>
          isMediaRequest(response.request()) &&
          /\.m4s$/.test(new URL(response.url()).pathname) &&
          response.status() === 200,
      );
      await openAndPlay(page);
      expect((await segment).status()).toBe(200);
    });
  });
});
