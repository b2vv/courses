import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse, delay } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiTransport, unnamedStatusCode } from './transport';

const url = `${apiOrigin()}/api/v1/courses`;
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

async function call(init: RequestInit) {
  const response = await fetch(url, init);
  return { data: await response.json().catch(() => undefined), status: response.status };
}

test('returns the response for the caller to interpret, whatever the status is', async () => {
  server.use(http.get(url, () => HttpResponse.json({ ok: 1 }, { status: 418 })));
  const outcome = await createApiTransport()('probe', call, 200);
  expect(outcome).toEqual({ ok: true, status: 418, data: { ok: 1 } });
});

test('separates a caller cancellation from a deadline', async () => {
  server.use(
    http.get(url, async () => {
      await delay(50);
      return HttpResponse.json({});
    }),
  );
  const timedOut = await createApiTransport({ timeoutMs: 5 })('probe', call, 200);
  expect(timedOut).toEqual({ ok: false, code: 'request_timeout' });

  const controller = new AbortController();
  const pending = createApiTransport()('probe', call, 200, controller.signal);
  controller.abort();
  expect(await pending).toEqual({ ok: false, code: 'request_cancelled' });
});

test('treats an SSR request that was abandoned as a cancellation, not a fault', async () => {
  server.use(
    http.get(url, async () => {
      await delay(50);
      return HttpResponse.json({});
    }),
  );
  const request = new AbortController();
  const pending = createApiTransport({ requestSignal: request.signal })('probe', call, 200);
  request.abort();
  expect(await pending).toEqual({ ok: false, code: 'request_cancelled' });
});

test('reports an unreachable dependency and a malformed body differently', async () => {
  server.use(http.get(url, () => HttpResponse.error()));
  expect(await createApiTransport()('probe', call, 200)).toEqual({
    ok: false,
    code: 'dependency_unavailable',
  });

  server.use(http.get(url, () => HttpResponse.json({})));
  const malformed = await createApiTransport()(
    'probe',
    () => Promise.reject(new SyntaxError('Unexpected token')),
    200,
  );
  expect(malformed).toEqual({ ok: false, code: 'invalid_response' });
});

test.each([
  [500, 'dependency_unavailable'],
  [502, 'dependency_unavailable'],
  [429, 'dependency_unavailable'],
  [418, 'invalid_response'],
  [404, 'invalid_response'],
] as const)('classifies an unnamed %i as %s', (status, code) => {
  expect(unnamedStatusCode(status)).toBe(code);
});
