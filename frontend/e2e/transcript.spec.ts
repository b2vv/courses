import { randomUUID } from 'node:crypto';
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

const moduleID = randomUUID();
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

  const published = await request.post(`/api/v1/admin/courses/${course.id}/status`, {
    data: { expectedUpdatedAt: tree.updatedAt, status: 'published' },
    headers,
  });
  expect(published.status(), await published.text()).toBe(200);

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
