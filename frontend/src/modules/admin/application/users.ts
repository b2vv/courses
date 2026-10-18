import {
  makeAdminUserQuery,
  validAdministrationPassword,
  validAdminUserId,
} from '../domain/administration';
import type {
  AdministrationResult,
  AdminUserQuery,
  WritableAccountStatus,
} from '../domain/administration';
import type { AdminUserRepository } from '../ports/AdminUserRepository';

const invalid = { ok: false, error: { code: 'invalid_request' } } as const;

export function createUserAdministration(repository: AdminUserRepository) {
  return {
    list(input: Partial<AdminUserQuery>, signal?: AbortSignal) {
      const query = makeAdminUserQuery(input);
      return query.ok ? repository.list(query.value, signal) : Promise.resolve(query);
    },
    setStatus(
      id: string,
      status: WritableAccountStatus,
      signal?: AbortSignal,
    ): Promise<AdministrationResult<null>> {
      return validAdminUserId(id) && (status === 'active' || status === 'blocked')
        ? repository.setStatus(id, status, signal)
        : Promise.resolve(invalid);
    },
    delete(id: string, signal?: AbortSignal): Promise<AdministrationResult<null>> {
      return validAdminUserId(id) ? repository.delete(id, signal) : Promise.resolve(invalid);
    },
    resetPassword(
      id: string,
      password: string,
      signal?: AbortSignal,
    ): Promise<AdministrationResult<null>> {
      return validAdminUserId(id) && validAdministrationPassword(password)
        ? repository.resetPassword(id, password, signal)
        : Promise.resolve(invalid);
    },
  };
}
