import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { promisify } from 'node:util';
import { expect, test, type APIRequestContext, type Page } from '@playwright/test';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-asr-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const speechVideo = path.join(fixtureDirectory, 'e2e-speech-640x360.mp4');
const silentVideo = path.join(fixtureDirectory, 'e2e-640x360.mp4');
const videoOnly = path.join(fixtureDirectory, 'e2e-640x360-silent.mp4');
const speechTrack = path.join(fixtureDirectory, 'e2e-speech-dub.m4a');
const courses: { id: string; updatedAt: string }[] = [];
const execute = promisify(execFile);
const repositoryRoot = fileURLToPath(new URL('../../', import.meta.url));

type Seed = { id: string; updatedAt: string; lessonID: string };

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

async function pauseRecognition() {
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'redis',
      'sh',
      '-c',
      'export REDISCLI_AUTH="$REDIS_PASSWORD"; redis-cli SET "asynq:{asr}:paused" "$(date +%s)" > /dev/null',
    ],
    { cwd: repositoryRoot, timeout: 30_000 },
  );
}

async function resumeRecognition() {
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'redis',
      'sh',
      '-c',
      'export REDISCLI_AUTH="$REDIS_PASSWORD"; redis-cli DEL "asynq:{asr}:paused" > /dev/null',
    ],
    { cwd: repositoryRoot, timeout: 30_000 },
  );
}

async function seed(
  request: APIRequestContext,
  baseURL: string,
  suffix: string,
  transcriptLanguage: string | null = 'uk',
): Promise<Seed> {
  const headers = origin(baseURL);
  const courseID = randomUUID();
  const lessonID = randomUUID();
  const moduleID = randomUUID();
  const slug = `${runNamespace}-${suffix}-${courseID.slice(0, 8)}`;
  const draft = {
    slug,
    title: `ASR ${suffix} ${runNamespace}`,
    summary: 'ASR browser journey.',
    description: 'A run-scoped course for ASR acceptance.',
    instructorId: adminID,
  };

  const created = await request.post('/api/v1/admin/courses', { data: draft, headers });
  expect(created.status(), await created.text()).toBe(201);
  const course = token((await created.json()) as object);

  const saved = await request.put(`/api/v1/admin/courses/${course.id}`, {
    data: {
      expectedUpdatedAt: course.updatedAt,
      course: draft,
      modules: [
        {
          id: moduleID,
          title: `Модуль ${runNamespace}`,
          lessons: [{ id: lessonID, title: `Урок ${suffix}` }],
        },
      ],
    },
    headers,
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const savedCourse = token((await saved.json()) as object);
  if (transcriptLanguage !== null) {
    const transcript = await request.put(`/api/v1/admin/lessons/${lessonID}/transcript`, {
      data: {
        language: transcriptLanguage,
        segments: [
          {
            id: randomUUID(),
            startMs: 0,
            endMs: 1_000,
            text: 'Початковий текст уроку',
            documentIds: [],
          },
        ],
      },
      headers,
    });
    expect(transcript.status(), await transcript.text()).toBe(200);
  }
  courses.push(savedCourse);
  return { ...savedCourse, lessonID };
}

async function openEditor(page: Page, courseID: string) {
  await page.goto(`/admin/courses/${courseID}`);
  // Modules/lessons live on the Lessons tab of the course editor (view-3 layout); reveal it first.
  await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
  await expect(page.getByRole('group', { name: 'Модуль 1' })).toBeVisible();
}

async function upload(page: Page, fixture: string) {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.url().includes('/api/v1/admin/lessons/') &&
      candidate.url().endsWith('/video') &&
      candidate.request().method() === 'POST',
  );
  await page
    .locator('input[type="file"]:not(.admin-audio-field input)')
    .first()
    .setInputFiles(fixture);
  expect((await response).status()).toBe(201);
  await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
}

/** Adds an uploaded audio track to the lesson's shown video (multi-audio S5-L2b UI). */
async function addAudioTrack(page: Page, language: string, label: string, file: string) {
  const section = page.getByRole('region', { name: 'Аудіодоріжки' });
  await section.getByRole('button', { name: /Мова доріжки/ }).click();
  await page.getByRole('option', { name: language }).click();
  await section.getByRole('textbox', { name: 'Назва доріжки' }).fill(label);
  const minted = page.waitForResponse(
    (candidate) =>
      candidate.url().includes('/audio-tracks') && candidate.request().method() === 'POST',
  );
  await section.locator('input[type="file"]').setInputFiles(file);
  expect((await minted).status()).toBe(201);
  return section;
}

type DraftRead = {
  readonly state: string | null;
  readonly updatedAt: string | null;
  readonly segments: number;
  readonly text: string;
};

async function readDraft(request: APIRequestContext, lessonID: string): Promise<DraftRead> {
  const response = await request.get(`/api/v1/admin/lessons/${lessonID}/draft`);
  expect(response.status(), await response.text()).toBe(200);
  const body: unknown = await response.json();
  if (typeof body !== 'object' || body === null)
    throw new Error('Draft response must be an object');
  const state: unknown = Reflect.get(body, 'state');
  const updatedAt: unknown = Reflect.get(body, 'updatedAt');
  const segments: unknown = Reflect.get(body, 'segments');
  const lines: unknown[] = Array.isArray(segments) ? segments : [];
  return {
    state: typeof state === 'string' ? state : null,
    updatedAt: typeof updatedAt === 'string' ? updatedAt : null,
    segments: lines.length,
    text: lines
      .map((line) => {
        const text: unknown =
          typeof line === 'object' && line !== null ? Reflect.get(line, 'text') : undefined;
        return typeof text === 'string' ? text : '';
      })
      .join(' '),
  };
}

async function waitForDraft(page: Page, state: 'ready' | 'no_speech') {
  await expect(page.locator(`[data-draft-state="${state}"]`)).toBeVisible({ timeout: 240_000 });
}

async function waitForRegeneratedDraft(page: Page) {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.url().includes('/draft/regenerate') && candidate.request().method() === 'POST',
  );
  await page
    .getByRole('button', { name: /Перегенерувати транскрипт|Regenerate transcript/ })
    .click();
  expect((await response).status()).toBe(202);
  await expect(
    page.locator('[data-draft-state="pending"], [data-draft-state="running"]'),
  ).toBeVisible({
    timeout: 30_000,
  });
  await waitForDraft(page, 'ready');
}

async function startRecognition(page: Page) {
  const button = page.getByRole('button', {
    name: /Згенерувати транскрипт|Generate transcript/,
  });
  const inProgress = page.locator('[data-draft-state="pending"], [data-draft-state="running"]');
  let shouldClick = false;
  await expect
    .poll(
      async () => {
        if ((await button.isVisible()) && (await button.isEnabled())) {
          shouldClick = true;
          return true;
        }
        return (await inProgress.count()) > 0;
      },
      { timeout: 30_000, intervals: [250] },
    )
    .toBe(true);
  if (shouldClick) await button.click();
}

async function selectLanguage(page: Page, label: 'Auto' | 'English') {
  const select = page.getByRole('button', { name: /Мова розпізнавання|Recognition language/ });
  await select.click();
  await expect(page.getByRole('option')).toHaveCount(3);
  await expect(page.getByRole('option', { name: 'Auto' })).toBeVisible();
  await expect(page.getByRole('option', { name: 'English' })).toBeVisible();
  await expect(page.getByRole('option', { name: 'Українська' })).toBeVisible();
  await page.getByRole('option', { name: label }).click();
}

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    for (const course of courses) {
      const response = await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
        data: { expectedUpdatedAt: course.updatedAt },
        headers: origin(baseURL),
      });
      expect(response.status(), await response.text()).toBe(204);
    }
  } finally {
    await context.close();
  }
});

test.describe('ASR acceptance with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('C13 shows no speech for a silent video', async ({ page, baseURL }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'silent');
    await openEditor(page, course.id);
    await upload(page, silentVideo);
    await startRecognition(page);
    await waitForDraft(page, 'no_speech');
    await expect(page.locator('[data-draft-state="ready"]')).toHaveCount(0);
  });

  test('C03 observes recognition in progress and then a ready draft', async ({ page, baseURL }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'speech');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(page);
    await expect
      .poll(
        async () =>
          (await page
            .locator('[data-draft-state="pending"], [data-draft-state="running"]')
            .count()) > 0,
        { timeout: 120_000, intervals: [500] },
      )
      .toBe(true);
    await waitForDraft(page, 'ready');
  });

  test('C11 sends no language for auto and the selected language otherwise', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'language');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(page);
    await waitForDraft(page, 'ready');

    await selectLanguage(page, 'Auto');
    const autoRequest = page.waitForRequest(
      (request) => request.url().includes('/draft/regenerate') && request.method() === 'POST',
    );
    await waitForRegeneratedDraft(page);
    const autoRequestValue = await autoRequest;
    const autoBodyText = autoRequestValue.postData();
    const autoBody: unknown = autoBodyText === null ? undefined : JSON.parse(autoBodyText);
    expect(
      autoBody === undefined ||
        (typeof autoBody === 'object' && autoBody !== null && !('language' in autoBody)),
    ).toBe(true);

    await waitForDraft(page, 'ready');
    await selectLanguage(page, 'English');
    const languageRequest = page.waitForRequest(
      (request) => request.url().includes('/draft/regenerate') && request.method() === 'POST',
    );
    await waitForRegeneratedDraft(page);
    const request = await languageRequest;
    const bodyText = request.postData();
    if (bodyText === null) throw new Error('Language regeneration request must have a body');
    const body: unknown = JSON.parse(bodyText);
    if (typeof body !== 'object' || body === null || !('language' in body)) {
      throw new Error('Language regeneration request must carry language');
    }
    const language = Reflect.get(body, 'language');
    expect(language).toBe('en');
  });

  test('C06 marks a ready draft stale after replacing the video', async ({ page, baseURL }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'stale');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(page);
    await waitForDraft(page, 'ready');
    // The stale state exists only until the next automatic recognition of the new generation.
    // Pause that queue so the observable window stays open; the data is still a real replacement.
    await pauseRecognition();
    try {
      await upload(page, silentVideo);
      await expect(page.getByText(/старшого відео|older video/)).toBeVisible({ timeout: 120_000 });
    } finally {
      await resumeRecognition();
    }
  });

  test('C17 keeps editor text mounted and offers the recognized draft', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'editor', 'en');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const text = page.getByRole('textbox', { name: 'Текст репліки 1' });
    await text.fill('Незбережений текст автора');
    await startRecognition(page);
    await waitForDraft(page, 'ready');
    await expect(text).toHaveValue('Незбережений текст автора');
    const discard = page.getByRole('button', { name: /Не відновлювати|Discard/ });
    if (await discard.isVisible()) await discard.click();
    await page
      .getByRole('button', { name: /Завантажити чернетку в редактор|Load draft into editor/ })
      .click();
    await expect(
      page.getByText(/Чернетка розпізнавання готова|A recognized draft is ready/),
    ).toBeVisible();
    await page.getByRole('button', { name: /^(Завантажити чернетку|Load the draft)$/ }).click();
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(/.+/);
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).not.toHaveValue(
      'Незбережений текст автора',
    );
  });

  test('C18 offers stale recovery after an unsaved draft is regenerated', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'recovery', 'en');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(page);
    await waitForDraft(page, 'ready');
    await page
      .getByRole('button', { name: /Завантажити чернетку в редактор|Load draft into editor/ })
      .click();
    await page.getByRole('button', { name: /^(Завантажити чернетку|Load the draft)$/ }).click();
    const text = page.getByRole('textbox', { name: 'Текст репліки 1' });
    await text.fill('Незбережений варіант після чернетки');
    await text.blur();
    await selectLanguage(page, 'Auto');
    await waitForRegeneratedDraft(page);
    await page.reload();
    // A reload resets the editor tab to Overview; reopen Lessons to reach the draft notice.
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    await expect(
      page.getByText(
        /Незбережені зміни поверх застарілого тексту|Unsaved changes over text that has moved on/,
      ),
    ).toBeVisible({ timeout: 30_000 });
  });

  test('S6 a video-only lesson is recognised from its uploaded default track', async ({
    page,
    baseURL,
  }) => {
    // Video transcode + audio transcode + the 60 s ASR sweep + recognition.
    test.setTimeout(420_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'video-only', null);
    await openEditor(page, course.id);
    const minted = page.waitForResponse(
      (candidate) =>
        candidate.url().endsWith(`/api/v1/admin/lessons/${course.lessonID}/video`) &&
        candidate.request().method() === 'POST',
    );
    await page
      .locator('input[type="file"]:not(.admin-audio-field input)')
      .first()
      .setInputFiles(videoOnly);
    expect((await minted).status()).toBe(201);
    await expect(page.locator('[data-video-state="awaiting_audio"]')).toBeVisible({
      timeout: 240_000,
    });
    await addAudioTrack(page, 'Англійська', 'English dub', speechTrack);
    await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });

    // Before S6 this lesson failed with audio_extraction_failed: the video source has no sound.
    // Delivery of a lesson's first draft into an open editor is outside S6 (D-S6-6), so the
    // result is read from the server and then shown after a reload.
    await expect
      .poll(async () => (await readDraft(page.request, course.lessonID)).state, {
        timeout: 240_000,
        intervals: [2_000],
      })
      .toBe('ready');
    expect((await readDraft(page.request, course.lessonID)).segments).toBeGreaterThan(0);
    await page.reload();
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    await waitForDraft(page, 'ready');
  });

  test('S6 a default-track change makes the open draft stale and replaces it unaided', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(480_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'default-change');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(page);
    await waitForDraft(page, 'ready');
    const original = await readDraft(page.request, course.lessonID);

    const section = await addAudioTrack(page, 'Англійська', 'Dub', speechTrack);
    await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(2, {
      timeout: 240_000,
    });
    // Hold recognition while the banner is observed: a sweep right after the switch could claim
    // the replacement before the re-read draft reaches the browser (C06's pattern).
    const stale = page.getByText(/іншої аудіодоріжки|another audio track/);
    await pauseRecognition();
    try {
      await section
        .getByRole('button', { name: 'Зробити основною доріжку Dub (Англійська)' })
        .click();
      // The editor learns at once that its draft heard another track (no reload, no Regenerate)…
      await expect(stale).toBeVisible({ timeout: 30_000 });
    } finally {
      await resumeRecognition();
    }
    // …and the server's automatic replacement lands in the same open editor.
    await expect(stale).toBeHidden({ timeout: 300_000 });
    await waitForDraft(page, 'ready');
    const replaced = await readDraft(page.request, course.lessonID);
    expect(replaced.state).toBe('ready');
    expect(replaced.updatedAt).not.toBe(original.updatedAt);
    // It heard the dub, not the video's own speech ("Welcome to the course…").
    expect(replaced.text).toMatch(/dub|languages/i);
    expect(original.text).not.toMatch(/dub|languages/i);
  });
});
