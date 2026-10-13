import { randomUUID } from 'node:crypto';
import { expect, test, type Browser, type BrowserContext, type Locator } from '@playwright/test';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authFixtures, authStatePath, trackPageRequests } from './auth';
import { contractField } from './contract';

const runNamespace = `e2e-admin-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';

let finishRequestTracking = () => Promise.resolve();

test.beforeEach(({ page }) => {
  finishRequestTracking = trackPageRequests(page);
});

test.afterEach(async () => {
  await finishRequestTracking();
});

type Draft = {
  slug: string;
  title: string;
  summary: string;
  description: string;
};

type CreatedCourse = { id: string; updatedAt: string };

async function fillDraft(dialog: Locator, draft: Draft) {
  await dialog.getByLabel('Назва').fill(draft.title);
  await dialog.getByLabel('Slug').fill(draft.slug);
  await dialog.getByLabel('Короткий опис').fill(draft.summary);
  await dialog.getByLabel('Опис', { exact: true }).fill(draft.description);
}

function createdCourse(payload: unknown): CreatedCourse {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('Admin create response must be an object');
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('Admin create response must carry id and updatedAt');
  return { id, updatedAt };
}

async function openAdminContext(browser: Browser, baseURL: string): Promise<BrowserContext> {
  return browser.newContext({
    baseURL,
    storageState: authStatePath('admin'),
  });
}

async function createAdminDraft(
  context: BrowserContext,
  baseURL: string,
  draft: Draft,
): Promise<CreatedCourse> {
  const response = await context.request.post('/api/v1/admin/courses', {
    data: { ...draft, instructorId: adminID },
    headers: { Origin: new URL(baseURL).origin },
  });
  expect(response.status()).toBe(201);
  return createdCourse(await response.json());
}

async function deleteAdminDraft(context: BrowserContext, baseURL: string, course: CreatedCourse) {
  const response = await context.request.delete(`/api/v1/admin/courses/${course.id}`, {
    data: { expectedUpdatedAt: course.updatedAt },
    headers: { Origin: new URL(baseURL).origin },
  });
  expect(response.status()).toBe(204);
}

test.describe('with a stored admin session', () => {
  test.use({ storageState: authStatePath('admin') });

  test('P20 creates, authors, publishes, and archives a course', async ({ page }) => {
    const draft: Draft = {
      slug: `${runNamespace}-admin`,
      title: `A14 адмін ${runNamespace}`,
      summary: 'Короткий опис браузерної перевірки адміністратора.',
      description: 'Повний опис курсу, створеного у браузерній подорожі A14.',
    };
    const moduleTitle = `Модуль ${runNamespace}`;
    const lessonTitle = `Урок ${runNamespace}`;

    await page.goto('/admin');
    await expect(page.getByRole('heading', { name: 'Адміністрування' })).toBeVisible();

    await page.getByRole('button', { name: 'Створити курс', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: 'Створити курс' });
    await fillDraft(dialog, draft);
    await dialog.getByRole('button', { name: 'Створити', exact: true }).click();

    await expect(page).toHaveURL(/\/admin\/courses\/[0-9a-f-]+$/);
    const editorPath = new URL(page.url()).pathname;
    await expect(page.getByRole('heading', { name: draft.title })).toBeVisible();

    await page.getByRole('button', { name: 'Додати модуль', exact: true }).click();
    await page.getByLabel('Назва модуля 1').fill(moduleTitle);
    await page.getByRole('button', { name: 'Додати урок до модуля 1', exact: true }).click();
    await page.getByLabel('Назва уроку 1 у модулі 1').fill(lessonTitle);
    await page.getByRole('button', { name: 'Зберегти зміни', exact: true }).click();
    await expect(page.getByRole('status').filter({ hasText: 'Зміни збережено.' })).toBeVisible();
    await expect(
      page.getByRole('status').filter({ hasText: /Відсутні транскрипти/ }),
    ).toBeVisible();

    await page.getByRole('button', { name: 'Опублікувати', exact: true }).click();
    await expect(
      page.getByRole('status').filter({ hasText: 'Статус курсу оновлено.' }),
    ).toBeVisible();
    await expect(page.getByText('Опубліковано', { exact: true })).toBeVisible();

    await page.goto(`/courses?q=${encodeURIComponent(draft.title)}`);
    await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
    await page.getByRole('heading', { name: draft.title, exact: true }).getByRole('link').click();
    await expect(page).toHaveURL((url) => url.pathname === `/courses/${draft.slug}`);
    await expect(page.getByRole('heading', { level: 1, name: draft.title })).toBeVisible();
    await expect(page.getByRole('heading', { level: 3, name: moduleTitle })).toBeVisible();
    await expect(page.getByText(lessonTitle, { exact: true })).toBeVisible();

    await page.goto('/admin?status=published');
    const publishedRow = page.getByRole('row').filter({ hasText: draft.title });
    await expect(publishedRow).toContainText('Опубліковано');

    await page.goto(editorPath);
    await page.getByRole('button', { name: 'Архівувати', exact: true }).click();
    await expect(
      page.getByRole('status').filter({ hasText: 'Статус курсу оновлено.' }),
    ).toBeVisible();
    await expect(page.getByText('В архіві', { exact: true })).toBeVisible();

    await page.goto(`/courses?q=${encodeURIComponent(draft.title)}`);
    await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
    await expect(page.getByRole('heading', { name: draft.title, exact: true })).toHaveCount(0);
  });
});

test.describe('with a stored instructor session', () => {
  test.use({ storageState: authStatePath('instructor') });

  test('P21 manages only owned courses and is refused on a foreign course', async ({
    page,
    browser,
    baseURL,
  }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const foreignDraft: Draft = {
      slug: `${runNamespace}-foreign`,
      title: `A14 чужий ${runNamespace}`,
      summary: 'Курс іншого власника для перевірки відмови.',
      description: 'Цей курс не належить викладачу з браузерної подорожі.',
    };
    const ownDraft: Draft = {
      slug: `${runNamespace}-instructor`,
      title: `A14 викладач ${runNamespace}`,
      summary: 'Короткий опис браузерної перевірки викладача.',
      description: 'Повний опис курсу, створеного викладачем у подорожі A14.',
    };
    const changedTitle = `${ownDraft.title} оновлено`;
    const adminContext = await openAdminContext(browser, baseURL);
    const foreignCourse = await createAdminDraft(adminContext, baseURL, foreignDraft);
    try {
      await page.goto('/admin');
      await expect(page.getByText(foreignDraft.title, { exact: true })).toHaveCount(0);

      await page.getByRole('button', { name: 'Створити курс', exact: true }).click();
      const dialog = page.getByRole('dialog', { name: 'Створити курс' });
      await expect(dialog.getByText('Власник', { exact: true })).toBeVisible();
      await expect(dialog.getByText(authFixtures.instructor.name, { exact: true })).toBeVisible();
      await expect(dialog.getByRole('button', { name: /Власник/ })).toHaveCount(0);
      await fillDraft(dialog, ownDraft);
      await dialog.getByRole('button', { name: 'Створити', exact: true }).click();

      await expect(page.getByRole('heading', { name: ownDraft.title })).toBeVisible();
      const ownEditorPath = new URL(page.url()).pathname;
      await expect(
        page
          .locator('.admin-readonly-field')
          .getByText(authFixtures.instructor.name, { exact: true }),
      ).toBeVisible();
      await expect(page.getByRole('button', { name: /Власник/ })).toHaveCount(0);
      await page.getByLabel('Назва').fill(changedTitle);
      await page.getByRole('button', { name: 'Зберегти зміни', exact: true }).click();
      await expect(page.getByRole('status').filter({ hasText: 'Зміни збережено.' })).toHaveText(
        'Зміни збережено.',
      );
      await page.getByRole('button', { name: 'Опублікувати', exact: true }).click();
      await expect(page.getByText('Опубліковано', { exact: true })).toBeVisible();

      await page.goto('/admin?status=published');
      const ownRow = page.getByRole('row').filter({ hasText: changedTitle });
      await expect(ownRow).toContainText('Опубліковано');
      await expect(page.getByText(foreignDraft.title, { exact: true })).toHaveCount(0);

      await page.goto(ownEditorPath);
      await page.getByRole('button', { name: 'Архівувати', exact: true }).click();
      await expect(page.getByText('В архіві', { exact: true })).toBeVisible();

      const refusal = page.waitForResponse((response) => {
        return (
          new URL(response.url()).pathname === `/api/v1/admin/courses/${foreignCourse.id}` &&
          response.request().method() === 'GET'
        );
      });
      await page.goto(`/admin/courses/${foreignCourse.id}`);
      expect((await refusal).status()).toBe(403);
      await expect(
        page.getByRole('heading', { name: 'Не вдалося завантажити курси' }),
      ).toBeVisible();
      await expect(page.getByRole('heading', { name: foreignDraft.title })).toHaveCount(0);
    } finally {
      try {
        await deleteAdminDraft(adminContext, baseURL, foreignCourse);
      } finally {
        await adminContext.close();
      }
    }
  });
});

test.describe('with a stored student session', () => {
  test.use({ storageState: authStatePath('student') });

  test('R33 direct API access is refused and cannot mutate', async ({ page, browser, baseURL }) => {
    if (baseURL === undefined) throw new Error('Playwright baseURL is required');
    const refusedDraft: Draft = {
      slug: `${runNamespace}-refused`,
      title: `A14 відмова ${runNamespace}`,
      summary: 'Курс для перевірки серверної відмови.',
      description: 'Клієнтський гард не є механізмом авторизації.',
    };

    await page.goto('/courses');
    await expect(page.getByText(authFixtures.student.name, { exact: true })).toBeVisible();

    const readResponse = await page.request.get('/api/v1/admin/courses');
    expect(readResponse.status()).toBe(403);
    expect(await readResponse.json()).toMatchObject({ error: { code: 'forbidden' } });

    const mutationResponse = await page.request.post('/api/v1/admin/courses', {
      data: refusedDraft,
      headers: { Origin: new URL(baseURL).origin },
    });
    expect(mutationResponse.status()).toBe(403);
    expect(await mutationResponse.json()).toMatchObject({ error: { code: 'forbidden' } });

    const adminContext = await openAdminContext(browser, baseURL);
    try {
      const created = await createAdminDraft(adminContext, baseURL, refusedDraft);
      await deleteAdminDraft(adminContext, baseURL, created);
    } finally {
      await adminContext.close();
    }
  });
});
