import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';
import {
  expect,
  test,
  type APIRequestContext,
  type Locator,
  type Page,
  type Response,
} from '@playwright/test';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

// Multi-audio S7 (docs/multi-audio/plan-s7.md §4): a student picks the lesson's audio track in the
// player's gear menu, the choice is heard, and it is remembered for the next view.
const runNamespace = `e2e-audio-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const studentID = '40000000-0000-4000-8000-000000000001';
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
// 30 s, embedded 440 Hz; the uploaded track is 880 Hz of the same length (media.setup.ts).
const dualVideo = path.join(fixtureDirectory, 'e2e-640x360-30s.mp4');
const dubTrack = path.join(fixtureDirectory, 'e2e-tone-30s.m4a');
const singleVideo = path.join(fixtureDirectory, 'e2e-640x360.mp4');
const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));
const uploadedAudioPlaylist = /\/audio-[0-9a-f-]{36}\/index\.m3u8$/;
const anyAudioPlaylist = /\/audio(-[0-9a-f-]{36})?\/index\.m3u8$/;

// Grants keep ownership records without ON DELETE CASCADE, so the course goes through SQL, as in
// playback.spec.ts (module level: the heredoc terminator must stay in column zero).
const deleteCourseStatement = `exec psql -v ON_ERROR_STOP=1 -v "course_id=$1" -U "$POSTGRES_USER" -d "$POSTGRES_DB" <<'SQL'
DELETE FROM grant_audit WHERE grant_id IN (SELECT id FROM course_grants WHERE course_id = :'course_id');
DELETE FROM course_grants WHERE course_id = :'course_id';
DELETE FROM courses WHERE id = :'course_id';
SQL`;

type Course = { id: string; updatedAt: string; slug: string; dual: string; single: string };
let course: Course | undefined;

function token(payload: object): { id: string; updatedAt: string } {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string') {
    throw new Error('Admin response must carry id and updatedAt');
  }
  return { id, updatedAt };
}

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

async function seed(request: APIRequestContext, baseURL: string): Promise<Course> {
  const headers = origin(baseURL);
  const slug = runNamespace;
  const dual = randomUUID();
  const single = randomUUID();
  const draft = {
    slug,
    title: `Audio tracks ${runNamespace}`,
    summary: 'Audio track selection in the player.',
    description: 'A run-scoped course for multi-audio S7.',
    instructorId: adminID,
  };
  const created = await request.post('/api/v1/admin/courses', { data: draft, headers });
  expect(created.status(), await created.text()).toBe(201);
  const first = token((await created.json()) as object);
  // Registered before anything else can fail, so afterAll removes the course whatever happens.
  course = { ...first, slug, dual, single };
  const saved = await request.put(`/api/v1/admin/courses/${first.id}`, {
    data: {
      expectedUpdatedAt: first.updatedAt,
      course: draft,
      modules: [
        {
          id: randomUUID(),
          title: `Модуль ${runNamespace}`,
          lessons: [
            { id: dual, title: 'Two audio tracks' },
            { id: single, title: 'One audio track' },
          ],
        },
      ],
    },
    headers,
  });
  expect(saved.status(), await saved.text()).toBe(200);
  course = { ...course, ...token((await saved.json()) as object) };
  return course;
}

function lessonGroup(page: Page, number: number): Locator {
  return page
    .getByRole('group', { name: 'Модуль 1' })
    .getByRole('group', { name: `Урок ${number}` });
}

async function uploadVideo(page: Page, lesson: Locator, lessonID: string, file: string) {
  const minted = page.waitForResponse(
    (candidate) =>
      candidate.url().endsWith(`/api/v1/admin/lessons/${lessonID}/video`) &&
      candidate.request().method() === 'POST',
  );
  await lesson.locator('input[type="file"]:not(.admin-audio-field input)').setInputFiles(file);
  expect((await minted).status()).toBe(201);
  await expect(lesson.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
}

async function addDub(page: Page, lesson: Locator) {
  const section = lesson.getByRole('region', { name: 'Аудіодоріжки' });
  await section.getByRole('button', { name: /Мова доріжки/ }).click();
  await page.getByRole('option', { name: 'Англійська' }).click();
  await section.getByRole('textbox', { name: 'Назва доріжки' }).fill('English dub');
  const minted = page.waitForResponse(
    (candidate) =>
      candidate.url().includes('/audio-tracks') && candidate.request().method() === 'POST',
  );
  await section.locator('input[type="file"]').setInputFiles(dubTrack);
  expect((await minted).status()).toBe(201);
  await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(2, {
    timeout: 240_000,
  });
}

async function publishAndGrant(request: APIRequestContext, baseURL: string, courseID: string) {
  const headers = origin(baseURL);
  const current = await request.get(`/api/v1/admin/courses/${courseID}`);
  expect(current.status(), await current.text()).toBe(200);
  const { updatedAt } = token((await current.json()) as object);
  const published = await request.post(`/api/v1/admin/courses/${courseID}/status`, {
    data: { status: 'published', expectedUpdatedAt: updatedAt },
    headers,
  });
  expect(published.status(), await published.text()).toBe(200);
  const granted = await request.post(`/api/v1/admin/courses/${courseID}/grants`, {
    data: { userId: studentID, reason: 'e2e multi-audio S7 access for the student' },
    headers,
  });
  expect(granted.status(), await granted.text()).toBe(201);
}

async function play(page: Page, lessonID: string) {
  if (course === undefined) throw new Error('Audio course was not created');
  await page.goto(`/learn/${course.slug}/${lessonID}`);
  // A click is the gesture an unmuted play() needs (plan §4, _wip-s7-probe/README.md).
  await page.getByRole('button', { name: 'Відтворити' }).click();
  await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
}

type Peak = { hz: number; db: number; state: string; tracks: number; time: number };

/**
 * Taps the video's decoded audio for an analyser. Never connected to the destination, so the
 * player's own output path is unchanged.
 */
async function tapAudio(page: Page) {
  await page.locator('video').evaluate((node) => {
    if (!(node instanceof HTMLVideoElement)) throw new Error('Video element is unavailable');
    const stream = (node as HTMLVideoElement & { captureStream(): MediaStream }).captureStream();
    const context = new AudioContext();
    const analyser = context.createAnalyser();
    analyser.fftSize = 8192;
    context.createMediaStreamSource(stream).connect(analyser);
    Object.assign(window, { s7Tap: { context, analyser, stream } });
  });
}

async function peak(page: Page): Promise<Peak> {
  return page.evaluate(() => {
    const tap = (
      window as unknown as {
        s7Tap: { context: AudioContext; analyser: AnalyserNode; stream: MediaStream };
      }
    ).s7Tap;
    const bins = new Float32Array(tap.analyser.frequencyBinCount);
    tap.analyser.getFloatFrequencyData(bins);
    let top = 0;
    for (let index = 1; index < bins.length; index++) {
      if ((bins[index] ?? -Infinity) > (bins[top] ?? -Infinity)) top = index;
    }
    const video = document.querySelector('video');
    return {
      hz: (top * tap.context.sampleRate) / tap.analyser.fftSize,
      db: bins[top] ?? -Infinity,
      state: tap.context.state,
      tracks: tap.stream.getAudioTracks().length,
      time: video?.currentTime ?? -1,
    };
  });
}

function isTone(sample: Peak, hz: number): boolean {
  return Math.abs(sample.hz - hz) <= 12 && sample.db >= -60;
}

/** Waits for `count` consecutive samples (300 ms apart) of the tone; returns ms to the first. */
async function waitForTone(page: Page, hz: number, count: number, withinMs: number) {
  const started = Date.now();
  let run = 0;
  let firstAt = -1;
  const seen: Peak[] = [];
  while (Date.now() - started <= withinMs) {
    const sample = await peak(page);
    seen.push(sample);
    if (isTone(sample, hz)) {
      if (run === 0) firstAt = Date.now() - started;
      run += 1;
      if (run >= count) return { firstAt, seen };
    } else {
      run = 0;
    }
    await page.waitForTimeout(300);
  }
  throw new Error(
    `No ${count} consecutive ${hz} Hz samples in ${withinMs} ms: ${JSON.stringify(seen)}`,
  );
}

test.afterAll(async () => {
  if (course === undefined) return;
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
      'audio-cleanup',
      course.id,
    ],
    { cwd: repositoryRoot, timeout: 60_000 },
  );
});

test.describe('multi-audio S7 player audio tracks', () => {
  test.describe.configure({ mode: 'serial' });

  test('S7 an author publishes a lesson with two audio tracks and one with one', async ({
    browser,
    baseURL,
  }) => {
    // Two video transcodes and one audio transcode.
    test.setTimeout(600_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
    try {
      const page = await context.newPage();
      const seeded = await seed(page.request, baseURL);
      await page.goto(`/admin/courses/${seeded.id}`);
      await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
      const dual = lessonGroup(page, 1);
      await uploadVideo(page, dual, seeded.dual, dualVideo);
      await addDub(page, dual);
      await uploadVideo(page, lessonGroup(page, 2), seeded.single, singleVideo);
      await publishAndGrant(page.request, baseURL, seeded.id);
    } finally {
      await context.close();
    }
  });

  test.describe('as a student', () => {
    test.use({ storageState: authStatePath('student') });

    test('S7 switches to the dub, hears it without stopping and starts on it after a reload', async ({
      page,
    }) => {
      test.setTimeout(300_000);
      if (course === undefined) throw new Error('Audio course was not created');
      const dubResponses: Response[] = [];
      const audioPlaylists: string[] = [];
      // Listeners first, so no request made by an action below can be missed.
      page.on('response', (response) => {
        const pathname = new URL(response.url()).pathname;
        if (/\/audio-[0-9a-f-]{36}\//.test(pathname)) dubResponses.push(response);
      });
      page.on('request', (request) => {
        const pathname = new URL(request.url()).pathname;
        if (anyAudioPlaylist.test(pathname)) audioPlaylists.push(pathname);
      });

      await play(page, course.dual);
      await tapAudio(page);
      const baseline = await waitForTone(page, 440, 3, 10_000);
      expect(baseline.seen.at(-1)?.state).toBe('running');
      expect(baseline.seen.at(-1)?.tracks).toBe(1);
      expect(dubResponses).toHaveLength(0);

      const settings = page.getByRole('button', { name: 'Налаштування' });
      await settings.click();
      const audio = page.getByRole('radiogroup', { name: 'Аудіо' });
      await expect(audio.getByRole('radio')).toHaveCount(2);
      await expect(audio.getByRole('radio', { name: 'Original' })).toBeChecked();
      const before = (await peak(page)).time;
      expect(before).toBeLessThanOrEqual(10);
      const clickedAt = Date.now();
      await audio.locator('label', { hasText: 'English dub' }).click();
      await expect(audio.getByRole('radio', { name: 'English dub' })).toBeChecked();
      await expect
        .poll(
          () =>
            dubResponses.some(
              (response) =>
                uploadedAudioPlaylist.test(new URL(response.url()).pathname) &&
                response.status() === 200,
            ),
          { timeout: 10_000 },
        )
        .toBe(true);
      await expect
        .poll(
          () =>
            dubResponses.some(
              (response) =>
                new URL(response.url()).pathname.endsWith('.m4s') && response.status() === 200,
            ),
          { timeout: 10_000 },
        )
        .toBe(true);
      const firstSegmentMs = Date.now() - clickedAt;
      const switched = await waitForTone(page, 880, 5, 10_000);
      const audibleMs = Date.now() - clickedAt;
      const after = await peak(page);
      expect(after.time).toBeGreaterThan(before);
      await expect(page.getByRole('button', { name: 'Пауза' })).toBeVisible();
      console.log(
        `S7 switch: first dub segment ${firstSegmentMs} ms, 880 Hz run ended ${audibleMs} ms after the click (first 880 Hz sample ${switched.firstAt} ms into the wait), currentTime ${before.toFixed(2)} -> ${after.time.toFixed(2)} s`,
      );

      // The choice is remembered: a new view starts on the dub before the default is requested.
      await page.keyboard.press('Escape');
      audioPlaylists.length = 0;
      await play(page, course.dual);
      await expect.poll(() => audioPlaylists.length, { timeout: 15_000 }).toBeGreaterThan(0);
      expect(audioPlaylists[0]).toMatch(uploadedAudioPlaylist);
      await tapAudio(page);
      await waitForTone(page, 880, 3, 10_000);
      await page.getByRole('button', { name: 'Налаштування' }).click();
      await expect(
        page.getByRole('radiogroup', { name: 'Аудіо' }).getByRole('radio', { name: 'English dub' }),
      ).toBeChecked();
    });

    test('S7 the audio group is keyboard reachable and returns focus on Escape', async ({
      page,
    }) => {
      test.setTimeout(120_000);
      if (course === undefined) throw new Error('Audio course was not created');
      await play(page, course.dual);
      const settings = page.getByRole('button', { name: 'Налаштування' });
      await settings.focus();
      await page.keyboard.press('Enter');
      const audio = page.getByRole('radiogroup', { name: 'Аудіо' });
      await expect(audio).toBeVisible();
      // A fresh browser context has no remembered language: the default is selected, and an arrow
      // moves within the group, selecting as it goes.
      const original = audio.getByRole('radio', { name: 'Original' });
      await expect(original).toBeChecked();
      await original.focus();
      await expect(original).toBeFocused();
      await page.keyboard.press('ArrowDown');
      const dub = audio.getByRole('radio', { name: 'English dub' });
      await expect(dub).toBeFocused();
      await expect(dub).toBeChecked();
      await page.keyboard.press('Escape');
      await expect(audio).toBeHidden();
      await expect(settings).toBeFocused();
    });

    test('S7 a lesson with one audio track has no audio group', async ({ page }) => {
      test.setTimeout(120_000);
      if (course === undefined) throw new Error('Audio course was not created');
      await play(page, course.single);
      await page.getByRole('button', { name: 'Налаштування' }).click();
      await expect(page.getByRole('radiogroup', { name: 'Швидкість' })).toBeVisible();
      await expect(page.getByRole('radiogroup', { name: 'Аудіо' })).toHaveCount(0);
    });
  });
});
