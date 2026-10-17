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
  await page.getByRole('radio', { name: /Lessons|Уроки/ }).click();
  await expect(page.getByRole('group', { name: 'Модуль 1' })).toBeVisible();
}

async function upload(page: Page, fixture: string) {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.url().includes('/api/v1/admin/lessons/') &&
      candidate.url().endsWith('/video') &&
      candidate.request().method() === 'POST',
  );
  await page.locator('input[type="file"]').first().setInputFiles(fixture);
  expect((await response).status()).toBe(201);
  await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
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
    await page.getByRole('radio', { name: /Lessons|Уроки/ }).click();
    await expect(
      page.getByText(
        /Незбережені зміни поверх застарілого тексту|Unsaved changes over text that has moved on/,
      ),
    ).toBeVisible({ timeout: 30_000 });
  });
});
