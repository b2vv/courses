export type AccountStatus = 'active' | 'blocked' | 'deleted';
export type AdminUserRole = 'student' | 'instructor' | 'admin';
export type AdminUserSort = 'email' | 'name' | 'recent';
export type WritableAccountStatus = Exclude<AccountStatus, 'deleted'>;
export type AdministrationErrorCode =
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'forbidden'
  | 'cannot_target_self'
  | 'admin_immutable'
  | 'last_admin'
  | 'user_not_found'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type AdministrationError = { readonly code: AdministrationErrorCode };
export type AdministrationResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: AdministrationError };
export type AdminUser = {
  readonly id: string;
  readonly email: string;
  readonly displayName: string;
  readonly role: AdminUserRole;
  readonly status: AccountStatus;
  readonly updatedAt: string;
};
export type AdminUserQuery = {
  readonly q: string;
  readonly limit: number;
  readonly offset: number;
  readonly status?: AccountStatus;
  readonly role?: AdminUserRole;
  readonly sort?: AdminUserSort;
};
export type AdminUserPage = {
  readonly items: readonly AdminUser[];
  readonly pagination: { readonly limit: number; readonly offset: number; readonly total: number };
};

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const accountStatuses: readonly AccountStatus[] = ['active', 'blocked', 'deleted'];
const adminUserRoles: readonly AdminUserRole[] = ['student', 'instructor', 'admin'];
const adminUserSorts: readonly AdminUserSort[] = ['email', 'name', 'recent'];

export function validAdminUserId(value: string): boolean {
  return uuidPattern.test(value) && value !== '00000000-0000-0000-0000-000000000000';
}

export function validAdministrationPassword(password: string): boolean {
  const bytes = new TextEncoder().encode(password).byteLength;
  return bytes > 0 && bytes <= 1024;
}

export function makeAdminUserQuery(
  input: Partial<AdminUserQuery>,
): AdministrationResult<AdminUserQuery> {
  const query = {
    q: (input.q ?? '').trim(),
    limit: input.limit ?? 20,
    offset: input.offset ?? 0,
    status: input.status,
    role: input.role,
    sort: input.sort,
  };
  if (
    !Number.isInteger(query.limit) ||
    query.limit < 1 ||
    query.limit > 50 ||
    !Number.isInteger(query.offset) ||
    query.offset < 0 ||
    query.offset > 10000 ||
    [...query.q].length > 100 ||
    query.q.includes('\u0000') ||
    (query.status !== undefined && !accountStatuses.includes(query.status)) ||
    (query.role !== undefined && !adminUserRoles.includes(query.role)) ||
    (query.sort !== undefined && !adminUserSorts.includes(query.sort))
  )
    return { ok: false, error: { code: 'invalid_request' } };
  return {
    ok: true,
    value: {
      q: query.q,
      limit: query.limit,
      offset: query.offset,
      ...(query.status === undefined ? {} : { status: query.status }),
      ...(query.role === undefined ? {} : { role: query.role }),
      ...(query.sort === undefined ? {} : { sort: query.sort }),
    },
  };
}
