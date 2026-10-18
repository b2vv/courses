import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse, delay } from 'msw';
import { setupServer } from 'msw/node';
import { createApiSessionRepository } from './session-repository';
import { apiOrigin } from '../../../adapters/api/api-origin';

const signedIn = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Студентка',
  emailVerified: false,
  twoFactorEnabled: false,
  role: 'student',
};
const credentials = { email: 'student@courseforge.local', password: 'secret' };
const envelope = (code: string) => ({
  error: { code, message: 'x', requestId: 'a'.repeat(32) },
});
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('maps a successful login to the domain user', async () => {
  server.use(http.post(`${apiOrigin()}/api/v1/auth/login`, () => HttpResponse.json(signedIn)));
  const result = await createApiSessionRepository().logIn(credentials);
  expect(result).toEqual({ ok: true, value: signedIn });
});

test('sends rememberMe as part of the login command', async () => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/auth/login`, async ({ request }) => {
      expect(await request.json()).toEqual({ ...credentials, rememberMe: true });
      return HttpResponse.json(signedIn);
    }),
  );
  expect(await createApiSessionRepository().logIn({ ...credentials, rememberMe: true })).toEqual({
    ok: true,
    value: signedIn,
  });
});

test('registers a user and retains the unverified state from the API', async () => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/auth/register`, () =>
      HttpResponse.json(signedIn, { status: 201 }),
    ),
  );
  expect(
    await createApiSessionRepository().register({
      email: signedIn.email,
      password: 'password1',
      displayName: signedIn.displayName,
    }),
  ).toEqual({ ok: true, value: signedIn });
});

test.each([
  ['/api/v1/auth/register', 409, 'email_taken', 'email_taken'],
  ['/api/v1/auth/password/reset', 400, 'invalid_account_token', 'invalid_account_token'],
  ['/api/v1/auth/password/reset', 410, 'token_used', 'token_used'],
  ['/api/v1/auth/password/reset', 422, 'weak_password', 'weak_password'],
  ['/api/v1/auth/password/forgot', 429, 'rate_limited', 'rate_limited'],
])('maps self-service business error %s %i', async (path, status, code, expected) => {
  server.use(
    http.post(`${apiOrigin()}${path}`, () => HttpResponse.json(envelope(code), { status })),
  );
  const repository = createApiSessionRepository();
  const result =
    path === '/api/v1/auth/register'
      ? await repository.register({
          email: signedIn.email,
          password: 'password1',
          displayName: 'Name',
        })
      : path === '/api/v1/auth/password/forgot'
        ? await repository.forgotPassword({ email: signedIn.email })
        : await repository.resetPassword({ token: 'token', password: 'password1' });
  expect(result).toEqual({ ok: false, error: { code: expected } });
});

test('keeps password-reset requests neutral and validates empty success bodies', async () => {
  server.use(
    http.post(
      `${apiOrigin()}/api/v1/auth/password/forgot`,
      () => new HttpResponse(null, { status: 200 }),
    ),
    http.post(`${apiOrigin()}/api/v1/auth/verify-email`, () =>
      HttpResponse.json({ unexpected: true }, { status: 204 }),
    ),
  );
  const repository = createApiSessionRepository();
  expect(await repository.forgotPassword({ email: 'missing@courseforge.local' })).toEqual({
    ok: true,
    value: null,
  });
  // A 204 cannot have a JSON response body in Fetch; Orval exposes it as undefined.
  expect(await repository.verifyEmail({ token: 'token' })).toEqual({ ok: true, value: null });
});

test.each([
  [401, 'invalid_credentials', 'invalid_credentials'],
  [400, 'invalid_request', 'invalid_request'],
  [403, 'origin_forbidden', 'origin_forbidden'],
  [503, 'dependency_unavailable', 'dependency_unavailable'],
  [500, 'internal_error', 'dependency_unavailable'],
])('maps login status %i to %s', async (status, code, expected) => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/auth/login`, () =>
      HttpResponse.json(envelope(code), { status }),
    ),
  );
  const result = await createApiSessionRepository().logIn(credentials);
  expect(result.ok).toBe(false);
  if (!result.ok) expect(result.error.code).toBe(expected);
});

test('refuses a login response whose body is not an authenticated user', async () => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/auth/login`, () =>
      HttpResponse.json({ ...signedIn, role: 'owner' }),
    ),
  );
  const result = await createApiSessionRepository().logIn(credentials);
  expect(result.ok).toBe(false);
  if (!result.ok) expect(result.error.code).toBe('invalid_response');
});

test('refuses an authenticated user response without emailVerified', async () => {
  const { emailVerified: _emailVerified, ...withoutVerification } = signedIn;
  server.use(
    http.get(`${apiOrigin()}/api/v1/auth/me`, () => HttpResponse.json(withoutVerification)),
  );
  const result = await createApiSessionRepository().currentUser();
  expect(result).toEqual({ ok: false, error: { code: 'invalid_response' } });
});

test('reports an anonymous viewer as success carrying null, not as an error', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/auth/me`, () =>
      HttpResponse.json(envelope('authentication_required'), { status: 401 }),
    ),
  );
  expect(await createApiSessionRepository().currentUser()).toEqual({ ok: true, value: null });
});

test('keeps a dependency failure on the session read distinct from being anonymous', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/auth/me`, () =>
      HttpResponse.json(envelope('dependency_unavailable'), { status: 503 }),
    ),
  );
  const result = await createApiSessionRepository().currentUser();
  expect(result.ok).toBe(false);
  if (!result.ok) expect(result.error.code).toBe('dependency_unavailable');
});

test('accepts the empty 204 body that logout returns', async () => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/auth/logout`, () => new HttpResponse(null, { status: 204 })),
  );
  expect(await createApiSessionRepository().logOut()).toEqual({ ok: true, value: null });
});

test('separates a timeout from a caller cancellation', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/auth/me`, async () => {
      await delay(50);
      return HttpResponse.json(signedIn);
    }),
  );
  const timedOut = await createApiSessionRepository({ timeoutMs: 5 }).currentUser();
  expect(timedOut.ok).toBe(false);
  if (!timedOut.ok) expect(timedOut.error.code).toBe('request_timeout');

  const controller = new AbortController();
  const pending = createApiSessionRepository().currentUser(controller.signal);
  controller.abort();
  const cancelled = await pending;
  expect(cancelled.ok).toBe(false);
  if (!cancelled.ok) expect(cancelled.error.code).toBe('request_cancelled');
});

test('reports an unreachable API as a dependency failure', async () => {
  server.use(http.post(`${apiOrigin()}/api/v1/auth/login`, () => HttpResponse.error()));
  const result = await createApiSessionRepository().logIn(credentials);
  expect(result.ok).toBe(false);
  if (!result.ok) expect(result.error.code).toBe('dependency_unavailable');
});
