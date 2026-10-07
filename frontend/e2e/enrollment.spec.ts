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

declare global {
  interface Window {
    axe: typeof axe;
  }
}

const runNamespace = `e2e-enrollment-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
const studentID = '40000000-0000-4000-8000-000000000001';
type CourseToken = { id: string; updatedAt: string; slug: string; lessonID: string; title: string };
const published: CourseToken[] = [];

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
  const initial = token((await created.json()) as object, slug, lessonID, title);
  const saved = await request.put(`/api/v1/admin/courses/${initial.id}`, {
    data: {
      expectedUpdatedAt: initial.updatedAt,
      course: draft,
      modules: [
        {
          id: randomUUID(),
          title: `Модуль ${suffix}`,
          lessons: [{ id: lessonID, title: `Урок ${suffix}`, description: '' }],
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
  published.push(course);
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

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const admin = await openAs(browser, baseURL, 'admin');
  try {
    for (const course of published) {
      const response = await admin.request.post(`/api/v1/admin/courses/${course.id}/status`, {
        data: { expectedUpdatedAt: course.updatedAt, status: 'archived' },
        headers: origin(baseURL),
      });
      expect(response.status(), await response.text()).toBe(200);
    }
  } finally {
    await admin.close();
  }
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
    const lesson = editor
      .getByRole('group', { name: 'Модуль 1', exact: true })
      .getByRole('group', { name: 'Урок 1', exact: true });
    const minted = editor.waitForResponse(
      (response) =>
        new URL(response.url()).pathname === `/api/v1/admin/lessons/${course.lessonID}/video` &&
        response.request().method() === 'POST',
    );
    await lesson
      .locator('input[type="file"]')
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
