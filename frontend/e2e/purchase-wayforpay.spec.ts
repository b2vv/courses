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

const runNamespace = `e2e-wayforpay-${Date.now().toString(36)}-${randomUUID().slice(0, 8)}`;
const adminID = '40000000-0000-4000-8000-000000000003';
const merchant = 'e2e_test_merchant';
const secret = 'e2e-wayforpay-secret';
const amount = '1890.00';
const published: Array<{ id: string; updatedAt: string }> = [];

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

function callback(
  orderReference: string,
  merchantSignature = signature(
    merchant,
    orderReference,
    amount,
    'UAH',
    'e2e',
    '44****1212',
    'Approved',
    '1100',
  ),
) {
  return {
    merchantAccount: merchant,
    orderReference,
    merchantSignature,
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

async function createPaidCourse(admin: BrowserContext, baseURL: string, label: string) {
  const lesson = { id: randomUUID(), title: `Оплачуваний урок ${label}` };
  const slug = `${runNamespace}-${label}`;
  const draft = {
    slug,
    title: `WayForPay ${label}`,
    summary: 'Offline WayForPay e2e course.',
    description: 'A run-scoped paid course.',
    instructorId: adminID,
  };
  const modules = [{ id: randomUUID(), title: `Модуль ${label}`, lessons: [lesson] }];
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
  ).json()) as {
    updatedAt: string;
  };
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
  const publishedToken = token((await publishedResponse.json()) as object);
  published.push(publishedToken);
  return { slug, courseId: publishedToken.id, lessonId: lesson.id };
}

test.use({ storageState: authStatePath('student') });

test.afterAll(async ({ browser, baseURL }) => {
  if (baseURL === undefined) return;
  const admin = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
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

test('Pay24 callback vectors and Pay25 offline status-check purchase unlock the course', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  let callbackCourse: { slug: string; courseId: string; lessonId: string };
  let statusCourse: { slug: string; courseId: string; lessonId: string };
  try {
    callbackCourse = await createPaidCourse(admin, baseURL, 'callback');
    statusCourse = await createPaidCourse(admin, baseURL, 'status');
  } finally {
    await admin.close();
  }
  if (callbackCourse === undefined || statusCourse === undefined)
    throw new Error('courses were not created');

  const start = await page.request.post(
    `/api/v1/learn/courses/${callbackCourse.courseId}/purchase`,
    {
      headers: { ...origin(baseURL), 'Idempotency-Key': `callback-${runNamespace}` },
    },
  );
  expect(start.status(), await start.text()).toBe(201);
  const started = (await start.json()) as { id: string; status: string };
  expect(started.status).toBe('pending');

  const bad = await page.request.post('/api/v1/payments/wayforpay/callback', {
    data: callback(started.id, 'bad-signature'),
  });
  expect(bad.status()).toBe(401);
  await expect
    .poll(async () => (await page.request.get(`/api/v1/learn/orders/${started.id}`)).json())
    .toMatchObject({ status: 'pending' });

  const valid = await page.request.post('/api/v1/payments/wayforpay/callback', {
    data: callback(started.id),
  });
  expect(valid.status()).toBe(200);
  const acknowledgement = (await valid.json()) as {
    orderReference: string;
    status: string;
    time: number;
    signature: string;
  };
  expect(acknowledgement).toMatchObject({ orderReference: started.id, status: 'accept' });
  expect(acknowledgement.signature).toBe(
    signature(started.id, 'accept', String(acknowledgement.time)),
  );
  await expect
    .poll(async () => (await page.request.get(`/api/v1/learn/orders/${started.id}`)).json())
    .toMatchObject({ status: 'paid' });

  const replay = await page.request.post('/api/v1/payments/wayforpay/callback', {
    data: callback(started.id),
  });
  expect(replay.status()).toBe(200);
  await expect
    .poll(async () => (await page.request.get(`/api/v1/learn/orders/${started.id}`)).json())
    .toMatchObject({ status: 'paid' });

  await page.goto(`/learn/${callbackCourse.slug}/${callbackCourse.lessonId}`);
  await expect(page.getByRole('region', { name: 'Обговорення' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Купити курс' })).toHaveCount(0);

  const beforeStub = (await (await page.request.get('http://localhost:8090/debug')).json()) as {
    payCount: number;
    statusCount: number;
  };
  await page.goto(`/learn/${statusCourse.slug}/${statusCourse.lessonId}`);
  await page.getByRole('button', { name: 'Купити курс' }).click();
  await expect(page).toHaveURL(new RegExp(`/orders/[0-9a-f-]{36}[?]course=${statusCourse.slug}$`));
  await expect(
    page.getByText('Чекаємо підтвердження оплати. Сторінка оновиться сама.'),
  ).toBeVisible();
  const payButton = page.getByRole('button', { name: 'Перейти до оплати' });
  expect((await payButton.boundingBox())?.height ?? 0).toBeGreaterThanOrEqual(44);
  expect(await axeViolations(page)).toEqual([]);
  await payButton.click();
  await expect(page).toHaveURL(new RegExp(`/orders/[0-9a-f-]{36}[?]course=${statusCourse.slug}$`));
  await expect(page.locator('[data-status="paid"]')).toBeVisible({ timeout: 15_000 });
  const debug = (await (await page.request.get('http://localhost:8090/debug')).json()) as {
    payCount: number;
    statusCount: number;
  };
  expect(debug.payCount).toBeGreaterThan(beforeStub.payCount);
  expect(debug.statusCount).toBeGreaterThan(beforeStub.statusCount);
  expect(await axeViolations(page)).toEqual([]);
  await page.goto(`/learn/${statusCourse.slug}/${statusCourse.lessonId}`);
  await expect(page.getByRole('region', { name: 'Обговорення' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Купити курс' })).toHaveCount(0);
});

// SA05-T1 — the SA05/M2 fix reuses one pending order per (user, course). Pay13 proves a fresh
// Idempotency-Key resumes the SAME order; this test carries that resumed order all the way to a
// paid grant, after the buyer opened the provider and abandoned it, and proves the grant lands
// exactly once (no second chargeable order). It complements Pay24 (which pays a never-abandoned
// order) by exercising the reused-pending path end to end.
test('Pay26 SA05-T1 a resumed pending order after an abandoned provider pays and grants exactly once', async ({
  page,
  browser,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const admin = await browser.newContext({ baseURL, storageState: authStatePath('admin') });
  let course: { slug: string; courseId: string; lessonId: string };
  try {
    course = await createPaidCourse(admin, baseURL, 'resume');
  } finally {
    await admin.close();
  }

  // Attempt 1 opens the one pending order and is sent to the provider (checkout opened), then left
  // without paying.
  const first = await page.request.post(`/api/v1/learn/courses/${course.courseId}/purchase`, {
    headers: { ...origin(baseURL), 'Idempotency-Key': `resume-a-${runNamespace}` },
  });
  expect(first.status(), await first.text()).toBe(201);
  const firstOrder = (await first.json()) as { id: string; status: string };
  expect(firstOrder.status).toBe('pending');
  const opened = await page.request.post(`/api/v1/learn/orders/${firstOrder.id}/checkout`, {
    headers: origin(baseURL),
    data: { courseSlug: course.slug },
  });
  expect(opened.status(), await opened.text()).toBe(200);

  // Attempt 2 mints a fresh Idempotency-Key (as the frontend does per attempt): it resumes the SAME
  // pending order, never a second order that could also be charged (SA05/M2).
  const resumed = await page.request.post(`/api/v1/learn/courses/${course.courseId}/purchase`, {
    headers: { ...origin(baseURL), 'Idempotency-Key': `resume-b-${runNamespace}` },
  });
  expect(resumed.status(), await resumed.text()).toBe(201);
  const resumedOrder = (await resumed.json()) as { id: string; status: string };
  expect(resumedOrder.id).toBe(firstOrder.id);
  expect(resumedOrder.status).toBe('pending');

  // Re-opening checkout on the resumed (already-opened) order is refused: a pre-existing,
  // owner-accepted no-expiry limitation — the buyer completes through the provider it was sent to.
  const reopen = await page.request.post(`/api/v1/learn/orders/${firstOrder.id}/checkout`, {
    headers: origin(baseURL),
    data: { courseSlug: course.slug },
  });
  expect(reopen.status(), await reopen.text()).toBe(409);
  expect(((await reopen.json()) as { error: { code: string } }).error.code).toBe(
    'checkout_already_opened',
  );

  // The provider confirms the payment out-of-band (callback): the resumed order reaches paid.
  const paid = await page.request.post('/api/v1/payments/wayforpay/callback', {
    data: callback(firstOrder.id),
  });
  expect(paid.status(), await paid.text()).toBe(200);
  await expect
    .poll(async () => (await page.request.get(`/api/v1/learn/orders/${firstOrder.id}`)).json())
    .toMatchObject({ status: 'paid' });

  // The grant lands exactly once: the course unlocks, a further purchase is refused as already
  // owned (so there is no second order to charge), and a replayed callback stays paid (idempotent).
  await page.goto(`/learn/${course.slug}/${course.lessonId}`);
  await expect(page.getByRole('region', { name: 'Обговорення' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Купити курс' })).toHaveCount(0);

  const again = await page.request.post(`/api/v1/learn/courses/${course.courseId}/purchase`, {
    headers: { ...origin(baseURL), 'Idempotency-Key': `resume-c-${runNamespace}` },
  });
  expect(again.status(), await again.text()).toBe(409);
  expect(((await again.json()) as { error: { code: string } }).error.code).toBe(
    'course_already_owned',
  );

  const replay = await page.request.post('/api/v1/payments/wayforpay/callback', {
    data: callback(firstOrder.id),
  });
  expect(replay.status()).toBe(200);
  await expect
    .poll(async () => (await page.request.get(`/api/v1/learn/orders/${firstOrder.id}`)).json())
    .toMatchObject({ status: 'paid' });
});
