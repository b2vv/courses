import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import { z } from 'zod';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';
import { deleteE2ECourse } from './course-cleanup';

// Scoped by a per-run token so the catalog window this test reads cannot contain another run's
// publications: offset pagination is documented as unstable during publishing, so asserting a
// name at a global position would be asserting something the product never promised (B-13).
const runNamespace = `e2e-transcript-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
// Module state is local to this Playwright worker, including when a failed test stops it.
const cleanupCourseIDs = new Set<string>();
const adminID = '40000000-0000-4000-8000-000000000003';

// The SHORTEST fixture the media setup generates: two seconds, so a real transcode costs seconds
// rather than the five minutes the playback journey pays. Two open measurements — the browser
// parsing our WebVTT, and keyboard traversal over real lines — both need a <video>, which only
// exists when a lesson has one. This is the cheapest lesson that truly has one.
const shortVideo = path.join(
  import.meta.dirname,
  '..',
  'playwright',
  '.fixtures',
  'e2e-640x360.mp4',
);

// The seeded student fixture (student@courseforge.local); after enrollment, content reads need membership.
const studentID = '40000000-0000-4000-8000-000000000001';

const moduleID = randomUUID();
const videoLessonID = randomUUID();
const videoModuleID = randomUUID();
const lessonID = randomUUID();
const firstSegment = randomUUID();
const secondSegment = randomUUID();

// instructorId belongs to the draft itself, not only to the create call: the tree save validates
// the owner too, and omitting it there answers invalid_owner.
const draft = {
  slug: `${runNamespace}`,
  title: `Транскрипт ${runNamespace}`,
  summary: 'Курс для браузерної перевірки транскрипту.',
  description: 'Повний опис курсу, створеного браузерною подорожжю транскрипту.',
  instructorId: adminID,
};

function token(payload: object): { id: string; updatedAt: string } {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string') {
    throw new Error('Admin response must carry id and updatedAt');
  }
  return { id, updatedAt };
}

async function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

/**
 * Everything this journey needs is created through the product's own API — course, module, lesson
 * and transcript. Writing rows behind the API's back would prove reading of data the product
 * cannot produce.
 *
 * Deliberately no video: the transcript surface does not depend on one, and a lesson without a
 * video keeps this spec out of the media chain, where a single fixture costs a five-minute
 * transcode. The player renders its empty state beside a fully working transcript.
 */
async function seed(request: APIRequestContext, baseURL: string) {
  const headers = await origin(baseURL);
  const created = await request.post('/api/v1/admin/courses', {
    data: draft,
    headers,
  });
  expect(created.status(), await created.text()).toBe(201);
  const body: unknown = await created.json();
  if (typeof body !== 'object' || body === null || !('id' in body) || typeof body.id !== 'string')
    throw new Error('Course response must carry id');
  cleanupCourseIDs.add(body.id);
  const course = token(body);

  const saved = await request.put(`/api/v1/admin/courses/${course.id}`, {
    data: {
      expectedUpdatedAt: course.updatedAt,
      course: draft,
      modules: [
        {
          id: moduleID,
          title: `Модуль ${runNamespace}`,
          lessons: [{ id: lessonID, title: `Урок ${runNamespace}` }],
        },
      ],
    },
    headers,
  });
  // The body travels into the failure message: a bare status tells you the seed broke but not
  // which rule it broke, and this seed has four steps that all answer 400 the same way.
  expect(saved.status(), await saved.text()).toBe(200);
  const tree = token((await saved.json()) as object);

  const publishedResponse = await request.post(`/api/v1/admin/courses/${course.id}/status`, {
    data: { expectedUpdatedAt: tree.updatedAt, status: 'published' },
    headers,
  });
  expect(publishedResponse.status(), await publishedResponse.text()).toBe(200);
  token((await publishedResponse.json()) as object);

  // After the enrollment phase, lesson content requires membership: enroll the reading student as
  // an (admin-granted) active member of this freshly created course. The fixture user is enrolled,
  // not the content gate loosened.
  const enrolled = await request.post(`/api/v1/admin/courses/${course.id}/grants`, {
    data: { userId: studentID, reason: 'e2e content access for the reading student' },
    headers,
  });
  expect(enrolled.status(), await enrolled.text()).toBe(201);

  const transcript = await request.put(`/api/v1/admin/lessons/${lessonID}/transcript`, {
    data: {
      language: 'uk',
      segments: [
        {
          id: firstSegment,
          startMs: 0,
          endMs: 5_000,
          text: `Перша репліка ${runNamespace} про канали`,
          documentIds: [],
        },
        {
          id: secondSegment,
          startMs: 65_000,
          endMs: 70_000,
          text: `Друга репліка ${runNamespace} про горутини`,
          documentIds: [],
        },
      ],
    },
    headers,
  });
  expect(transcript.status(), await transcript.text()).toBe(200);
  return course;
}

declare global {
  interface Window {
    axe: typeof axe;
  }
}

async function transcriptAxe(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () => {
    const region = document.querySelector('.transcript');
    if (region === null) return { violations: ['the transcript region is absent'], incomplete: [] };
    const results = await window.axe.run(region, {
      runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
    });
    const names = (items: typeof results.violations) =>
      items.flatMap((item) => item.nodes.map((node) => `${item.id} ${node.target.join(' ')}`));
    return { violations: names(results.violations), incomplete: names(results.incomplete) };
  });
}

test.afterAll(async () => {
  test.setTimeout(60_000);
  const results = await Promise.allSettled([...cleanupCourseIDs].map(deleteE2ECourse));
  const failures = results.flatMap((result) =>
    result.status === 'rejected' ? [result.reason] : [],
  );
  if (failures.length > 0) throw new AggregateError(failures, 'Course cleanup failed');
});

test.describe('with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('an author writes a transcript and a student reads it on the lesson', async ({
    page,
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    await seed(page.request, baseURL);

    // The student rides the SHARED session: the login budget is already at its ceiling, so a new
    // form login here would break B19 for every other spec in the run.
    const context = await openStudent(browser, baseURL);
    try {
      const studentPage = await context.newPage();
      await studentPage.goto(`/learn/${draft.slug}/${lessonID}`);

      const transcript = studentPage.getByRole('region', { name: 'Транскрипт' });
      await expect(transcript.getByRole('listitem').first()).toContainText('Перша репліка');
      // Timestamps are rendered for every line, whether or not seeking is possible.
      await expect(transcript.getByText('01:05')).toBeVisible();

      // Search runs over the whole lesson because the read is never paginated.
      await transcript.getByRole('searchbox').fill('горутини');
      await expect(transcript.getByRole('listitem')).toHaveCount(1);
      await expect(transcript.getByRole('listitem').first()).toContainText('Друга репліка');

      await transcript.getByRole('searchbox').fill('слова-якого-немає');
      await expect(transcript.getByText('Нічого не знайдено.')).toBeVisible();
      await transcript.getByRole('searchbox').fill('');

      // A25-A27 measured on the real surface rather than on a component in isolation: the panel
      // shares a page with the player, and contrast and focus are properties of that page.
      expect(await transcriptAxe(studentPage)).toEqual({ violations: [], incomplete: [] });

      // Without a video there is no seek, so a line is TEXT rather than a disabled button: a
      // disabled control drops out of keyboard traversal and would make the text itself
      // unreachable for exactly the readers a transcript exists for. Asserting the absence of a
      // dead control is the honest measurement here; the keyboard journey over real lines belongs
      // where a video makes seeking possible.
      await expect(transcript.getByRole('button', { name: /Перейти до/ })).toHaveCount(0);
      await expect(transcript.getByText('Перша репліка')).toBeVisible();

      // Dark theme and 200% text zoom must not cut the lines off or overlap them.
      await studentPage.emulateMedia({ colorScheme: 'dark' });
      await studentPage.setViewportSize({ width: 640, height: 720 });
      await studentPage.addStyleTag({ content: 'html { font-size: 200%; }' });
      await expect(transcript.getByRole('listitem').first()).toBeVisible();
      expect(await transcriptAxe(studentPage)).toEqual({ violations: [], incomplete: [] });

      // The WebVTT anchor is NOT here, and that is a finding rather than an omission: this
      // journey deliberately has no video so it stays out of the media chain, and a caption track
      // lives on the <video> element, which the player does not render for a lesson without one.
      // Measuring cues therefore requires a real transcode, so the anchor belongs to a lesson that
      // already pays for one. Asserting it here would mean asserting against an element that does
      // not exist — the serialiser would stay unjudged by anything but our own parser.
    } finally {
      await context.close();
    }
  });
});

async function openStudent(browser: Browser, baseURL: string) {
  return browser.newContext({ baseURL, storageState: authStatePath('student') });
}

test.describe('with a real two-second video', () => {
  test.use({ storageState: authStatePath('admin') });

  test('the browser parses our WebVTT, and lines are reachable by keyboard', async ({
    page,
    browser,
    baseURL,
  }) => {
    test.setTimeout(300_000);
    page.setDefaultTimeout(5_000);
    page.setDefaultNavigationTimeout(10_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const headers = { Origin: new URL(baseURL).origin };
    const slug = `${runNamespace}-video`;

    const created = await page.request.post('/api/v1/admin/courses', {
      data: { ...draft, slug, title: `Субтитри ${runNamespace}` },
      headers,
    });
    expect(created.status(), await created.text()).toBe(201);
    const body: unknown = await created.json();
    if (typeof body !== 'object' || body === null || !('id' in body) || typeof body.id !== 'string')
      throw new Error('Course response must carry id');
    cleanupCourseIDs.add(body.id);
    const course = token(body);

    const saved = await page.request.put(`/api/v1/admin/courses/${course.id}`, {
      data: {
        expectedUpdatedAt: course.updatedAt,
        course: { ...draft, slug, title: `Субтитри ${runNamespace}` },
        modules: [
          {
            id: videoModuleID,
            title: `Модуль ${runNamespace}`,
            lessons: [{ id: videoLessonID, title: `Урок з відео ${runNamespace}` }],
          },
        ],
      },
      headers,
    });
    expect(saved.status(), await saved.text()).toBe(200);
    const tree = token((await saved.json()) as object);

    // Uploaded through the editor, the way an author does it — the upload contract belongs to
    // Phase 5 and this journey must not invent a second way into it.
    await page.goto(`/admin/courses/${course.id}`);
    // Per-lesson video upload lives on the Lessons tab (view-3 layout).
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const lesson = page
      .getByRole('group', { name: 'Модуль 1', exact: true })
      .getByRole('group', { name: 'Урок 1', exact: true });
    await expect(
      lesson.getByRole('button', { name: /^(Завантажити відео|Замінити відео)$/ }),
    ).toBeVisible();
    const minted = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/lessons/${videoLessonID}/video` &&
        response.request().method() === 'POST',
    );
    await lesson
      .locator('input[type="file"]:not(.admin-audio-field input)')
      .setInputFiles(shortVideo);
    expect((await minted).status()).toBe(201);
    await expect
      .poll(
        async () => {
          const listed = await page.request.get(`/api/v1/admin/courses/${course.id}/videos`);
          if (listed.status() !== 200) return `http ${listed.status()}`;
          const body = (await listed.json()) as { items?: { state?: unknown }[] };
          return body.items?.[0]?.state ?? 'absent';
        },
        { timeout: 240_000, intervals: [1_000] },
      )
      .toBe('ready');

    // Keep the embedded default language until the populated Ukrainian orphan tab is activated.
    const videosResponse = await page.request.get(`/api/v1/admin/courses/${course.id}/videos`);
    expect(videosResponse.status()).toBe(200);
    const video = z
      .object({
        items: z.array(z.object({ videoId: z.string(), lessonId: z.string() })),
      })
      .parse(await videosResponse.json())
      .items.find((item) => item.lessonId === videoLessonID);
    if (!video) throw new Error('Missing Ready video');
    const tracksResponse = await page.request.get(
      `/api/v1/admin/videos/${video.videoId}/audio-tracks`,
    );
    expect(tracksResponse.status()).toBe(200);
    const embedded = z
      .object({
        items: z.array(z.object({ trackId: z.string(), kind: z.string() })),
      })
      .parse(await tracksResponse.json())
      .items.find((track) => track.kind === 'embedded');
    if (!embedded) throw new Error('Missing embedded track');

    const publishedResponse = await page.request.post(`/api/v1/admin/courses/${course.id}/status`, {
      data: { expectedUpdatedAt: tree.updatedAt, status: 'published' },
      headers,
    });
    expect(publishedResponse.status(), await publishedResponse.text()).toBe(200);
    token((await publishedResponse.json()) as object);

    // Enroll the reading student: lesson content needs membership after the enrollment phase.
    const videoEnrolled = await page.request.post(`/api/v1/admin/courses/${course.id}/grants`, {
      data: { userId: studentID, reason: 'e2e content access for the reading student' },
      headers,
    });
    expect(videoEnrolled.status(), await videoEnrolled.text()).toBe(201);

    const written = await page.request.put(`/api/v1/admin/lessons/${videoLessonID}/transcript`, {
      data: {
        language: 'uk',
        segments: [
          // Spaced across a second boundary: both lines inside the first second would render the
          // same 00:00 label, and the assertion below would then be ambiguous rather than wrong.
          { id: randomUUID(), startMs: 0, endMs: 900, text: 'Перша секунда', documentIds: [] },
          { id: randomUUID(), startMs: 1000, endMs: 1900, text: 'Друга секунда', documentIds: [] },
        ],
      },
      headers,
    });
    expect(written.status(), await written.text()).toBe(200);

    // F1: the Ready video keeps its default embedded track, so the populated Ukrainian
    // transcript starts in an inactive, force-mounted panel. Activating it used to hang
    // Chromium in the native container-query visibility check. A material preserves the
    // reported course composition; the single lesson starts expanded (D4).
    const material = await page.request.post(`/api/v1/admin/courses/${course.id}/documents`, {
      headers,
      data: {
        id: randomUUID(),
        label: 'Renderer regression material',
        url: 'https://example.test/transcript-renderer',
      },
    });
    expect(material.status()).toBe(201);
    await page.setViewportSize({ width: 390, height: 844 });
    await page.reload({ timeout: 10_000 });
    await page.getByRole('tab', { name: 'Уроки', exact: true }).click({ timeout: 5_000 });
    const ukrainian = lesson.getByRole('tab', { name: 'Українська', exact: true });
    await expect(ukrainian).toHaveAttribute('aria-selected', 'false', { timeout: 5_000 });
    await ukrainian.click({ timeout: 5_000 });
    await expect(ukrainian).toHaveAttribute('aria-selected', 'true', { timeout: 5_000 });
    await expect(page.locator('.transcript-editor-lines > li:visible')).toHaveCount(2, {
      timeout: 5_000,
    });
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(
      'Перша секунда',
      { timeout: 5_000 },
    );

    // D4 keeps the active populated panel mounted through collapse and expansion.
    const toggle = lesson.locator('.admin-lesson-summary button[aria-expanded][aria-controls]');
    await toggle.click({ timeout: 5_000 });
    await expect(toggle).toHaveAttribute('aria-expanded', 'false', { timeout: 5_000 });
    await toggle.click({ timeout: 5_000 });
    await expect(lesson.locator('.transcript-editor-lines > li:visible')).toHaveCount(2, {
      timeout: 5_000,
    });

    // O7 predicts the embedded language before Play; now match the authored transcript.
    const patched = await page.request.patch(
      `/api/v1/admin/videos/${video.videoId}/audio-tracks/${embedded.trackId}`,
      {
        headers,
        data: { language: 'uk', label: 'Українська' },
      },
    );
    expect(patched.status(), await patched.text()).toBe(200);

    const context = await openStudent(browser, baseURL);
    context.setDefaultTimeout(5_000);
    context.setDefaultNavigationTimeout(10_000);
    try {
      const studentPage = await context.newPage();
      await studentPage.goto(`/learn/${slug}/${videoLessonID}`);
      const transcript = studentPage.getByRole('region', { name: 'Транскрипт' });
      await expect(transcript.getByText('Перша секунда')).toBeVisible();

      // THE EXTERNAL ANCHOR. A round trip through our own parser proves only that we agree with
      // ourselves; here the browser parses the file, so the judge is not our code. This is the one
      // measurement in the phase that passes through none of it.
      await expect(studentPage.locator('video track')).toHaveCount(0);
      await studentPage.getByRole('button', { name: 'Відтворити' }).click();
      await expect(studentPage.locator('section.lesson-player')).toHaveAttribute(
        'data-caption-selection',
        'confirmed:uk',
      );
      await studentPage.getByRole('button', { name: 'Налаштування' }).click();
      const captionsToggle = studentPage.getByRole('switch', { name: 'Субтитри', exact: true });
      await studentPage.locator('label').filter({ has: captionsToggle }).click();
      await expect(captionsToggle).toBeChecked();
      await studentPage.keyboard.press('Escape');
      await expect
        .poll(() =>
          studentPage.evaluate(() => {
            const element = document.querySelector('video track');
            if (!(element instanceof HTMLTrackElement)) return null;
            return {
              language: element.srclang,
              mode: element.track.mode,
              kind: element.track.kind,
              cues: Array.from(element.track.cues ?? []).map((cue) =>
                cue instanceof VTTCue ? cue.text : '',
              ),
            };
          }),
        )
        .toEqual({
          language: 'uk',
          mode: 'showing',
          kind: 'captions',
          cues: ['Перша секунда', 'Друга секунда'],
        });

      // Seeking becomes possible once the duration is known.
      const firstLine = transcript.getByRole('button', { name: 'Перейти до 00:00' });
      await expect(firstLine).toBeVisible();
      await transcript.getByRole('searchbox').focus();
      await studentPage.keyboard.press('Tab');
      await expect(studentPage.locator(':focus')).toHaveAttribute('aria-label', /Перейти до/);
    } finally {
      await context.close();
    }
  });
});

// --- the author's own surface: the editor, and the local copy under it ----------------------
// Until now no browser journey reached the transcript EDITOR: every spec above writes the
// transcript through the API and then reads it as a student. That left two things unmeasured —
// the editor itself, and the IndexedDB adapter behind its crash recovery, which unit tests
// exercise only through the in-memory store. Both need a real browser, so both live here.

const editorSlug = `${runNamespace}-editor`;
const editorModuleID = randomUUID();
const editorLessonID = randomUUID();
const editorSegment = randomUUID();

const editorDraft = {
  slug: editorSlug,
  title: `Редактор ${runNamespace}`,
  summary: 'Курс для браузерної перевірки редактора транскрипту.',
  description: 'Повний опис курсу, створеного браузерною подорожжю редактора.',
  instructorId: adminID,
};

/**
 * A DRAFT course, never published: the editor is an authoring surface, and leaving this one out
 * of the catalog keeps the published-course count that `recovery.spec` pins untouched.
 */
async function seedDraftCourse(request: APIRequestContext, baseURL: string) {
  const headers = await origin(baseURL);
  const created = await request.post('/api/v1/admin/courses', { data: editorDraft, headers });
  expect(created.status(), await created.text()).toBe(201);
  const body: unknown = await created.json();
  if (typeof body !== 'object' || body === null || !('id' in body) || typeof body.id !== 'string')
    throw new Error('Course response must carry id');
  cleanupCourseIDs.add(body.id);
  const course = token(body);

  const saved = await request.put(`/api/v1/admin/courses/${course.id}`, {
    data: {
      expectedUpdatedAt: course.updatedAt,
      course: editorDraft,
      modules: [
        {
          id: editorModuleID,
          title: `Модуль ${runNamespace}`,
          lessons: [{ id: editorLessonID, title: `Урок ${runNamespace}` }],
        },
      ],
    },
    headers,
  });
  expect(saved.status(), await saved.text()).toBe(200);

  const transcript = await request.put(`/api/v1/admin/lessons/${editorLessonID}/transcript`, {
    data: {
      language: 'uk',
      segments: [
        { id: editorSegment, startMs: 0, endMs: 5_000, text: 'Серверний текст', documentIds: [] },
      ],
    },
    headers,
  });
  expect(transcript.status(), await transcript.text()).toBe(200);
  return course;
}

/** What the recovery adapter actually put on disk, read from the page's own origin. */
async function storedSnapshotKeys(page: Page): Promise<string[]> {
  return page.evaluate(
    () =>
      new Promise<string[]>((resolve, reject) => {
        const request = indexedDB.open('courses-transcript-history');
        request.addEventListener('error', () => {
          reject(request.error ?? new Error('open failed'));
        });
        request.addEventListener('success', () => {
          const database = request.result;
          if (!database.objectStoreNames.contains('snapshots')) {
            resolve([]);
            return;
          }
          const keys = database
            .transaction('snapshots', 'readonly')
            .objectStore('snapshots')
            .getAllKeys();
          keys.addEventListener('success', () => {
            resolve(keys.result.map(String));
          });
          keys.addEventListener('error', () => {
            reject(keys.error ?? new Error('read failed'));
          });
        });
      }),
  );
}

/** Read the persisted snapshot payload, not just its IndexedDB key. */
async function storedSnapshotLanguage(page: Page, key: string): Promise<string | undefined> {
  return page.evaluate(
    (snapshotKey) =>
      new Promise<string | undefined>((resolve, reject) => {
        const request = indexedDB.open('courses-transcript-history');
        request.addEventListener('error', () => {
          reject(request.error ?? new Error('open failed'));
        });
        request.addEventListener('success', () => {
          const database = request.result;
          if (!database.objectStoreNames.contains('snapshots')) {
            resolve(undefined);
            return;
          }
          const snapshot = database
            .transaction('snapshots', 'readonly')
            .objectStore('snapshots')
            .get(snapshotKey);
          snapshot.addEventListener('success', () => {
            const value = snapshot.result as { readonly language?: unknown } | undefined;
            resolve(typeof value?.language === 'string' ? value.language : undefined);
          });
          snapshot.addEventListener('error', () => {
            reject(snapshot.error ?? new Error('read failed'));
          });
        });
      }),
    key,
  );
}

test.describe('the transcript editor, with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('unsaved lines survive a reload and are offered back, then a save clears them', async ({
    page,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const course = await seedDraftCourse(page.request, baseURL);

    await page.goto(`/admin/courses/${course.id}`);
    // The per-lesson transcript editor lives on the Lessons tab (view-3 layout).
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const text = page.getByRole('textbox', { name: 'Текст репліки 1' });
    await expect(text).toHaveValue('Серверний текст');

    // Type, then leave the field. Leaving is what captures — the author is most likely to leave
    // the page next.
    await text.fill('Правка, яку ніхто не зберіг');
    await text.blur();

    // The adapter really wrote. The key carries the author and lesson; per the 2026-09-24
    // decision in work/archive/brief-a13-language.md, the language lives in the snapshot itself. This is
    // the assertion the unit tests cannot make: they run against the in-memory store.
    const ownSnapshotKey = `${adminID}:${editorLessonID}`;
    await expect.poll(() => storedSnapshotKeys(page)).toEqual([ownSnapshotKey]);
    await expect.poll(() => storedSnapshotLanguage(page, ownSnapshotKey)).toBe('uk');

    // The reload that used to lose the work without a word.
    await page.reload();
    // The editor tab is local view state, so a reload lands back on Overview — reopen Lessons.
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();

    // The server's text is what the editor shows: recovery OFFERS, it never restores by itself.
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(
      'Серверний текст',
    );
    await page.getByRole('button', { name: 'Відновити зміни' }).click();
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(
      'Правка, яку ніхто не зберіг',
    );

    await page.getByRole('button', { name: 'Зберегти транскрипт' }).click();
    await expect(page.getByText('Транскрипт збережено.')).toBeVisible();

    // Only a save the server accepted drops the local copy — and it drops it for real, not just
    // out of the prompt.
    await expect.poll(() => storedSnapshotKeys(page)).toEqual([]);

    await page.reload();
    await page.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    await expect(page.getByRole('textbox', { name: 'Текст репліки 1' })).toHaveValue(
      'Правка, яку ніхто не зберіг',
    );
    await expect(page.getByRole('button', { name: 'Відновити зміни' })).toHaveCount(0);

    // A snapshot outlives the session on disk, so a change of viewer must remove what is no
    // longer theirs. Proved by PLANTING a foreign snapshot rather than by signing out: this spec
    // rides the shared admin session, and signing out would revoke it for every other spec in the
    // run. Planting exercises the same sweep on the same real store.
    await page.evaluate(
      ([foreignKey, ownKey]) =>
        new Promise<void>((resolve, reject) => {
          const request = indexedDB.open('courses-transcript-history');
          request.addEventListener('error', () => {
            reject(request.error ?? new Error('open failed'));
          });
          request.addEventListener('success', () => {
            const transaction = request.result.transaction('snapshots', 'readwrite');
            const store = transaction.objectStore('snapshots');
            for (const key of [foreignKey, ownKey]) {
              store.put({ key, capturedAt: 1, expectedUpdatedAt: undefined, lines: [] });
            }
            transaction.addEventListener('complete', () => {
              resolve();
            });
            transaction.addEventListener('error', () => {
              reject(transaction.error ?? new Error('plant failed'));
            });
          });
        }),
      ['40000000-0000-4000-8000-000000000009:other-lesson', `${adminID}:other-lesson`],
    );

    await page.reload();
    await expect.poll(() => storedSnapshotKeys(page)).toEqual([`${adminID}:other-lesson`]);
  });
});
