import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

// Scoped by a per-run token so the catalog window this test reads cannot contain another run's
// publications: offset pagination is documented as unstable during publishing, so asserting a
// name at a global position would be asserting something the product never promised (B-13).
const runNamespace = `e2e-transcript-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
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

// Published courses live on after the test that made them, and the public catalog counts them.
// A recovery assertion elsewhere pins that count to the seed, so a journey that publishes and
// walks away breaks a test it never touches. Archiving removes a course from the catalog; a
// published course cannot be deleted at all (only drafts can), so archiving is the whole cleanup.
const published: { id: string; updatedAt: string }[] = [];

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
  const course = token((await created.json()) as object);

  const saved = await request.put(`/api/v1/admin/courses/${course.id}`, {
    data: {
      expectedUpdatedAt: course.updatedAt,
      course: draft,
      modules: [
        {
          id: moduleID,
          title: `Модуль ${runNamespace}`,
          lessons: [{ id: lessonID, title: `Урок ${runNamespace}`, description: '' }],
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
  published.push(token((await publishedResponse.json()) as object));

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

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined || published.length === 0) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    for (const course of published) {
      await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { expectedUpdatedAt: course.updatedAt, status: 'archived' },
        headers: { Origin: new URL(baseURL).origin },
      });
    }
  } finally {
    await context.close();
  }
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
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const headers = { Origin: new URL(baseURL).origin };
    const slug = `${runNamespace}-video`;

    const created = await page.request.post('/api/v1/admin/courses', {
      data: { ...draft, slug, title: `Субтитри ${runNamespace}` },
      headers,
    });
    expect(created.status(), await created.text()).toBe(201);
    const course = token((await created.json()) as object);

    const saved = await page.request.put(`/api/v1/admin/courses/${course.id}`, {
      data: {
        expectedUpdatedAt: course.updatedAt,
        course: { ...draft, slug, title: `Субтитри ${runNamespace}` },
        modules: [
          {
            id: videoModuleID,
            title: `Модуль ${runNamespace}`,
            lessons: [
              { id: videoLessonID, title: `Урок з відео ${runNamespace}`, description: '' },
            ],
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
    const minted = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/lessons/${videoLessonID}/video` &&
        response.request().method() === 'POST',
    );
    await page
      .getByRole('group', { name: 'Модуль 1' })
      .getByRole('group', { name: 'Урок 1' })
      .locator('input[type="file"]')
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

    const publishedResponse = await page.request.post(`/api/v1/admin/courses/${course.id}/status`, {
      data: { expectedUpdatedAt: tree.updatedAt, status: 'published' },
      headers,
    });
    expect(publishedResponse.status(), await publishedResponse.text()).toBe(200);
    published.push(token((await publishedResponse.json()) as object));

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

    const context = await openStudent(browser, baseURL);
    try {
      const studentPage = await context.newPage();
      await studentPage.goto(`/learn/${slug}/${videoLessonID}`);
      const transcript = studentPage.getByRole('region', { name: 'Транскрипт' });
      await expect(transcript.getByText('Перша секунда')).toBeVisible();

      // THE EXTERNAL ANCHOR. A round trip through our own parser proves only that we agree with
      // ourselves; here the browser parses the file, so the judge is not our code. This is the one
      // measurement in the phase that passes through none of it.
      const captions = await studentPage.evaluate(async () => {
        const video = document.querySelector('video');
        if (video === null) return { tracks: 0, cues: 0, kind: '' };
        const track = video.textTracks[0];
        if (track === undefined) return { tracks: video.textTracks.length, cues: 0, kind: '' };
        track.mode = 'showing';
        await new Promise((resolve) => setTimeout(resolve, 1_000));
        return { tracks: video.textTracks.length, cues: track.cues?.length ?? 0, kind: track.kind };
      });
      expect(captions.tracks, 'the video carries a caption track').toBe(1);
      expect(captions.kind).toBe('captions');
      expect(captions.cues, 'the browser parsed our WebVTT into cues').toBe(2);

      // Seeking becomes possible once the duration is known, and only then does a line become a
      // control. Keyboard traversal over real lines is measurable exactly here.
      await studentPage.getByRole('button', { name: 'Відтворити' }).click();
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
