import { describe, expect, test } from '@rstest/core';
import { MutationObserver, QueryClient, QueryObserver } from '@tanstack/react-query';
import { createAuthoring } from '../../modules/admin';
import { asBrandedId } from '../../contracts/course';
import type { AdminCourseQuery, CourseTree, CourseTreeEdit } from '../../contracts/course';
import type { AdminCourseRepository } from '../../modules/admin';
import {
  adminCourseListQueryKey,
  adminCourseTreeQueryKey,
  adminListStaleMs,
  AuthoringQueryError,
  createAdminQueries,
  instructorListKey,
  instructorListStaleMs,
  invalidateAfterCourseDelete,
  invalidateAfterCourseWrite,
} from './admin-queries';

const tree: CourseTree = {
  id: asBrandedId('10000000-0000-4000-8000-000000000001', 'course'),
  slug: 'go-basics',
  title: 'Go',
  summary: 'Практика',
  description: 'Опис',
  status: 'draft',
  instructorId: '40000000-0000-4000-8000-000000000001',
  createdAt: '2026-09-01T00:00:00.000000Z',
  updatedAt: '2026-09-01T00:00:00.000001Z',
  publishedAt: null,
  price: null,
  studentCount: 0,
  modules: [],
};
const edit: CourseTreeEdit = {
  slug: tree.slug,
  title: tree.title,
  summary: tree.summary,
  description: tree.description,
  instructorId: tree.instructorId,
  modules: [],
};
const query: AdminCourseQuery = {
  limit: 20,
  offset: 0,
  status: 'draft',
  search: '',
  sort: 'recent',
};

function repository(overrides: Partial<AdminCourseRepository> = {}): AdminCourseRepository {
  const base: AdminCourseRepository = {
    list: async (input) => ({
      ok: true,
      value: {
        items: [],
        pagination: { limit: input.limit, offset: input.offset, total: 0 },
        statusCounts: { total: 0, draft: 0, published: 0, archived: 0 },
      },
    }),
    load: async () => ({ ok: true, value: tree }),
    listInstructors: async () => ({ ok: true, value: [] }),
    create: async () => ({ ok: true, value: tree }),
    save: async () => ({ ok: true, value: tree }),
    transition: async () => ({ ok: true, value: tree }),
    delete: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

function clientForTest() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

describe('admin query boundary', () => {
  test('keeps list and reference-data staleness explicit in query option shapes', () => {
    const queries = createAdminQueries(createAuthoring(repository()));
    expect(queries.list(query).queryKey).toEqual([
      'admin',
      'courses',
      'list',
      'draft',
      '',
      'recent',
      20,
      0,
    ]);
    expect(queries.list(query).staleTime).toBe(adminListStaleMs);
    expect(queries.instructors().queryKey).toEqual(instructorListKey);
    expect(queries.instructors().staleTime).toBe(instructorListStaleMs);
    expect(queries.detail(tree.id).queryKey).toEqual(adminCourseTreeQueryKey(tree.id));
  });

  test('surfaces a typed domain refusal without caching it as data', async () => {
    const queries = createAdminQueries(
      createAuthoring(
        repository({
          list: async () => ({ ok: false, error: { code: 'forbidden' } }),
        }),
      ),
    );
    const client = clientForTest();
    await expect(client.fetchQuery(queries.list(query))).rejects.toThrow(AuthoringQueryError);
    const error = await client.fetchQuery(queries.list(query)).catch((cause: unknown) => cause);
    expect(error instanceof AuthoringQueryError && error.code).toBe('forbidden');
    expect(client.getQueryData(adminCourseListQueryKey(query))).toBeUndefined();
  });

  test('invalidates list and tree after writes and removes a deleted tree', async () => {
    const client = clientForTest();
    const listKey = adminCourseListQueryKey(query);
    const treeKey = adminCourseTreeQueryKey(tree.id);
    client.setQueryData(listKey, { items: [], pagination: { limit: 20, offset: 0, total: 0 } });
    client.setQueryData(treeKey, tree);

    await invalidateAfterCourseWrite(client, tree.id);
    expect(client.getQueryState(listKey)?.isInvalidated).toBe(true);
    expect(client.getQueryState(treeKey)?.isInvalidated).toBe(true);

    await invalidateAfterCourseDelete(client, tree.id);
    expect(client.getQueryData(treeKey)).toBeUndefined();
  });

  test('a course_modified mutation refetches through the named write helper', async () => {
    let reads = 0;
    const queries = createAdminQueries(
      createAuthoring(
        repository({
          load: async () => {
            reads += 1;
            return { ok: true, value: tree };
          },
          save: async () => ({ ok: false, error: { code: 'course_modified' } }),
        }),
      ),
    );
    const client = clientForTest();
    const listKey = adminCourseListQueryKey(query);
    const treeKey = adminCourseTreeQueryKey(tree.id);
    client.setQueryData(listKey, { items: [], pagination: { limit: 20, offset: 0, total: 0 } });
    const detail = new QueryObserver(client, queries.detail(tree.id));
    const unsubscribe = detail.subscribe(() => {});
    await client.ensureQueryData(queries.detail(tree.id));
    expect(reads).toBe(1);
    const mutation = new MutationObserver(client, queries.save(client));

    await expect(mutation.mutate({ current: tree, edit })).rejects.toMatchObject({
      code: 'course_modified',
    });
    expect(client.getQueryState(listKey)?.isInvalidated).toBe(true);
    expect(client.getQueryData(treeKey)).toEqual(tree);
    expect(reads).toBe(2);
    unsubscribe();
  });
});
