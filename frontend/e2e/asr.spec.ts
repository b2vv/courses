import { randomUUID } from 'node:crypto';
import path from 'node:path';
import {
  expect,
  test,
  type APIRequestContext,
  type Page,
  type Locator,
  type Request,
} from '@playwright/test';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { z } from 'zod';
import { writeFile } from 'node:fs/promises';
import { authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-asr-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const fixtureDirectory = path.join(import.meta.dirname, '..', 'playwright', '.fixtures');
const speechVideo = path.join(fixtureDirectory, 'e2e-speech-640x360.mp4');
const silentVideo = path.join(fixtureDirectory, 'e2e-640x360.mp4');
const videoOnly = path.join(fixtureDirectory, 'e2e-640x360-silent.mp4');
const speechTrack = path.join(fixtureDirectory, 'e2e-speech-dub.m4a');
const courses: { id: string; updatedAt: string; lessonID: string }[] = [];

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
  const course = token(
    z.object({ id: z.string(), updatedAt: z.string() }).parse(await created.json()),
  );

  courses.push(Object.assign(course, { lessonID }));
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
  const savedCourse = token(
    z.object({ id: z.string(), updatedAt: z.string() }).parse(await saved.json()),
  );
  Object.assign(course, savedCourse);
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
  return { ...savedCourse, lessonID };
}

async function openEditor(page: Page, courseID: string) {
  await page.goto(`/admin/courses/${courseID}`);
  // Modules/lessons live on the Lessons tab of the course editor (view-3 layout); reveal it first.
  await page.getByRole('tab', { name: 'Уроки' }).click();
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
  const minted = await response;
  expect(minted.status()).toBe(201);
  const ticket = z.object({ videoId: z.string().uuid() }).parse(await minted.json());
  await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
  return ticket.videoId;
}

/** Real UI upload; en-GB only rewrites the mint metadata, never the audio bytes. */
async function addAudioTrack(
  page: Page,
  label: string,
  options: {
    british?: boolean;
    language?: 'en' | 'de';
    multilingual?: boolean;
    videoOnly?: boolean;
  } = {},
) {
  const section = page.getByRole('region', { name: 'Аудіодоріжки' });
  const form = section.locator('fieldset');
  await form.getByRole('button', { name: /Мова доріжки/ }).click();
  await page
    .getByRole('option', {
      name: options.language === 'de' ? 'Німецька' : 'Англійська',
      exact: true,
    })
    .click();
  await form.getByRole('textbox', { name: 'Назва доріжки' }).fill(label);
  if (options.multilingual) {
    const control = form.getByRole('switch', { name: 'Кілька мов', exact: true });
    await control.focus();
    if (!(await control.isChecked())) await control.press('Space');
    await expect(control).toBeChecked();
  }
  if (options.british) {
    await page.route('**/api/v1/admin/videos/*/audio-tracks', async (route) => {
      if (route.request().method() !== 'POST') return route.continue();
      const body = z.record(z.string(), z.unknown()).parse(route.request().postDataJSON());
      await route.continue({ postData: JSON.stringify({ ...body, language: 'en-GB' }) });
    });
  }
  const minted = page.waitForResponse(
    (r) => r.url().endsWith('/audio-tracks') && r.request().method() === 'POST',
  );
  await form.locator('input[type="file"]').setInputFiles(speechTrack);
  const response = await minted;
  expect(response.status(), await response.text()).toBe(201);
  const ticket = z.object({ trackId: z.string().uuid() }).parse(await response.json());
  await expect(section.locator('li[data-audio-track-state="ready"]')).toHaveCount(
    options.videoOnly ? 1 : 2,
    { timeout: 240_000 },
  );
  const list = await page.request.get(response.url());
  expect(list.status()).toBe(200);
  const tracks = z
    .object({
      items: z.array(
        z.object({ trackId: z.string().uuid(), language: z.string(), multilingual: z.boolean() }),
      ),
    })
    .parse(await list.json()).items;
  const track = tracks.find((item) => item.trackId === ticket.trackId);
  if (!track) throw new Error('Minted track missing from inventory');
  expect(track.language.toLowerCase()).toBe(options.british ? 'en-gb' : (options.language ?? 'en'));
  if (options.multilingual) expect(track.multilingual).toBe(true);
  return track;
}

const segmentSchema = z.object({
  id: z.string().uuid(),
  startMs: z.number().int().nonnegative(),
  endMs: z.number().int().nonnegative(),
  text: z.string().min(1),
});
const transcriptPutSchema = z.object({ language: z.string(), segments: z.array(segmentSchema) });
async function readTranscript(request: APIRequestContext, lessonID: string, language: string) {
  const response = await request.get(
    `/api/v1/learn/lessons/${lessonID}/transcript?language=${language}`,
  );
  expect(response.status(), await response.text()).toBe(200);
  return transcriptPutSchema.extend({ updatedAt: z.string() }).parse(await response.json());
}
async function expectTranscriptTexts(
  request: APIRequestContext,
  lessonID: string,
  language: string,
  texts: string[],
) {
  const transcript = await readTranscript(request, lessonID, language);
  expect(transcript.language).toBe(language);
  expect(transcript.segments.map((line) => line.text)).toEqual(texts);
}

const draftSchema = z.object({
  assetId: z.string().uuid(),
  trackId: z.string().uuid(),
  trackLanguage: z.string(),
  trackMultilingual: z.boolean(),
  state: z.enum(['queued', 'running', 'ready', 'failed', 'no_speech']),
  updatedAt: z.string(),
  stale: z.boolean(),
  segments: z.array(segmentSchema),
});
async function readDrafts(request: APIRequestContext, lessonID: string) {
  const response = await request.get(`/api/v1/admin/lessons/${lessonID}/drafts`);
  expect(response.status(), await response.text()).toBe(200);
  const body: unknown = await response.json();
  return z.object({ drafts: z.array(draftSchema) }).parse(body).drafts;
}
async function draftFor(request: APIRequestContext, lessonID: string, trackId: string) {
  return (await readDrafts(request, lessonID)).find((draft) => draft.trackId === trackId);
}
function selectedPanel(page: Page) {
  // Only the selected, visible language panel; the course editor also has outer tabpanels.
  return page
    .getByRole('tabpanel')
    .filter({ has: page.getByRole('heading', { name: 'Транскрипт уроку', exact: true }) })
    .last();
}
async function selectTab(page: Page, name: string | RegExp) {
  await page.getByRole('tablist', { name: 'Мови транскрипту' }).getByRole('tab', { name }).click();
  return selectedPanel(page);
}
async function waitForDraft(panel: Locator, state: 'ready' | 'no_speech') {
  await expect(panel.locator(`[data-draft-state="${state}"]`)).toBeVisible({ timeout: 300_000 });
}
async function startRecognition(panel: Locator) {
  const button = panel.getByRole('button', { name: 'Згенерувати транскрипт', exact: true });
  await expect
    .poll(
      async () =>
        (await button.isVisible()) || (await panel.locator('[data-draft-state]').count()) > 0,
      { timeout: 30_000 },
    )
    .toBe(true);
  if (await button.isVisible()) await button.click();
}
async function waitForRegeneratedDraft(page: Page, panel: Locator) {
  const response = page.waitForResponse(
    (r) => /\/drafts\/[^/]+\/regenerate$/.test(r.url()) && r.request().method() === 'POST',
  );
  await panel.getByRole('button', { name: 'Перегенерувати транскрипт', exact: true }).click();
  const value = await response;
  expect(value.status()).toBe(202);
  expect(value.request().postData()).toBeNull();
  await expect(
    panel.locator('[data-draft-state="pending"], [data-draft-state="running"]'),
  ).toBeVisible({ timeout: 30_000 });
  await waitForDraft(panel, 'ready');
}
async function loadDraft(panel: Locator) {
  await panel.getByRole('button', { name: 'Завантажити чернетку в редактор', exact: true }).click();
  await panel.getByRole('button', { name: 'Завантажити чернетку', exact: true }).click();
}
async function confirmEnglish(page: Page) {
  const before = network.length;
  const selected = page
    .getByRole('tablist', { name: 'Мови транскрипту' })
    .getByRole('tab', { selected: true });
  const selectedID = await selected.getAttribute('id');
  const dialog = page.getByRole('dialog', { name: 'Підтвердьте мову транскрипту' });
  await dialog.getByRole('button', { name: /Мова доріжки/ }).click();
  await page.getByRole('option', { name: 'Англійська', exact: true }).click();
  await dialog.getByRole('button', { name: 'Перейменувати доріжку й зберегти' }).click();
  await expect(dialog).toBeHidden();
  await expect(selected).toHaveAttribute('id', selectedID ?? '');
  expect(
    network
      .slice(before)
      .filter((entry) => entry.method === 'PUT' && entry.path.startsWith('/api/v1/admin/courses/')),
  ).toEqual([]);
}
async function saveTranscript(page: Page, panel: Locator, lessonID: string, und = false) {
  const response = page.waitForResponse(
    (r) => r.url().endsWith(`/lessons/${lessonID}/transcript`) && r.request().method() === 'PUT',
  );
  await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
  if (und) await confirmEnglish(page);
  const value = await response;
  expect(value.status(), await value.text()).toBe(200);
  return z
    .object({ language: z.string(), segments: z.array(segmentSchema) })
    .parse(value.request().postDataJSON());
}
async function snapshotKeys(page: Page, clear = false): Promise<string[]> {
  return page.evaluate(
    ({ clear, prefix }) =>
      new Promise<string[]>((resolve, reject) => {
        const opening = indexedDB.open('courses-transcript-history');
        opening.onerror = () => reject(opening.error);
        opening.onsuccess = () => {
          const db = opening.result;
          if (!db.objectStoreNames.contains('snapshots')) {
            db.close();
            resolve([]);
            return;
          }
          const tx = db.transaction('snapshots', clear ? 'readwrite' : 'readonly');
          const store = tx.objectStore('snapshots');
          const read = store.getAllKeys();
          let keys: string[] = [];
          read.onsuccess = () => {
            keys = read.result.map(String).filter((key) => prefix.some((p) => key.startsWith(p)));
            if (clear) for (const key of keys) store.delete(key);
          };
          tx.oncomplete = () => {
            db.close();
            resolve(keys.sort());
          };
          tx.onerror = () => {
            db.close();
            reject(tx.error);
          };
        };
      }),
    { clear, prefix: courses.map((course) => `${adminID}:${course.lessonID}:`) },
  );
}

async function renameTrack(page: Page, label: string, language: string) {
  const before = network.length;
  const selected = page
    .getByRole('tablist', { name: 'Мови транскрипту' })
    .getByRole('tab', { selected: true });
  const selectedID = await selected.getAttribute('id');
  const row = page
    .getByRole('region', { name: 'Аудіодоріжки' })
    .getByRole('listitem')
    .filter({ has: page.getByText(label, { exact: true }) });
  await row.getByRole('button', { name: /^Змінити доріжку/ }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: /Мова доріжки/ }).click();
  await page.getByRole('option', { name: language, exact: true }).click();
  const response = page.waitForResponse(
    (r) => r.url().includes('/audio-tracks/') && r.request().method() === 'PATCH',
  );
  await dialog.getByRole('button', { name: 'Зберегти', exact: true }).click();
  const patched = await response;
  expect(patched.status()).toBe(200);
  await expect(dialog).toBeHidden();
  await expect(selected).toHaveAttribute('id', selectedID ?? '');
  // The PATCH answer arrives before the tracks refetch re-renders the tabs; a Save clicked in
  // between still carries the old language and M3-R10 correctly refuses it (gate 2026-11-17).
  // Wait for what the author sees: this track's tab (React Aria id `…-tab-track:<id>`) is
  // named after the new language.
  const trackID = new URL(patched.url()).pathname.split('/').at(-1) ?? '';
  await expect(
    page
      .getByRole('tablist', { name: 'Мови транскрипту' })
      .locator(`[role="tab"][id$="-tab-track:${trackID}"]`),
  ).toHaveAccessibleName(new RegExp(`^${language}`));
  expect(
    network
      .slice(before)
      .filter((entry) => entry.method === 'PUT' && entry.path.startsWith('/api/v1/admin/courses/')),
  ).toEqual([]);
}
async function renameEmbedded(page: Page, language: string) {
  const row = page.getByRole('region', { name: 'Аудіодоріжки' }).getByRole('listitem').first();
  const label = await row.locator('strong').innerText();
  await renameTrack(page, label, language);
}

const network: { method: string; path: string; body: unknown; status?: number }[] = [];
test.beforeEach(async ({ page }) => {
  network.length = 0;
  page.on('request', (request: Request) => {
    const url = new URL(request.url());
    if (!/^\/api\/v1\/admin\/(lessons|videos|courses)(\/|$)/.test(url.pathname)) return;
    const raw = request.postData();
    network.push({
      method: request.method(),
      path: url.pathname,
      body: raw === null ? null : JSON.parse(raw),
    });
  });
  page.on('response', (response) => {
    const entry = [...network]
      .reverse()
      .find(
        (entry) =>
          entry.path === new URL(response.url()).pathname &&
          entry.method === response.request().method() &&
          entry.status === undefined,
      );
    if (entry) entry.status = response.status();
  });
});
test.afterEach(async ({ page }, info) => {
  const log = info.outputPath('requests.json');
  await writeFile(log, JSON.stringify(network, null, 2));
  await info.attach('requests', { path: log, contentType: 'application/json' });
  if (info.status !== info.expectedStatus) {
    const screenshot = info.outputPath('failure.png');
    await page.screenshot({ path: screenshot, fullPage: true });
    await info.attach('failure', { path: screenshot, contentType: 'image/png' });
  }
  // Clear only this spec's course-scoped keys after the proof (also on assertion failure).
  if (!page.isClosed() && new URL(page.url()).protocol.startsWith('http'))
    await snapshotKeys(page, true);
});

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  const failures: string[] = [];
  try {
    for (const course of courses) {
      try {
        const current = await context.request.get(`/api/v1/admin/courses/${course.id}`);
        expect(current.status(), await current.text()).toBe(200);
        const fresh = token(
          z.object({ id: z.string(), updatedAt: z.string() }).parse(await current.json()),
        );
        const response = await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
          data: { expectedUpdatedAt: fresh.updatedAt },
          headers: origin(baseURL),
        });
        expect(response.status(), await response.text()).toBe(204);
      } catch (error) {
        failures.push(`${course.id}: ${error instanceof Error ? error.message : String(error)}`);
      }
    }
  } finally {
    await context.close();
  }
  expect(failures, 'Every run-scoped course must be cleaned up').toEqual([]);
});

test.describe('ASR acceptance with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin'), actionTimeout: 30_000 });

  test('C13 shows no speech for a silent video', async ({ page, baseURL }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'silent');
    await openEditor(page, course.id);
    await upload(page, silentVideo);
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'no_speech');
    await expect(page.locator('[data-draft-state="ready"]')).toHaveCount(0);
  });

  test('C03 observes recognition in progress and then a ready draft', async ({ page, baseURL }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'speech');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(selectedPanel(page));
    await expect
      .poll(
        async () =>
          (await page
            .locator('[data-draft-state="pending"], [data-draft-state="running"]')
            .count()) > 0,
        { timeout: 120_000, intervals: [500] },
      )
      .toBe(true);
    await waitForDraft(selectedPanel(page), 'ready');
  });

  test('C11 the recognition hint comes from the track language and Regenerate sends no body', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(600_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'language', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const panel = selectedPanel(page);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    await expect(panel.getByRole('button', { name: /Мова розпізнавання/ })).toHaveCount(0);
    const original = (await readDrafts(page.request, course.lessonID))[0];
    if (!original) throw new Error('Expected embedded draft');
    expect(original.trackLanguage).toBe('und');
    await renameEmbedded(page, 'Англійська');
    await waitForRegeneratedDraft(page, panel);
    expect((await draftFor(page.request, course.lessonID, original.trackId))?.trackLanguage).toBe(
      'en',
    );
  });

  test('C06 after replacing the video the old draft is not offered', async ({ page, baseURL }) => {
    test.setTimeout(600_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'replacement', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'ready');
    const original = (await readDrafts(page.request, course.lessonID))[0];
    await upload(page, silentVideo);
    await expect
      .poll(
        async () =>
          (await readDrafts(page.request, course.lessonID)).some(
            (draft) => draft.trackId === original?.trackId,
          ),
        { timeout: 30_000 },
      )
      .toBe(false);
    await expect(
      selectedPanel(page).getByRole('button', {
        name: 'Завантажити чернетку в редактор',
        exact: true,
      }),
    ).toHaveCount(0);
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'no_speech');
  });

  test('C17 keeps editor text mounted and offers the recognized draft', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'editor', 'und');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const text = selectedPanel(page).getByRole('textbox', { name: 'Текст репліки 1' });
    await text.fill('Незбережений текст автора');
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'ready');
    await expect(text).toHaveValue('Незбережений текст автора');
    const discard = selectedPanel(page).getByRole('button', { name: /Не відновлювати|Discard/ });
    if (await discard.isVisible()) await discard.click();
    await selectedPanel(page)
      .getByRole('button', { name: /Завантажити чернетку в редактор|Load draft into editor/ })
      .click();
    await expect(
      selectedPanel(page).getByText(/Чернетка розпізнавання готова|A recognized draft is ready/),
    ).toBeVisible();
    await selectedPanel(page)
      .getByRole('button', { name: /^(Завантажити чернетку|Load the draft)$/ })
      .click();
    await expect(selectedPanel(page).getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(
      /.+/,
    );
    await expect(
      selectedPanel(page).getByRole('textbox', { name: 'Текст репліки 1' }),
    ).not.toHaveValue('Незбережений текст автора');
  });

  test('C18 offers stale recovery after an unsaved draft is regenerated', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(240_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'recovery', 'und');
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'ready');
    await selectedPanel(page)
      .getByRole('button', { name: /Завантажити чернетку в редактор|Load draft into editor/ })
      .click();
    await selectedPanel(page)
      .getByRole('button', { name: /^(Завантажити чернетку|Load the draft)$/ })
      .click();
    const text = selectedPanel(page).getByRole('textbox', { name: 'Текст репліки 1' });
    await text.fill('Незбережений варіант після чернетки');
    await text.blur();
    await waitForRegeneratedDraft(page, selectedPanel(page));
    await page.reload();
    // A reload resets the editor tab to Overview; reopen Lessons to reach the draft notice.
    await page.getByRole('tab', { name: 'Уроки' }).click();
    await expect(
      selectedPanel(page).getByText(
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
    const track = await addAudioTrack(page, 'English dub', { videoOnly: true });
    await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });

    // Before S6 this lesson failed with audio_extraction_failed: the video source has no sound.
    // Delivery of a lesson's first draft into an open editor is outside S6 (D-S6-6), so the
    // result is read from the server and then shown after a reload.
    await expect
      .poll(async () => (await draftFor(page.request, course.lessonID, track.trackId))?.state, {
        timeout: 240_000,
        intervals: [2_000],
      })
      .toBe('ready');
    expect(
      (await draftFor(page.request, course.lessonID, track.trackId))?.segments.length,
    ).toBeGreaterThan(0);
    await page.reload();
    await page.getByRole('tab', { name: 'Уроки' }).click();
    await waitForDraft(selectedPanel(page), 'ready');
  });

  test('S6 every Ready track has its own draft; renaming makes its draft stale until Regenerate', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(900_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'per-track', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    await startRecognition(selectedPanel(page));
    await waitForDraft(selectedPanel(page), 'ready');
    const original = (await readDrafts(page.request, course.lessonID))[0];
    if (!original) throw new Error('Expected embedded draft');
    const track = await addAudioTrack(page, 'Dub');
    const panel = await selectTab(page, /^Англійська/);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    const dub = await draftFor(page.request, course.lessonID, track.trackId);
    expect(await readDrafts(page.request, course.lessonID)).toHaveLength(2);
    expect(dub?.segments.map((line) => line.text).join(' ')).toMatch(/dubbed/i);
    expect((await draftFor(page.request, course.lessonID, original.trackId))?.updatedAt).toBe(
      original.updatedAt,
    );
    await renameTrack(page, 'Dub', 'Французька');
    await expect(
      panel.getByText(
        'Чернетка розпізнана для іншої мови чи режиму доріжки — натисніть Regenerate',
      ),
    ).toBeVisible();
    await expect(
      panel.getByRole('button', { name: 'Завантажити чернетку в редактор', exact: true }),
    ).toHaveCount(0);
    const stale = await draftFor(page.request, course.lessonID, track.trackId);
    expect(stale?.stale).toBe(true);
    expect(stale?.updatedAt).toBe(dub?.updatedAt);
    await waitForRegeneratedDraft(page, panel);
    const regenerated = await draftFor(page.request, course.lessonID, track.trackId);
    expect(regenerated?.trackLanguage).toBe('fr');
    expect(regenerated?.updatedAt).not.toBe(dub?.updatedAt);
    expect(regenerated?.stale).toBe(false);
  });
  test('M3-R8 R2 R4 success and R6 two-track drafts, saves and isolated reload recovery', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(900_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'two-track', null);
    await openEditor(page, course.id);
    const assetId = await upload(page, speechVideo);
    const initialTracks = await page.request.get(`/api/v1/admin/videos/${assetId}/audio-tracks`);
    expect(initialTracks.status()).toBe(200);
    const embeddedTrack = z
      .object({ items: z.array(z.object({ trackId: z.string().uuid(), isDefault: z.boolean() })) })
      .parse(await initialTracks.json())
      .items.find((track) => track.isDefault);
    if (!embeddedTrack) throw new Error('Expected the embedded default track in media inventory');
    const dubTrack = await addAudioTrack(page, 'British dub', { british: true });
    await expect(
      page.getByRole('region', { name: 'Аудіодоріжки' }).getByText('en-GB', { exact: true }),
    ).toBeVisible();
    let panel = await selectTab(page, /^Невідома мова/);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    panel = await selectTab(page, 'en-gb');
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    const drafts = await readDrafts(page.request, course.lessonID);
    expect(drafts).toHaveLength(2);
    const embedded = drafts.find((draft) => draft.trackId === embeddedTrack.trackId);
    const dub = drafts.find((draft) => draft.trackId === dubTrack.trackId);
    if (!embedded || !dub) throw new Error('Both real track drafts must exist');
    expect(embedded.trackLanguage).toBe('und');
    expect(dub.trackLanguage).toBe('en-gb');
    const embeddedText = embedded.segments.map((line) => line.text).join(' ');
    const dubText = dub.segments.map((line) => line.text).join(' ');
    const embeddedTexts = embedded.segments.map((line) => line.text);
    const dubTexts = dub.segments.map((line) => line.text);
    expect(embeddedText).toMatch(/Welcome/i);
    expect(embeddedText).not.toMatch(/dubbed/i);
    expect(dubText).toMatch(/dubbed/i);
    expect(dubText).not.toMatch(/Welcome/i);

    panel = await selectTab(page, /^Невідома мова/);
    await loadDraft(panel);
    const beforeSave = network.length;
    const englishPut = await saveTranscript(page, panel, course.lessonID, true);
    expect(englishPut.language).toBe('en');
    expect(englishPut.segments.map((line) => line.text)).toEqual(embeddedTexts);
    const writes = network
      .slice(beforeSave)
      .filter((entry) => entry.method === 'PATCH' || entry.method === 'PUT');
    expect(writes.map((entry) => [entry.method, entry.status])).toEqual([
      ['PATCH', 200],
      ['PUT', 200],
    ]);
    await expect(
      page.getByRole('tablist', { name: 'Мови транскрипту' }).getByRole('tab', { selected: true }),
    ).toHaveText(/^Англійська/);
    panel = await selectTab(page, 'en-gb');
    await loadDraft(panel);
    const britishPut = await saveTranscript(page, panel, course.lessonID);
    expect(britishPut.language).toBe('en-gb');
    expect(britishPut.segments.map((line) => line.text)).toEqual(dubTexts);
    const inventory = await page.request.get(
      `/api/v1/admin/lessons/${course.lessonID}/transcripts`,
    );
    expect(inventory.status()).toBe(200);
    expect(
      z
        .object({ transcripts: z.array(z.object({ language: z.string() })) })
        .parse(await inventory.json())
        .transcripts.map((item) => item.language)
        .sort(),
    ).toEqual(['en', 'en-gb']);
    const tracks = await page.request.get(`/api/v1/admin/videos/${embedded.assetId}/audio-tracks`);
    expect(tracks.status()).toBe(200);
    const trackList = z
      .object({ items: z.array(z.object({ trackId: z.string(), language: z.string() })) })
      .parse(await tracks.json());
    expect(trackList.items.find((track) => track.trackId === embedded.trackId)?.language).toBe(
      'en',
    );
    await expectTranscriptTexts(page.request, course.lessonID, 'en', embeddedTexts);
    await expectTranscriptTexts(page.request, course.lessonID, 'en-gb', dubTexts);

    const own = 'Перший незбережений рядок';
    const other = 'Другий незбережений рядок';
    panel = await selectTab(page, /^Англійська/);
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(own);
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
    panel = await selectTab(page, 'en-gb');
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(other);
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
    const ownKey = `${adminID}:${course.lessonID}:track:${embedded.trackId}`;
    const otherKey = `${adminID}:${course.lessonID}:track:${dub.trackId}`;
    await expect.poll(() => snapshotKeys(page)).toEqual([ownKey, otherKey].sort());
    await openEditor(page, course.id);
    panel = await selectTab(page, /^Англійська/);
    await panel.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      own,
    );
    await expect(
      panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
    ).not.toHaveValue(other);
    panel = await selectTab(page, 'en-gb');
    // Restore proves this tab contains its own snapshot, then reload offers it again to Discard.
    await panel.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      other,
    );
    await expect(
      panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
    ).not.toHaveValue(own);
    await openEditor(page, course.id);
    panel = await selectTab(page, 'en-gb');
    await panel.getByRole('button', { name: 'Не відновлювати' }).click();
    await expect.poll(() => snapshotKeys(page)).toEqual([ownKey]);
    await openEditor(page, course.id);
    panel = await selectTab(page, 'en-gb');
    await expect(panel.getByRole('button', { name: 'Відновити зміни' })).toHaveCount(0);
    panel = await selectTab(page, /^Англійська/);
    await panel.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      own,
    );
    await expect.poll(() => snapshotKeys(page)).toEqual([ownKey]);
    await saveTranscript(page, panel, course.lessonID);
    await expect.poll(() => snapshotKeys(page)).toEqual([]);
  });

  test('M3-R4 PATCH succeeds and PUT fails, reload restores the track key and Save retries', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(600_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'put-failure', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    let panel = selectedPanel(page);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    const draft = (await readDrafts(page.request, course.lessonID))[0];
    if (!draft) throw new Error('Expected embedded draft');
    await loadDraft(panel);
    const text = 'Рядок переживає помилку збереження';
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(text);
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
    const endpoint = `**/api/v1/admin/lessons/${course.lessonID}/transcript`;
    await page.route(endpoint, (route) =>
      route.request().method() === 'PUT'
        ? route.fulfill({
            status: 503,
            contentType: 'application/json',
            body: JSON.stringify({ error: { code: 'dependency_unavailable' } }),
          })
        : route.continue(),
    );
    const failed = page.waitForResponse(
      (r) =>
        r.request().method() === 'PUT' &&
        r.url().endsWith(`/lessons/${course.lessonID}/transcript`),
    );
    const beforeSave = network.length;
    await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
    await confirmEnglish(page);
    expect((await failed).status()).toBe(503);
    await expect(panel.getByText('Сервіс недоступний. Спробуйте ще раз.')).toBeVisible();
    expect(
      network
        .slice(beforeSave)
        .filter((entry) => entry.method === 'PATCH' || entry.method === 'PUT')
        .map((entry) => [entry.method, entry.status]),
    ).toEqual([
      ['PATCH', 200],
      ['PUT', 503],
    ]);
    await expect(
      page.getByRole('tablist', { name: 'Мови транскрипту' }).getByRole('tab', { selected: true }),
    ).toHaveText(/^Англійська/);
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      text,
    );
    const key = `${adminID}:${course.lessonID}:track:${draft.trackId}`;
    await expect.poll(() => snapshotKeys(page)).toEqual([key]);
    await page.unroute(endpoint);
    await openEditor(page, course.id);
    panel = await selectTab(page, /^Англійська/);
    await panel.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      text,
    );
    const put = await saveTranscript(page, panel, course.lessonID);
    expect(put.language).toBe('en');
    expect(put.segments[0]?.text).toBe(text);
    await expect.poll(() => snapshotKeys(page)).toEqual([]);
  });

  test('M3-R14 a track minted with the multilingual switch gets a ready draft', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(600_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seed(page.request, baseURL, 'multilingual', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const track = await addAudioTrack(page, 'Multilingual dub', { multilingual: true });
    const panel = await selectTab(page, /Англійська.*Кілька мов/);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    const draft = await draftFor(page.request, course.lessonID, track.trackId);
    expect(draft?.trackMultilingual).toBe(true);
    expect(draft?.segments.length).toBeGreaterThan(0);
  });
  test('M3-R10 a foreign track rename after Load refuses Save without a transcript PUT', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(420_000);
    if (!baseURL) throw new Error('baseURL required');
    const course = await seed(page.request, baseURL, 'verify-rename', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const panel = selectedPanel(page);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    const draft = (await readDrafts(page.request, course.lessonID))[0];
    if (!draft) throw new Error('Missing embedded draft');
    await loadDraft(panel);
    const lines = await Promise.all(
      (await panel.getByRole('textbox', { name: /^Текст репліки/ }).all()).map((field) =>
        field.inputValue(),
      ),
    );
    const renamed = await page.request.patch(
      `/api/v1/admin/videos/${draft.assetId}/audio-tracks/${draft.trackId}`,
      {
        headers: origin(baseURL),
        data: { language: 'fr', label: 'Foreign rename', multilingual: false },
      },
    );
    expect(renamed.status(), await renamed.text()).toBe(200);
    const before = network.length;
    await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
    await expect(
      panel.getByText('Доріжку змінено або видалено. Оновіть доріжки й спробуйте ще.'),
    ).toBeVisible();
    expect(
      network.slice(before).filter((r) => r.method === 'PUT' && r.path.endsWith('/transcript')),
    ).toEqual([]);
    expect(
      await Promise.all(
        (await panel.getByRole('textbox', { name: /^Текст репліки/ }).all()).map((field) =>
          field.inputValue(),
        ),
      ),
    ).toEqual(lines);
  });

  test('M3-R10 GET tracks 503 keeps lines and sends no PUT; retry saves', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(420_000);
    if (!baseURL) throw new Error('baseURL required');
    const course = await seed(page.request, baseURL, 'verify-retry', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    const panel = selectedPanel(page);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    await loadDraft(panel);
    const lines = await Promise.all(
      (await panel.getByRole('textbox', { name: /^Текст репліки/ }).all()).map((field) =>
        field.inputValue(),
      ),
    );
    const endpoint = '**/api/v1/admin/videos/*/audio-tracks';
    await page.route(endpoint, (route) =>
      route.request().method() === 'GET'
        ? route.fulfill({
            status: 503,
            contentType: 'application/json',
            body: JSON.stringify({ error: { code: 'dependency_unavailable' } }),
          })
        : route.continue(),
    );
    const before = network.length;
    await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
    await expect(panel.getByText('Не вдалося перевірити доріжку — спробуйте ще')).toBeVisible();
    expect(
      network.slice(before).filter((r) => r.method === 'PUT' && r.path.endsWith('/transcript')),
    ).toEqual([]);
    expect(
      await Promise.all(
        (await panel.getByRole('textbox', { name: /^Текст репліки/ }).all()).map((field) =>
          field.inputValue(),
        ),
      ),
    ).toEqual(lines);
    await page.unroute(endpoint);
    const put = await saveTranscript(page, panel, course.lessonID, true);
    expect(put.segments.map((line) => line.text)).toEqual(lines);
    await expectTranscriptTexts(page.request, course.lessonID, 'en', lines);
  });

  test('M3-R11 a-d one track saves fresh ids through fr pl en and restored es without changing prior languages', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(600_000);
    if (!baseURL) throw new Error('baseURL required');
    const course = await seed(page.request, baseURL, 'new-segment-ids', null);
    await openEditor(page, course.id);
    await upload(page, speechVideo);
    let panel = selectedPanel(page);
    await startRecognition(panel);
    await waitForDraft(panel, 'ready');
    await loadDraft(panel);
    const initialDraft = (await readDrafts(page.request, course.lessonID))[0];
    if (!initialDraft) throw new Error('Expected the single embedded draft');
    const draftTexts = initialDraft.segments.map((line) => line.text);
    const en = await saveTranscript(page, panel, course.lessonID, true);
    expect(en.language).toBe('en');
    expect(en.segments.map((line) => line.text)).toEqual(draftTexts);
    let previous = en;
    const saved = new Map([['en', draftTexts]]);
    const check = async (put: typeof en, language: string, expectedTexts: string[]) => {
      expect(put.language).toBe(language);
      expect(put.segments.map((line) => line.text)).toEqual(expectedTexts);
      for (const line of put.segments)
        expect(previous.segments.map((s) => s.id)).not.toContain(line.id);
      for (const [prior, texts] of saved)
        if (prior !== language)
          await expectTranscriptTexts(page.request, course.lessonID, prior, texts);
      saved.set(language, expectedTexts);
      await expectTranscriptTexts(page.request, course.lessonID, language, expectedTexts);
      previous = put;
    };
    await renameEmbedded(page, 'Французька');
    await loadDraft(panel);
    await check(await saveTranscript(page, panel, course.lessonID), 'fr', draftTexts);
    // (b) reopen saved fr, edit, then move to an absent language.
    await openEditor(page, course.id);
    panel = await selectTab(page, /^Французька/);
    const polishText = 'Варіант польською';
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(polishText);
    await renameEmbedded(page, 'Польська');
    const polishTexts = [polishText, ...draftTexts.slice(1)];
    await check(await saveTranscript(page, panel, course.lessonID), 'pl', polishTexts);
    // (c) en exists as an orphan, so replacing it requires explicit confirmation.
    const replacementText = 'Заміна англійського тексту';
    await panel
      .getByRole('textbox', { name: 'Текст репліки 1', exact: true })
      .fill(replacementText);
    await renameEmbedded(page, 'Англійська');
    const replacing = page.waitForResponse(
      (r) =>
        r.request().method() === 'PUT' &&
        r.url().endsWith(`/lessons/${course.lessonID}/transcript`),
    );
    await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
    await expect(page.getByRole('dialog').getByText(/уже є — замінити/)).toBeVisible();
    await page.getByRole('dialog').getByRole('button', { name: 'Замінити', exact: true }).click();
    const replaced = await replacing;
    expect(replaced.status(), await replaced.text()).toBe(200);
    const replacementTexts = [replacementText, ...draftTexts.slice(1)];
    await check(
      transcriptPutSchema.parse(replaced.request().postDataJSON()),
      'en',
      replacementTexts,
    );
    // (d) capture a dirty rename, reload and restore the real track snapshot.
    const restoredText = 'Відновлений іспанський варіант';
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(restoredText);
    await renameEmbedded(page, 'Іспанська');
    await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
    await expect.poll(() => snapshotKeys(page)).not.toEqual([]);
    await openEditor(page, course.id);
    panel = await selectTab(page, /^Іспанська/);
    await panel.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true })).toHaveValue(
      restoredText,
    );
    const restoredTexts = [restoredText, ...draftTexts.slice(1)];
    await check(await saveTranscript(page, panel, course.lessonID), 'es', restoredTexts);
  });

  test('M3-R12a typing during first processing is offered on Ready without blur or reload', async ({
    page,
    baseURL,
  }) => {
    test.setTimeout(420_000);
    if (!baseURL) throw new Error('baseURL required');
    const course = await seed(page.request, baseURL, 'first-processing', null);
    await openEditor(page, course.id);
    let release = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const endpoint = `**/api/v1/admin/courses/${course.id}/videos`;
    await page.route(endpoint, async (route) => {
      const response = await route.fetch();
      const body = z
        .object({ items: z.array(z.object({ state: z.string() })) })
        .parse(await response.json());
      if (body.items.some((item) => item.state === 'ready')) await gate;
      await route.fulfill({ response });
    });
    const text = 'Написано під час першої обробки';
    try {
      await page
        .locator('input[type="file"]:not(.admin-audio-field input)')
        .first()
        .setInputFiles(speechVideo);
      await expect(
        page.locator('[data-video-state="queued"], [data-video-state="processing"]'),
      ).toBeVisible({ timeout: 30_000 });
      const panel = selectedPanel(page);
      await panel.getByRole('button', { name: 'Додати репліку', exact: true }).click();
      const field = panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true });
      await field.fill(text);
      await expect(field).toBeFocused();
      release();
      await expect(page.locator('[data-video-state="ready"]')).toBeVisible({ timeout: 240_000 });
      await expect(
        page
          .getByRole('tablist', { name: 'Мови транскрипту' })
          .getByRole('tab', { selected: true }),
      ).toHaveText(/^Невідома мова/);
      const readyPanel = selectedPanel(page);
      await readyPanel.getByRole('button', { name: 'Відновити зміни' }).click();
      await expect(
        readyPanel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
      ).toHaveValue(text);
    } finally {
      release();
      await page.unroute(endpoint);
    }
  });

  for (const matching of [true, false]) {
    test(`M3-R12b replacement offers detached lines in ${matching ? 'matching non-default' : 'default fallback'} track tab and saves without changing the original`, async ({
      page,
      baseURL,
    }) => {
      test.setTimeout(600_000);
      if (!baseURL) throw new Error('baseURL required');
      const course = await seed(
        page.request,
        baseURL,
        matching ? 'replace-matching' : 'replace-fallback',
        'es',
      );
      await openEditor(page, course.id);
      const oldAssetId = await upload(page, speechVideo);
      const readAsset = async (videoId: string) => {
        const inventory = await page.request.get(`/api/v1/admin/courses/${course.id}/videos`);
        expect(inventory.status()).toBe(200);
        const ready = z
          .object({ items: z.array(z.object({ videoId: z.string().uuid(), state: z.string() })) })
          .parse(await inventory.json())
          .items.filter((video) => video.videoId === videoId);
        expect(ready).toHaveLength(1);
        const asset = ready[0];
        if (!asset) throw new Error('Uploaded asset missing from media inventory');
        expect(asset.state).toBe('ready');
        return asset.videoId;
      };
      await readAsset(oldAssetId);
      await renameEmbedded(page, 'Іспанська');
      // Bind the saved es base before editing and renaming to a new destination language.
      await openEditor(page, course.id);
      await selectTab(page, /^Іспанська/);
      const original = await readTranscript(page.request, course.lessonID, 'es');
      const panel = selectedPanel(page);
      const text = matching ? 'Знімок тієї самої мови' : 'Знімок у default-вкладці';
      await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(text);
      await renameEmbedded(page, 'Німецька');
      await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).focus();
      await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
      await expect.poll(() => snapshotKeys(page)).not.toEqual([]);
      const [snapshotKey] = await snapshotKeys(page);
      if (!snapshotKey) throw new Error('Expected the dirty track snapshot');
      const readSnapshot = async () =>
        page.evaluate(
          (key) =>
            new Promise<unknown>((resolve, reject) => {
              const opening = indexedDB.open('courses-transcript-history');
              opening.onerror = () => reject(opening.error);
              opening.onsuccess = () => {
                const db = opening.result;
                const read = db
                  .transaction('snapshots', 'readonly')
                  .objectStore('snapshots')
                  .get(key);
                read.onsuccess = () => {
                  db.close();
                  resolve(read.result);
                };
                read.onerror = () => {
                  db.close();
                  reject(read.error);
                };
              };
            }),
          snapshotKey,
        );
      await expect
        .poll(readSnapshot)
        .toMatchObject({ language: 'de', baseLanguage: 'es', lines: [{ text }] });
      await test.info().attach('detached-source', {
        body: JSON.stringify(await readSnapshot()),
        contentType: 'application/json',
      });
      const newAssetId = await upload(page, silentVideo);
      // The old Ready asset remains visible while its replacement is processing.
      await expect(
        page
          .getByRole('tablist', { name: 'Мови транскрипту' })
          .getByRole('tab', { selected: true }),
      ).toHaveText(/^Невідома мова/, { timeout: 240_000 });
      await readAsset(newAssetId);
      expect(newAssetId).not.toBe(oldAssetId);
      await renameEmbedded(page, 'Французька');
      const added = await addAudioTrack(page, 'Replacement dub', {
        language: matching ? 'de' : 'en',
      });
      const inventory = await page.request.get(`/api/v1/admin/videos/${newAssetId}/audio-tracks`);
      expect(inventory.status()).toBe(200);
      const tracks = z
        .object({
          items: z.array(
            z.object({ trackId: z.string(), language: z.string(), isDefault: z.boolean() }),
          ),
        })
        .parse(await inventory.json()).items;
      expect(tracks).toHaveLength(2);
      expect(tracks.find((track) => track.isDefault)?.language).toBe('fr');
      expect(tracks.find((track) => track.trackId === added.trackId)).toMatchObject({
        language: matching ? 'de' : 'en',
        isDefault: false,
      });
      const defaultPanel = await selectTab(page, /^Французька/);
      if (matching)
        await expect(defaultPanel.getByRole('button', { name: 'Відновити зміни' })).toHaveCount(0);
      else
        await expect(defaultPanel.getByRole('button', { name: 'Відновити зміни' })).toBeVisible();
      const otherPanel = await selectTab(page, matching ? /^Німецька/ : /^Англійська/);
      if (!matching)
        await expect(otherPanel.getByRole('button', { name: 'Відновити зміни' })).toHaveCount(0);
      const replacementPanel = matching ? otherPanel : await selectTab(page, /^Французька/);
      await expect(replacementPanel.getByRole('button', { name: 'Відновити зміни' })).toBeVisible();
      await replacementPanel.getByRole('button', { name: 'Відновити зміни' }).click();
      await expect(
        replacementPanel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
      ).toHaveValue(text);
      const put = await saveTranscript(page, replacementPanel, course.lessonID);
      expect(put.language).toBe(matching ? 'de' : 'fr');
      expect(put.segments.map((line) => line.text)).toEqual([text]);
      await expectTranscriptTexts(page.request, course.lessonID, matching ? 'de' : 'fr', [text]);
      expect(await readTranscript(page.request, course.lessonID, 'es')).toEqual(original);
      await test.info().attach('replacement-identity', {
        body: JSON.stringify({ oldAssetId, newAssetId, tracks, destination: put.language }),
        contentType: 'application/json',
      });
    });
  }

  for (const restored of [false, true]) {
    test(`M3-R13 ${restored ? 'b restored T1 snapshot' : 'a open T1 editor'} refuses Save after foreign T2`, async ({
      page,
      baseURL,
    }) => {
      test.setTimeout(420_000);
      if (!baseURL) throw new Error('baseURL required');
      const course = await seed(
        page.request,
        baseURL,
        restored ? 'conflict-restored' : 'conflict-open',
        'en',
      );
      await openEditor(page, course.id);
      await upload(page, speechVideo);
      await renameEmbedded(page, 'Англійська');
      // Reopen to bind the saved en base T1 before either author edits.
      await openEditor(page, course.id);
      let panel = await selectTab(page, /^Англійська/);
      const own = 'Рядки редактора A поверх T1';
      await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(own);
      if (restored) {
        await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).blur();
        await expect.poll(() => snapshotKeys(page)).not.toEqual([]);
      }
      const t1 = await readTranscript(page.request, course.lessonID, 'en');
      const foreign = 'Чужий збережений текст T2';
      const t2 = await page.request.put(`/api/v1/admin/lessons/${course.lessonID}/transcript`, {
        headers: origin(baseURL),
        data: {
          language: 'en',
          expectedUpdatedAt: t1.updatedAt,
          segments: [{ id: randomUUID(), startMs: 0, endMs: 1000, text: foreign, documentIds: [] }],
        },
      });
      expect(t2.status(), await t2.text()).toBe(200);
      if (restored) {
        await openEditor(page, course.id);
        panel = await selectTab(page, /^Англійська/);
        await panel.getByRole('button', { name: 'Відновити зміни' }).click();
      }
      const response = page.waitForResponse(
        (r) =>
          r.request().method() === 'PUT' &&
          r.url().endsWith(`/lessons/${course.lessonID}/transcript`),
      );
      await panel.getByRole('button', { name: 'Зберегти транскрипт', exact: true }).click();
      const conflict = await response;
      expect(conflict.status()).toBe(409);
      expect(
        z.object({ error: z.object({ code: z.string() }) }).parse(await conflict.json()).error.code,
      ).toBe('transcript_modified');
      expect(
        z.object({ expectedUpdatedAt: z.string() }).parse(conflict.request().postDataJSON())
          .expectedUpdatedAt,
      ).toBe(t1.updatedAt);
      await expect(
        panel.getByText('Транскрипт змінили, поки ви редагували. Перезавантажте сторінку.'),
      ).toBeVisible();
      await expect(
        panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
      ).toHaveValue(own);
      await expectTranscriptTexts(page.request, course.lessonID, 'en', [foreign]);
    });
  }

  for (const language of ['en', 'und']) {
    test(`M3-R15 ${language} draft switch invalidates Load until Regenerate without replacing editor lines`, async ({
      page,
      baseURL,
    }) => {
      test.setTimeout(600_000);
      if (!baseURL) throw new Error('baseURL required');
      const course = await seed(page.request, baseURL, `mode-${language}`, null);
      await openEditor(page, course.id);
      await upload(page, speechVideo);
      const panel = selectedPanel(page);
      await startRecognition(panel);
      await waitForDraft(panel, 'ready');
      if (language === 'en') {
        await renameEmbedded(page, 'Англійська');
        await waitForRegeneratedDraft(page, panel);
      }
      const draft = (await readDrafts(page.request, course.lessonID))[0];
      if (!draft) throw new Error('Expected ready draft');
      expect(draft.trackLanguage).toBe(language);
      await panel.getByRole('button', { name: 'Додати репліку', exact: true }).click();
      const own = 'Авторський текст не замінюється перемикачем';
      await panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }).fill(own);
      const row = page.getByRole('region', { name: 'Аудіодоріжки' }).getByRole('listitem').first();
      const control = row.getByRole('switch');
      await control.focus();
      await control.press('Space');
      await expect(control).toBeChecked();
      await expect(
        panel.getByText(
          'Чернетка розпізнана для іншої мови чи режиму доріжки — натисніть Regenerate',
        ),
      ).toBeVisible();
      await expect(
        panel.getByRole('button', { name: 'Завантажити чернетку в редактор', exact: true }),
      ).toHaveCount(0);
      await expect(
        panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
      ).toHaveValue(own);
      await waitForRegeneratedDraft(page, panel);
      const regenerated = await draftFor(page.request, course.lessonID, draft.trackId);
      expect(regenerated?.updatedAt).not.toBe(draft.updatedAt);
      expect(regenerated?.trackMultilingual).toBe(true);
      await expect(
        panel.getByRole('button', { name: 'Завантажити чернетку в редактор', exact: true }),
      ).toBeEnabled();
      await expect(
        panel.getByRole('textbox', { name: 'Текст репліки 1', exact: true }),
      ).toHaveValue(own);
    });
  }
});
