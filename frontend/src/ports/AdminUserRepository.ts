import type {
  AdministrationResult,
  AdminUserPage,
  AdminUserQuery,
  WritableAccountStatus,
} from '../domain/identity/administration';

export interface AdminUserRepository {
  list(query: AdminUserQuery, signal?: AbortSignal): Promise<AdministrationResult<AdminUserPage>>;
  setStatus(
    id: string,
    status: WritableAccountStatus,
    signal?: AbortSignal,
  ): Promise<AdministrationResult<null>>;
  delete(id: string, signal?: AbortSignal): Promise<AdministrationResult<null>>;
  resetPassword(
    id: string,
    password: string,
    signal?: AbortSignal,
  ): Promise<AdministrationResult<null>>;
}
