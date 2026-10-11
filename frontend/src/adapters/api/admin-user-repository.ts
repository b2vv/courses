import type {
  AdministrationError,
  AdministrationResult,
  AdminUserPage,
  AdminUserQuery,
} from '../../domain/identity/administration';
import type { AdminUserRepository } from '../../ports/AdminUserRepository';
import {
  deleteAdminUser,
  listAdminUsers,
  resetAdminUserPassword,
  setAdminUserStatus,
} from './generated/courses';
import { errorSchema } from './course-schema';
import { adminUserPageSchema } from './admin-user-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

const invalid = {
  ok: false,
  error: { code: 'invalid_response' },
} satisfies AdministrationResult<never>;

function wireError(response: WireResponse): AdministrationError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'invalid_request':
        if (response.status === 400) return { code: 'invalid_request' };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'origin_forbidden':
      case 'forbidden':
        if (response.status === 403) return { code: parsed.data.error.code };
        break;
      case 'user_not_found':
        if (response.status === 404) return { code: 'user_not_found' };
        break;
      case 'cannot_target_self':
      case 'admin_immutable':
      case 'last_admin':
        if (response.status === 409) return { code: parsed.data.error.code };
        break;
      case 'dependency_unavailable':
        if (response.status === 503) return { code: 'dependency_unavailable' };
        break;
      case 'internal_error':
        if (response.status === 500) return { code: 'dependency_unavailable' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

function userPage(raw: unknown, query: AdminUserQuery): AdministrationResult<AdminUserPage> {
  const parsed = adminUserPageSchema.safeParse(raw);
  if (
    !parsed.success ||
    parsed.data.pagination.limit !== query.limit ||
    parsed.data.pagination.offset !== query.offset ||
    parsed.data.items.length > query.limit ||
    parsed.data.items.length > Math.max(0, parsed.data.pagination.total - query.offset) ||
    (query.status === undefined && parsed.data.items.some((user) => user.status === 'deleted')) ||
    (query.status !== undefined && parsed.data.items.some((user) => user.status !== query.status))
  )
    return invalid;
  return { ok: true, value: parsed.data };
}

/** One repository per router request; generated transport types never leave this adapter. */
export function createApiAdminUserRepository(options: TransportOptions = {}): AdminUserRepository {
  const send = createApiTransport(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => AdministrationResult<T>,
    caller?: AbortSignal,
  ): Promise<AdministrationResult<T>> {
    const outcome = await send(operation, execute, expected, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === expected) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }

  return {
    list(query, signal) {
      const params = {
        limit: query.limit,
        offset: query.offset,
        ...(query.q === '' ? {} : { q: query.q }),
        ...(query.status === undefined ? {} : { status: query.status }),
        ...(query.role === undefined ? {} : { role: query.role }),
        ...(query.sort === undefined ? {} : { sort: query.sort }),
      };
      return request<AdminUserPage>(
        'admin.users.list',
        (init) => listAdminUsers(params, init),
        200,
        (raw) => userPage(raw, query),
        signal,
      );
    },
    setStatus(id, status, signal) {
      return request<null>(
        'admin.users.status',
        (init) => setAdminUserStatus(id, { status }, init),
        200,
        () => ({ ok: true, value: null }),
        signal,
      );
    },
    delete(id, signal) {
      return request<null>(
        'admin.users.delete',
        (init) => deleteAdminUser(id, init),
        204,
        () => ({ ok: true, value: null }),
        signal,
      );
    },
    resetPassword(id, password, signal) {
      return request<null>(
        'admin.users.password',
        (init) => resetAdminUserPassword(id, { password }, init),
        204,
        () => ({ ok: true, value: null }),
        signal,
      );
    },
  };
}
