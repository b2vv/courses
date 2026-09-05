import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { createAuthoring } from '../../application/admin/authoring';
import { asBrandedId } from '../../domain/course/course';
import type {
  AdminCourseDraft,
  AuthoringErrorCode,
  CourseTree,
  CourseTreeEdit,
} from '../../domain/course/course';
import { apiOrigin } from './api-origin';
import { createApiAdminCourseRepository } from './admin-course-repository';

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
  modules: [],
};
const draft: AdminCourseDraft = {
  slug: tree.slug,
  title: tree.title,
  summary: tree.summary,
  description: tree.description,
  instructorId: tree.instructorId,
};
const edit: CourseTreeEdit = { ...draft, instructorId: tree.instructorId, modules: [] };
const envelope = (code: string) => ({
  error: { code, message: 'x', requestId: 'a'.repeat(32) },
});
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

type Operation = 'list' | 'load' | 'create' | 'save' | 'transition' | 'delete';

function invokeFailure(operation: Operation, status: number, code: string) {
  const response = () => HttpResponse.json(envelope(code), { status });
  const root = `${apiOrigin()}/api/v1/admin/courses`;
  const authoring = createAuthoring(createApiAdminCourseRepository());
  switch (operation) {
    case 'list':
      server.use(http.get(root, response));
      return authoring.list({});
    case 'load':
      server.use(http.get(`${root}/${tree.id}`, response));
      return authoring.load(tree.id);
    case 'create':
      server.use(http.post(root, response));
      return authoring.create(draft);
    case 'save':
      server.use(http.put(`${root}/${tree.id}`, response));
      return authoring.save(tree, edit);
    case 'transition':
      server.use(http.post(`${root}/${tree.id}/status`, response));
      return authoring.transition(tree, 'published');
    case 'delete':
      server.use(http.delete(`${root}/${tree.id}`, response));
      return authoring.delete(tree);
  }
}

const failures: readonly {
  operation: Operation;
  status: number;
  code: string;
  expected: AuthoringErrorCode;
}[] = [
  { operation: 'create', status: 400, code: 'invalid_request', expected: 'invalid_request' },
  { operation: 'create', status: 400, code: 'invalid_owner', expected: 'invalid_owner' },
  {
    operation: 'list',
    status: 401,
    code: 'authentication_required',
    expected: 'authentication_required',
  },
  {
    operation: 'create',
    status: 403,
    code: 'origin_forbidden',
    expected: 'origin_forbidden',
  },
  { operation: 'list', status: 403, code: 'forbidden', expected: 'forbidden' },
  {
    operation: 'load',
    status: 403,
    code: 'course_forbidden',
    expected: 'course_forbidden',
  },
  {
    operation: 'load',
    status: 404,
    code: 'course_not_found',
    expected: 'course_not_found',
  },
  { operation: 'create', status: 409, code: 'slug_taken', expected: 'slug_taken' },
  { operation: 'save', status: 409, code: 'slug_immutable', expected: 'slug_immutable' },
  {
    operation: 'transition',
    status: 409,
    code: 'invalid_status_transition',
    expected: 'invalid_status_transition',
  },
  {
    operation: 'delete',
    status: 409,
    code: 'course_not_deletable',
    expected: 'course_not_deletable',
  },
  { operation: 'save', status: 409, code: 'course_modified', expected: 'course_modified' },
  {
    operation: 'list',
    status: 503,
    code: 'dependency_unavailable',
    expected: 'dependency_unavailable',
  },
  {
    operation: 'list',
    status: 500,
    code: 'internal_error',
    expected: 'dependency_unavailable',
  },
];

describe('admin API repository', () => {
  test.each(failures)('maps $status $code from $operation to $expected', async (failure) => {
    const result = await invokeFailure(failure.operation, failure.status, failure.code);
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe(failure.expected);
  });

  test('maps a generated list response explicitly into branded domain courses', async () => {
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/courses`, ({ request }) => {
        const url = new URL(request.url);
        expect(url.searchParams.get('limit')).toBe('20');
        expect(url.searchParams.get('offset')).toBe('0');
        return HttpResponse.json({
          items: [
            {
              id: tree.id,
              slug: tree.slug,
              title: tree.title,
              status: tree.status,
              instructorId: tree.instructorId,
              updatedAt: tree.updatedAt,
            },
          ],
          pagination: { limit: 20, offset: 0, total: 1 },
        });
      }),
    );
    const result = await createAuthoring(createApiAdminCourseRepository()).list({});
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.value.items[0]?.id).toBe(tree.id);
  });

  test('refuses a malformed successful tree instead of leaking wire data', async () => {
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/courses/${tree.id}`, () =>
        HttpResponse.json({ ...tree, modules: [{ id: 'not-a-uuid', title: 'x', lessons: [] }] }),
      ),
    );
    expect(await createAuthoring(createApiAdminCourseRepository()).load(tree.id)).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  });
});
