import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiPurchaseRepository } from './purchase-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const orderId = '60000000-0000-4000-8000-000000000001';
const order = {
  id: orderId,
  courseId,
  amount: { minor: 189000, currency: 'UAH' },
  status: 'pending',
  createdAt: '2026-10-01T10:00:00Z',
  paidAt: null,
};
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('starts a purchase with the idempotency key and maps the order', async () => {
  let key: string | null = null;
  let accept: string | null = null;
  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/courses/${courseId}/purchase`, ({ request }) => {
      key = request.headers.get('Idempotency-Key');
      accept = request.headers.get('Accept');
      return HttpResponse.json(order, { status: 201 });
    }),
  );

  const result = await createApiPurchaseRepository().start(courseId, 'attempt-1');

  expect(result).toEqual({ ok: true, value: order });
  expect(key).toBe('attempt-1');
  // The transport's own headers survive next to the key.
  expect(accept).toBe('application/json');
});

test('reads the order status and refuses a malformed order', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/orders/${orderId}`, () =>
      HttpResponse.json({ ...order, status: 'paid', paidAt: '2026-10-01T10:01:00Z' }),
    ),
  );
  expect(await createApiPurchaseRepository().status(orderId)).toEqual({
    ok: true,
    value: { ...order, status: 'paid', paidAt: '2026-10-01T10:01:00Z' },
  });

  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/orders/${orderId}`, () =>
      HttpResponse.json({ ...order, status: 'refunded' }),
    ),
  );
  expect(await createApiPurchaseRepository().status(orderId)).toEqual({
    ok: false,
    error: { code: 'invalid_response' },
  });
});

test('names the purchase conflicts the backend answers', async () => {
  const conflict = (code: string, status: number) =>
    HttpResponse.json({ error: { code, message: 'x', requestId: 'a'.repeat(32) } }, { status });

  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/courses/${courseId}/purchase`, () =>
      conflict('course_already_owned', 409),
    ),
    http.get(`${apiOrigin()}/api/v1/learn/orders/${orderId}`, () =>
      conflict('order_not_found', 404),
    ),
  );

  expect(await createApiPurchaseRepository().start(courseId, 'attempt-2')).toEqual({
    ok: false,
    error: { code: 'course_already_owned' },
  });
  expect(await createApiPurchaseRepository().status(orderId)).toEqual({
    ok: false,
    error: { code: 'order_not_found' },
  });
});

test('maps a hosted checkout form and sends the course slug', async () => {
  let requestBody: unknown;
  const form = {
    action: 'https://secure.wayforpay.com/pay',
    fields: [
      { name: 'merchantAccount', value: 'merchant' },
      { name: 'merchantSignature', value: 'signature' },
    ],
  };
  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/orders/${orderId}/checkout`, async ({ request }) => {
      requestBody = await request.json();
      return HttpResponse.json(form);
    }),
  );

  expect(await createApiPurchaseRepository().checkout(orderId, 'go-basics')).toEqual({
    ok: true,
    value: form,
  });
  expect(requestBody).toEqual({ courseSlug: 'go-basics' });
});

test('rejects an unsafe or empty checkout form', async () => {
  for (const response of [
    { action: 'http://secure.wayforpay.com/pay', fields: [{ name: 'x', value: 'y' }] },
    { action: 'https://example.test/pay', fields: [{ name: 'x', value: 'y' }] },
    { action: 'https://secure.wayforpay.com/pay', fields: [] },
  ]) {
    server.use(
      http.post(`${apiOrigin()}/api/v1/learn/orders/${orderId}/checkout`, () =>
        HttpResponse.json(response),
      ),
    );
    expect(await createApiPurchaseRepository().checkout(orderId, 'go-basics')).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  }
});

test('names checkout not-found and unavailable responses', async () => {
  const error = (code: string, status: number) =>
    HttpResponse.json({ error: { code, message: 'x', requestId: 'a'.repeat(32) } }, { status });
  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/orders/${orderId}/checkout`, () =>
      error('order_not_found', 404),
    ),
  );
  expect(await createApiPurchaseRepository().checkout(orderId, 'go-basics')).toEqual({
    ok: false,
    error: { code: 'order_not_found' },
  });
  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/orders/${orderId}/checkout`, () =>
      error('dependency_unavailable', 503),
    ),
  );
  expect(await createApiPurchaseRepository().checkout(orderId, 'go-basics')).toEqual({
    ok: false,
    error: { code: 'dependency_unavailable' },
  });
});
