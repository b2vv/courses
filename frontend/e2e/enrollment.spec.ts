import { randomUUID } from 'node:crypto';
import path from 'node:path';
import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import axe from 'axe-core';
import type {
  AdminCourseTree,
  CourseVideos,
  ErrorResponse,
  ErrorResponseError,
  MediaUploadTicket,
  MediaVideo,
  MembershipState,
  PlaybackSession,
} from '../src/adapters/api/generated/models';
import { authFixtures, authStatePath } from './auth';
import { contractField } from './contract';
import { deleteE2ECourse } from './course-cleanup';

declare global {
  interface Window {
    axe: typeof axe;
  }
}

const runNamespace = `e2e-enrollment-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
// Module state is local to this Playwright worker, including when a failed test stops it.
const cleanupCourseIDs = new Set<string>();
const instructorID = '40000000-0000-4000-8000-000000000002';
const studentID = '40000000-0000-4000-8000-000000000001';
type CourseToken = { id: string; updatedAt: string; slug: string; lessonID: string; title: string };

function responseObject(value: unknown): object {
  if (typeof value !== 'object' || value === null) throw new Error('Expected an API object');
  return value;
}

/** Follow the actual worker playlists; never guess a segment filename. */
async function mediaSegment(request: APIRequestContext, baseURL: string, lessonID: string) {
  const minted = await request.post('/api/v1/playback/sessions', {
    data: { lessonId: lessonID },
    headers: origin(baseURL),
  });
  expect(minted.status(), await minted.text()).toBe(201);
  const session = responseObject(await minted.json());
  const playbackId = contractField<PlaybackSession>(session, 'playbackId');
  const mediaPath = contractField<PlaybackSession>(session, 'mediaPath');
  if (typeof playbackId !== 'string' || typeof mediaPath !== 'string') {
    throw new Error('Playback session must carry playbackId and mediaPath');
  }
  const master = new URL(mediaPath, baseURL);
  const prefix = `/media/p/${playbackId}/`;
  expect(master.pathname).toMatch(new RegExp(`^${prefix}hls/[0-9a-f-]+/master\\.m3u8$`));
  const masterResponse = await request.get(master.toString());
  expect(masterResponse.status()).toBe(200);
  const variant = (await masterResponse.text())
    .split(/\r?\n/)
    .map((line) => line.trim())
    .find((line) => line !== '' && !line.startsWith('#'));
  if (variant === undefined) throw new Error('Master playlist has no rendition');
  const playlist = new URL(variant, master);
  expect(playlist.origin).toBe(master.origin);
  expect(playlist.pathname.startsWith(prefix)).toBe(true);
  const listed = await request.get(playlist.toString());
  expect(listed.status()).toBe(200);
  const segment = (await listed.text())
    .split(/\r?\n/)
    .map((line) => line.trim())
    .find((line) => line !== '' && !line.startsWith('#'));
  if (segment === undefined) throw new Error('Rendition playlist has no media segment');
  const url = new URL(segment, playlist);
  expect(url.origin).toBe(master.origin);
  expect(url.pathname.startsWith(prefix)).toBe(true);
  expect(url.pathname).toMatch(/\.m4s$/);
  return url.toString();
}

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}
function token(payload: object, slug: string, lessonID: string, title: string): CourseToken {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('course token is missing');
  return { id, updatedAt, slug, lessonID, title };
}
function errorCode(payload: unknown) {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('error response is not an object');
  const error = contractField<ErrorResponse>(payload, 'error');
  if (typeof error !== 'object' || error === null) throw new Error('error response has no error');
  const code = contractField<ErrorResponseError>(error, 'code');
  if (typeof code !== 'string') throw new Error('error response has no code');
  return code;
}
async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}
async function createFreeCourse(request: APIRequestContext, baseURL: string, suffix: string) {
  const slug = `${runNamespace}-${suffix}`;
  const title = `Enrollment ${suffix} ${runNamespace}`;
  const lessonID = randomUUID();
  const draft = {
    slug,
    title,
    summary: 'Run-scoped enrollment e2e course.',
    description: 'A published free course for enrollment browser journeys.',
    instructorId: instructorID,
  };
  const created = await request.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const body: unknown = await created.json();
  if (typeof body !== 'object' || body === null || !('id' in body) || typeof body.id !== 'string')
    throw new Error('Course response must carry id');
  cleanupCourseIDs.add(body.id);
  const initial = token(body, slug, lessonID, title);
  const saved = await request.put(`/api/v1/admin/courses/${initial.id}`, {
    data: {
      expectedUpdatedAt: initial.updatedAt,
      course: draft,
      modules: [
        {
          id: randomUUID(),
          title: `Модуль ${suffix}`,
          lessons: [{ id: lessonID, title: `Урок ${suffix}` }],
        },
      ],
    },
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const stored = token((await saved.json()) as object, slug, lessonID, title);
  const status = await request.post(`/api/v1/admin/courses/${stored.id}/status`, {
    data: { expectedUpdatedAt: stored.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(status.status(), await status.text()).toBe(200);
  const course = token((await status.json()) as object, slug, lessonID, title);
  return course;
}
async function enroll(request: APIRequestContext, baseURL: string, courseID: string) {
  const response = await request.post(`/api/v1/learn/courses/${courseID}/enroll`, {
    headers: origin(baseURL),
  });
  expect(response.status(), await response.text()).toBe(201);
}
async function wcagViolations(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () =>
    (await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] })).violations.map(
      (item) => item.id,
    ),
  );
}
async function overflowAtDoubleText(page: Page) {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });
  return page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
}

test.afterAll(async () => {
  test.setTimeout(60_000);
  const results = await Promise.allSettled([...cleanupCourseIDs].map(deleteE2ECourse));
  const failures = results.flatMap((result) =>
    result.status === 'rejected' ? [result.reason] : [],
  );
  if (failures.length > 0) throw new AggregateError(failures, 'Course cleanup failed');
});

test('En04 En10 nginx media bytes require membership except for a published preview', async ({
  browser,
  baseURL,
}) => {
  test.setTimeout(300_000);
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'media-preview');
    const editor = await admin.newPage();
    await editor.goto(`/admin/courses/${course.id}`);
    await editor.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const lesson = editor
      .getByRole('group', { name: 'Модуль 1', exact: true })
      .getByRole('group', { name: 'Урок 1', exact: true });
    await expect(
      lesson.getByRole('button', { name: /^(Завантажити відео|Замінити відео)$/ }),
    ).toBeVisible();
    const minted = editor.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/lessons/${course.lessonID}/video` &&
        response.request().method() === 'POST',
    );
    await lesson
      .locator('input[type="file"]:not(.admin-audio-field input)')
      .setInputFiles(
        path.join(import.meta.dirname, '..', 'playwright', '.fixtures', 'e2e-640x360.mp4'),
      );
    const upload = await minted;
    expect(upload.status()).toBe(201);
    const videoId = contractField<MediaUploadTicket>(
      responseObject(await upload.json()),
      'videoId',
    );
    if (typeof videoId !== 'string') throw new Error('Upload ticket has no videoId');
    await expect
      .poll(
        async () => {
          const videos = await admin.request.get(`/api/v1/admin/courses/${course.id}/videos`);
          expect(videos.status()).toBe(200);
          const items = contractField<CourseVideos>(responseObject(await videos.json()), 'items');
          if (!Array.isArray(items)) throw new Error('Course videos has no items');
          for (const item of items) {
            const video = responseObject(item);
            if (contractField<MediaVideo>(video, 'videoId') === videoId) {
              return contractField<MediaVideo>(video, 'state');
            }
          }
          return 'absent';
        },
        { timeout: 240_000, intervals: [1_000] },
      )
      .toBe('ready');

    await enroll(student.request, baseURL, course.id);
    const memberSegment = await mediaSegment(student.request, baseURL, course.lessonID);
    const allowed = await student.request.get(memberSegment);
    expect(allowed.status()).toBe(200);
    const bytes = await allowed.body();
    expect(bytes.length).toBeGreaterThan(0);

    // Keep the same authenticated session, playbackId and existing file. No takeover or logout
    // can explain this 403: only membership changed; this GET traverses nginx auth_request.
    const revoked = await admin.request.post(
      `/api/v1/admin/courses/${course.id}/participants/${studentID}/revoke`,
      { headers: origin(baseURL) },
    );
    expect(revoked.status(), await revoked.text()).toBe(200);
    expect((await student.request.get(memberSegment)).status()).toBe(403);

    await editor.reload();
    await editor.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const previewLabel = 'Відкритий до покупки — урок 1, модуль 1';
    await expect(editor.getByRole('checkbox', { name: previewLabel })).not.toBeChecked();
    await editor.getByText(previewLabel, { exact: true }).click();
    await expect(editor.getByRole('checkbox', { name: previewLabel })).toBeChecked();
    const saved = editor.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/courses/${course.id}` &&
        response.request().method() === 'PUT',
    );
    await editor.getByRole('button', { name: 'Зберегти зміни' }).click();
    const updated = await saved;
    expect(updated.status()).toBe(200);
    // afterAll must archive using the version written by the preview edit.
    Object.assign(
      course,
      token(responseObject(await updated.json()), course.slug, course.lessonID, course.title),
    );

    const membership = await student.request.get(`/api/v1/learn/courses/${course.id}/membership`);
    expect(membership.status()).toBe(200);
    expect(contractField<MembershipState>(responseObject(await membership.json()), 'state')).toBe(
      'revoked',
    );
    // The original slot still works once preview permits access: the earlier 403 was not TTL expiry.
    expect((await student.request.get(memberSegment)).status()).toBe(200);
    const previewSegment = await mediaSegment(student.request, baseURL, course.lessonID);
    const previewBytes = await student.request.get(previewSegment);
    expect(previewBytes.status()).toBe(200);
    expect(await previewBytes.body()).toEqual(bytes);
  } finally {
    await student.close();
    await admin.close();
  }
});

test('En01 free enrollment shows CTA, creates membership, and appears in My learning', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'en01');
    const page = await student.newPage();
    await page.goto(`/courses/${course.slug}`);
    await expect(page.getByRole('button', { name: 'Записатися' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Програма курсу' })).toBeVisible();
    expect(await wcagViolations(page)).toEqual([]);
    await page.getByRole('button', { name: 'Записатися' }).click();
    await expect(page.getByRole('heading', { name: 'Ви учасник' })).toBeVisible();
    const start = page.getByRole('link', { name: 'Почати навчання' });
    await expect(start).toHaveAttribute('href', `/learn/${course.slug}`);
    await page.goto('/learn');
    await expect(page.getByRole('heading', { name: 'Моє навчання' })).toBeVisible();
    await expect(page.getByText(course.title)).toBeVisible();
    expect(await overflowAtDoubleText(page)).toBe(false);
  } finally {
    await student.close();
    await admin.close();
  }
});

test('C3 a guest signs in from a course and returns to it; its author sees their own course', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const guest = await browser.newContext({ baseURL, viewport: { width: 390, height: 844 } });
  const student = await openAs(browser, baseURL, 'student');
  const instructor = await openAs(browser, baseURL, 'instructor');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'c3');
    // A guest gets the price and one sign-in action, never a write (C3 #1); on a phone the
    // action is on the first screen (C3 #5: this fixture, uk, 100% text, 390×844).
    const page = await guest.newPage();
    await page.goto(`/courses/${course.slug}`);
    const box = page.getByRole('complementary');
    await expect(box.getByText('Безкоштовно')).toBeVisible();
    await expect(box.getByRole('button', { name: 'Записатися' })).toHaveCount(0);
    const signIn = box.getByRole('link', { name: 'Увійти, щоб записатися' });
    const target = `/login?returnTo=${encodeURIComponent(`/courses/${course.slug}`)}`;
    await expect(signIn).toHaveAttribute('href', target);
    const bounds = await signIn.boundingBox();
    expect((bounds?.y ?? Infinity) + (bounds?.height ?? 0)).toBeLessThanOrEqual(844);
    expect(await wcagViolations(page)).toEqual([]);
    // The return target survives the sign-in page: a signed-in visitor of that address
    // continues to the same course (the e2e login budget is spent, so no new form login).
    const back = await student.newPage();
    await back.goto(target);
    await back.getByRole('button', { name: 'Продовжити' }).click();
    await expect(back).toHaveURL(new RegExp(`/courses/${course.slug}$`));
    await expect(back.getByRole('button', { name: 'Записатися' })).toBeVisible();
    // The fixture's instructor is its author: their course, with a way to edit it (owner).
    const own = await instructor.newPage();
    await own.goto(`/courses/${course.slug}`);
    await expect(own.getByRole('heading', { name: 'Це ваш курс' })).toBeVisible();
    await expect(own.getByRole('link', { name: 'Редагувати курс' })).toHaveAttribute(
      'href',
      `/teach/courses/${course.id}`,
    );
    await expect(own.getByRole('button', { name: 'Записатися' })).toHaveCount(0);
  } finally {
    await instructor.close();
    await student.close();
    await guest.close();
    await admin.close();
  }
});

test('Lo19 an administrator sees a free course but never the enrol call to action', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'lo19');
    const page = await admin.newPage();
    await page.goto(`/courses/${course.slug}`);
    // Exempt: the administrator sees the course and can start learning.
    await expect(page.getByRole('heading', { name: 'Програма курсу' })).toBeVisible();
    const start = page.getByRole('link', { name: 'Почати навчання' });
    await expect(start).toHaveAttribute('href', `/learn/${course.slug}`);
    // G8: not a learner — the enrol button is absent.
    await expect(page.getByRole('button', { name: 'Записатися' })).toHaveCount(0);
    expect(await wcagViolations(page)).toEqual([]);
  } finally {
    await admin.close();
  }
});

test('C3b an administrator starts learning inside the app; crossing to the admin side is one load', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'c3b');
    const page = await admin.newPage();
    await page.goto(`/courses/${course.slug}`);
    let documents = 0;
    page.on('request', (request) => {
      if (request.resourceType() === 'document' && request.frame() === page.mainFrame()) {
        documents += 1;
      }
    });
    // Course → learning: a move inside the client (C3b-1, row A4).
    await page.getByRole('link', { name: 'Почати навчання' }).click();
    await expect(page).toHaveURL(new RegExp(`/learn/${course.slug}(/|$)`));
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    expect(documents).toBe(0);

    // Client → admin side: exactly one document load (C3b-2).
    await page.getByRole('link', { name: 'В адмінку' }).first().click();
    await expect(page).toHaveURL(/\/admin$/);
    await expect(page.locator('[data-surface="admin"]').first()).toBeVisible();
    expect(documents).toBe(1);
    // Inside the admin side: no load.
    await page.goto(`/admin/courses/${course.id}`);
    documents = 0;
    await page.getByRole('link', { name: 'Перегляд' }).click();
    await expect(page).toHaveURL(new RegExp(`/courses/${course.slug}$`));
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(course.title);
    // «Перегляд» crosses back to the client: one load.
    expect(documents).toBe(1);
    await page.goto('/admin');
    documents = 0;
    await page.getByRole('link', { name: 'Учні' }).first().click();
    await expect(page).toHaveURL(/\/admin\/learners$/);
    expect(documents).toBe(0);
    await page.getByRole('link', { name: 'На клієнт' }).click();
    await expect(page).toHaveURL(/\/$/);
    expect(documents).toBe(1);
  } finally {
    await admin.close();
  }
});

test('En03 keeps course metadata public while a non-member free lesson is denied', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'en03');
    const page = await student.newPage();
    await page.goto(`/courses/${course.slug}`);
    await expect(page.getByText(course.title)).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Програма курсу' })).toBeVisible();
    const learning = await page.request.get(`/api/v1/learn/lessons/${course.lessonID}/comments`);
    expect(learning.status()).toBe(403);
  } finally {
    await student.close();
    await admin.close();
  }
});

test('En05 En08 blocked member retains content, loses communication, and roster toggles state', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  const instructor = await openAs(browser, baseURL, 'instructor');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'en05');
    await enroll(student.request, baseURL, course.id);
    const teach = await instructor.newPage();
    await teach.goto(`/teach/courses/${course.id}`);
    const roster = teach.getByRole('region', { name: 'Учасники' });
    await expect(roster.getByText(authFixtures.student.name)).toBeVisible();
    await expect(roster.getByText('0% пройдено')).toBeVisible();
    expect(await wcagViolations(teach)).toEqual([]);
    await roster.getByRole('button', { name: 'Заблокувати' }).click();
    await expect(roster.getByText('Заблокований')).toBeVisible();
    const studentPage = await student.newPage();
    await studentPage.goto(`/learn/${course.slug}/${course.lessonID}`);
    await expect(studentPage.getByRole('region', { name: 'Обговорення' })).toBeVisible();
    const rejected = studentPage.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/learn/lessons/${course.lessonID}/comments` &&
        response.request().method() === 'POST',
    );
    const discussion = studentPage.getByRole('region', { name: 'Обговорення' });
    await discussion.getByRole('textbox', { name: 'Коментар' }).fill(`blocked ${runNamespace}`);
    await discussion.getByRole('button', { name: 'Опублікувати' }).click();
    const response = await rejected;
    expect(response.status()).toBe(403);
    expect(errorCode(await response.json())).toBe('enrollment_blocked');
    // En11: the refusal must be understandable — the blocked member sees the NAMED message,
    // not a generic error (L7d). Assert the rendered text, not only the network code.
    await expect(discussion.getByText(/Вас заблоковано в цьому курсі/)).toBeVisible();
    await roster.getByRole('button', { name: 'Розблокувати' }).click();
    await expect(roster.getByText('Активний')).toBeVisible();
    expect(await overflowAtDoubleText(teach)).toBe(false);
  } finally {
    await instructor.close();
    await student.close();
    await admin.close();
  }
});

test('En04 En15 revoke cuts learning access and re-enroll restores one roster row', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'en04');
    await enroll(student.request, baseURL, course.id);
    const revoke = await admin.request.post(
      `/api/v1/admin/courses/${course.id}/participants/${studentID}/revoke`,
      { headers: origin(baseURL) },
    );
    expect(revoke.status(), await revoke.text()).toBe(200);
    const denied = await student.request.get(`/api/v1/learn/lessons/${course.lessonID}/comments`);
    expect(denied.status()).toBe(403);
    const page = await student.newPage();
    await page.goto(`/courses/${course.slug}`);
    await page.getByRole('button', { name: 'Записатися' }).click();
    await expect(page.getByRole('heading', { name: 'Ви учасник' })).toBeVisible();
    // En15 single-row: re-enrolling after revoke restores exactly one active roster row, not a
    // second. The roster endpoint is /roster and its authority is owner∨admin, so read it as the
    // instructor (owner).
    const verify = await openAs(browser, baseURL, 'instructor');
    try {
      const check = await verify.request.get(`/api/v1/teach/courses/${course.id}/roster`);
      expect(check.status(), await check.text()).toBe(200);
      const body = (await check.json()) as { entries?: Array<{ userId?: string; state?: string }> };
      const mine = body.entries?.filter((entry) => entry.userId === studentID) ?? [];
      expect(mine).toHaveLength(1);
      expect(mine[0]?.state).toBe('active');
    } finally {
      await verify.close();
    }
  } finally {
    await student.close();
    await admin.close();
  }
});

test('En07 En14 participant reports resolve in the admin queue and an instructor self-enrolls in another course', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  const instructor = await openAs(browser, baseURL, 'instructor');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'en07');
    await enroll(student.request, baseURL, course.id);
    const studentPage = await student.newPage();
    await studentPage.goto(`/courses/${course.slug}`);
    await studentPage.getByRole('button', { name: 'Поскаржитися на викладача' }).click();
    await studentPage
      .getByRole('textbox', { name: 'Причина скарги' })
      .fill(`report ${runNamespace}`);
    await studentPage.getByRole('button', { name: 'Надіслати скаргу' }).click();
    await expect(studentPage.getByText('Скаргу надіслано')).toBeVisible();
    const queue = await admin.newPage();
    await queue.goto('/admin/participant-reports');
    await expect(queue.getByRole('heading', { name: 'Черга скарг учасників' })).toBeVisible();
    const report = queue.locator('article').filter({ hasText: `report ${runNamespace}` });
    await expect(report).toBeVisible();
    expect(await wcagViolations(queue)).toEqual([]);
    await report.getByRole('button', { name: 'Вирішити' }).click();
    await expect(report).toHaveCount(0);
    // The resolved report leaves the default (open) view but is retrievable under the
    // Resolved filter, where it carries the resolved pill and no live Resolve control.
    await queue.getByRole('radio', { name: 'Вирішені' }).click();
    const resolvedReport = queue.locator('article').filter({ hasText: `report ${runNamespace}` });
    await expect(resolvedReport).toBeVisible();
    await expect(resolvedReport).toContainText('Вирішена');
    await expect(resolvedReport.getByRole('button', { name: 'Вирішити' })).toHaveCount(0);
    const teach = await instructor.newPage();
    await teach.goto(`/teach/courses/${course.id}`);
    const studentRow = teach.locator('article').filter({ hasText: authFixtures.student.name });
    await studentRow.getByRole('button', { name: 'Поскаржитися' }).click();
    await teach
      .getByRole('textbox', { name: 'Причина скарги' })
      .fill(`instructor report ${runNamespace}`);
    await teach.getByRole('button', { name: 'Надіслати скаргу' }).click();
    await expect(teach.getByText('Скаргу надіслано')).toBeVisible();
    const selfEnroll = await instructor.request.post(`/api/v1/learn/courses/${course.id}/enroll`, {
      headers: origin(baseURL),
    });
    expect(selfEnroll.status(), await selfEnroll.text()).toBe(201);
  } finally {
    await instructor.close();
    await student.close();
    await admin.close();
  }
});

test('En16 blocked member keeps media bytes (CanConsume) but loses comments (CanCommunicate)', async ({
  browser,
  baseURL,
}) => {
  test.setTimeout(300_000);
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const student = await openAs(browser, baseURL, 'student');
  const instructor = await openAs(browser, baseURL, 'instructor');
  try {
    const course = await createFreeCourse(admin.request, baseURL, 'block-consume');
    // Upload a lesson video and wait until it is ready (same production path as En04).
    const editor = await admin.newPage();
    await editor.goto(`/admin/courses/${course.id}`);
    await editor.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    const lesson = editor
      .getByRole('group', { name: 'Модуль 1', exact: true })
      .getByRole('group', { name: 'Урок 1', exact: true });
    const minted = editor.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/lessons/${course.lessonID}/video` &&
        response.request().method() === 'POST',
    );
    await lesson
      .locator('input[type="file"]:not(.admin-audio-field input)')
      .setInputFiles(
        path.join(import.meta.dirname, '..', 'playwright', '.fixtures', 'e2e-640x360.mp4'),
      );
    const upload = await minted;
    expect(upload.status()).toBe(201);
    const videoId = contractField<MediaUploadTicket>(
      responseObject(await upload.json()),
      'videoId',
    );
    if (typeof videoId !== 'string') throw new Error('Upload ticket has no videoId');
    await expect
      .poll(
        async () => {
          const videos = await admin.request.get(`/api/v1/admin/courses/${course.id}/videos`);
          expect(videos.status()).toBe(200);
          const items = contractField<CourseVideos>(responseObject(await videos.json()), 'items');
          if (!Array.isArray(items)) throw new Error('Course videos has no items');
          for (const item of items) {
            const video = responseObject(item);
            if (contractField<MediaVideo>(video, 'videoId') === videoId) {
              return contractField<MediaVideo>(video, 'state');
            }
          }
          return 'absent';
        },
        { timeout: 240_000, intervals: [1_000] },
      )
      .toBe('ready');

    await enroll(student.request, baseURL, course.id);
    const memberSegment = await mediaSegment(student.request, baseURL, course.lessonID);
    // Baseline: an active member consumes media bytes (warms the authorized slot).
    expect((await student.request.get(memberSegment)).status()).toBe(200);

    // The owning instructor blocks the member (lecturer's communication control, not revoke).
    const blocked = await instructor.request.post(
      `/api/v1/teach/courses/${course.id}/participants/${studentID}/block`,
      { headers: origin(baseURL) },
    );
    expect(blocked.status(), await blocked.text()).toBe(200);

    // CanCommunicate is LOST: the same session can no longer post a comment.
    const comment = await student.request.post(
      `/api/v1/learn/lessons/${course.lessonID}/comments`,
      {
        data: { body: `blocked ${runNamespace}`, positionMs: null },
        headers: origin(baseURL),
      },
    );
    expect(comment.status()).toBe(403);
    expect(errorCode(await comment.json())).toBe('enrollment_blocked');

    // CanConsume is RETAINED: the SAME gated media segment still serves bytes through nginx
    // auth_request. Same session, playbackId and file as the warmed slot; only communication
    // changed. This is the contract distinction a comment-read does not prove — blocking removes
    // CanCommunicate without removing CanConsume (contrast En04, where revoke 403s this exact slot).
    const stillAllowed = await student.request.get(memberSegment);
    expect(stillAllowed.status(), await stillAllowed.text()).toBe(200);
    expect((await stillAllowed.body()).length).toBeGreaterThan(0);

    // Unblock restores communication; consumption was never interrupted.
    const unblocked = await instructor.request.post(
      `/api/v1/teach/courses/${course.id}/participants/${studentID}/unblock`,
      { headers: origin(baseURL) },
    );
    expect(unblocked.status(), await unblocked.text()).toBe(200);
  } finally {
    await instructor.close();
    await student.close();
    await admin.close();
  }
});
