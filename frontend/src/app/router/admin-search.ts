import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';
import type { AdminCourseQuery, AdminCourseSort, CourseStatus } from '../../contracts/course';
import type {
  AccountStatus,
  AdminUserQuery,
  AdminUserRole,
  AdminUserSort,
} from '../../modules/admin';

const pageSize = 20;
const maximumPage = Math.floor(10000 / pageSize) + 1;
const statusSchema = z.enum(['draft', 'published', 'archived']);
const sortSchema = z.enum(['recent', 'title', 'students', 'price']);
const userStatusSchema = z.enum(['active', 'blocked', 'deleted']);
const userRoleSchema = z.enum(['student', 'instructor', 'admin']);
const userSortSchema = z.enum(['email', 'name', 'recent']);
const searchTermSchema = z
  .string()
  .refine((value) => [...value].length <= 100 && !value.includes('\u0000'))
  .transform((value) => value.trim());
const pageSchema = z
  .union([
    z.number().int(),
    z
      .string()
      .regex(/^[1-9][0-9]*$/)
      .transform(Number),
  ])
  .pipe(z.number().int().min(1).max(maximumPage));

/** Raw known URL fields. They remain unknown until each field is parsed independently. */
export type AdminSearch = {
  readonly status?: unknown;
  readonly search?: unknown;
  readonly sort?: unknown;
  readonly page?: unknown;
};

export type AdminListState = {
  readonly status?: CourseStatus;
  readonly search: string;
  readonly sort: AdminCourseSort;
  readonly page: number;
};

export type AdminUserSearch = {
  readonly status?: unknown;
  readonly role?: unknown;
  readonly sort?: unknown;
  readonly q?: unknown;
  readonly page?: unknown;
};

export type AdminUserListState = {
  readonly status?: AccountStatus;
  readonly role?: AdminUserRole;
  readonly sort: AdminUserSort;
  readonly q: string;
  readonly page: number;
};

export function parseAdminSearch(search: SearchSchemaInput & Partial<AdminSearch>): AdminSearch {
  return validateAdminSearch(search);
}

/**
 * Echoes known parameters exactly as received. Defaults and refusal do not belong in
 * validateSearch: either would let the router rewrite or reject an otherwise usable address.
 */
export function validateAdminSearch(search: unknown): AdminSearch {
  if (typeof search !== 'object' || search === null) return {};
  return {
    ...(Object.hasOwn(search, 'status') ? { status: Reflect.get(search, 'status') } : {}),
    ...(Object.hasOwn(search, 'search') ? { search: Reflect.get(search, 'search') } : {}),
    ...(Object.hasOwn(search, 'sort') ? { sort: Reflect.get(search, 'sort') } : {}),
    ...(Object.hasOwn(search, 'page') ? { page: Reflect.get(search, 'page') } : {}),
  };
}

export function parseAdminUserSearch(
  search: SearchSchemaInput & Partial<AdminUserSearch>,
): AdminUserSearch {
  return validateAdminUserSearch(search);
}

export function validateAdminUserSearch(search: unknown): AdminUserSearch {
  if (typeof search !== 'object' || search === null) return {};
  return {
    ...(Object.hasOwn(search, 'status') ? { status: Reflect.get(search, 'status') } : {}),
    ...(Object.hasOwn(search, 'role') ? { role: Reflect.get(search, 'role') } : {}),
    ...(Object.hasOwn(search, 'sort') ? { sort: Reflect.get(search, 'sort') } : {}),
    ...(Object.hasOwn(search, 'q') ? { q: Reflect.get(search, 'q') } : {}),
    ...(Object.hasOwn(search, 'page') ? { page: Reflect.get(search, 'page') } : {}),
  };
}

/** Applies defaults once and drops each unusable field without affecting the others. */
export function adminListState(search: AdminSearch): AdminListState {
  const status = statusSchema.safeParse(search.status);
  const term = searchTermSchema.safeParse(search.search ?? '');
  const sort = sortSchema.safeParse(search.sort);
  const page = pageSchema.safeParse(search.page);
  return {
    ...(status.success ? { status: status.data } : {}),
    search: term.success ? term.data : '',
    sort: sort.success ? sort.data : 'recent',
    page: page.success ? page.data : 1,
  };
}

export function adminCourseQuery(state: AdminListState): AdminCourseQuery {
  return {
    limit: pageSize,
    offset: (state.page - 1) * pageSize,
    search: state.search,
    sort: state.sort,
    ...(state.status === undefined ? {} : { status: state.status }),
  };
}

export function adminUserListState(search: AdminUserSearch): AdminUserListState {
  const status = userStatusSchema.safeParse(search.status);
  const role = userRoleSchema.safeParse(search.role);
  const sort = userSortSchema.safeParse(search.sort);
  const q = searchTermSchema.safeParse(search.q ?? '');
  const page = pageSchema.safeParse(search.page);
  return {
    ...(status.success ? { status: status.data } : {}),
    ...(role.success ? { role: role.data } : {}),
    // Default sort is `name`: the account column header (the person's name) is the one
    // the UI offers and shows active on load. `email` stays a valid hand-typed value and
    // the backend's own default, so adminUserQuery still omits it to mean "server default".
    sort: sort.success ? sort.data : 'name',
    q: q.success ? q.data : '',
    page: page.success ? page.data : 1,
  };
}

export function adminUserQuery(state: AdminUserListState): AdminUserQuery {
  return {
    q: state.q,
    limit: pageSize,
    offset: (state.page - 1) * pageSize,
    ...(state.status === undefined ? {} : { status: state.status }),
    ...(state.role === undefined ? {} : { role: state.role }),
    ...(state.sort === 'email' ? {} : { sort: state.sort }),
  };
}
