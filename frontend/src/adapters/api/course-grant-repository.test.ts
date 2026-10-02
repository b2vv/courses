import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiCourseGrantRepository } from './course-grant-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const userId = '40000000-0000-4000-8000-000000000001';
const adminId = '40000000-0000-4000-8000-000000000003';
const url = `${apiOrigin()}/api/v1/admin/courses/${courseId}/grants`;
const grant = {
  id: '70000000-0000-4000-8000-000000000001',
  userId,
  courseId,
  source: 'admin',
  state: 'active',
  grantedAt: '2026-10-01T10:00:00Z',
  grantedBy: adminId,
};
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('Pay10 sends the user and the reason, and maps the active grant', async () => {
  let body: unknown;
  server.use(
    http.post(url, async ({ request }) => {
      body = await request.json();
      return HttpResponse.json(grant, { status: 201 });
    }),
  );
  expect(await createApiCourseGrantRepository().grant(courseId, userId, 'Scholarship')).toEqual({
    ok: true,
    value: { id: grant.id, userId, courseId, grantedAt: grant.grantedAt },
  });
  expect(body).toEqual({ userId, reason: 'Scholarship' });
});

test('names a duplicate and a refusal, and rejects a grant for someone else', async () => {
  const conflict = (code: string, status: number) =>
    HttpResponse.json({ error: { code, message: 'x', requestId: 'a'.repeat(32) } }, { status });
  const repository = createApiCourseGrantRepository();

  server.use(http.post(url, () => conflict('course_already_owned', 409)));
  expect(await repository.grant(courseId, userId, 'r')).toEqual({
    ok: false,
    error: { code: 'course_already_owned' },
  });
  server.use(http.post(url, () => conflict('grant_forbidden', 403)));
  expect(await repository.grant(courseId, userId, 'r')).toEqual({
    ok: false,
    error: { code: 'grant_forbidden' },
  });
  server.use(
    http.post(url, () => HttpResponse.json({ ...grant, userId: adminId }, { status: 201 })),
  );
  expect(await repository.grant(courseId, userId, 'r')).toEqual({
    ok: false,
    error: { code: 'invalid_response' },
  });
});
