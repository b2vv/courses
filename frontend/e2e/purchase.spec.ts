import { randomUUID } from 'node:crypto';
import { expect, test, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

declare global {
  interface Window {
    axe: typeof axe;
  }
}

const runNamespace = `e2e-purchase-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
type CourseToken = { id: string; updatedAt: string };
const published: CourseToken[] = [];

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

async function wcagViolations(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () => {
    const results = await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] });
    return results.violations.map((item) => ({
      id: item.id,
      targets: item.nodes.map((node) => node.target.join(' ')),
    }));
  });
}

/** 200% text zoom on a phone width: nothing may push the page sideways. */
async function overflowAtDoubleText(page: Page) {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.addStyleTag({ content: 'html { font-size: 200% !important; }' });
  return page.evaluate(() =>
    Array.from(document.querySelectorAll('body *'))
      .filter((element) => element.getBoundingClientRect().right > window.innerWidth + 1)
      .map((element) => `${element.tagName}.${element.className}`),
  );
}

test.use({ storageState: authStatePath('student') });

// Published courses are counted by the catalog: whoever publishes one archives it again.
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
  } finally {
    await context.close();
  }
});

test('Pay11 Pay13 Pay10 Pay09 a student without access sees the outline, opens the preview and starts a purchase — 44px, axe light/dark, keyboard, 200% text', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const slug = `${runNamespace}-paid`;
  const lockedLesson = { id: randomUUID(), title: `Закритий урок ${runNamespace}` };
  // The preview is the SECOND lesson: the defect fixed in slice 1 read only the first.
  const previewLesson = { id: randomUUID(), title: `Відкритий урок ${runNamespace}` };
  let courseId = '';
  const admin = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    const draft = {
      slug,
      title: `Платний курс ${runNamespace}`,
      summary: 'A run-scoped paid course.',
      description: 'A paid course with a preview on its second lesson.',
      instructorId: adminID,
    };
    const created = await admin.request.post('/api/v1/admin/courses', {
      data: draft,
      headers: origin(baseURL),
    });
    expect(created.status(), await created.text()).toBe(201);
    const course = token((await created.json()) as object);
    courseId = course.id;
    const modules = [
      {
        id: randomUUID(),
        title: `Модуль ${runNamespace}`,
        lessons: [lockedLesson, previewLesson],
      },
    ];
    const structured = await admin.request.put(`/api/v1/admin/courses/${course.id}`, {
      data: { expectedUpdatedAt: course.updatedAt, course: draft, modules },
      headers: origin(baseURL),
    });
    expect(structured.status(), await structured.text()).toBe(200);

    // The preview and the price go through the editor the author uses: the preview with the
    // course form, the first price on its own afterwards (the server refuses it combined).
    const editor = await admin.newPage();
    await editor.goto(`/admin/courses/${course.id}`);
    // The per-lesson preview toggle lives on the Lessons tab; the price lives on Overview (view-3).
    await editor.getByRole('tab', { name: /Lessons|Уроки/ }).click();
    // React Aria hides the native input; the label is what a person clicks.
    await editor.getByText('Відкритий до покупки — урок 2, модуль 1').click();
    await expect(
      editor.getByRole('checkbox', { name: 'Відкритий до покупки — урок 2, модуль 1' }),
    ).toBeChecked();
    await editor.getByRole('tab', { name: /Overview|Огляд/ }).click();
    const amount = editor.getByRole('textbox', { name: 'Ціна, грн' });
    await expect(amount).toBeDisabled();
    await editor.getByRole('button', { name: 'Зберегти зміни' }).click();
    await expect(amount).toBeEnabled();
    await amount.fill('1890');
    await editor.getByRole('button', { name: 'Зберегти ціну' }).click();
    await expect(editor.getByText(/Поточна ціна: 1\s?890/)).toBeVisible();
    expect(await wcagViolations(editor)).toEqual([]);

    const stored = await admin.request.get(`/api/v1/admin/courses/${course.id}`);
    expect(stored.status(), await stored.text()).toBe(200);
    const storedTree = (await stored.json()) as {
      updatedAt: string;
      price: unknown;
      modules: Array<{ lessons: Array<{ isPreview?: boolean }> }>;
    };
    expect(storedTree.price).toEqual({ minor: 189000, currency: 'UAH' });
    expect(storedTree.modules[0]?.lessons.map((lesson) => lesson.isPreview)).toEqual([false, true]);
    const status = await admin.request.post(`/api/v1/admin/courses/${course.id}/status`, {
      data: { expectedUpdatedAt: storedTree.updatedAt, status: 'published' },
      headers: origin(baseURL),
    });
    expect(status.status(), await status.text()).toBe(200);
    published.push(token((await status.json()) as object));
  } finally {
    await admin.close();
  }

  const contentRequests: string[] = [];
  // Content of the locked lesson: its transcript and comments by path, its playback
  // session by body (the session path carries no lesson id).
  page.on('request', (request) => {
    const path = new URL(request.url()).pathname;
    const byPath =
      path === `/api/v1/learn/lessons/${lockedLesson.id}/transcript` ||
      path === `/api/v1/learn/lessons/${lockedLesson.id}/comments`;
    const byBody =
      path === '/api/v1/playback/sessions' && (request.postData() ?? '').includes(lockedLesson.id);
    if (byPath || byBody) contentRequests.push(path);
  });

  // Pay11 — the locked lesson: structure, price and the purchase offer, never its content.
  await page.goto(`/learn/${slug}/${lockedLesson.id}`);
  await expect(page.getByRole('heading', { level: 1, name: lockedLesson.title })).toBeVisible();
  await expect(page.getByText(/1\s?890/)).toBeVisible();
  const outline = page.getByRole('navigation', { name: 'Зміст курсу' });
  await expect(
    outline.getByRole('link', { name: new RegExp(`${lockedLesson.title}.*Після покупки`) }),
  ).toBeVisible();
  await expect(
    outline.getByRole('link', { name: new RegExp(`${previewLesson.title}.*Відкритий урок`) }),
  ).toBeVisible();
  const buy = page.getByRole('button', { name: 'Купити курс' });
  await expect(buy).toBeVisible();
  const box = await buy.boundingBox();
  expect(box?.height ?? 0).toBeGreaterThanOrEqual(44);
  expect(await wcagViolations(page)).toEqual([]);
  await page.emulateMedia({ colorScheme: 'dark' });
  expect(await wcagViolations(page)).toEqual([]);
  await page.emulateMedia({ colorScheme: 'light' });
  expect(await overflowAtDoubleText(page)).toEqual([]);
  await expect(page.getByRole('button', { name: 'Купити курс' })).toBeVisible();
  const zoomedBox = await page.getByRole('button', { name: 'Купити курс' }).boundingBox();
  expect(zoomedBox?.height ?? 0).toBeGreaterThanOrEqual(44);
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.reload();
  await expect(page.getByRole('heading', { level: 1, name: lockedLesson.title })).toBeVisible();
  expect(contentRequests).toEqual([]);

  // Pay11 — the preview lesson opens: no purchase offer, and its discussion is served.
  await outline.getByRole('link', { name: new RegExp(previewLesson.title) }).click();
  // A lesson without a video names that state instead of its title, so the outline's
  // current marker is what says which lesson is open.
  await expect(
    outline.getByRole('link', { name: new RegExp(`${previewLesson.title}.*Поточний урок`) }),
  ).toBeVisible();
  await expect(page.getByRole('button', { name: 'Купити курс' })).toHaveCount(0);
  await expect(
    page.getByRole('region', { name: 'Обговорення' }).getByText('Коментарів ще немає.'),
  ).toBeVisible();

  // Pay13 — buying by keyboard leads to the order page, which names the pending state.
  await outline.getByRole('link', { name: new RegExp(lockedLesson.title) }).click();
  await page.getByRole('button', { name: 'Купити курс' }).focus();
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(new RegExp(`/orders/[0-9a-f-]{36}\\?course=${slug}$`));
  await expect(
    page.getByText('Чекаємо підтвердження оплати. Сторінка оновиться сама.'),
  ).toBeVisible();
  await expect(page.getByRole('link', { name: 'Повернутися до курсу' })).toHaveAttribute(
    'href',
    `/learn/${slug}`,
  );
  expect(await wcagViolations(page)).toEqual([]);
  expect(await overflowAtDoubleText(page)).toEqual([]);
  await expect(page.getByRole('link', { name: 'Повернутися до курсу' })).toBeVisible();
  expect(contentRequests).toEqual([]);

  // SA05/M2 — one pending order per course. Returning to the locked lesson and buying again resumes
  // the SAME order (the frontend mints a fresh Idempotency-Key per attempt), never a second order
  // that could also be paid and charge the buyer twice.
  const firstOrderId = page.url().match(/\/orders\/([0-9a-f-]{36})/)?.[1];
  expect(firstOrderId).toBeTruthy();
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(`/learn/${slug}/${lockedLesson.id}`);
  await page.getByRole('button', { name: 'Купити курс' }).click();
  await expect(page).toHaveURL(new RegExp(`/orders/${firstOrderId}\\?course=${slug}$`));

  // Pay10 — an administrator grants the student access in the editor; the locked lesson opens.
  const grantor = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    const editor = await grantor.newPage();
    await editor.goto(`/admin/courses/${courseId}`);
    // The grant tool lives on the Access tab of the course editor (view-3 layout).
    await editor.getByRole('tab', { name: /Access|Доступ/ }).click();
    const grant = editor.getByRole('region', { name: 'Видати доступ до курсу' });
    await grant
      .getByRole('combobox', { name: 'Користувач' })
      .pressSequentially('student@courseforge');
    await editor
      .getByRole('option', { name: 'Студентка Курсу — student@courseforge.local' })
      .click();
    await grant.getByRole('textbox', { name: 'Причина' }).fill(`e2e ${runNamespace}`);
    await grant.getByRole('button', { name: 'Видати доступ' }).click();
    await expect(
      grant.getByText('Доступ видано: Студентка Курсу — student@courseforge.local.'),
    ).toBeVisible();
    expect(await wcagViolations(editor)).toEqual([]);
    // A second grant for the same learner is refused by name, not duplicated.
    await grant.getByRole('textbox', { name: 'Причина' }).fill('again');
    await grant.getByRole('button', { name: 'Видати доступ' }).click();
    await expect(grant.getByRole('alert')).toHaveText('У цього користувача вже є доступ до курсу.');
  } finally {
    await grantor.close();
  }
  await page.setViewportSize({ width: 1280, height: 720 });
  await page.goto(`/learn/${slug}/${lockedLesson.id}`);
  await expect(page.getByRole('region', { name: 'Обговорення' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Купити курс' })).toHaveCount(0);
  await expect(page.getByText('Після покупки')).toHaveCount(0);
});
