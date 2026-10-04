import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiTeachRepository } from './teach-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const reviewId = '75000000-0000-4000-8000-000000000001';
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
const error = (code: string) => ({ error: { code, message: code, requestId: 'a'.repeat(32) } });
test('maps review moderation remove and dismiss outcomes', async () => {
  for (const action of ['removeReview', 'dismissReviewReport'] as const) {
    const path = action === 'removeReview' ? 'remove' : 'dismiss-report';
    // The backend refuses removal by anyone but an admin with review_moderation_forbidden (G9), and a
    // dismissal by anyone but the owner or an admin with course_forbidden (requireModerationAccess).
    const forbidden =
      action === 'removeReview' ? 'review_moderation_forbidden' : 'course_forbidden';
    server.use(
      http.post(
        `${apiOrigin()}/api/v1/teach/reviews/${reviewId}/${path}`,
        () => new HttpResponse(null, { status: 204 }),
      ),
    );
    expect(
      await createApiTeachRepository()[action](reviewId, { reason: 'moderation_reviewed' }),
    ).toEqual({ ok: true, value: undefined });
    server.use(
      http.post(`${apiOrigin()}/api/v1/teach/reviews/${reviewId}/${path}`, () =>
        HttpResponse.json(error(forbidden), { status: 403 }),
      ),
    );
    expect(
      await createApiTeachRepository()[action](reviewId, { reason: 'moderation_reviewed' }),
    ).toEqual({ ok: false, error: { code: forbidden } });
  }
});
test('parses a review item in the moderation queue', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/teach/courses/${courseId}/moderation-queue`, () =>
      HttpResponse.json({
        items: [
          {
            kind: 'review',
            comment: null,
            review: {
              id: reviewId,
              authorDisplayName: 'Олена',
              stars: 5,
              body: 'Детальний відгук про курс.',
              completedLessons: 1,
              totalLessons: 2,
              createdAt: '2026-10-01T09:00:00Z',
              editedAt: null,
            },
            openReports: 2,
          },
        ],
        nextCursor: '',
      }),
    ),
  );
  const result = await createApiTeachRepository().listModerationQueue(courseId);
  expect(result.ok).toBe(true);
  if (result.ok) expect(result.value.items[0]?.kind).toBe('review');
});

test('rejects a malformed review row in the moderation queue', async () => {
  // A negative star count would crash the queue ('★'.repeat throws), so the row is rejected at the edge.
  for (const review of [
    {
      id: reviewId,
      authorDisplayName: 'Олена',
      stars: -1,
      body: 'x',
      completedLessons: 0,
      totalLessons: 1,
      createdAt: '2026-10-01T09:00:00Z',
      editedAt: null,
    },
    {
      id: 'not-a-uuid',
      authorDisplayName: null,
      stars: 3,
      body: 'x',
      completedLessons: 0,
      totalLessons: 1,
      createdAt: '2026-10-01T09:00:00Z',
      editedAt: null,
    },
    {
      id: reviewId,
      authorDisplayName: null,
      stars: 3,
      completedLessons: 0,
      totalLessons: 1,
      createdAt: '2026-10-01T09:00:00Z',
      editedAt: null,
    },
  ]) {
    server.use(
      http.get(`${apiOrigin()}/api/v1/teach/courses/${courseId}/moderation-queue`, () =>
        HttpResponse.json({
          items: [{ kind: 'review', comment: null, review, openReports: 1 }],
          nextCursor: '',
        }),
      ),
    );
    expect(await createApiTeachRepository().listModerationQueue(courseId)).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  }
});
