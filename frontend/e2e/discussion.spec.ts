import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-discussion-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
// The seeded student fixture; after the enrollment phase, commenting needs active membership.
const studentID = '40000000-0000-4000-8000-000000000001';
const shortVideo = path.join(
  import.meta.dirname,
  '..',
  'playwright',
  '.fixtures',
  'e2e-640x360.mp4',
);

type CourseToken = { id: string; updatedAt: string };
const published: CourseToken[] = [];
const drafts: CourseToken[] = [];

function token(payload: object): CourseToken {
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

async function createCourse(
  request: APIRequestContext,
  baseURL: string,
  slug: string,
  lessonID: string,
  title: string,
): Promise<CourseToken> {
  const draft = {
    slug,
    title,
    summary: 'Run-scoped discussion course.',
    description: 'A course created by the discussion browser journey.',
    instructorId: instructorID,
  };
  const created = await request.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const course = token((await created.json()) as object);
  const moduleID = randomUUID();
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
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const result = token((await saved.json()) as object);
  drafts.push(result);
  return result;
}

async function publishVideo(page: Page, course: CourseToken, lessonID: string, slug: string) {
  await page.goto(`/admin/courses/${course.id}`);
  const minted = page.waitForResponse(
    (response) =>
      new URL(response.url()).pathname === `/api/v1/admin/lessons/${lessonID}/video` &&
      response.request().method() === 'POST',
  );
  await page.locator('input[type="file"]').first().setInputFiles(shortVideo);
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
  const response = await page.request.post(`/api/v1/admin/courses/${course.id}/status`, {
    data: { expectedUpdatedAt: course.updatedAt, status: 'published' },
    headers: origin(new URL(page.url()).origin),
  });
  expect(response.status(), await response.text()).toBe(200);
  published.push(token((await response.json()) as object));
  // Commenting requires active membership after the enrollment phase: admin-grant the student onto
  // this course (enroll the fixture, not loosen the gate).
  const enrolled = await page.request.post(`/api/v1/admin/courses/${course.id}/grants`, {
    data: { userId: studentID, reason: 'e2e discussion access for the commenting student' },
    headers: origin(new URL(page.url()).origin),
  });
  expect(enrolled.status(), await enrolled.text()).toBe(201);
  const draftIndex = drafts.findIndex((candidate) => candidate.id === course.id);
  if (draftIndex >= 0) drafts.splice(draftIndex, 1);
  await page.goto(`/learn/${slug}/${lessonID}`);
}

async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}

async function publishThroughUI(page: Page, body: string, withMoment = false) {
  const panel = page.getByRole('region', { name: /Discussion|Обговорення/ });
  const composer = panel.getByRole('textbox', { name: /Comment|Коментар/ }).last();
  if (withMoment) {
    await page.evaluate(() => {
      const video = document.querySelector('video');
      if (video === null) throw new Error('video is absent');
      video.currentTime = 1;
      video.dispatchEvent(new Event('timeupdate'));
    });
    await panel
      .getByRole('button', { name: /Attach current moment|Прив[’']язати до поточного моменту/ })
      .click();
  }
  await composer.fill(body);
  await panel.getByRole('button', { name: /Publish|Опублікувати/ }).click();
  await expect(panel.getByText(body)).toBeVisible();
}

async function commentCard(page: Page, body: string) {
  const panel = page.getByRole('region', { name: /Discussion|Обговорення/ });
  return panel.locator('article').filter({ hasText: body }).first();
}

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    for (const course of published) {
      const response = await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { expectedUpdatedAt: course.updatedAt, status: 'archived' },
        headers: origin(baseURL),
      });
      expect(response.status(), await response.text()).toBe(200);
    }
    for (const course of drafts) {
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

test.describe('discussion branches', () => {
  test('K19 publishes a timestamped comment, seeks, and replies', async ({ browser, baseURL }) => {
    test.setTimeout(300_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lessonID = randomUUID();
    const slug = `${runNamespace}-published`;
    const admin = await openAs(browser, baseURL, 'admin');
    const student = await openAs(browser, baseURL, 'student');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        slug,
        lessonID,
        `Обговорення ${runNamespace}`,
      );
      await publishVideo(await admin.newPage(), course, lessonID, slug);
      const studentPage = await student.newPage();
      await studentPage.goto(`/learn/${slug}/${lessonID}`);
      await expect(studentPage.locator('video')).toBeVisible();
      const root = `K19 root ${runNamespace}`;
      await publishThroughUI(studentPage, root, true);

      const instructorPage = await instructor.newPage();
      await instructorPage.goto(`/learn/${slug}/${lessonID}`);
      const rootCard = await commentCard(instructorPage, root);
      await expect(rootCard).toBeVisible();
      await expect(rootCard.getByText('00:01')).toBeVisible();
      await expect(rootCard.getByRole('button', { name: /00:01/ })).toHaveCount(0);
      await instructorPage.getByRole('button', { name: 'Відтворити' }).click();
      const timestamp = rootCard.getByRole('button', { name: /00:01/ });
      await timestamp.click();
      await expect
        .poll(() =>
          instructorPage
            .locator('video')
            .evaluate((element) =>
              element instanceof HTMLVideoElement ? element.currentTime : -1,
            ),
        )
        .toBeGreaterThanOrEqual(0.9);
      await rootCard.getByRole('button', { name: /Reply|Відповісти/ }).click();
      await rootCard
        .getByRole('textbox', { name: /Reply|Відповідь/ })
        .fill(`K19 reply ${runNamespace}`);
      await rootCard.getByRole('button', { name: /Send reply|Надіслати відповідь/ }).click();
      await expect(rootCard.getByText(`K19 reply ${runNamespace}`)).toBeVisible();
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('K20 enforces author, course-owner, and admin moderation', async ({ browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lessonID = randomUUID();
    const slug = `${runNamespace}-moderation`;
    const admin = await openAs(browser, baseURL, 'admin');
    const student = await openAs(browser, baseURL, 'student');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        slug,
        lessonID,
        `Модерація ${runNamespace}`,
      );
      await publishVideo(await admin.newPage(), course, lessonID, slug);
      const studentPage = await student.newPage();
      await studentPage.goto(`/learn/${slug}/${lessonID}`);
      const discussion = studentPage.getByRole('region', { name: /Discussion|Обговорення/ });
      const own = `K20 own ${runNamespace}`;
      await publishThroughUI(studentPage, own);
      const ownCard = await commentCard(studentPage, own);
      await ownCard.getByRole('button', { name: /Edit|Змінити/ }).click();
      const editForm = discussion.locator('.discussion-edit-form');
      await editForm
        .getByRole('textbox', { name: /Comment|Коментар/ })
        .fill(`K20 edited ${runNamespace}`);
      await editForm.getByRole('button', { name: /Save|Зберегти/ }).click();
      const editedCard = await commentCard(studentPage, `K20 edited ${runNamespace}`);
      await expect(editedCard).toBeVisible();
      await editedCard.getByRole('button', { name: /Remove|Прибрати/ }).click();
      await expect(studentPage.getByText(own)).toHaveCount(0);

      const root = `K20 owner target ${runNamespace}`;
      await publishThroughUI(studentPage, root);
      const instructorPage = await instructor.newPage();
      await instructorPage.goto(`/learn/${slug}/${lessonID}`);
      const rootCard = await commentCard(instructorPage, root);
      await rootCard.getByRole('button', { name: /Reply|Відповісти/ }).click();
      await rootCard
        .getByRole('textbox', { name: /Reply|Відповідь/ })
        .fill(`K20 owner reply ${runNamespace}`);
      await rootCard.getByRole('button', { name: /Send reply|Надіслати відповідь/ }).click();
      await expect(rootCard.getByText(`K20 owner reply ${runNamespace}`)).toBeVisible();
      await rootCard.getByRole('button', { name: /Remove|Прибрати/ }).click();
      await expect(instructorPage.getByText(/Comment removed|Коментар видалено/)).toBeVisible();
      await expect(rootCard.getByText(root)).toHaveCount(0);

      const adminRoot = `K20 admin target ${runNamespace}`;
      await publishThroughUI(instructorPage, adminRoot);
      const adminPage = await admin.newPage();
      await adminPage.goto(`/learn/${slug}/${lessonID}`);
      const adminCard = await commentCard(adminPage, adminRoot);
      await adminCard.getByRole('button', { name: /Remove|Прибрати/ }).click();
      await expect(adminPage.getByText(adminRoot)).toHaveCount(0);
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('K21 gives the same access denial and 403 for discussion', async ({ browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lessonID = randomUUID();
    const slug = `${runNamespace}-draft`;
    const admin = await openAs(browser, baseURL, 'admin');
    const student = await openAs(browser, baseURL, 'student');
    try {
      await createCourse(admin.request, baseURL, slug, lessonID, `Чернетка ${runNamespace}`);
      const studentPage = await student.newPage();
      await studentPage.goto(`/learn/${slug}`);
      await expect(
        studentPage.getByRole('heading', {
          name: /Немає доступу до курсу|Course access unavailable/,
        }),
      ).toBeVisible();
      const response = await student.request.get(`/api/v1/learn/lessons/${lessonID}/comments`);
      expect(response.status()).toBe(403);
      const errorBody = (await response.json()) as { error?: { code?: string } };
      expect(errorBody.error?.code).toBe('course_forbidden');
    } finally {
      await student.close();
      await admin.close();
    }
  });

  test('K22 keeps discussion keyboard accessible, sized, zoom-safe, and axe-clean', async ({
    browser,
    baseURL,
  }) => {
    test.setTimeout(300_000);
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lessonID = randomUUID();
    const slug = `${runNamespace}-a11y`;
    const admin = await openAs(browser, baseURL, 'admin');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        slug,
        lessonID,
        `A11y ${runNamespace}`,
      );
      await publishVideo(await admin.newPage(), course, lessonID, slug);
      const page = await student.newPage();
      await page.route('**/api/v1/learn/lessons/*/comments*', async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 500));
        await route.continue();
      });
      await page.goto(`/learn/${slug}/${lessonID}`);
      const panel = page.getByRole('region', { name: /Discussion|Обговорення/ });
      await expect(panel.getByRole('status')).toBeVisible();
      await expect(panel).toBeVisible();
      const textBox = panel.getByRole('textbox', { name: /Comment|Коментар/ }).last();
      await textBox.focus();
      await textBox.press('Tab');
      await expect(
        panel.getByRole('button', {
          name: /Attach current moment|Прив[’']язати до поточного моменту/,
        }),
      ).toBeFocused();
      await page.keyboard.press('Enter');
      await expect(panel.getByText(/Attached at|Прив[’']язано до/)).toBeVisible();
      for (const button of await panel.getByRole('button').all()) {
        const box = await button.boundingBox();
        if (box === null) throw new Error('Discussion button is not rendered');
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.height).toBeGreaterThanOrEqual(44);
      }
      await page.addScriptTag({ content: axe.source });
      for (const scheme of ['light', 'dark'] as const) {
        await page.emulateMedia({ colorScheme: scheme });
        const result = await page.evaluate(async () => {
          const region = document.querySelector('.discussion-panel');
          if (region === null) return { violations: ['discussion panel absent'] };
          const audit = await window.axe.run(region, {
            runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          });
          return {
            violations: audit.violations.flatMap((item) =>
              item.nodes.map((node) => `${item.id} ${node.target.join(' ')}`),
            ),
          };
        });
        expect(result.violations, `${scheme} discussion axe`).toEqual([]);
      }
      await page.addStyleTag({ content: 'html { font-size: 200%; }' });
      expect(await panel.evaluate((node) => node.scrollWidth <= node.clientWidth + 1)).toBe(true);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
      ).toBe(true);
    } finally {
      await student.close();
      await admin.close();
    }
  });
});
