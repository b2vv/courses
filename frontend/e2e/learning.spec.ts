import { randomUUID } from 'node:crypto';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-learning-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
const studentID = '40000000-0000-4000-8000-000000000001';

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
          lessons: [{ id: randomUUID(), title: `Урок ${runNamespace}`, description: '' }],
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
