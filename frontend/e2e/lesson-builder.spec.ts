import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { promisify } from 'node:util';
import { expect, test, type APIRequestContext } from '@playwright/test';
import axe from 'axe-core';
import { authStatePath } from './auth';

const execute = promisify(execFile);
const runNamespace = `e2e-lb-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
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

function token(body: object): { id: string; updatedAt: string } {
  const value = body as { id: string; updatedAt: string };
  return { id: value.id, updatedAt: value.updatedAt };
}

type Lesson = { courseId: string; slug: string; lessonId: string };

async function createPublishedLesson(
  admin: APIRequestContext,
  baseURL: string,
  label: string,
): Promise<Lesson> {
  const slug = `${runNamespace}-${label}`;
  const draft = {
    slug,
    title: `Конструктор уроку ${label} ${runNamespace}`,
    summary: 'Run-scoped lesson-builder browser fixture.',
    description: 'A course created only for the lesson-builder browser journey.',
    instructorId: instructorID,
  };
  const created = await admin.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const initial = token((await created.json()) as object);
  const lessonId = randomUUID();
  const modules = [
    {
      id: randomUUID(),
      title: `Модуль ${label}`,
      lessons: [{ id: lessonId, title: `Урок ${label}` }],
    },
  ];
  const saved = await admin.put(`/api/v1/admin/courses/${initial.id}`, {
    data: { expectedUpdatedAt: initial.updatedAt, course: draft, modules },
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const current = token((await saved.json()) as object);
  const published = await admin.post(`/api/v1/admin/courses/${current.id}/status`, {
    data: { expectedUpdatedAt: current.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(published.status(), await published.text()).toBe(200);
  // Free published course: the student must be an active member to consume and to load images.
  await sql(
    `INSERT INTO course_grants(user_id,course_id,source,state,granted_at) VALUES (${literal(studentID)}::uuid,${literal(initial.id)}::uuid,'self','active',now()) ON CONFLICT (user_id,course_id) DO NOTHING`,
  );
  return { courseId: initial.id, slug, lessonId };
}

test('an admin builds a document, publishes it, and a member sees it rendered with its image', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('baseURL is required');
  const adminContext = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  const admin = await adminContext.newPage();
  const lesson = await createPublishedLesson(adminContext.request, baseURL, 'read');

  await admin.goto(`/admin/courses/${lesson.courseId}/lessons/${lesson.lessonId}`);
  await expect(admin.getByRole('heading', { name: 'Редактор уроку' })).toBeVisible();

  // Text section with a heading.
  await admin
    .getByRole('group', { name: 'Додати секцію' })
    .getByRole('button', { name: 'Текст' })
    .click();
  const textbox = admin.getByRole('textbox', { name: 'Текст' });
  await textbox.click();
  await textbox.pressSequentially('Замовляння практики');

  // Image section: upload the pixel and describe it.
  await admin
    .getByRole('group', { name: 'Додати секцію' })
    .getByRole('button', { name: 'Зображення' })
    .click();
  await admin.getByLabel('Завантажити зображення').setInputFiles({
    name: 'practice.png',
    mimeType: 'image/png',
    buffer: onePixelPng,
  });
  await expect(admin.locator('.lesson-editor-image-preview')).toBeVisible();
  await admin.getByLabel('Опис зображення').fill('Схема практики');

  // Task section.
  await admin
    .getByRole('group', { name: 'Додати секцію' })
    .getByRole('button', { name: 'Завдання' })
    .click();
  const taskBox = admin.getByRole('textbox', { name: 'Завдання' });
  await taskBox.click();
  await taskBox.pressSequentially('Виконайте вправу');

  await admin.getByRole('button', { name: 'Опублікувати' }).click();
  // No error surfaced by the save+publish.
  await expect(admin.getByRole('alert')).toHaveCount(0);

  // The member reads the published lesson.
  const studentContext = await browser.newContext({
    baseURL,
    storageState: authStatePath('student'),
  });
  const student = await studentContext.newPage();
  const imageResponse = student.waitForResponse((response) =>
    /\/media\/i\/[0-9a-f-]+\.png$/.test(new URL(response.url()).pathname),
  );
  await student.goto(`/learn/${lesson.slug}/${lesson.lessonId}`);

  const region = student.getByRole('region', { name: 'Матеріали уроку' });
  await expect(region.getByText('Замовляння практики')).toBeVisible();
  await expect(region.getByText('Виконайте вправу')).toBeVisible();
  const image = region.getByRole('img', { name: 'Схема практики' });
  await expect(image).toBeVisible();
  expect((await imageResponse).status()).toBe(200);
  // The image actually decoded in the browser, so the gated /media/i serve really delivered bytes.
  expect(await image.evaluate((node: HTMLImageElement) => node.naturalWidth)).toBeGreaterThan(0);

  await adminContext.close();
  await studentContext.close();
});

test('reordering a section by keyboard keeps focus on its control, and the editor passes axe at 200% zoom', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('baseURL is required');
  const adminContext = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  const admin = await adminContext.newPage();
  const lesson = await createPublishedLesson(adminContext.request, baseURL, 'a11y');

  await admin.goto(`/admin/courses/${lesson.courseId}/lessons/${lesson.lessonId}`);
  await expect(admin.getByRole('heading', { name: 'Редактор уроку' })).toBeVisible();
  const addGroup = admin.getByRole('group', { name: 'Додати секцію' });
  await addGroup.getByRole('button', { name: 'Відео' }).click();
  await addGroup.getByRole('button', { name: 'Зображення' }).click();

  // Focus the second section's move-up control and activate it by keyboard.
  const sections = admin.locator('.lesson-editor-section');
  await expect(sections).toHaveCount(2);
  const moveUp = sections.nth(1).getByRole('button', { name: 'Пересунути вгору' });
  await moveUp.focus();
  await admin.keyboard.press('Enter');

  // The image is now first, and focus never fell to <body> (the isDisabled trap) — it stays on a button.
  await expect(sections.first()).toHaveAttribute('data-type', 'image');
  const focusedTag = await admin.evaluate(() => document.activeElement?.tagName.toLowerCase());
  expect(focusedTag).toBe('button');

  await admin.addScriptTag({ content: axe.source });
  await admin.addStyleTag({ content: 'html { font-size: 200% !important; }' });
  const violations = await admin.evaluate(async () => {
    const result = await (window as unknown as { axe: typeof axe }).axe.run(document, {
      runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'],
    });
    return result.violations.map((violation) => violation.id);
  });
  expect(violations).toEqual([]);

  await adminContext.close();
});

test('a saved but unpublished draft stays invisible to the member', async ({
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('baseURL is required');
  const adminContext = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  const admin = await adminContext.newPage();
  const lesson = await createPublishedLesson(adminContext.request, baseURL, 'draft');

  await admin.goto(`/admin/courses/${lesson.courseId}/lessons/${lesson.lessonId}`);
  await admin
    .getByRole('group', { name: 'Додати секцію' })
    .getByRole('button', { name: 'Текст' })
    .click();
  const textbox = admin.getByRole('textbox', { name: 'Текст' });
  await textbox.click();
  await textbox.pressSequentially('Чернетка, яку не публікували');
  await admin.getByRole('button', { name: 'Зберегти чернетку' }).click();
  await expect(admin.getByRole('alert')).toHaveCount(0);

  const studentContext = await browser.newContext({
    baseURL,
    storageState: authStatePath('student'),
  });
  const student = await studentContext.newPage();
  await student.goto(`/learn/${lesson.slug}/${lesson.lessonId}`);
  // The lesson page loads (its outline is always present), but the never-published draft text is
  // absent: the member only ever sees published_body.
  await expect(student.getByRole('heading', { name: 'Зміст курсу' })).toBeVisible();
  await expect(student.getByText('Чернетка, яку не публікували')).toHaveCount(0);

  await adminContext.close();
  await studentContext.close();
});
