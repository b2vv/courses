import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { createApiEnrollmentRepository } from './enrollment-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const userId = '40000000-0000-4000-8000-000000000001';
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const refusal = (code: string, status: number) =>
  HttpResponse.json({ error: { code, message: 'x', requestId: 'a'.repeat(32) } }, { status });

test.each([
  [403, 'enrollment_forbidden'],
  [403, 'email_not_verified'],
  [403, 'admin_not_learner'],
  [409, 'course_not_enrollable'],
] as const)('names the enrol refusal %i %s', async (status, code) => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/learn/courses/${courseId}/enroll`, () =>
      refusal(code, status),
    ),
  );
  expect(await createApiEnrollmentRepository().enroll(courseId)).toEqual({
    ok: false,
    error: { code },
  });
});

// Blocking or unblocking someone who is not a member is an invalid transition, not a fault.
test.each(['block', 'unblock'] as const)('names grant_transition_invalid on %s', async (action) => {
  server.use(
    http.post(
      `${apiOrigin()}/api/v1/teach/courses/${courseId}/participants/${userId}/${action}`,
      () => refusal('grant_transition_invalid', 409),
    ),
  );
  expect(await createApiEnrollmentRepository()[action](courseId, userId)).toEqual({
    ok: false,
    error: { code: 'grant_transition_invalid' },
  });
});
