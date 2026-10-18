import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { createUserAdministration } from '../../application/admin/users';
import type {
  AdministrationError,
  AdministrationResult,
  AdminUserQuery,
  WritableAccountStatus,
} from '../../domain/admin/administration';
import { adminListStaleMs } from './admin-queries';

export const adminUserListKey = ['admin', 'users', 'list'] as const;

export class AdminUserQueryError extends Error {
  readonly code: AdministrationError['code'];
  constructor(error: AdministrationError) {
    super(error.code);
    this.name = 'AdminUserQueryError';
    this.code = error.code;
  }
}

function unwrap<T>(result: AdministrationResult<T>): T {
  if (!result.ok) throw new AdminUserQueryError(result.error);
  return result.value;
}

export function adminUserListQueryKey(query: AdminUserQuery) {
  return [...adminUserListKey, query.status ?? 'all', query.q, query.limit, query.offset] as const;
}

export async function invalidateAfterUserWrite(client: QueryClient) {
  await client.invalidateQueries({ queryKey: adminUserListKey });
}

export type SetUserStatusVariables = {
  readonly id: string;
  readonly status: WritableAccountStatus;
};
export type ResetUserPasswordVariables = { readonly id: string; readonly password: string };

export function createAdminUserQueries(
  administration: ReturnType<typeof createUserAdministration>,
) {
  return {
    administration,
    list: (query: AdminUserQuery) =>
      queryOptions({
        queryKey: adminUserListQueryKey(query),
        queryFn: async ({ signal }) => unwrap(await administration.list(query, signal)),
        staleTime: adminListStaleMs,
      }),
    setStatus: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ id, status }: SetUserStatusVariables) =>
          unwrap(await administration.setStatus(id, status)),
        onSuccess: () => invalidateAfterUserWrite(client),
      }),
    delete: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async (id: string) => unwrap(await administration.delete(id)),
        onSuccess: () => invalidateAfterUserWrite(client),
      }),
    resetPassword: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ id, password }: ResetUserPasswordVariables) =>
          unwrap(await administration.resetPassword(id, password)),
        onSuccess: () => invalidateAfterUserWrite(client),
      }),
    unwrap,
  };
}
