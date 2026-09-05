import { describe, expect, test } from '@rstest/core';
import { MutationObserver, QueryClient } from '@tanstack/react-query';
import { createUserAdministration } from '../../application/admin/users';
import type { AdminUserQuery } from '../../domain/identity/administration';
import type { AdminUserRepository } from '../../ports/AdminUserRepository';
import { adminListStaleMs } from './admin-queries';
import {
  AdminUserQueryError,
  adminUserListQueryKey,
  createAdminUserQueries,
  invalidateAfterUserWrite,
} from './admin-user-queries';

const userId = '40000000-0000-4000-8000-000000000001';
const query: AdminUserQuery = { q: 'student', limit: 20, offset: 0, status: 'active' };

function repository(overrides: Partial<AdminUserRepository> = {}): AdminUserRepository {
  const base: AdminUserRepository = {
    list: async (input) => ({
      ok: true,
      value: {
        items: [],
        pagination: { limit: input.limit, offset: input.offset, total: 0 },
      },
    }),
    setStatus: async () => ({ ok: true, value: null }),
    delete: async () => ({ ok: true, value: null }),
    resetPassword: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

function clientForTest() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

describe('admin user query boundary', () => {
  test('uses the shared list staleness and a complete stable query key', () => {
    const users = createAdminUserQueries(createUserAdministration(repository()));
    expect(users.list(query).queryKey).toEqual([
      'admin',
      'users',
      'list',
      'active',
      'student',
      20,
      0,
    ]);
    expect(users.list(query).staleTime).toBe(adminListStaleMs);
  });

  test('surfaces a typed refusal without caching it as list data', async () => {
    const users = createAdminUserQueries(
      createUserAdministration(
        repository({
          list: async () => ({ ok: false, error: { code: 'forbidden' } }),
        }),
      ),
    );
    const client = clientForTest();
    await expect(client.fetchQuery(users.list(query))).rejects.toThrow(AdminUserQueryError);
    expect(client.getQueryData(adminUserListQueryKey(query))).toBeUndefined();
  });

  test('invalidates every cached directory filter through the named helper', async () => {
    const client = clientForTest();
    const activeKey = adminUserListQueryKey(query);
    const deletedKey = adminUserListQueryKey({ ...query, q: '', status: 'deleted' });
    client.setQueryData(activeKey, { items: [], pagination: { limit: 20, offset: 0, total: 0 } });
    client.setQueryData(deletedKey, { items: [], pagination: { limit: 20, offset: 0, total: 0 } });
    await invalidateAfterUserWrite(client);
    expect(client.getQueryState(activeKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(deletedKey)?.isInvalidated).toBe(true);
  });

  test('all three mutations invalidate the directory through their option shape', async () => {
    const users = createAdminUserQueries(createUserAdministration(repository()));
    const client = clientForTest();
    const key = adminUserListQueryKey(query);
    const seed = () => {
      client.setQueryData(key, { items: [], pagination: { limit: 20, offset: 0, total: 0 } });
    };

    seed();
    await new MutationObserver(client, users.setStatus(client)).mutate({
      id: userId,
      status: 'blocked',
    });
    expect(client.getQueryState(key)?.isInvalidated).toBe(true);

    seed();
    await new MutationObserver(client, users.delete(client)).mutate(userId);
    expect(client.getQueryState(key)?.isInvalidated).toBe(true);

    seed();
    await new MutationObserver(client, users.resetPassword(client)).mutate({
      id: userId,
      password: 'new password',
    });
    expect(client.getQueryState(key)?.isInvalidated).toBe(true);
  });
});
