import { randomUUID } from 'node:crypto';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type {
  AdminCourseTree,
  Comment,
  ErrorResponse,
  ErrorResponseError,
} from '../src/adapters/api/generated/models';
import { authFixtures, authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-instructor-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
// The seeded student fixture; after the enrollment phase, reporting/commenting needs active membership.
const studentID = '40000000-0000-4000-8000-000000000001';

type CourseToken = { id: string; updatedAt: string };
type LessonFixture = { id: string; title: string };
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

function commentID(payload: object): string {
  const id = contractField<Comment>(payload, 'id');
  if (typeof id !== 'string') throw new Error('Comment response must carry id');
  return id;
}

function errorCode(payload: unknown): string {
  if (typeof payload !== 'object' || payload === null) {
    throw new Error('Error response must be an object');
  }
  const error = contractField<ErrorResponse>(payload, 'error');
  if (typeof error !== 'object' || error === null) {
    throw new Error('Error response must carry an error object');
  }
  const code = contractField<ErrorResponseError>(error, 'code');
  if (typeof code !== 'string') throw new Error('Error response must carry a code');
  return code;
}

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}

async function createCourse(
  request: APIRequestContext,
  baseURL: string,
  slug: string,
  title: string,
  lessons: readonly LessonFixture[],
): Promise<CourseToken> {
  const draft = {
    slug,
    title,
    summary: 'Run-scoped instructor workspace course.',
    description: 'A course created by the instructor workspace browser journey.',
    instructorId: instructorID,
  };
  const created = await request.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const course = token((await created.json()) as object);
  const saved = await request.put(`/api/v1/admin/courses/${course.id}`, {
    data: {
      expectedUpdatedAt: course.updatedAt,
      course: draft,
      modules: [
        {
          id: randomUUID(),
          title: `Модуль ${runNamespace}`,
          lessons: lessons.map((lesson) => ({ ...lesson, description: '' })),
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

async function publishCourse(
  request: APIRequestContext,
  baseURL: string,
  course: CourseToken,
): Promise<CourseToken> {
  const response = await request.post(`/api/v1/admin/courses/${course.id}/status`, {
    data: { expectedUpdatedAt: course.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(200);
  const result = token((await response.json()) as object);
  published.push(result);
  // The student reports/comments on these courses in the moderation journeys; reporting is
  // communication now, so enroll the student as an active member (enroll the fixture, not loosen).
  const enrolled = await request.post(`/api/v1/admin/courses/${result.id}/grants`, {
    data: { userId: studentID, reason: `instructor ${runNamespace}` },
    headers: origin(baseURL),
  });
  expect(enrolled.status(), await enrolled.text()).toBe(201);
  const draftIndex = drafts.findIndex((candidate) => candidate.id === course.id);
  if (draftIndex >= 0) drafts.splice(draftIndex, 1);
  return result;
}

async function publishComment(
  request: APIRequestContext,
  baseURL: string,
  lessonID: string,
  body: string,
  positionMs: number | null = null,
): Promise<string> {
  const response = await request.post(`/api/v1/learn/lessons/${lessonID}/comments`, {
    data: { body, positionMs },
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(201);
  return commentID((await response.json()) as object);
}

async function reportComment(
  request: APIRequestContext,
  baseURL: string,
  commentId: string,
): Promise<void> {
  const response = await request.post(`/api/v1/learn/comments/${commentId}/reports`, {
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(201);
}

function workspace(page: Page) {
  return page.locator('.teach-workspace');
}

function overview(page: Page) {
  return workspace(page).getByRole('region', { name: 'Коментарі курсу' });
}

function moderationQueue(page: Page) {
  return workspace(page).getByRole('region', { name: 'Черга модерації' });
}

function discussionCard(page: Page, body: string) {
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

test.describe('instructor workspace', () => {
  test('W01 shows comments newest-first with lesson, author, and post time', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const firstLesson = { id: randomUUID(), title: `Перший урок ${runNamespace}` };
    const secondLesson = { id: randomUUID(), title: `Другий урок ${runNamespace}` };
    const lessons = [firstLesson, secondLesson];
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-overview`,
        `Огляд ${runNamespace}`,
        lessons,
      );
      await publishCourse(admin.request, baseURL, course);
      const older = `W01 older ${runNamespace}`;
      const newer = `W01 newer ${runNamespace}`;
      await publishComment(instructor.request, baseURL, firstLesson.id, older);
      await new Promise<void>((resolve) => setTimeout(resolve, 20));
      await publishComment(instructor.request, baseURL, secondLesson.id, newer);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const rows = overview(page).locator('article');
      await expect(rows).toHaveCount(2);
      await expect(rows.nth(0)).toContainText(newer);
      await expect(rows.nth(0)).toContainText(secondLesson.title);
      await expect(rows.nth(0)).toContainText(authFixtures.instructor.name);
      await expect(rows.nth(1)).toContainText(older);
      await expect(rows.nth(1)).toContainText(firstLesson.title);
      await expect(rows.nth(1)).toContainText(authFixtures.instructor.name);
      const newerTime = rows.nth(0).locator('time');
      const olderTime = rows.nth(1).locator('time');
      await expect(newerTime).toHaveCount(1);
      await expect(olderTime).toHaveCount(1);
      await expect(newerTime).toBeVisible();
      await expect(olderTime).toBeVisible();
      await expect(newerTime).not.toHaveText('');
      await expect(olderTime).not.toHaveText('');
      const newerCreatedAt = await newerTime.getAttribute('datetime');
      const olderCreatedAt = await olderTime.getAttribute('datetime');
      if (newerCreatedAt === null || olderCreatedAt === null) {
        throw new Error('W01 rows must expose their publication time');
      }
      expect(Date.parse(newerCreatedAt)).toBeGreaterThan(Date.parse(olderCreatedAt));
    } finally {
      await instructor.close();
      await admin.close();
    }
  });

  test('W03/W04 keeps a duplicate report singular and orders the queue by report count', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lesson = { id: randomUUID(), title: `Репорти ${runNamespace}` };
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const slug = `${runNamespace}-report`;
      const course = await createCourse(admin.request, baseURL, slug, `Репорти ${runNamespace}`, [
        lesson,
      ]);
      await publishCourse(admin.request, baseURL, course);
      const body = `W03 target ${runNamespace}`;
      const id = await publishComment(instructor.request, baseURL, lesson.id, body);

      const studentPage = await student.newPage();
      await studentPage.goto(`/learn/${slug}/${lesson.id}`);
      const card = discussionCard(studentPage, body);
      const firstReport = studentPage.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          url.pathname === `/api/v1/learn/comments/${id}/reports` &&
          response.request().method() === 'POST'
        );
      });
      await card.getByRole('button', { name: /Report|Поскаржитись/ }).click();
      expect((await firstReport).status()).toBe(201);
      await expect(card.getByRole('button', { name: /Reported|Поскаржено/ })).toBeDisabled();
      await reportComment(student.request, baseURL, id);

      const priorityBody = `W04 priority ${runNamespace}`;
      const priorityID = await publishComment(instructor.request, baseURL, lesson.id, priorityBody);
      await reportComment(student.request, baseURL, priorityID);
      await reportComment(admin.request, baseURL, priorityID);

      const instructorPage = await instructor.newPage();
      await instructorPage.goto(`/teach/courses/${course.id}`);
      const queued = moderationQueue(instructorPage).locator('article');
      await expect(queued).toHaveCount(2);
      await expect(queued.nth(0)).toContainText(priorityBody);
      await expect(queued.nth(0)).toContainText('Скарг: 2');
      await expect(queued.nth(1)).toContainText(body);
      await expect(queued.nth(1)).toContainText('Скарг: 1');
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('W05 removes a queued comment into a placeholder and removes it from the queue', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lesson = { id: randomUUID(), title: `Прибирання ${runNamespace}` };
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-remove`,
        `Прибирання ${runNamespace}`,
        [lesson],
      );
      await publishCourse(admin.request, baseURL, course);
      const body = `W05 target ${runNamespace}`;
      const id = await publishComment(admin.request, baseURL, lesson.id, body);
      await reportComment(student.request, baseURL, id);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const queued = moderationQueue(page).locator('article').filter({ hasText: body });
      await queued.getByRole('button', { name: 'Прибрати коментар' }).click();
      await expect(queued).toHaveCount(0);
      const row = overview(page).locator('article').filter({ hasText: authFixtures.admin.name });
      await expect(row).toContainText(/Comment removed|Коментар видалено/);
      await expect(row).not.toContainText(body);
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('W06 dismisses a report with a reason while leaving the comment visible', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lesson = { id: randomUUID(), title: `Відхилення ${runNamespace}` };
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-dismiss`,
        `Відхилення ${runNamespace}`,
        [lesson],
      );
      await publishCourse(admin.request, baseURL, course);
      const body = `W06 target ${runNamespace}`;
      const id = await publishComment(instructor.request, baseURL, lesson.id, body);
      await reportComment(student.request, baseURL, id);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const queued = moderationQueue(page).locator('article').filter({ hasText: body });
      const dismissed = page.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          url.pathname === `/api/v1/teach/comments/${id}/dismiss-report` &&
          response.request().method() === 'POST'
        );
      });
      await queued.getByRole('button', { name: 'Відхилити скарги' }).click();
      const response = await dismissed;
      expect(response.status()).toBe(204);
      expect(response.request().postData()).toContain('"reason":"moderation_reviewed"');
      await expect(queued).toHaveCount(0);
      await expect(overview(page).getByText(body)).toBeVisible();
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('W07/W08 shows scoped empty states for an empty course and moderation queue', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-empty`,
        `Порожній ${runNamespace}`,
        [{ id: randomUUID(), title: `Порожній урок ${runNamespace}` }],
      );
      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      await expect(overview(page).getByText('Коментарів у курсі ще немає.')).toBeVisible();
      await expect(overview(page).locator('article')).toHaveCount(0);
      await expect(moderationQueue(page).getByText('Відкритих скарг немає.')).toBeVisible();
      await expect(moderationQueue(page).locator('article')).toHaveCount(0);
    } finally {
      await instructor.close();
      await admin.close();
    }
  });

  test('W14 gives a student the named not-found UI and course_forbidden API answer', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-forbidden`,
        `Заборонений ${runNamespace}`,
        [{ id: randomUUID(), title: `Закритий урок ${runNamespace}` }],
      );
      const page = await student.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      await expect(page.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeVisible();
      await expect(page.getByRole('heading', { name: 'Робоче місце інструктора' })).toHaveCount(0);
      const response = await student.request.get(`/api/v1/teach/courses/${course.id}/comments`);
      expect(response.status()).toBe(403);
      expect(errorCode(await response.json())).toBe('course_forbidden');
    } finally {
      await student.close();
      await admin.close();
    }
  });

  test('W16 lets an admin open an instructor-owned course workspace', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-admin`,
        `Адмін ${runNamespace}`,
        [{ id: randomUUID(), title: `Адмін-урок ${runNamespace}` }],
      );
      const page = await admin.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      await expect(page.getByRole('heading', { name: 'Робоче місце інструктора' })).toBeVisible();
      await expect(overview(page).getByText('Коментарів у курсі ще немає.')).toBeVisible();
      await expect(moderationQueue(page).getByText('Відкритих скарг немає.')).toBeVisible();
    } finally {
      await admin.close();
    }
  });

  test('W21 keeps the teach panels keyboard-visible, sized, zoom-safe, themed, and axe-clean', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const lesson = { id: randomUUID(), title: `A11y ${runNamespace}` };
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-a11y`,
        `A11y ${runNamespace}`,
        [lesson],
      );
      await publishCourse(admin.request, baseURL, course);
      const body = `W21 target ${runNamespace}`;
      const id = await publishComment(instructor.request, baseURL, lesson.id, body);
      await reportComment(student.request, baseURL, id);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const panel = moderationQueue(page);
      const dismiss = panel.getByRole('button', { name: 'Відхилити скарги' });
      const remove = panel.getByRole('button', { name: 'Прибрати коментар' });
      await dismiss.focus();
      await dismiss.press('Tab');
      await expect(remove).toBeFocused();
      const focus = await remove.evaluate((element) => {
        const style = getComputedStyle(element);
        return { style: style.outlineStyle, width: Number.parseFloat(style.outlineWidth) };
      });
      expect(focus.style).not.toBe('none');
      expect(focus.width).toBeGreaterThanOrEqual(2);
      for (const button of await workspace(page).getByRole('button').all()) {
        const box = await button.boundingBox();
        if (box === null) throw new Error('Teach button is not rendered');
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.height).toBeGreaterThanOrEqual(44);
      }

      await page.addScriptTag({ content: axe.source });
      for (const scheme of ['light', 'dark'] as const) {
        await page.emulateMedia({ colorScheme: scheme });
        const result = await page.evaluate(async () => {
          const region = document.querySelector('.teach-workspace');
          if (region === null) return { violations: ['teach workspace absent'] };
          const audit = await window.axe.run(region, {
            runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          });
          return {
            violations: audit.violations.flatMap((item) =>
              item.nodes.map((node) => `${item.id} ${node.target.join(' ')}`),
            ),
          };
        });
        expect(result.violations, `${scheme} teach axe`).toEqual([]);
      }
      await page.addStyleTag({ content: 'html { font-size: 200%; }' });
      expect(
        await workspace(page).evaluate((node) => node.scrollWidth <= node.clientWidth + 1),
      ).toBe(true);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
      ).toBe(true);
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });
});
