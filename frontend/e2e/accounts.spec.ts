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
const runNamespace = `e2e-accounts-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const initialEmail = `${runNamespace}@example.test`;
const changedEmail = `${runNamespace}-changed@example.test`;
const password = 'Accounts e2e password 1';
const resetPassword = 'Accounts e2e password 2';
const changedPassword = 'Accounts e2e password 3';
const instructorID = '40000000-0000-4000-8000-000000000002';
let course: { id: string; updatedAt: string; slug: string; lessonId: string } | undefined;

test.describe.configure({ mode: 'serial', timeout: 120_000 });

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

function token(payload: object) {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('admin course response must carry id and updatedAt');
  return { id, updatedAt };
}

async function openAs(browser: Browser, baseURL: string, role: 'admin') {
  return browser.newContext({ baseURL, storageState: authStatePath(role) });
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

async function createCourse(admin: APIRequestContext, baseURL: string) {
  const slug = `${runNamespace}-verification-gate`;
  const draft = {
    slug,
    title: `Accounts verification gate ${runNamespace}`,
    summary: 'Run-scoped account self-service fixture.',
    description: 'A course visible before email verification.',
    instructorId: instructorID,
  };
  const created = await admin.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const initial = token((await created.json()) as object);
  const lessonId = randomUUID();
  const saved = await admin.put(`/api/v1/admin/courses/${initial.id}`, {
    data: {
      expectedUpdatedAt: initial.updatedAt,
      course: draft,
      modules: [
        {
          id: randomUUID(),
          title: 'Account fixture',
          lessons: [{ id: lessonId, title: 'Visible lesson', description: '' }],
        },
      ],
    },
    headers: origin(baseURL),
  });
  expect(saved.status(), await saved.text()).toBe(200);
  const prepared = token((await saved.json()) as object);
  const published = await admin.post(`/api/v1/admin/courses/${prepared.id}/status`, {
    data: { expectedUpdatedAt: prepared.updatedAt, status: 'published' },
    headers: origin(baseURL),
  });
  expect(published.status(), await published.text()).toBe(200);
  return { ...token((await published.json()) as object), slug, lessonId };
}

function messageRecipients(message: unknown) {
  return JSON.stringify(message).toLowerCase();
}

function mailpitMessages(payload: unknown): unknown[] {
  if (typeof payload !== 'object' || payload === null || !('messages' in payload))
    throw new Error('Mailpit messages response has no messages array');
  const messages = payload.messages;
  if (!Array.isArray(messages)) throw new Error('Mailpit messages is not an array');
  return messages;
}

function messageID(message: unknown) {
  if (
    typeof message !== 'object' ||
    message === null ||
    !('ID' in message) ||
    typeof message.ID !== 'string'
  )
    throw new Error('Mailpit message has no ID');
  return message.ID;
}

async function messagesFor(email: string) {
  const list = await fetch('http://localhost:8025/api/v1/messages');
  if (!list.ok) throw new Error(`Mailpit list failed: ${list.status}`);
  return mailpitMessages(await list.json()).filter((item) =>
    messageRecipients(item).includes(email.toLowerCase()),
  );
}

function mailBody(payload: unknown) {
  if (typeof payload !== 'object' || payload === null)
    throw new Error('Mailpit message detail is not an object');
  const text = 'Text' in payload ? payload.Text : undefined;
  const html = 'HTML' in payload ? payload.HTML : undefined;
  const body = [text, html].filter((value): value is string => typeof value === 'string');
  if (body.length === 0) throw new Error('Mailpit message detail has no text or HTML body');
  return body.join('\n').replaceAll('&amp;', '&');
}

async function readMailLink(email: string, pathname: string) {
  let result = '';
  await expect
    .poll(
      async () => {
        const messages = await messagesFor(email);
        for (const message of messages) {
          const detail = await fetch(`http://localhost:8025/api/v1/message/${messageID(message)}`);
          if (!detail.ok) throw new Error(`Mailpit detail failed: ${detail.status}`);
          const text = mailBody(await detail.json());
          const link = text.match(/https?:\/\/[^"'\s<>]+/g)?.find((candidate) => {
            try {
              return new URL(candidate).pathname === pathname;
            } catch {
              return false;
            }
          });
          if (link !== undefined) {
            result = link;
            return true;
          }
        }
        return false;
      },
      { timeout: 30_000 },
    )
    .toBe(true);
  return result;
}

async function axeClean(page: Page) {
  await page.addScriptTag({ content: axe.source });
  const violations = await page.evaluate(async () => {
    const audit = await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] });
    return audit.violations.map((violation) => violation.id);
  });
  expect(violations).toEqual([]);
}

async function login(page: Page, email: string, value: string, rememberMe = false) {
  await page.goto('/login');
  await page.getByRole('textbox', { name: 'Електронна пошта' }).fill(email);
  await page.getByLabel('Пароль').fill(value);
  if (rememberMe) await page.getByRole('checkbox', { name: 'Запам’ятати мене' }).check();
  await page.getByRole('button', { name: 'Увійти' }).click();
  await expect(page.getByText('Accounts E2E', { exact: true })).toBeVisible();
}

test('Acc01–Acc08 register, gate unverified playback and mutation, verify, reset, and change account credentials', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await openAs(browser, baseURL, 'admin');
  try {
    course = await createCourse(admin.request, baseURL);
  } finally {
    await admin.close();
  }

  await page.goto('/register');
  await page.getByRole('textbox', { name: 'Ім’я' }).fill('Accounts E2E');
  await page.getByRole('textbox', { name: 'Електронна пошта' }).fill(initialEmail);
  await page.getByLabel('Пароль').fill(password);
  await page.getByRole('button', { name: 'Зареєструватися' }).click();
  await expect(page.getByText('Accounts E2E')).toBeVisible();
  await expect(page.getByLabel('Підтвердження пошти')).toBeVisible();
  await expect.poll(async () => (await messagesFor(initialEmail)).length).toBe(1);
  const verifyLink = await readMailLink(initialEmail, '/verify-email');

  await page.goto(`/courses/${course.slug}`);
  await expect(page.getByText('A course visible before email verification.')).toBeVisible();
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await expect(page.getByText('Accounts verification gate')).toBeVisible();
  await expect(
    page.getByRole('heading', { name: 'Підтвердьте пошту, щоб дивитися' }),
  ).toBeVisible();
  const denied = await page.request.post(`/api/v1/learn/courses/${course.id}/purchase`, {
    headers: { ...origin(baseURL), 'Idempotency-Key': randomUUID() },
  });
  expect(denied.status()).toBe(403);
  expect(await denied.json()).toMatchObject({ error: { code: 'email_not_verified' } });

  await page.goto(verifyLink);
  await expect(page.getByRole('heading', { name: 'Пошту підтверджено' })).toBeVisible();
  await page.goto('/');
  await expect(page.getByLabel('Підтвердження пошти')).toHaveCount(0);
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await expect(page.getByRole('heading', { name: 'Підтвердьте пошту, щоб дивитися' })).toHaveCount(
    0,
  );

  await page.goto('/password/forgot');
  await page.getByRole('textbox', { name: 'Електронна пошта' }).fill(initialEmail);
  await page.getByRole('button', { name: 'Надіслати посилання' }).click();
  await expect(page.getByRole('heading', { name: 'Перевірте пошту' })).toBeVisible();
  const resetSession = await browser.newContext({
    storageState: await page.context().storageState(),
  });
  const resetLink = await readMailLink(initialEmail, '/password/reset');
  await page.goto(resetLink);
  await page.getByLabel('Новий пароль').fill(resetPassword);
  await page.getByRole('button', { name: 'Зберегти пароль' }).click();
  await expect(page.getByText('Пароль оновлено')).toBeVisible();
  try {
    expect((await resetSession.request.get('/api/v1/auth/me')).status()).toBe(401);
  } finally {
    await resetSession.close();
  }
  await login(page, initialEmail, resetPassword);

  await page.goto('/settings/account');
  const otherPasswordSession = await browser.newContext({
    storageState: await page.context().storageState(),
  });
  const passwordForm = page.getByRole('heading', { name: 'Змінити пароль' }).locator('..');
  await passwordForm.getByLabel('Поточний пароль').fill(resetPassword);
  await passwordForm.getByLabel('Новий пароль').fill(changedPassword);
  await passwordForm.getByRole('button', { name: 'Змінити пароль' }).click();
  await expect(passwordForm.getByText('Пароль змінено')).toBeVisible();
  try {
    expect((await page.request.get('/api/v1/auth/me')).status()).toBe(200);
    expect((await otherPasswordSession.request.get('/api/v1/auth/me')).status()).toBe(401);
  } finally {
    await otherPasswordSession.close();
  }

  const emailForm = page.getByRole('heading', { name: 'Змінити електронну пошту' }).locator('..');
  await emailForm.getByRole('textbox', { name: 'Нова електронна пошта' }).fill(changedEmail);
  await emailForm.getByRole('button', { name: 'Змінити пошту' }).click();
  await expect(emailForm.getByText('Перевірте нову пошту')).toBeVisible();
  const changeLink = await readMailLink(changedEmail, '/email/change/confirm');
  expect(
    await page.request.get('/api/v1/auth/me').then((response) => response.json()),
  ).toMatchObject({
    email: initialEmail,
  });
  await page.goto(changeLink);
  await expect(page.getByRole('heading', { name: 'Пошту змінено' })).toBeVisible();
  expect(
    await page.request.get('/api/v1/auth/me').then((response) => response.json()),
  ).toMatchObject({
    email: changedEmail,
    emailVerified: true,
  });

  await page.request.post('/api/v1/auth/logout', { headers: origin(baseURL) });
  await login(page, changedEmail, changedPassword);
});

test('Acc10, Acc16, and Acc19 cover duplicate signup, cookie lifetimes, and axe', async ({
  page,
  context,
}) => {
  await page.goto('/register');
  await page.getByRole('textbox', { name: 'Ім’я' }).fill('Duplicate');
  await page.getByRole('textbox', { name: 'Електронна пошта' }).fill(changedEmail);
  await page.getByLabel('Пароль').fill(password);
  await page.getByRole('button', { name: 'Зареєструватися' }).click();
  await expect(page.getByRole('alert')).toContainText('Ця пошта вже зареєстрована.');

  for (const pathname of ['/register', '/verify-email', '/password/forgot', '/password/reset']) {
    await page.goto(pathname);
    await axeClean(page);
  }

  await login(page, changedEmail, changedPassword);
  const ordinary = (await context.cookies()).find(
    (cookie) => cookie.name === 'courseforge_session',
  );
  expect(ordinary?.expires).toBeGreaterThan(Date.now() / 1000 + 11 * 60 * 60);
  expect(ordinary?.expires).toBeLessThan(Date.now() / 1000 + 13 * 60 * 60);
  await page.goto('/settings/account');
  await axeClean(page);
  await page.request.post('/api/v1/auth/logout', {
    headers: { Origin: new URL(page.url()).origin },
  });
  await login(page, changedEmail, changedPassword, true);
  const remembered = (await context.cookies()).find(
    (cookie) => cookie.name === 'courseforge_session',
  );
  expect(remembered?.expires).toBeGreaterThan(Date.now() / 1000 + 71 * 60 * 60);
  expect(remembered?.expires).toBeLessThan(Date.now() / 1000 + 73 * 60 * 60);
});

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL !== undefined && course !== undefined) {
    const admin = await openAs(browser, baseURL, 'admin');
    try {
      const current = await admin.request.get(`/api/v1/admin/courses/${course.id}`);
      if (current.ok()) {
        const currentToken = token((await current.json()) as object);
        const removed = await admin.request.delete(`/api/v1/admin/courses/${course.id}`, {
          data: { expectedUpdatedAt: currentToken.updatedAt },
          headers: origin(baseURL),
        });
        expect(removed.status(), await removed.text()).toBe(204);
      }
    } finally {
      await admin.close();
    }
  }
  await sql(`DELETE FROM users WHERE email LIKE '${runNamespace.replaceAll("'", "''")}%'`);
});
