import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promisify } from 'node:util';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

declare global {
  interface Window {
    axe: typeof axe;
  }
}

const execute = promisify(execFile);
const runNamespace = `e2e-ratings-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const studentID = '40000000-0000-4000-8000-000000000001';
const instructorID = '40000000-0000-4000-8000-000000000002';
const adminID = '40000000-0000-4000-8000-000000000003';
const published: CourseToken[] = [];

type CourseToken = { id: string; updatedAt: string };
type CourseFixture = CourseToken & { slug: string; title: string; lessonId: string };

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

function token(payload: object): CourseToken {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('admin course response must carry id and updatedAt');
  return { id, updatedAt };
}

function literal(value: string) {
  return `'${value.replaceAll("'", "''")}'`;
}

async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}

async function axeViolations(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () => {
    const result = await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] });
    return result.violations.map((violation) => violation.id);
  });
}

async function createCourse(
  admin: APIRequestContext,
  baseURL: string,
  label: string,
  options: { paid?: boolean; lessons?: number } = {},
): Promise<CourseFixture> {
  const slug = `${runNamespace}-${label}`;
  const title = `Рейтинги ${label} ${runNamespace}`;
  const draft = {
    slug,
    title,
    summary: 'Run-scoped ratings browser fixture.',
    description: 'A course created only for the ratings browser journey.',
    instructorId: instructorID,
  };
  const created = await admin.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const initial = token((await created.json()) as object);
  const lessons = Array.from({ length: options.lessons ?? 1 }, (_, index) => ({
    id: randomUUID(),
    title: `Урок ${index + 1} ${label}`,
    description: '',
  }));
  const modules = [{ id: randomUUID(), title: `Модуль ${label}`, lessons }];
  const saved = await admin.put(`/api/v1/admin/courses/${initial.id}`, {
    data: { expectedUpdatedAt: initial.updatedAt, course: draft, modules },
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  let current = token((await saved.json()) as object);
  if (options.paid === true) {
    const priced = await admin.put(`/api/v1/admin/courses/${current.id}`, {
      data: {
        expectedUpdatedAt: current.updatedAt,
        course: { ...draft, price: { minor: 189000, currency: 'UAH' } },
        modules,
      },
      headers: origin(baseURL),
    });
    expect(priced.status(), await priced.text()).toBe(200);
    current = token((await priced.json()) as object);
  }
  const madePublic = await admin.post(`/api/v1/admin/courses/${current.id}/status`, {
    data: { expectedUpdatedAt: current.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(madePublic.status(), await madePublic.text()).toBe(200);
  const result = token((await madePublic.json()) as object);
  published.push(result);
  return { ...result, slug, title, lessonId: lessons[0]?.id ?? '' };
}

// The browser suite already uses Compose psql for fixture-only progress cleanup.  This seeds rows
// which cannot be made through a public account-creation API (notably the 25-author page), while
// all review writes exercised below continue to use the production HTTP/UI path.
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

async function seedProgress(course: CourseFixture, userID = studentID, completed = false) {
  await sql(
    `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at) VALUES (${literal(userID)}::uuid,${literal(course.lessonId)}::uuid,1,${completed ? 'now()' : 'NULL'},1,now()) ON CONFLICT (user_id,lesson_id) DO UPDATE SET position_seconds=EXCLUDED.position_seconds,completed_at=EXCLUDED.completed_at,client_seq=EXCLUDED.client_seq,updated_at=EXCLUDED.updated_at`,
  );
  // Progress implies enrollment after the enrollment phase: a learner with progress is an active
  // member, which writing/reporting a review (communication) now requires.
  await sql(
    `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES (${literal(userID)}::uuid,${literal(course.id)}::uuid,'self','active',now()) ON CONFLICT (user_id,course_id) DO NOTHING`,
  );
}

async function seedReview(course: CourseFixture, index: number, stars: number, body: string) {
  const author = randomUUID();
  await sql(
    `INSERT INTO users(id,email,display_name,role,password_hash) VALUES (${literal(author)}::uuid,${literal(`${runNamespace}-${index}@example.test`)},${literal(`Рецензент ${index}`)},'student','$argon2id$v=19$m=65536,t=3,p=4$abcdefghijklmnop$abcdefghijklmnopqrstuvwxabcdefghijklmnopQRSTUVWX')`,
  );
  await sql(
    `INSERT INTO discussion_reviews(id,course_id,author_id,stars,body,completed_lessons,total_lessons,created_at) VALUES (gen_random_uuid(),${literal(course.id)}::uuid,${literal(author)}::uuid,${stars},${literal(body)},0,1,now() + interval '${index} seconds')`,
  );
}

async function grant(
  admin: APIRequestContext,
  baseURL: string,
  course: CourseFixture,
  userId: string,
) {
  const response = await admin.post(`/api/v1/admin/courses/${course.id}/grants`, {
    data: { userId, reason: `ratings ${runNamespace}` },
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(201);
}

function reviewForm(page: Page) {
  return page.getByRole('heading', { name: 'Ваш відгук' }).locator('..');
}

async function writeReviewInUI(page: Page, stars: number, body: string) {
  const form = reviewForm(page);
  // React Aria hides the radio input inside a visible label and keeps the state itself: click the label.
  await form
    .locator('label.ui-radio')
    .filter({ hasText: `${stars} з 5` })
    .click();
  await expect(form.getByRole('radio', { name: `${stars} з 5` })).toBeChecked();
  await form.getByRole('textbox', { name: 'Текст відгуку' }).fill(body);
  await form.getByRole('button', { name: 'Зберегти' }).click();
  // The save must land: the form closes and the stored review is shown with its Edit action.
  await expect(form.getByRole('button', { name: 'Редагувати' })).toBeVisible();
  await expect(form.getByText(body)).toBeVisible();
}

test.use({ storageState: authStatePath('student') });

// Published courses affect catalog/recovery totals.  Archive every course this spec publishes.
test.afterAll(async ({ browser, baseURL }) => {
  // Cleanup is several API calls plus one SQL delete; under a loaded Docker host it outran the default 30 s.
  test.setTimeout(120_000);
  if (baseURL === undefined) return;
  const admin = await openAs(browser, baseURL, 'admin');
  try {
    // Archive first: published courses are counted by catalog.spec and recovery.spec. Each course re-reads its
    // current version, so a test that failed half-way (stale token) or already archived does not turn the
    // cleanup into a second failure.
    for (const course of published) {
      const current = await admin.request.get(`/api/v1/admin/courses/${course.id}`);
      if (!current.ok()) continue;
      const tree = (await current.json()) as object;
      const status = contractField<AdminCourseTree>(tree, 'status');
      const updatedAt = contractField<AdminCourseTree>(tree, 'updatedAt');
      if (status === 'archived' || typeof updatedAt !== 'string') continue;
      // Archiving is cleanup, not an assertion under test: tolerate a transient nginx api-zone 429
      // (the zone refills within a second). A hard failure here would leave the course published
      // and break recovery.spec's seeded-course count — a cleanup flake masquerading as a defect.
      let response = await admin.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { expectedUpdatedAt: updatedAt, status: 'archived' },
        headers: origin(baseURL),
      });
      for (let attempt = 0; attempt < 5 && response.status() === 429; attempt += 1) {
        await new Promise((resolve) => setTimeout(resolve, 600));
        response = await admin.request.post(`/api/v1/admin/courses/${course.id}/status`, {
          data: { expectedUpdatedAt: updatedAt, status: 'archived' },
          headers: origin(baseURL),
        });
      }
      expect(response.status(), await response.text()).toBe(200);
    }
  } finally {
    await admin.close();
    // Users are counted too: the admin users list is paged, and leftover reviewers pushed admin-users P08's
    // fixture rows off the first page. Remove every reviewer this run created (their reviews cascade).
    await sql(`DELETE FROM users WHERE email LIKE ${literal(`${runNamespace}-%@example.test`)}`);
  }
});

test('Rt01 writes a free-course review from learning and shows its frozen progress on the course page', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'rt01');
  } finally {
    await admin.close();
  }
  await seedProgress(course);
  const body = `Rt01 review body ${runNamespace}`;
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await writeReviewInUI(page, 4, body);
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText(body)).toBeVisible();
  await expect(page.getByText('пройдено 0% на момент відгуку')).toBeVisible();
});

test('Rt03 edits 4 stars to 2 and Rt04 deletes through the dialog then writes again', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'rt03-rt04');
  } finally {
    await admin.close();
  }
  await seedProgress(course);
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await writeReviewInUI(page, 4, `Rt03 original ${runNamespace}`);
  const own = reviewForm(page);
  await own.getByRole('button', { name: 'Редагувати' }).click();
  await own.locator('label.ui-radio').filter({ hasText: '2 з 5' }).click();
  await expect(own.getByRole('radio', { name: '2 з 5' })).toBeChecked();
  await own.getByRole('button', { name: 'Зберегти' }).click();
  await expect(own.getByRole('button', { name: 'Редагувати' })).toBeVisible();
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText('2 з 5')).toBeVisible();
  await expect(page.getByText(/змінено/)).toBeVisible();
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await reviewForm(page).getByRole('button', { name: 'Видалити' }).first().click();
  await page.getByRole('dialog').getByRole('button', { name: 'Видалити' }).click();
  await expect(page.getByText(`Rt03 original ${runNamespace}`)).toHaveCount(0);
  await writeReviewInUI(page, 5, `Rt04 rewrite body ${runNamespace}`);
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText(`Rt04 rewrite body ${runNamespace}`)).toBeVisible();
});

test('Rt05 an admin buys a paid course through the offline WayForPay stub and can review after progress', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'rt05', { paid: true });
  } finally {
    await admin.close();
  }
  const page = await (await openAs(browser, baseURL, 'admin')).newPage();
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await expect(page.getByRole('button', { name: 'Купити курс' })).toBeVisible();
  await page.getByRole('button', { name: 'Купити курс' }).click();
  await page.getByRole('button', { name: 'Перейти до оплати' }).click();
  await expect(page.locator('[data-status="paid"]')).toBeVisible({ timeout: 15_000 });
  await seedProgress(course, adminID);
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await writeReviewInUI(page, 4, `Rt05 paid review ${runNamespace}`);
});

test('Rt06 Rt07 Rt08 show New course below threshold and ★ 4,7 (3) on this run’s catalog card and course page', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'threshold');
  } finally {
    await admin.close();
  }
  await seedReview(course, 1, 4, `Threshold first body ${runNamespace}`);
  await seedReview(course, 2, 5, `Threshold second body ${runNamespace}`);
  await page.goto('/courses');
  const card = page.locator('article').filter({ hasText: course.title });
  await expect(card.getByText('Новий курс')).toBeVisible();
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText('Новий курс').first()).toBeVisible();
  await seedReview(course, 3, 5, `Threshold third body ${runNamespace}`);
  await page.goto('/courses');
  await expect(card.getByText('★ 4,7 (3)')).toBeVisible();
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText('★ 4,7 (3)').first()).toBeVisible();
});

test('Rt09 Rt21 Rt22 and report: report reaches the owner queue, owner has no remove button, admin removal hides it and resets the card', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'moderation');
  } finally {
    await admin.close();
  }
  for (const [index, stars] of [4, 5, 5].entries())
    await seedReview(course, index + 10, stars, `Moderation ${index} ${runNamespace}`);
  // The student reports a review below; reporting is communication, so enroll the student as an
  // active member of this course (enroll the fixture, not loosen the gate).
  await sql(
    `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES (${literal(studentID)}::uuid,${literal(course.id)}::uuid,'self','active',now()) ON CONFLICT (user_id,course_id) DO NOTHING`,
  );
  await page.goto(`/courses/${course.slug}`);
  const target = page.getByText(`Moderation 0 ${runNamespace}`).locator('..');
  await target.getByRole('button', { name: 'Поскаржитись' }).click();
  // The sent state stays focusable (no native disabled, which would drop focus to <body>); it is labelled.
  await expect(target.getByRole('button', { name: 'Скаргу надіслано' })).toBeVisible();
  const instructor = await openAs(browser, baseURL, 'instructor');
  const ownerPage = await instructor.newPage();
  await ownerPage.goto(`/teach/courses/${course.id}`);
  const queue = ownerPage
    .getByRole('region', { name: 'Черга модерації' })
    .locator('article')
    .filter({ hasText: `Moderation 0 ${runNamespace}` });
  await expect(queue).toBeVisible();
  await expect(queue.getByRole('button', { name: 'Зняти' })).toHaveCount(0);
  const administrator = await openAs(browser, baseURL, 'admin');
  const adminPage = await administrator.newPage();
  await adminPage.goto(`/teach/courses/${course.id}`);
  await adminPage
    .getByRole('region', { name: 'Черга модерації' })
    .locator('article')
    .filter({ hasText: `Moderation 0 ${runNamespace}` })
    .getByRole('button', { name: 'Зняти' })
    .click();
  // Wait for the removal to land before reading the public page (navigating at once could race it).
  await expect(
    adminPage
      .getByRole('region', { name: 'Черга модерації' })
      .locator('article')
      .filter({ hasText: `Moderation 0 ${runNamespace}` }),
  ).toHaveCount(0);
  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText(`Moderation 0 ${runNamespace}`)).toHaveCount(0);
  await page.goto('/courses');
  await expect(
    page.locator('article').filter({ hasText: course.title }).getByText('Новий курс'),
  ).toBeVisible();
  await instructor.close();
  await administrator.close();
});

test('Rt14 pages 20 + 20 + 5 by keyboard, keeps focus on «Показати ще» between pages, focuses the first final addition, and announces completion', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'pagination');
  } finally {
    await admin.close();
  }
  // 45 reviews make an intermediate page: a natively disabled button would lose focus to <body> there, while
  // the last page restores focus on its own and would hide that defect.
  for (let index = 0; index < 45; index += 1)
    // Indices 200–244: every scenario seeds a disjoint index range (the reviewer email embeds it).
    await seedReview(course, index + 200, 5, `Rt14 review ${index} ${runNamespace}`);
  await page.goto(`/courses/${course.slug}`);
  const list = page.getByRole('list', { name: 'Відгуки про курс' });
  const items = list.getByRole('listitem');
  await expect(items).toHaveCount(20);
  await expect(items.first()).toContainText(`Rt14 review 44 ${runNamespace}`);
  const more = page.getByRole('button', { name: 'Показати ще' });
  await more.focus();
  await page.keyboard.press('Enter');
  await expect(items).toHaveCount(40);
  await expect(more).toBeFocused();
  await page.keyboard.press('Enter');
  await expect(items).toHaveCount(45);
  await expect(more).toHaveCount(0);
  await expect(items.nth(40)).toBeFocused();
  await expect(page.locator('output[aria-live="polite"]')).toContainText('Усі відгуки показано');
});

test('Rt26 Rt30 keeps reviews for an archived grant holder, makes public course 404, and restores the card rating after republish', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'archive');
    for (let index = 0; index < 3; index += 1)
      await seedReview(course, index + 60, 5, `Archive review ${index} ${runNamespace}`);
    await grant(admin.request, baseURL, course, studentID);
    const archived = await admin.request.post(`/api/v1/admin/courses/${course.id}/status`, {
      data: { expectedUpdatedAt: course.updatedAt, status: 'archived' },
      headers: origin(baseURL),
    });
    expect(archived.status(), await archived.text()).toBe(200);
    course = { ...course, ...token((await archived.json()) as object) };
  } finally {
    await admin.close();
  }
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await expect(page.getByText(`Archive review 0 ${runNamespace}`)).toBeVisible();
  const missing = await page.request.get(`/api/v1/courses/${course.slug}/reviews`);
  expect(missing.status()).toBe(404);
  const publisher = await openAs(browser, baseURL, 'admin');
  const restored = await publisher.request.post(`/api/v1/admin/courses/${course.id}/status`, {
    data: { expectedUpdatedAt: course.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(restored.status(), await restored.text()).toBe(200);
  const restoredToken = token((await restored.json()) as object);
  const publishedIndex = published.findIndex((candidate) => candidate.id === course.id);
  if (publishedIndex < 0) throw new Error('published course was not registered for cleanup');
  published[publishedIndex] = restoredToken;
  await publisher.close();
  await page.goto('/courses');
  await expect(
    page.locator('article').filter({ hasText: course.title }).getByText('★ 5,0 (3)'),
  ).toBeVisible();
});

test('cross-origin review PUT is refused with 403', async ({ browser, baseURL }) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'origin');
  } finally {
    await admin.close();
  }
  await seedProgress(course);
  const response = await (
    await openAs(browser, baseURL, 'student')
  ).request.put(`/api/v1/learn/courses/${course.id}/review`, {
    data: { stars: 4, body: `Foreign Origin body ${runNamespace}` },
    headers: { Origin: 'https://foreign.example' },
  });
  expect(response.status()).toBe(403);
});

test('a11y audits the public reviews, open learning form, and teach queue review item', async ({
  browser,
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  let course: CourseFixture;
  try {
    course = await createCourse(admin.request, baseURL, 'axe');
  } finally {
    await admin.close();
  }
  await seedProgress(course);
  for (let index = 0; index < 3; index += 1)
    await seedReview(course, index + 90, 5, `Axe review ${index} ${runNamespace}`);
  await page.goto(`/courses/${course.slug}`);
  expect(await axeViolations(page)).toEqual([]);
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  expect(await axeViolations(page)).toEqual([]);
  const reviewID = (await page.request
    .get(`/api/v1/courses/${course.slug}/reviews`)
    .then((response) => response.json())) as { items: Array<{ id: string }> };
  const report = await page.request.post(`/api/v1/learn/reviews/${reviewID.items[0]?.id}/reports`, {
    headers: origin(baseURL),
  });
  expect(report.status(), await report.text()).toBe(204);
  const adminPage = await (await openAs(browser, baseURL, 'admin')).newPage();
  await adminPage.goto(`/teach/courses/${course.id}`);
  expect(await axeViolations(adminPage)).toEqual([]);
});
