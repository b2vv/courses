import { createHmac, randomUUID } from 'node:crypto';
import { expect, test, type BrowserContext, type Page } from '@playwright/test';
import axe from 'axe-core';
import type { AdminCourseTree } from '../src/adapters/api/generated/models';
import { authStatePath } from './auth';
import { contractField } from './contract';

declare global {
  interface Window {
    axe: typeof axe;
  }
}

const runNamespace = `e2e-refund-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const merchant = 'e2e_test_merchant';
const secret = 'e2e-wayforpay-secret';
const amount = '1890.00';

function origin(baseURL: string) {
  return { Origin: new URL(baseURL).origin };
}

function token(payload: object) {
  const id = contractField<AdminCourseTree>(payload, 'id');
  const updatedAt = contractField<AdminCourseTree>(payload, 'updatedAt');
  if (typeof id !== 'string' || typeof updatedAt !== 'string')
    throw new Error('invalid course token');
  return { id, updatedAt };
}

function signature(...values: string[]) {
  return createHmac('md5', secret).update(values.join(';')).digest('hex');
}

function approvedCallback(orderReference: string) {
  return {
    merchantAccount: merchant,
    orderReference,
    merchantSignature: signature(
      merchant,
      orderReference,
      amount,
      'UAH',
      'e2e',
      '44****1212',
      'Approved',
      '1100',
    ),
    amount,
    currency: 'UAH',
    authCode: 'e2e',
    cardPan: '44****1212',
    transactionStatus: 'Approved',
    reasonCode: '1100',
  };
}

async function axeViolations(page: Page) {
  await page.addScriptTag({ content: axe.source });
  return page.evaluate(async () => {
    const result = await window.axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21aa'] });
    return result.violations.map((item) => item.id);
  });
}

async function createPaidCourse(admin: BrowserContext, baseURL: string) {
  const slug = runNamespace;
  const draft = {
    slug,
    title: `Refund ${runNamespace}`,
    summary: 'Refund e2e course.',
    description: 'A run-scoped paid course for the refund queue.',
    instructorId: adminID,
  };
  const modules = [
    {
      id: randomUUID(),
      title: 'Модуль',
      lessons: [{ id: randomUUID(), title: 'Урок' }],
    },
  ];
  const created = await admin.request.post('/api/v1/admin/courses', {
    data: draft,
    headers: origin(baseURL),
  });
  expect(created.status(), await created.text()).toBe(201);
  const createdToken = token((await created.json()) as object);
  const structured = await admin.request.put(`/api/v1/admin/courses/${createdToken.id}`, {
    data: { expectedUpdatedAt: createdToken.updatedAt, course: draft, modules },
    headers: origin(baseURL),
  });
  expect(structured.status(), await structured.text()).toBe(200);
  const afterStructure = (await (
    await admin.request.get(`/api/v1/admin/courses/${createdToken.id}`)
  ).json()) as { updatedAt: string };
  const priced = await admin.request.put(`/api/v1/admin/courses/${createdToken.id}`, {
    data: {
      expectedUpdatedAt: afterStructure.updatedAt,
      course: { ...draft, price: { minor: 189000, currency: 'UAH' } },
      modules,
    },
    headers: origin(baseURL),
  });
  expect(priced.status(), await priced.text()).toBe(200);
  const pricedTree = (await priced.json()) as { updatedAt: string };
  const publishedResponse = await admin.request.post(
    `/api/v1/admin/courses/${createdToken.id}/status`,
    {
      data: { expectedUpdatedAt: pricedTree.updatedAt, status: 'published' },
      headers: origin(baseURL),
    },
  );
  expect(publishedResponse.status(), await publishedResponse.text()).toBe(200);
  return { courseId: createdToken.id };
}

test.use({ storageState: authStatePath('student') });

test('Rf-e2e refund queue surfaces a grant conflict, revoke and resolve clear it', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await browser.newContext({ baseURL, storageState: authStatePath('admin') });

  let courseId: string;
  let userID: string;
  try {
    ({ courseId } = await createPaidCourse(admin, baseURL));

    // Student starts a purchase (pending order) BEFORE any grant exists, so CanStartOrder allows it.
    const start = await page.request.post(`/api/v1/learn/courses/${courseId}/purchase`, {
      headers: { ...origin(baseURL), 'Idempotency-Key': `refund-${runNamespace}` },
    });
    expect(start.status(), await start.text()).toBe(201);
    const order = (await start.json()) as { id: string };

    // The student's own id, read from their session, to grant them access as admin.
    const me = (await (await page.request.get('/api/v1/auth/me')).json()) as { id: string };
    userID = me.id;

    // Admin grants access while the order is pending — now the settle will conflict.
    const granted = await admin.request.post(`/api/v1/admin/courses/${courseId}/grants`, {
      data: { userId: userID, reason: 'granted while payment pending (refund e2e)' },
      headers: origin(baseURL),
    });
    expect(granted.status(), await granted.text()).toBe(201);

    // A valid Approved callback settles the order; the grant insert conflicts with the admin grant,
    // so the order is flagged refund_required with reason grant_conflict (production path).
    const callback = await page.request.post('/api/v1/payments/wayforpay/callback', {
      data: approvedCallback(order.id),
    });
    expect(callback.status(), await callback.text()).toBe(200);
  } finally {
    await admin.close();
  }

  const adminPage = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  try {
    const refunds = await adminPage.newPage();
    await refunds.goto('/admin/refunds');
    await expect(refunds.getByRole('heading', { name: 'Повернення коштів' })).toBeVisible();

    // A1: the flagged order is listed, scoped to this run's course, with the grant-conflict reason.
    const row = refunds.locator('li', { hasText: courseId });
    await expect(row).toBeVisible();
    await expect(row.getByText('Доступ уже був')).toBeVisible();

    const violations = await axeViolations(refunds);
    expect(violations).toEqual([]);

    // A2: revoke the access from the row's confirm dialog. A successful revoke closes the dialog
    // (the deeper effect — grant → revoked — is proven at integration, Rf4/Rf6).
    await row.getByRole('button', { name: 'Відкликати доступ' }).click();
    const dialog = refunds.getByRole('dialog');
    await dialog.getByRole('button', { name: 'Відкликати доступ' }).click();
    await expect(dialog).toHaveCount(0);
    await expect(refunds.getByRole('alert')).toHaveCount(0);

    // A14: once the grant is revoked, the queue refetches and the row no longer offers revoke —
    // only resolve remains. (The row stays until resolve; revoke and resolve are independent.)
    await expect(row.getByRole('button', { name: 'Відкликати доступ' })).toHaveCount(0);

    // A3: resolve removes the row from the queue.
    await row.getByRole('button', { name: 'Позначити розглянутим' }).click();
    await expect(refunds.locator('li', { hasText: courseId })).toHaveCount(0);
  } finally {
    await adminPage.close();
  }
});
