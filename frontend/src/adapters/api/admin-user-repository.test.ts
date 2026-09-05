import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { createUserAdministration } from '../../application/admin/users';
import type { AdministrationErrorCode } from '../../domain/identity/administration';
import { apiOrigin } from './api-origin';
import { createApiAdminUserRepository } from './admin-user-repository';

const user = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Student',
  role: 'student' as const,
  status: 'active' as const,
  updatedAt: '2026-09-01T00:00:00.000001Z',
};
const envelope = (code: string) => ({
  error: { code, message: 'x', requestId: 'a'.repeat(32) },
});
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

type Operation = 'list' | 'status' | 'delete' | 'password';

function invokeFailure(operation: Operation, status: number, code: string) {
  const response = () => HttpResponse.json(envelope(code), { status });
  const root = `${apiOrigin()}/api/v1/admin/users`;
  const users = createUserAdministration(createApiAdminUserRepository());
  switch (operation) {
    case 'list':
      server.use(http.get(root, response));
      return users.list({});
    case 'status':
      server.use(http.post(`${root}/${user.id}/status`, response));
      return users.setStatus(user.id, 'blocked');
    case 'delete':
      server.use(http.delete(`${root}/${user.id}`, response));
      return users.delete(user.id);
    case 'password':
      server.use(http.post(`${root}/${user.id}/password`, response));
      return users.resetPassword(user.id, 'new password');
  }
}

const failures: readonly {
  operation: Operation;
  status: number;
  code: string;
  expected: AdministrationErrorCode;
}[] = [
  { operation: 'list', status: 400, code: 'invalid_request', expected: 'invalid_request' },
  {
    operation: 'list',
    status: 401,
    code: 'authentication_required',
    expected: 'authentication_required',
  },
  {
    operation: 'status',
    status: 403,
    code: 'origin_forbidden',
    expected: 'origin_forbidden',
  },
  { operation: 'list', status: 403, code: 'forbidden', expected: 'forbidden' },
  {
    operation: 'status',
    status: 404,
    code: 'user_not_found',
    expected: 'user_not_found',
  },
  {
    operation: 'status',
    status: 409,
    code: 'cannot_target_self',
    expected: 'cannot_target_self',
  },
  {
    operation: 'delete',
    status: 409,
    code: 'admin_immutable',
    expected: 'admin_immutable',
  },
  { operation: 'status', status: 409, code: 'last_admin', expected: 'last_admin' },
  {
    operation: 'list',
    status: 503,
    code: 'dependency_unavailable',
    expected: 'dependency_unavailable',
  },
  {
    operation: 'list',
    status: 500,
    code: 'internal_error',
    expected: 'dependency_unavailable',
  },
];

describe('admin user API repository', () => {
  test.each(failures)('maps $status $code from $operation to $expected', async (failure) => {
    const result = await invokeFailure(failure.operation, failure.status, failure.code);
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe(failure.expected);
  });

  test('maps a generated list response explicitly into domain users', async () => {
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/users`, ({ request }) => {
        const url = new URL(request.url);
        expect(url.searchParams.get('limit')).toBe('20');
        expect(url.searchParams.get('offset')).toBe('0');
        expect(url.searchParams.get('q')).toBe('Student');
        expect(url.searchParams.get('status')).toBe('active');
        return HttpResponse.json({
          items: [user],
          pagination: { limit: 20, offset: 0, total: 1 },
        });
      }),
    );
    const result = await createUserAdministration(createApiAdminUserRepository()).list({
      q: ' Student ',
      status: 'active',
    });
    expect(result).toEqual({
      ok: true,
      value: { items: [user], pagination: { limit: 20, offset: 0, total: 1 } },
    });
  });

  test('refuses malformed successful data instead of leaking generated models', async () => {
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/users`, () =>
        HttpResponse.json({
          items: [{ ...user, status: 'disabled' }],
          pagination: { limit: 20, offset: 0, total: 1 },
        }),
      ),
    );
    expect(await createUserAdministration(createApiAdminUserRepository()).list({})).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  });

  test('serializes each mutation through the generated client', async () => {
    const root = `${apiOrigin()}/api/v1/admin/users/${user.id}`;
    server.use(
      http.post(`${root}/status`, async ({ request }) => {
        expect(await request.json()).toEqual({ status: 'blocked' });
        return new HttpResponse(null, { status: 200 });
      }),
      http.delete(root, () => new HttpResponse(null, { status: 204 })),
      http.post(`${root}/password`, async ({ request }) => {
        expect(await request.json()).toEqual({ password: 'new password' });
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const users = createUserAdministration(createApiAdminUserRepository());
    expect(await users.setStatus(user.id, 'blocked')).toEqual({ ok: true, value: null });
    expect(await users.delete(user.id)).toEqual({ ok: true, value: null });
    expect(await users.resetPassword(user.id, 'new password')).toEqual({
      ok: true,
      value: null,
    });
  });
});
