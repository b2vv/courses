import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  expect,
  test,
  type APIRequestContext,
  type Browser,
  type Locator,
  type Page,
} from '@playwright/test';
import { promisify } from 'node:util';
import axe from 'axe-core';
import type {
  AdminCourseTree,
  ErrorResponse,
  ErrorResponseError,
  MediaUploadTicket,
} from '../src/adapters/api/generated/models';
import { authStatePath, trackPageRequests } from './auth';
import { contractField } from './contract';

const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const runNamespace = `e2e-media-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const moduleID = randomUUID();
const lessons = {
  main: { id: randomUUID(), title: `Основний урок ${runNamespace}`, index: 1 },
  twoRung: { id: randomUUID(), title: `Дві сходинки ${runNamespace}`, index: 2 },
  subRung: { id: randomUUID(), title: `Мала роздільність ${runNamespace}`, index: 3 },
  restart: { id: randomUUID(), title: `Перезапуск ${runNamespace}`, index: 4 },
  videoOnly: { id: randomUUID(), title: `Відео без звуку ${runNamespace}`, index: 5 },
} as const;

type CourseToken = { readonly id: string; readonly updatedAt: string };

let course: CourseToken | undefined;
let mainVideoId: string | undefined;
let finishRequestTracking = () => Promise.resolve();

async function compose(...args: string[]) {
  return execute('docker', ['compose', ...args], {
    cwd: repositoryRoot,
    timeout: 240_000,
    maxBuffer: 4 * 1024 * 1024,
  });
}

async function worker(...args: string[]): Promise<string> {
  const result = await compose('exec', '-T', 'media-worker', ...args);
  return result.stdout;
}

function courseToken(payload: unknown): CourseToken {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('Admin course response must be an object');
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('Admin course response must carry id and updatedAt');
  return { id, updatedAt };
}

function videoID(payload: unknown): string {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('Media ticket response must be an object');
  const id = contractField<MediaUploadTicket>(payload, 'videoId');
  if (typeof id !== 'string') throw new Error('Media ticket response must carry videoId');
  return id;
}

function errorCode(payload: unknown): string | undefined {
  if (typeof payload !== 'object' || payload === null) return undefined;
  const error = contractField<ErrorResponse>(payload, 'error');
  if (typeof error !== 'object' || error === null) return undefined;
  const code = contractField<ErrorResponseError>(error, 'code');
  return typeof code === 'string' ? code : undefined;
}

function fixture(name: '640x360' | '1280x720' | '320x240' | '1280x720-5min' | '640x360-silent') {
  return path.join(fixtureDirectory, `e2e-${name}.mp4`);
}

const audioFixture = path.join(fixtureDirectory, 'e2e-tone.m4a');

function trackID(payload: unknown): string {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('Audio ticket response must be an object');
  const id: unknown = Reflect.get(payload, 'trackId');
  if (typeof id !== 'string') throw new Error('Audio ticket response must carry trackId');
  return id;
}

function audioSection(field: Locator): Locator {
  return field.getByRole('region', { name: 'Аудіодоріжки' });
}

/** Picks a language and a name, then the file; returns the minted track id. */
async function uploadAudio(
  page: Page,
  section: Locator,
  videoId: string,
  language: string,
  label: string,
): Promise<string> {
  await section.getByRole('button', { name: /Мова доріжки/ }).click();
  await page.getByRole('option', { name: language }).click();
  await section.getByRole('textbox', { name: 'Назва доріжки' }).fill(label);
  const minted = page.waitForResponse(
    (response) =>
      new URL(response.url()).pathname === `/api/v1/admin/videos/${videoId}/audio-tracks` &&
      response.request().method() === 'POST',
  );
  await section.locator('input[type="file"]').setInputFiles(audioFixture);
  const response = await minted;
  expect(response.status()).toBe(201);
  return trackID(await response.json());
}

async function audioRenditions(videoId: string): Promise<string[]> {
  return playlistLines(await readPlaylist(videoId, 'master.m3u8')).filter((line) =>
    line.startsWith('#EXT-X-MEDIA:TYPE=AUDIO'),
  );
}

declare global {
  interface Window {
    axe: typeof axe;
  }
}

async function axeViolations(page: Page, section: Locator): Promise<string[]> {
  await page.addScriptTag({ content: axe.source });
  return section.evaluate(async (element) => {
    const results = await window.axe.run(element, {
      runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
    });
    return results.violations.flatMap((item) =>
      item.nodes.map((node) => `${item.id} ${node.target.join(' ')}`),
    );
  });
}

function lessonField(page: Page, index: number): Locator {
  return page.getByRole('group', { name: `Урок ${index}` });
}

async function installProgressRecorder(field: Locator) {
  await field.evaluate((element) => {
    const values: number[] = [];
    Reflect.set(element, '__mediaProgressValues', values);
    const record = () => {
      for (const progress of element.querySelectorAll('[role="progressbar"][aria-valuenow]')) {
        const value = Number(progress.getAttribute('aria-valuenow'));
        if (Number.isFinite(value)) values.push(value);
      }
    };
    new MutationObserver(record).observe(element, {
      attributes: true,
      attributeFilter: ['aria-valuenow'],
      childList: true,
      subtree: true,
    });
    record();
  });
}

async function recordedProgress(field: Locator): Promise<number[]> {
  return field.evaluate((element) => {
    const value: unknown = Reflect.get(element, '__mediaProgressValues');
    return Array.isArray(value)
      ? value.filter((item): item is number => typeof item === 'number')
      : [];
  });
}

async function upload(page: Page, lesson: (typeof lessons)[keyof typeof lessons], file: string) {
  const field = lessonField(page, lesson.index);
  await expect(field.getByRole('button', { name: 'Завантажити відео' })).toBeVisible();
  const minted = page.waitForResponse((response) => {
    return (
      new URL(response.url()).pathname === `/api/v1/admin/lessons/${lesson.id}/video` &&
      response.request().method() === 'POST'
    );
  });
  await field.locator('input[type="file"]:not(.admin-audio-field input)').setInputFiles(file);
  const response = await minted;
  expect(response.status()).toBe(201);
  return { field, videoId: videoID(await response.json()) };
}

async function readPlaylist(videoId: string, relative: string): Promise<string> {
  return worker('cat', `/media/hls/${videoId}/${relative}`);
}

function playlistLines(playlist: string): string[] {
  return playlist
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);
}

function bandwidth(stream: string): string {
  const match = /(?:^|[:,])BANDWIDTH=(\d+)(?:,|$)/.exec(stream);
  if (match?.[1] === undefined) throw new Error('EXT-X-STREAM-INF must carry BANDWIDTH');
  return match[1];
}

async function assertPlaylistBundle(
  videoId: string,
  variants: readonly string[],
): Promise<{ readonly bandwidth360: string }> {
  const master = await readPlaylist(videoId, 'master.m3u8');
  const audio = await readPlaylist(videoId, 'audio/index.m3u8');
  const videoPlaylists = await Promise.all(
    variants.map((variant) => readPlaylist(videoId, `${variant}/index.m3u8`)),
  );
  const masterLines = playlistLines(master);
  const streams = masterLines.filter((line) => line.startsWith('#EXT-X-STREAM-INF:'));
  const audioMedia = masterLines.filter((line) => line.startsWith('#EXT-X-MEDIA:TYPE=AUDIO'));
  expect(streams).toHaveLength(variants.length);
  expect(audioMedia).toHaveLength(1);
  const group = /GROUP-ID="([^"]+)"/.exec(audioMedia[0] ?? '')?.[1];
  expect(group).toBeTruthy();
  for (const variant of variants) expect(masterLines).toContain(`${variant}/index.m3u8`);
  for (const stream of streams) {
    expect(stream).toContain('CODECS=');
    expect(stream).toContain('RESOLUTION=');
    expect(stream).toMatch(/(?:^|[:,])BANDWIDTH=\d+(?:,|$)/);
    expect(stream).toContain(`AUDIO="${group}"`);
  }
  const allPlaylists = [master, audio, ...videoPlaylists].join('\n');
  expect(allPlaylists).not.toContain('AVERAGE-BANDWIDTH');
  expect(allPlaylists).not.toContain('FRAME-RATE');
  expect(master).not.toContain('#EXT-X-INDEPENDENT-SEGMENTS');
  expect(audio).not.toContain('#EXT-X-INDEPENDENT-SEGMENTS');
  for (const playlist of videoPlaylists) expect(playlist).toContain('#EXT-X-INDEPENDENT-SEGMENTS');

  const rootEntries = playlistLines(await worker('ls', '-1', `/media/hls/${videoId}`));
  expect(rootEntries.filter((entry) => entry === 'audio')).toHaveLength(1);
  for (const variant of variants) {
    const codecTypes = playlistLines(
      await worker(
        'ffprobe',
        '-v',
        'error',
        '-show_entries',
        'stream=codec_type',
        '-of',
        'csv=p=0',
        `/media/hls/${videoId}/${variant}/index.m3u8`,
      ),
    );
    // The row's subject is that audio lives in the shared group rather than muxed into a
    // video variant; how many times ffprobe lists the video stream is not the contract.
    expect(codecTypes).not.toContain('audio');
    expect(codecTypes).toContain('video');
  }
  return { bandwidth360: bandwidth(streams[0] ?? '') };
}

async function expectPrivateMedia(
  adminRequest: APIRequestContext,
  browser: Browser,
  baseURL: string,
  publishedVideoId: string,
) {
  const anonymous = await browser.newContext({ baseURL });
  const paths = ['/media', '/media/', `/media/hls/${publishedVideoId}/master.m3u8`];
  try {
    for (const mediaPath of paths) {
      expect((await anonymous.request.get(mediaPath)).status()).toBe(403);
      expect((await adminRequest.get(mediaPath)).status()).toBe(403);
    }
  } finally {
    await anonymous.close();
  }
}

test.describe('the first real video through the whole chain', () => {
  test.describe.configure({ mode: 'serial', timeout: 360_000 });
  test.use({ storageState: authStatePath('admin') });

  test.beforeAll(async ({ browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const context = await browser.newContext({
      baseURL,
      storageState: authStatePath('admin'),
    });
    const draft = {
      slug: runNamespace,
      title: `M16 відео ${runNamespace}`,
      summary: 'Курс для наскрізної перевірки завантаження та транскодування.',
      description: 'Одноразова браузерна фікстура M16 з чотирма збереженими уроками.',
      instructorId: adminID,
    };
    try {
      const createdResponse = await context.request.post('/api/v1/admin/courses', {
        data: draft,
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(createdResponse.status()).toBe(201);
      const created = courseToken(await createdResponse.json());
      const savedResponse = await context.request.put(`/api/v1/admin/courses/${created.id}`, {
        data: {
          expectedUpdatedAt: created.updatedAt,
          course: draft,
          modules: [
            {
              id: moduleID,
              title: `Модуль ${runNamespace}`,
              lessons: Object.values(lessons).map((lesson) => ({
                id: lesson.id,
                title: lesson.title,
              })),
            },
          ],
        },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(savedResponse.status()).toBe(200);
      course = courseToken(await savedResponse.json());
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

  test.afterAll(async ({ browser, baseURL }) => {
    if (course === undefined || baseURL === undefined) return;
    const context = await browser.newContext({
      baseURL,
      storageState: authStatePath('admin'),
    });
    try {
      const response = await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
        data: { expectedUpdatedAt: course.updatedAt },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(response.status()).toBe(204);
    } finally {
      await context.close();
    }
    await compose('restart', 'media-worker');
    await compose('up', '-d', '--wait', 'media-worker');
  });

  test('P04 P08 B05 B25 and R13 upload a one-rung video and keep its bytes private', async ({
    page,
    browser,
    baseURL,
  }) => {
    if (course === undefined || baseURL === undefined)
      throw new Error('Media course was not created');
    await page.goto(`/admin/courses/${course.id}`);
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const field = lessonField(page, lessons.main.index);
    await installProgressRecorder(field);
    const uploaded = await upload(page, lessons.main, fixture('640x360'));
    mainVideoId = uploaded.videoId;
    await expect(uploaded.field.locator('[data-video-state]')).toHaveAttribute(
      'data-video-state',
      'ready',
      { timeout: 120_000 },
    );
    expect((await recordedProgress(field)).some((value) => value > 0)).toBe(true);

    const master = await readPlaylist(uploaded.videoId, 'master.m3u8');
    const streams = playlistLines(master).filter((line) => line.startsWith('#EXT-X-STREAM-INF:'));
    expect(streams).toHaveLength(1);
    expect(streams[0]).toContain('RESOLUTION=640x360');
    expect(master).toContain('360p/index.m3u8');
    await worker('test', '!', '-e', `/media/hls/${uploaded.videoId}/720p`);
    await worker('test', '!', '-e', `/media/hls/${uploaded.videoId}/1080p`);

    for (const directory of ['360p', 'audio'] as const) {
      const files = playlistLines(
        await worker('ls', '-1', `/media/hls/${uploaded.videoId}/${directory}`),
      );
      expect(files).toContain('index.m3u8');
      expect(files.filter((file) => /^init_.*\.mp4$/.test(file))).toHaveLength(1);
      expect(files.some((file) => /^segment_\d+\.m4s$/.test(file))).toBe(true);
    }
    await worker('test', '-f', `/media/thumbnails/${uploaded.videoId}.jpg`);

    const metrics = await worker('curl', '-fsS', 'http://127.0.0.1:8081/metrics');
    const readyCount =
      /^courses_transcoding_job_duration_seconds_count\{outcome="ready"\} ([\d.]+)$/m.exec(
        metrics,
      )?.[1];
    expect(Number(readyCount)).toBeGreaterThanOrEqual(1);
    await expectPrivateMedia(page.request, browser, baseURL, uploaded.videoId);

    await page.reload();
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    await expect(
      lessonField(page, lessons.main.index).locator('[data-video-state]'),
    ).toHaveAttribute('data-video-state', 'ready');
  });

  test('S5-L2b adds a second audio track, makes it the default and deletes the original', async ({
    page,
  }) => {
    if (course === undefined || mainVideoId === undefined)
      throw new Error('The ready main video was not uploaded');
    const videoId = mainVideoId;
    await page.goto(`/admin/courses/${course.id}`);
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const field = lessonField(page, lessons.main.index);
    const section = audioSection(field);
    // The source carried a sound stream, so the worker created one embedded Ready track.
    await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(1);
    expect(await audioRenditions(videoId)).toHaveLength(1);
    expect(await axeViolations(page, section)).toEqual([]);

    await uploadAudio(page, section, videoId, 'Англійська', 'English');
    await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(2, {
      timeout: 120_000,
    });
    await expect.poll(() => audioRenditions(videoId)).toHaveLength(2);

    await section
      .getByRole('button', { name: 'Зробити основною доріжку English (Англійська)' })
      .click();
    const english = section.locator('li').filter({ hasText: 'English' });
    await expect(english.getByText('За замовчуванням')).toBeVisible();
    await expect
      .poll(async () =>
        (await audioRenditions(videoId)).find((line) => line.includes('DEFAULT=YES')),
      )
      .toContain('NAME="English"');

    const original = section.locator('li').filter({ hasNotText: 'English' });
    await original.getByRole('button', { name: /^Видалити доріжку/ }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Видалити доріжку' }).click();
    await expect(section.locator('li')).toHaveCount(1);
    await expect.poll(() => audioRenditions(videoId)).toHaveLength(1);
    // The one Ready track left on a Ready asset is kept (D-L2-4).
    await expect(english.getByRole('button', { name: /^Видалити доріжку/ })).toHaveCount(0);
    await expect(field.locator('[data-video-state]')).toHaveAttribute('data-video-state', 'ready');
  });

  test('S5-L2b a video-only upload waits for audio and turns ready without navigation', async ({
    page,
  }) => {
    if (course === undefined) throw new Error('Media course was not created');
    await page.goto(`/admin/courses/${course.id}`);
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const uploaded = await upload(page, lessons.videoOnly, fixture('640x360-silent'));
    const state = uploaded.field.locator('[data-video-state]');
    await expect(state).toHaveAttribute('data-video-state', 'awaiting_audio', {
      timeout: 120_000,
    });
    const section = audioSection(uploaded.field);
    await expect(section.getByText(/Перша готова аудіодоріжка/)).toBeVisible();

    // Keyboard path: the language select opens and picks from the keyboard alone; with no
    // track yet, the first offered language is Ukrainian.
    await section.getByRole('button', { name: /Мова доріжки/ }).focus();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('option', { name: 'Українська' })).toBeFocused();
    await page.keyboard.press('Enter');
    await expect(section.getByRole('button', { name: /Мова доріжки/ })).toContainText('Українська');
    await section.getByRole('textbox', { name: 'Назва доріжки' }).fill('Українська');
    const minted = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname ===
          `/api/v1/admin/videos/${uploaded.videoId}/audio-tracks` &&
        response.request().method() === 'POST',
    );
    await section.locator('input[type="file"]').setInputFiles(audioFixture);
    expect((await minted).status()).toBe(201);

    // The worker promotes the asset with its first Ready track; the editor sees it unaided.
    await expect(state).toHaveAttribute('data-video-state', 'ready', { timeout: 120_000 });
    await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(1);
    expect(await audioRenditions(uploaded.videoId)).toHaveLength(1);
  });

  test('B26 B27 B30 pins both ladder shapes and their one shared audio rendition', async ({
    page,
  }) => {
    if (course === undefined) throw new Error('Media course was not created');
    await page.goto(`/admin/courses/${course.id}`);
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const twoRung = await upload(page, lessons.twoRung, fixture('1280x720'));
    await expect(twoRung.field.locator('[data-video-state]')).toHaveAttribute(
      'data-video-state',
      'ready',
      { timeout: 120_000 },
    );
    const twoRungEvidence = await assertPlaylistBundle(twoRung.videoId, ['360p', '720p']);

    const subRung = await upload(page, lessons.subRung, fixture('320x240'));
    await expect(subRung.field.locator('[data-video-state]')).toHaveAttribute(
      'data-video-state',
      'ready',
      { timeout: 120_000 },
    );
    const subRungEvidence = await assertPlaylistBundle(subRung.videoId, ['360p']);
    // Declared peak for the 360p rung = its maxrate (B-19: 600k) + the shared 128k audio +
    // container overhead, so it is pinned by the ladder constant, not the source. Both the
    // two-rung and the sub-360p upload declare the same 360p peak.
    expect(twoRungEvidence.bandwidth360).toBe('800800');
    expect(subRungEvidence.bandwidth360).toBe(twoRungEvidence.bandwidth360);
    expect(await readPlaylist(subRung.videoId, 'master.m3u8')).toContain('RESOLUTION=320x240');
    const resolution = await worker(
      'ffprobe',
      '-v',
      'error',
      '-select_streams',
      'v:0',
      '-show_entries',
      'stream=width,height',
      '-of',
      'csv=p=0:s=x',
      `/media/hls/${subRung.videoId}/360p/index.m3u8`,
    );
    // ffprobe repeats a stream entry per playlist segment, so the contract is the set of
    // reported resolutions, not the line count: the sub-360p source must never be enlarged.
    expect([...new Set(playlistLines(resolution))]).toEqual(['320x240']);
  });

  test('R26 refuses a student direct API mint without relying on the client guard', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const context = await browser.newContext({
      baseURL,
      storageState: authStatePath('student'),
    });
    const page = await context.newPage();
    const finishStudentTracking = trackPageRequests(page);
    try {
      await page.goto('/courses');
      await expect(page.getByRole('button', { name: 'Завантажити відео' })).toHaveCount(0);
      const response = await page.request.post(
        `/api/v1/admin/lessons/${lessons.restart.id}/video`,
        {
          data: { declaredBytes: 12, originalName: 'refused.mp4' },
          headers: { Origin: new URL(baseURL).origin },
        },
      );
      expect(response.status()).toBe(403);
      expect(errorCode(await response.json())).toBe('forbidden');
    } finally {
      await finishStudentTracking();
      await context.close();
    }
  });

  test('P05 R35 observes processing before restart and reaches ready without navigation', async ({
    page,
  }) => {
    if (course === undefined) throw new Error('Media course was not created');
    await page.goto(`/admin/courses/${course.id}`);
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const uploaded = await upload(page, lessons.restart, fixture('1280x720-5min'));
    const state = uploaded.field.locator('[data-video-state]');
    await expect(state).toHaveAttribute('data-video-state', 'processing', { timeout: 60_000 });
    await compose('restart', 'media-worker');
    await compose('up', '-d', '--wait', 'media-worker');
    await expect(state).toHaveAttribute('data-video-state', 'ready', { timeout: 180_000 });
  });
});
