import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promisify } from 'node:util';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree, Comment } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

const execute = promisify(execFile);
const runNamespace = `e2e-learning-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
const studentID = '40000000-0000-4000-8000-000000000001';

function literal(value: string) {
  return `'${value.replaceAll("'", "''")}'`;
}
// Fixture-only: seed progress so the student is review-eligible. Review and comment WRITES below
// all take the production HTTP path; only this precondition uses Compose psql (as ratings.spec does).
async function sql(statement: string) {
  await execute('docker', [
    'compose',
    'exec',
    '-T',
    'postgres',
    'psql',
    '-v',
    'ON_ERROR_STOP=1',
    '-U',
    'courses',
    '-d',
    'courses',
    '-c',
    statement,
  ]);
}
// Put a comment or review into the moderation queue without a second logged-in member: the open
// report is a fixture precondition (reporter is the instructor), exactly as ratings.spec seeds rows
// it cannot make through a public account. The block path itself runs through the real UI + API.
async function seedCommentReport(commentId: string) {
  await sql(
    `INSERT INTO discussion_comment_reports(comment_id,reporter_id) VALUES (${literal(commentId)}::uuid,${literal(instructorID)}::uuid) ON CONFLICT (comment_id,reporter_id) DO NOTHING`,
  );
}
async function seedReportedReview(courseId: string, reviewId: string, body: string) {
  await sql(
    `INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) VALUES (${literal(reviewId)}::uuid,${literal(courseId)}::uuid,${literal(studentID)}::uuid,2,${literal(body)},0,1,now())`,
  );
  await sql(
    `INSERT INTO discussion_review_reports(review_id,reporter_id) VALUES (${literal(reviewId)}::uuid,${literal(instructorID)}::uuid) ON CONFLICT (review_id,reporter_id) DO NOTHING`,
  );
}

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

async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}

async function createCourse(
  request: APIRequestContext,
  baseURL: string,
  slug: string,
  title: string,
  lessonId: string = randomUUID(),
): Promise<CourseToken> {
  const draft = {
    slug,
    title,
    summary: 'Run-scoped learning overview course.',
    description: 'A course created by the learner-overview browser journey.',
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
          lessons: [{ id: lessonId, title: `Урок ${runNamespace}` }],
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

async function publishComment(
  request: APIRequestContext,
  baseURL: string,
  lessonId: string,
  body: string,
): Promise<string> {
  const response = await request.post(`/api/v1/learn/lessons/${lessonId}/comments`, {
    data: { body, positionMs: null },
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(201);
  const id = contractField<Comment>((await response.json()) as object, 'id');
  if (typeof id !== 'string') throw new Error('Comment response must carry id');
  return id;
}

function queueRow(page: Page, text: string) {
  return page.getByRole('region', { name: 'Черга модерації' }).locator('article').filter({
    hasText: text,
  });
}

// Publish, then enrol the seeded student as an active member so the course has a learner to list.
async function publishWithLearner(
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
  const enrolled = await request.post(`/api/v1/admin/courses/${result.id}/grants`, {
    data: { userId: studentID, reason: `learning ${runNamespace}` },
    headers: origin(baseURL),
  });
  expect(enrolled.status(), await enrolled.text()).toBe(201);
  const draftIndex = drafts.findIndex((candidate) => candidate.id === course.id);
  if (draftIndex >= 0) drafts.splice(draftIndex, 1);
  return result;
}

function learnersSection(page: Page) {
  return page.locator('section[aria-labelledby="teach-learners-heading"]');
}

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const context = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    for (const course of published) {
      await context.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { expectedUpdatedAt: course.updatedAt, status: 'archived' },
        headers: origin(baseURL),
      });
    }
    for (const course of drafts) {
      await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
        data: { expectedUpdatedAt: course.updatedAt },
        headers: origin(baseURL),
      });
    }
  } finally {
    await context.close();
  }
});

test.describe('learner overview', () => {
  test('Lo10 links a learner row through to the course roster', async ({ browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const title = `Огляд ${runNamespace}`;
      const course = await createCourse(admin.request, baseURL, `${runNamespace}-lo10`, title);
      await publishWithLearner(admin.request, baseURL, course);

      const page = await instructor.newPage();
      await page.goto(`/teach/learners?courseId=${course.id}`);
      const section = learnersSection(page);
      const link = section.getByRole('link', { name: title });
      await expect(link).toBeVisible();
      await link.click();
      await expect(page).toHaveURL(new RegExp(`/teach/courses/${course.id}`));
    } finally {
      await instructor.close();
      await admin.close();
    }
  });

  test('Lo11 keeps the overview keyboard-visible, sized, themed, zoom-safe, and axe-clean', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const title = `A11y ${runNamespace}`;
      const course = await createCourse(admin.request, baseURL, `${runNamespace}-lo11`, title);
      await publishWithLearner(admin.request, baseURL, course);

      const page = await instructor.newPage();
      await page.goto(`/teach/learners?courseId=${course.id}`);
      const section = learnersSection(page);
      await expect(section.getByRole('link', { name: title })).toBeVisible();

      // Keyboard: the search box reaches the submit button, which shows a visible focus ring.
      const search = section.getByRole('searchbox', { name: 'Пошук за ім’ям' });
      await search.focus();
      await expect(search).toBeFocused();
      const submit = section.getByRole('button', { name: 'Шукати' });
      await submit.focus();
      const focus = await submit.evaluate((element) => {
        const style = getComputedStyle(element);
        return { style: style.outlineStyle, width: Number.parseFloat(style.outlineWidth) };
      });
      expect(focus.style).not.toBe('none');
      expect(focus.width).toBeGreaterThanOrEqual(2);

      // Every control in the overview meets the 44px target.
      for (const button of await section.getByRole('button').all()) {
        const box = await button.boundingBox();
        if (box === null) throw new Error('A learner-overview button is not rendered');
        expect(box.width).toBeGreaterThanOrEqual(44);
        expect(box.height).toBeGreaterThanOrEqual(44);
      }

      await page.addScriptTag({ content: axe.source });
      for (const scheme of ['light', 'dark'] as const) {
        await page.emulateMedia({ colorScheme: scheme });
        const result = await page.evaluate(async () => {
          const region = document.querySelector(
            'section[aria-labelledby="teach-learners-heading"]',
          );
          if (region === null) return { violations: ['learner overview absent'] };
          const audit = await window.axe.run(region, {
            runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
          });
          return {
            violations: audit.violations.flatMap((item) =>
              item.nodes.map((node) => `${item.id} ${node.target.join(' ')}`),
            ),
          };
        });
        expect(result.violations, `${scheme} learner-overview axe`).toEqual([]);
      }

      await page.emulateMedia({ colorScheme: 'light' });
      await page.addStyleTag({ content: 'html { font-size: 200%; }' });
      expect(await section.evaluate((node) => node.scrollWidth <= node.clientWidth + 1)).toBe(true);
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1),
      ).toBe(true);
    } finally {
      await instructor.close();
      await admin.close();
    }
  });
});

test.describe('author block from the moderation queue', () => {
  test('Lo12 blocks a comment author from the queue so they can no longer communicate', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    const student = await openAs(browser, baseURL, 'student');
    try {
      const lessonId = randomUUID();
      const course = await createCourse(
        admin.request,
        baseURL,
        `${runNamespace}-lo12`,
        `Lo12 ${runNamespace}`,
        lessonId,
      );
      await publishWithLearner(admin.request, baseURL, course);
      const body = `Lo12 student comment ${runNamespace}`;
      const commentId = await publishComment(student.request, baseURL, lessonId, body);
      await seedCommentReport(commentId);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const row = queueRow(page, body);
      await row.getByRole('button', { name: 'Заблокувати автора' }).click();
      // The transition refetches the queue and the control flips to Unblock.
      await expect(row.getByRole('button', { name: 'Розблокувати автора' })).toBeVisible();

      // CanCommunicate: the blocked author can no longer post, but can still read (CanConsume).
      const blocked = await student.request.post(`/api/v1/learn/lessons/${lessonId}/comments`, {
        data: { body: `after block ${runNamespace}`, positionMs: null },
        headers: origin(baseURL),
      });
      expect(blocked.status()).toBe(403);
      expect(await blocked.text()).toContain('enrollment_blocked');
      const read = await student.request.get(`/api/v1/learn/lessons/${lessonId}/comments`);
      expect(read.status()).toBe(200);
    } finally {
      await student.close();
      await instructor.close();
      await admin.close();
    }
  });

  test('Lo13 blocks a review author from the queue; the review stays visible', async ({
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const admin = await openAs(browser, baseURL, 'admin');
    const instructor = await openAs(browser, baseURL, 'instructor');
    try {
      const slug = `${runNamespace}-lo13`;
      const course = await createCourse(admin.request, baseURL, slug, `Lo13 ${runNamespace}`);
      await publishWithLearner(admin.request, baseURL, course);
      const reviewId = randomUUID();
      const body = `Lo13 student review ${runNamespace}`;
      await seedReportedReview(course.id, reviewId, body);

      const page = await instructor.newPage();
      await page.goto(`/teach/courses/${course.id}`);
      const row = queueRow(page, body);
      await row.getByRole('button', { name: 'Заблокувати автора' }).click();
      await expect(row.getByRole('button', { name: 'Розблокувати автора' })).toBeVisible();

      // The toggle must round-trip: the flipped control is itself clickable. A pending flag that
      // never cleared once left Unblock inert and the block one-way (Courses 2 review, LR5).
      await row.getByRole('button', { name: 'Розблокувати автора' }).click();
      await expect(row.getByRole('button', { name: 'Заблокувати автора' })).toBeVisible();
      // Re-block so the state under test is "blocked" for the visibility check below.
      await row.getByRole('button', { name: 'Заблокувати автора' }).click();
      await expect(row.getByRole('button', { name: 'Розблокувати автора' })).toBeVisible();

      // Lo13: blocking the author does NOT remove the review — it stays in the public list.
      const listed = await admin.request.get(`/api/v1/courses/${slug}/reviews`);
      expect(listed.status()).toBe(200);
      const reviews = (await listed.json()) as { items: ReadonlyArray<{ id: string }> };
      expect(reviews.items.some((review) => review.id === reviewId)).toBe(true);
    } finally {
      await instructor.close();
      await admin.close();
    }
  });
});
