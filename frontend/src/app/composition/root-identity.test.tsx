import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { RouterProvider } from '@tanstack/react-router';
import { requestHandler } from '@tanstack/react-start/server';
import { render, screen } from '@testing-library/react';
import { delay, http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../adapters/api/api-origin';
import { getRouter } from '../../router';
import { sessionQueryKey } from './identity-queries';
import type { RootIdentity } from './root-identity';

const apiBaseUrl = apiOrigin();
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

afterEach(() => server.resetHandlers());

afterAll(() => server.close());

type LoadedRequest = {
  router: ReturnType<typeof getRouter>;
  identity: RootIdentity;
  rootStatus: string;
  rootError: unknown;
  cachedSession: unknown;
  responseText: string;
};

async function loadRequest(cookie?: string): Promise<LoadedRequest> {
  let loaded: Omit<LoadedRequest, 'responseText'> | undefined;
  const handle = requestHandler(async () => {
    const router = getRouter();
    await router.load();
    const rootMatch = router.state.matches[0];
    if (rootMatch === undefined) throw new Error('The root route did not match');
    const identity = rootMatch.context.identity;
    loaded = {
      router,
      identity,
      rootStatus: rootMatch.status,
      rootError: rootMatch.error,
      cachedSession: router.options.context.queryClient.getQueryData(sessionQueryKey),
    };
    const name = identity.status === 'signed-in' ? identity.user.displayName : null;
    return Response.json({ name });
  });
  const request =
    cookie === undefined
      ? new Request('http://courseforge.test/')
      : new Request('http://courseforge.test/', { headers: { Cookie: cookie } });
  if (cookie !== undefined && request.headers.get('cookie') === null) {
    // happy-dom applies the browser's forbidden-header rule. A server Request
    // retains Cookie, so restore that server-only shape for requestHandler.
    Object.defineProperty(request, 'headers', {
      value: new Headers({ Cookie: cookie }),
    });
  }
  Object.defineProperty(request, 'signal', {
    value: new AbortController().signal,
  });
  const response = await handle(request, {});
  if (loaded === undefined) throw new Error('The SSR route load did not finish');
  return { ...loaded, responseText: await response.text() };
}

const unavailableEnvelope = {
  error: { code: 'dependency_unavailable', message: 'x', requestId: 'a'.repeat(32) },
};

test('G10 keeps the root route successful when identity returns 503', async () => {
  server.use(
    http.get(`${apiBaseUrl}/api/v1/auth/me`, () =>
      HttpResponse.json(unavailableEnvelope, { status: 503 }),
    ),
  );

  const result = await loadRequest('courseforge_session=valid');

  expect(result.identity).toEqual({ status: 'unavailable' });
  expect(result.rootStatus).toBe('success');
  expect(result.rootError).toBeUndefined();
});

test('G10 limits a stalled SSR identity read to the one-second server deadline', async () => {
  server.use(
    http.get(`${apiBaseUrl}/api/v1/auth/me`, async () => {
      await delay(2_000);
      return HttpResponse.json(unavailableEnvelope, { status: 503 });
    }),
  );
  const startedAt = performance.now();

  const result = await loadRequest('courseforge_session=valid');

  expect(result.identity).toEqual({ status: 'unavailable' });
  expect(result.rootStatus).toBe('success');
  expect(result.rootError).toBeUndefined();
  const durationMs = performance.now() - startedAt;
  expect(durationMs).toBeGreaterThanOrEqual(900);
  expect(durationMs).toBeLessThan(1_800);
});

test('G15 renders the public route anonymously when the API returns an unknown role', async () => {
  server.use(
    http.get(`${apiBaseUrl}/api/v1/auth/me`, () =>
      HttpResponse.json({
        id: '40000000-0000-4000-8000-000000000001',
        email: 'student@courseforge.local',
        displayName: 'Студентка',
        role: 'owner',
      }),
    ),
  );

  const result = await loadRequest('courseforge_session=valid');
  render(<RouterProvider router={result.router} />);

  expect(result.identity).toEqual({ status: 'unavailable' });
  expect(result.rootError).toBeUndefined();
  expect(screen.getByRole('link', { name: 'Увійти' })).toBeTruthy();
  expect(screen.queryByRole('alert')).toBeNull();
});

test('G22 makes no identity request without the session cookie and caches anonymous', async () => {
  let identityReads = 0;
  server.use(
    http.get(`${apiBaseUrl}/api/v1/auth/me`, () => {
      identityReads += 1;
      return HttpResponse.json(unavailableEnvelope, { status: 503 });
    }),
  );

  const result = await loadRequest('courseforge_locale=en');

  expect(result.identity).toEqual({ status: 'anonymous' });
  expect(result.cachedSession).toBe(null);
  expect(identityReads).toBe(0);
});

test('G23 keeps concurrent SSR identities isolated by request and response', async () => {
  const users = [
    {
      id: '40000000-0000-4000-8000-000000000001',
      email: 'student@courseforge.local',
      displayName: 'Запит A',
      role: 'student',
      emailVerified: true,
      twoFactorEnabled: false,
    },
    {
      id: '40000000-0000-4000-8000-000000000002',
      email: 'admin@courseforge.local',
      displayName: 'Запит B',
      role: 'admin',
      emailVerified: true,
      twoFactorEnabled: false,
    },
  ] as const;
  let arrivals = 0;
  let releaseBoth = () => {};
  const bothArrived = new Promise<void>((resolve) => {
    releaseBoth = resolve;
  });
  const originalFetch = globalThis.fetch;
  const identityFetch: typeof fetch = async (input, init) => {
    const requestedUrl = input instanceof Request ? input.url : input.toString();
    if (requestedUrl !== `${apiBaseUrl}/api/v1/auth/me`) return originalFetch(input, init);
    const requestHeaders = init?.headers ?? (input instanceof Request ? input.headers : undefined);
    const cookie = requestHeaders === undefined ? null : new Headers(requestHeaders).get('cookie');
    const user =
      cookie === 'courseforge_session=sessionA'
        ? users[0]
        : cookie === 'courseforge_session=sessionB'
          ? users[1]
          : undefined;
    arrivals += 1;
    if (arrivals === users.length) releaseBoth();
    await bothArrived;
    if (user === undefined) return HttpResponse.json(unavailableEnvelope, { status: 503 });
    return HttpResponse.json(user);
  };
  Object.defineProperty(globalThis, 'fetch', { configurable: true, value: identityFetch });

  let requestA: LoadedRequest;
  let requestB: LoadedRequest;
  try {
    [requestA, requestB] = await Promise.all([
      loadRequest('courseforge_session=sessionA'),
      loadRequest('courseforge_session=sessionB'),
    ]);
  } finally {
    Object.defineProperty(globalThis, 'fetch', { configurable: true, value: originalFetch });
  }

  expect(requestA.identity).toEqual({ status: 'signed-in', user: users[0] });
  expect(requestB.identity).toEqual({ status: 'signed-in', user: users[1] });
  expect(requestA.cachedSession).toEqual(users[0]);
  expect(requestB.cachedSession).toEqual(users[1]);
  expect(requestA.responseText).toContain('Запит A');
  expect(requestA.responseText).not.toContain('Запит B');
  expect(requestB.responseText).toContain('Запит B');
  expect(requestB.responseText).not.toContain('Запит A');
});
