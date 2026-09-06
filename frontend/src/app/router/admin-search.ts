import type { SearchSchemaInput } from '@tanstack/react-router';
import { z } from 'zod';
import type { AdminCourseQuery, CourseStatus } from '../../domain/course/course';
import type { AccountStatus, AdminUserQuery } from '../../domain/identity/administration';

const pageSize = 20;
const maximumPage = Math.floor(10000 / pageSize) + 1;
const statusSchema = z.enum(['draft', 'published', 'archived']);
const userStatusSchema = z.enum(['active', 'blocked', 'deleted']);
const userQuerySchema = z
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
  readonly page?: unknown;
};

export type AdminListState = {
  readonly status?: CourseStatus;
  readonly page: number;
};

export type AdminUserSearch = {
  readonly status?: unknown;
  readonly q?: unknown;
  readonly page?: unknown;
};

export type AdminUserListState = {
  readonly status?: AccountStatus;
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
    ...(Object.hasOwn(search, 'q') ? { q: Reflect.get(search, 'q') } : {}),
    ...(Object.hasOwn(search, 'page') ? { page: Reflect.get(search, 'page') } : {}),
  };
}

/** Applies defaults once and drops each unusable field without affecting the others. */
export function adminListState(search: AdminSearch): AdminListState {
  const status = statusSchema.safeParse(search.status);
  const page = pageSchema.safeParse(search.page);
  return {
    ...(status.success ? { status: status.data } : {}),
    page: page.success ? page.data : 1,
  };
}

export function adminCourseQuery(state: AdminListState): AdminCourseQuery {
  return {
    limit: pageSize,
    offset: (state.page - 1) * pageSize,
    ...(state.status === undefined ? {} : { status: state.status }),
  };
}

export function adminUserListState(search: AdminUserSearch): AdminUserListState {
  const status = userStatusSchema.safeParse(search.status);
  const q = userQuerySchema.safeParse(search.q ?? '');
  const page = pageSchema.safeParse(search.page);
  return {
    ...(status.success ? { status: status.data } : {}),
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
  };
}
