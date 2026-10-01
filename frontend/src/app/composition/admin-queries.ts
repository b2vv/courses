import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { createAuthoring } from '../../application/admin/authoring';
import type {
  AdminCourseDraft,
  AdminCourseQuery,
  AuthoringError,
  AuthoringResult,
  CourseId,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
  CoursePrice,
} from '../../domain/course/course';

export const adminListStaleMs = 30_000;
export const instructorListStaleMs = 300_000;
export const adminCourseListKey = ['admin', 'courses', 'list'] as const;
export const instructorListKey = ['admin', 'instructors'] as const;

export class AuthoringQueryError extends Error {
  readonly code: AuthoringError['code'];
  constructor(error: AuthoringError) {
    super(error.code);
    this.name = 'AuthoringQueryError';
    this.code = error.code;
  }
}

function unwrap<T>(result: AuthoringResult<T>): T {
  if (!result.ok) throw new AuthoringQueryError(result.error);
  return result.value;
}

export function adminCourseListQueryKey(query: AdminCourseQuery) {
  return [...adminCourseListKey, query.status ?? 'all', query.limit, query.offset] as const;
}

export function adminCourseTreeQueryKey(id: CourseId) {
  return ['admin', 'courses', 'tree', id] as const;
}

export async function invalidateAfterCourseWrite(client: QueryClient, id: CourseId) {
  await Promise.all([
    client.invalidateQueries({ queryKey: adminCourseListKey }),
    client.invalidateQueries({ queryKey: adminCourseTreeQueryKey(id), exact: true }),
  ]);
}

export async function invalidateAfterCourseDelete(client: QueryClient, id: CourseId) {
  client.removeQueries({ queryKey: adminCourseTreeQueryKey(id), exact: true });
  await client.invalidateQueries({ queryKey: adminCourseListKey });
}

export type SaveCourseVariables = {
  readonly current: CourseTree;
  readonly edit: CourseTreeEdit;
};
export type SetCoursePriceVariables = {
  readonly current: CourseTree;
  readonly price: CoursePrice | null;
};
export type TransitionCourseVariables = {
  readonly current: CourseTree;
  readonly status: CourseStatus;
};

export function createAdminQueries(authoring: ReturnType<typeof createAuthoring>) {
  const refreshConflict = (
    client: QueryClient,
    error: Error,
    courseId: CourseId,
  ): Promise<void> | undefined =>
    error instanceof AuthoringQueryError && error.code === 'course_modified'
      ? invalidateAfterCourseWrite(client, courseId)
      : undefined;

  return {
    authoring,
    list: (query: AdminCourseQuery) =>
      queryOptions({
        queryKey: adminCourseListQueryKey(query),
        queryFn: async ({ signal }) => unwrap(await authoring.list(query, signal)),
        staleTime: adminListStaleMs,
      }),
    detail: (id: CourseId) =>
      queryOptions({
        queryKey: adminCourseTreeQueryKey(id),
        queryFn: async ({ signal }) => unwrap(await authoring.load(id, signal)),
      }),
    instructors: () =>
      queryOptions({
        queryKey: instructorListKey,
        queryFn: async ({ signal }) => unwrap(await authoring.listInstructors(signal)),
        staleTime: instructorListStaleMs,
      }),
    create: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async (draft: AdminCourseDraft) => unwrap(await authoring.create(draft)),
        onSuccess: (course) => invalidateAfterCourseWrite(client, course.id),
      }),
    save: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ current, edit }: SaveCourseVariables) =>
          unwrap(await authoring.save(current, edit)),
        onSuccess: (course) => invalidateAfterCourseWrite(client, course.id),
        onError: (error, { current }) => refreshConflict(client, error, current.id),
      }),
    setPrice: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ current, price }: SetCoursePriceVariables) =>
          unwrap(await authoring.setPrice(current, price)),
        onSuccess: (course) => invalidateAfterCourseWrite(client, course.id),
        onError: (error, { current }) => refreshConflict(client, error, current.id),
      }),
    transition: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ current, status }: TransitionCourseVariables) =>
          unwrap(await authoring.transition(current, status)),
        onSuccess: (course) => invalidateAfterCourseWrite(client, course.id),
        onError: (error, { current }) => refreshConflict(client, error, current.id),
      }),
    delete: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async (course: CourseTree) => unwrap(await authoring.delete(course)),
        onSuccess: (_value, course) => invalidateAfterCourseDelete(client, course.id),
        onError: (error, course) => refreshConflict(client, error, course.id),
      }),
    unwrap,
  };
}
