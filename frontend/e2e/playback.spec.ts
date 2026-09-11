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
import { authStatePath, trackPageRequests } from './auth';

const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const namespace = `e2e-playback-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const crossCourseNamespace = `${namespace}-cross-course`;
const adminID = '40000000-0000-4000-8000-000000000003';
const moduleID = randomUUID();
const lessonID = randomUUID();
const secondLessonID = randomUUID();
const secondModuleID = randomUUID();
const thirdLessonID = randomUUID();
const fourthLessonID = randomUUID();
const crossCourseModuleID = randomUUID();
const crossCourseLessonID = randomUUID();

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
let finishRequestTracking = () => Promise.resolve();

function object(payload: unknown, name: string): object {
  if (typeof payload !== 'object' || payload === null) throw new Error(`${name} must be an object`);
  return payload;
}

function courseToken(payload: unknown): CourseToken {
  const record = object(payload, 'Admin course response');
  const id = Reflect.get(record, 'id');
  const updatedAt = Reflect.get(record, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('Admin course response must carry id and updatedAt');
  return { id, updatedAt };
}

function playbackSession(payload: unknown): PlaybackSession {
  const record = object(payload, 'Playback session');
  const playbackId = Reflect.get(record, 'playbackId');
  const mediaPath = Reflect.get(record, 'mediaPath');
  const clientSeq = Reflect.get(record, 'clientSeq');
  const heartbeatSeconds = Reflect.get(record, 'heartbeatSeconds');
  const ownershipTtlSeconds = Reflect.get(record, 'ownershipTtlSeconds');
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
  const id = Reflect.get(object(payload, 'Media ticket'), 'videoId');
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
  const position = Reflect.get(body, 'positionSeconds');
  const clientSeq = Reflect.get(body, 'clientSeq');
  const reason = Reflect.get(body, 'reason');
  if (
    Reflect.get(body, 'lessonId') !== expectedLessonID ||
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

function heartbeatPosition(response: Response, expectedLessonID: string): number | null {
  return heartbeatPayload(response, expectedLessonID)?.positionSeconds ?? null;
}

function learningLessonPosition(payload: unknown, expectedLessonID: string): number {
  const modules = Reflect.get(object(payload, 'Learning course'), 'modules');
  if (!Array.isArray(modules)) throw new Error('Learning course modules must be an array');
  for (const module of modules) {
    if (typeof module !== 'object' || module === null) continue;
    const lessons = Reflect.get(module, 'lessons');
    if (!Array.isArray(lessons)) continue;
    for (const lesson of lessons) {
      if (typeof lesson !== 'object' || lesson === null) continue;
      if (Reflect.get(lesson, 'id') !== expectedLessonID) continue;
      const position = Reflect.get(lesson, 'position');
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
  await lesson.locator('input[type="file"]').setInputFiles(file);
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

function isMediaRequest(request: Request): boolean {
  return new URL(request.url()).pathname.startsWith('/media/p/');
}

async function videoTime(page: Page): Promise<number> {
  return page.locator('video').evaluate((node) => {
    if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
    return node.currentTime;
  });
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
    const draft = {
      slug: namespace,
      title: `T7 playback ${namespace}`,
      summary: 'A real HLS playback fixture.',
      description: 'Run-scoped course for the T7 browser recovery table.',
      instructorId: adminID,
    };
    try {
      const created = await context.request.post('/api/v1/admin/courses', {
        data: draft,
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(created.status()).toBe(201);
      course = courseToken(await created.json());
      const saved = await context.request.put(`/api/v1/admin/courses/${course.id}`, {
        data: {
          expectedUpdatedAt: course.updatedAt,
          course: draft,
          modules: [
            {
              id: moduleID,
              title: `Module one ${namespace}`,
              lessons: [
                {
                  id: lessonID,
                  title: `First lesson ${namespace}`,
                  description: 'HLS ownership and recovery.',
                },
                {
                  id: secondLessonID,
                  title: `Second lesson ${namespace}`,
                  description: 'The next curriculum step.',
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
                  description: 'A later curriculum step.',
                },
                {
                  id: fourthLessonID,
                  title: `Fourth lesson ${namespace}`,
                  description: 'The final curriculum step.',
                },
              ],
            },
          ],
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
      await uploadReady(page, path.join(fixtureDirectory, 'e2e-1280x720-100s.mp4'), {
        lessonID: fourthLessonID,
        moduleNumber: 2,
        lessonNumber: 2,
      });
      const published = await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { status: 'published', expectedUpdatedAt: course.updatedAt },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(published.status()).toBe(200);
      course = courseToken(await published.json());

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
                    description: 'Proves ownership is not scoped to a course.',
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
      await uploadReady(page, path.join(fixtureDirectory, 'e2e-1280x720-100s.mp4'), {
        courseID: crossCourse.id,
        lessonID: crossCourseLessonID,
        moduleNumber: 1,
        lessonNumber: 1,
      });
      const crossCoursePublished = await context.request.post(
        `/api/v1/admin/courses/${crossCourse.id}/status`,
        {
          data: { status: 'published', expectedUpdatedAt: crossCourse.updatedAt },
          headers: { Origin: new URL(baseURL).origin },
        },
      );
      expect(crossCoursePublished.status()).toBe(200);
      crossCourse = courseToken(await crossCoursePublished.json());
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
      await expect(page.getByRole('button', { name: /Якість/ })).toBeVisible();
      await page.getByRole('button', { name: /Якість/ }).click();
      await page.getByRole('option', { name: '360p' }).click();
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
      expect(heartbeats).toBeGreaterThanOrEqual(4);
      expect(deniedSegments).toBe(0);
      expect(await videoTime(page)).toBeGreaterThan(25);
    });

    test('P04 restores the position after a reload', async ({ page }) => {
      await openAndPlay(page);
      const saved = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 40;
      });
      await setPosition(page, 40);
      await saved;

      await page.reload();
      await openAndPlay(page);
      const restored = await videoTime(page);
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

      const secondContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const secondPage = await secondContext.newPage();
        await openAndPlay(secondPage);
        const restored = await videoTime(secondPage);
        expect(restored).toBeGreaterThanOrEqual(39);
        expect(restored).toBeLessThanOrEqual(41);
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

      await page.reload();
      await expect(page.getByRole('slider', { name: 'Позиція відтворення' })).toHaveValue('0');
      await openAndPlay(page);
      expect(await videoTime(page)).toBeLessThanOrEqual(1);
    });

    test('B18 keeps the last recorded position across a network break', async ({ page }) => {
      await openAndPlay(page);
      const saved = page.waitForResponse((response) => {
        const heartbeat = heartbeatPayload(response, lessonID);
        return heartbeat?.reason === 'seek' && heartbeat.positionSeconds === 40;
      });
      await setPosition(page, 40);
      await saved;
      await page.getByRole('button', { name: 'Пауза' }).click();

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
        await openAndPlay(recoveredPage);
        const restored = await videoTime(recoveredPage);
        expect(restored).toBeGreaterThanOrEqual(39);
        expect(restored).toBeLessThanOrEqual(41);
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
      const writes = await page.locator('video').evaluate((video) => {
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
      const restoreWrite = writes.find(
        ({ assigned }, index) => index > 0 && Math.abs(assigned - preRemint) <= 1,
      );
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
      expect(restoreWrite?.assigned).toBeGreaterThanOrEqual(preRemint - 1);
      expect(restoreWrite?.assigned).toBeLessThanOrEqual(preRemint + 1);
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

    test('P13 B15 B32 takeover stops the first device and denies the same segment immediately', async ({
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
      await openAndPlay(page);
      const segmentPath = new URL((await served).url()).pathname;
      const beforeTakeover = await page.request.get(segmentPath);
      expect(beforeTakeover.status()).toBe(200);

      const secondContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('student'),
      });
      try {
        const secondPage = await secondContext.newPage();
        await openAndPlay(secondPage);
        const afterTakeover = await page.request.get(segmentPath);
        expect(afterTakeover.status()).toBe(403);
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

    test('R04 archive stops a started student view', async ({ page, browser, baseURL }) => {
      if (baseURL === undefined || course === undefined)
        throw new Error('Playback course fixture is unavailable');
      // Remember a segment the edge served while the course was public. Asking for that exact
      // path again after the archive is what proves the bytes are cut - stronger than waiting
      // for the player to happen to request one, and independent of what it does next.
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
      // The same path the edge served a moment ago now refuses. Waiting for the player to catch
      // a 403 itself no longer works and should not: a seek queues a heartbeat, that heartbeat
      // is refused first, and the player tears itself down before any segment goes out. Stopping
      // sooner is better behaviour - but it left this row watching for something that had
      // stopped happening, which is how a real guarantee quietly loses its proof.
      const refused = await page.request.get(servedSegment);
      expect(refused.status()).toBe(403);
      await expect(
        page.getByRole('heading', {
          name: 'Перегляд зупинено. Курс більше недоступний.',
        }),
      ).toBeVisible();
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
