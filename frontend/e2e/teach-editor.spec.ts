import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promisify } from 'node:util';
import {
  expect,
  test,
  type APIRequestContext,
  type APIResponse,
  type Browser,
} from '@playwright/test';
import { authStatePath } from './auth';

const execute = promisify(execFile);
const runNamespace = `e2e-te-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const instructorID = '40000000-0000-4000-8000-000000000002';
const studentID = '40000000-0000-4000-8000-000000000001';

// A 1×1 PNG: valid magic bytes and decodable dimensions, enough for the upload and the serve path.
const onePixelPng = Buffer.from(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
  'base64',
);

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

function literal(value: string) {
  return `'${value.replaceAll("'", "''")}'`;
}

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

function updatedAt(payload: unknown): string {
  if (
    typeof payload !== 'object' ||
    payload === null ||
    !('updatedAt' in payload) ||
    typeof payload.updatedAt !== 'string'
  )
    throw new Error('Response must carry updatedAt');
  return payload.updatedAt;
}

function token(payload: unknown): { id: string; updatedAt: string } {
  if (
    typeof payload !== 'object' ||
    payload === null ||
    !('id' in payload) ||
    typeof payload.id !== 'string'
  )
    throw new Error('Course response must carry id');
  return { id: payload.id, updatedAt: updatedAt(payload) };
}

async function openAs(browser: Browser, baseURL: string, role: 'student' | 'instructor' | 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
}

type Lesson = { courseId: string; slug: string; lessonId: string; title: string };

async function createPublishedLesson(
  admin: APIRequestContext,
  baseURL: string,
  label: string,
  ownerID = instructorID,
): Promise<Lesson> {
  const slug = `${runNamespace}-${label}`;
  const draft = {
    slug,
    title: `Teach editor ${label} ${runNamespace}`,
    summary: 'Run-scoped teach-editor browser fixture.',
    description: 'A course created only for the teach-editor browser journey.',
    instructorId: ownerID,
  };
  const created = await admin.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const initial = token(await created.json());
  const lessonId = randomUUID();
  const title = `Урок ${label} ${runNamespace}`;
  const modules = [
    {
      id: randomUUID(),
      title: `Модуль ${label} ${runNamespace}`,
      lessons: [{ id: lessonId, title }],
    },
  ];
  const saved = await admin.put(`/api/v1/admin/courses/${initial.id}`, {
    data: { expectedUpdatedAt: initial.updatedAt, course: draft, modules },
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const current = token(await saved.json());
  const published = await admin.post(`/api/v1/admin/courses/${current.id}/status`, {
    data: { expectedUpdatedAt: current.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(published.status(), await published.text()).toBe(200);
  // Match lesson-builder: an active member can verify the published reader path.
  await sql(
    `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES (${literal(studentID)}::uuid,${literal(initial.id)}::uuid,'self','active',now()) ON CONFLICT (user_id,course_id) DO NOTHING`,
  );
  return { courseId: initial.id, slug, lessonId, title };
}

async function expectForbidden(response: APIResponse, code: 'forbidden' | 'course_forbidden') {
  expect(response.status(), await response.text()).toBe(403);
  const body: unknown = await response.json();
  expect(body).toMatchObject({ error: { code } });
}

test('T9 owner opens a lesson from teach, edits and publishes it for a member', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const instructor = await openAs(browser, baseURL, 'instructor');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const lesson = await createPublishedLesson(admin.request, baseURL, 'owner');
    const page = await instructor.newPage();
    const adminRequests: string[] = [];
    page.on('request', (request) => {
      const path = new URL(request.url()).pathname;
      if (path.startsWith('/admin/') || path.startsWith('/api/v1/admin/')) adminRequests.push(path);
    });
    await page.goto(`/teach/courses/${lesson.courseId}`);
    const link = page.getByRole('link', { name: lesson.title, exact: true });
    await expect(link).toBeVisible();
    await expect(link).toHaveAttribute(
      'href',
      `/teach/courses/${lesson.courseId}/lessons/${lesson.lessonId}`,
    );
    await link.click();
    await expect(page).toHaveURL(
      new URL(`/teach/courses/${lesson.courseId}/lessons/${lesson.lessonId}`, baseURL).href,
    );
    await expect(page.getByRole('heading', { name: 'Редактор уроку' })).toBeVisible();
    await page
      .getByRole('group', { name: 'Додати секцію' })
      .getByRole('button', { name: 'Текст' })
      .click();
    const content = `Практика teach editor ${runNamespace}`;
    const textbox = page.getByRole('textbox', { name: 'Текст' });
    await textbox.click();
    await textbox.pressSequentially(content);
    const published = page.waitForResponse(
      (response) =>
        new URL(response.url()).pathname ===
          `/api/v1/teach/lessons/${lesson.lessonId}/document/publish` &&
        response.request().method() === 'POST',
    );
    await page.getByRole('button', { name: 'Опублікувати' }).click();
    expect((await published).status()).toBe(200);
    await expect(page.getByText('Урок опубліковано', { exact: true })).toBeVisible();
    await expect(page.getByRole('alert')).toHaveCount(0);

    const reloaded = await instructor.request.get(
      `/api/v1/teach/lessons/${lesson.lessonId}/document`,
    );
    expect(reloaded.status(), await reloaded.text()).toBe(200);
    const body: unknown = await reloaded.json();
    expect(body).toMatchObject({
      document: {
        sections: [
          {
            type: 'text',
            content: { blocks: [{ type: 'paragraph', spans: [{ text: content }] }] },
          },
        ],
      },
    });
    expect(adminRequests).toEqual([]);

    const reader = await student.newPage();
    await reader.goto(`/learn/${lesson.slug}/${lesson.lessonId}`);
    await expect(
      reader.getByRole('region', { name: 'Матеріали уроку' }).getByText(content, { exact: true }),
    ).toBeVisible();
  } finally {
    await student.close();
    await instructor.close();
    await admin.close();
  }
});

test('T5 non-owner instructor is refused by every teach operation; T6 student cannot read the tree', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  const instructor = await openAs(browser, baseURL, 'instructor');
  const student = await openAs(browser, baseURL, 'student');
  try {
    const otherOwnerID = randomUUID();
    // Only the owner identity is needed; reuse a valid seed hash without logging it or signing in.
    await sql(`INSERT INTO users(id,email,display_name,role,password_hash,email_verified)
      SELECT ${literal(otherOwnerID)}::uuid,${literal(`${runNamespace}-owner@example.test`)},
        ${literal(`Owner ${runNamespace}`)},'instructor',password_hash,true
      FROM users WHERE id=${literal(instructorID)}::uuid`);
    const lesson = await createPublishedLesson(admin.request, baseURL, 'non-owner', otherOwnerID);
    const document = { schemaVersion: 1, locale: 'uk', sections: [] };
    // A real saved token makes both denied mutations valid requests, not parser failures.
    const saved = await admin.request.put(`/api/v1/admin/lessons/${lesson.lessonId}/document`, {
      data: { expectedUpdatedAt: null, document },
      headers: origin(baseURL),
    });
    expect(saved.status(), await saved.text()).toBe(200);
    const expectedUpdatedAt = updatedAt(await saved.json());
    const path = `/api/v1/teach/lessons/${lesson.lessonId}`;
    await expectForbidden(await instructor.request.get(`${path}/document`), 'forbidden');
    await expectForbidden(
      await instructor.request.put(`${path}/document`, {
        data: { expectedUpdatedAt, document },
        headers: origin(baseURL),
      }),
      'forbidden',
    );
    await expectForbidden(
      await instructor.request.post(`${path}/document/publish`, {
        data: { expectedUpdatedAt },
        headers: origin(baseURL),
      }),
      'forbidden',
    );
    await expectForbidden(
      await instructor.request.post(`${path}/images`, {
        multipart: {
          file: { name: `${runNamespace}.png`, mimeType: 'image/png', buffer: onePixelPng },
        },
        headers: origin(baseURL),
      }),
      'forbidden',
    );
    await expectForbidden(
      await instructor.request.get(`/api/v1/teach/courses/${lesson.courseId}`),
      'course_forbidden',
    );
    await expectForbidden(
      await student.request.get(`/api/v1/teach/courses/${lesson.courseId}`),
      'forbidden',
    );
  } finally {
    await student.close();
    await instructor.close();
    await admin.close();
  }
});
