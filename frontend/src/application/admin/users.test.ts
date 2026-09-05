import { describe, expect, test } from '@rstest/core';
import type { AdminUserRepository } from '../../ports/AdminUserRepository';
import { createUserAdministration } from './users';

const userId = '40000000-0000-4000-8000-000000000001';

function repository(overrides: Partial<AdminUserRepository> = {}): AdminUserRepository {
  const base: AdminUserRepository = {
    list: async (query) => ({
      ok: true,
      value: {
        items: [],
        pagination: { limit: query.limit, offset: query.offset, total: 0 },
      },
    }),
    setStatus: async () => ({ ok: true, value: null }),
    delete: async () => ({ ok: true, value: null }),
    resetPassword: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

describe('admin user use cases', () => {
  test('normalizes list input before crossing the repository seam', async () => {
    let received = '';
    const users = createUserAdministration(
      repository({
        list: async (query) => {
          received = query.q;
          return {
            ok: true,
            value: {
              items: [],
              pagination: { limit: query.limit, offset: query.offset, total: 0 },
            },
          };
        },
      }),
    );
    expect(await users.list({ q: '  Ada  ' })).toMatchObject({ ok: true });
    expect(received).toBe('Ada');
  });

  test('refuses malformed mutation input before crossing the repository seam', async () => {
    let calls = 0;
    const users = createUserAdministration(
      repository({
        setStatus: async () => {
          calls += 1;
          return { ok: true, value: null };
        },
        delete: async () => {
          calls += 1;
          return { ok: true, value: null };
        },
        resetPassword: async () => {
          calls += 1;
          return { ok: true, value: null };
        },
      }),
    );
    expect(await users.setStatus('not-a-uuid', 'blocked')).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(await users.delete('not-a-uuid')).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(await users.resetPassword(userId, '')).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
    expect(calls).toBe(0);
  });

  test('returns repository business refusals unchanged', async () => {
    const users = createUserAdministration(
      repository({
        setStatus: async () => ({ ok: false, error: { code: 'last_admin' } }),
      }),
    );
    expect(await users.setStatus(userId, 'blocked')).toEqual({
      ok: false,
      error: { code: 'last_admin' },
    });
  });
});
