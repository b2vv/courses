import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiTeachRepository } from './teach-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const reviewId = '75000000-0000-4000-8000-000000000001';
const userId = '90000000-0000-4000-8000-000000000001';
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
            authorBlocked: false,
          },
        ],
        nextCursor: '',
      }),
    ),
  );
  const result = await createApiTeachRepository().listModerationQueue(courseId);
  expect(result.ok).toBe(true);
  if (result.ok) expect(result.value.items[0]?.kind).toBe('review');
  if (result.ok) expect(result.value.items[0]?.authorBlocked).toBe(false);
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

const learnerCourse = {
  courseId,
  slug: 'intro-course',
  title: 'Intro',
  state: 'active',
  progressPercent: 42,
  enrolledAt: '2026-10-01T09:00:00Z',
};
// The wire row carries an email the domain must never expose (Lo07). The adapter drops it by
// explicit mapping; the equality below asserts the mapped row has no email at all.
const wireRow = {
  userId,
  displayName: '', // a deleted account — anonymised, not a parse failure
  lastActivityAt: null,
  email: 'leak@example.com',
  courses: [learnerCourse],
};
const mappedRow = { userId, displayName: '', lastActivityAt: null, courses: [learnerCourse] };
const learnerPayload = { entries: [wireRow], nextCursor: '' };

test('maps a learner overview page for both scopes and normalises the empty cursor', async () => {
  for (const [scope, path] of [
    ['teach', 'teach'],
    ['admin', 'admin'],
  ] as const) {
    server.use(
      http.get(`${apiOrigin()}/api/v1/${path}/learners`, () => HttpResponse.json(learnerPayload)),
    );
    const result = await createApiTeachRepository().listLearners(scope, {});
    expect(result).toEqual({ ok: true, value: { entries: [mappedRow], nextCursor: null } });
    // The email on the wire row does not leak into the domain (Lo07).
    const [first] = result.ok ? result.value.entries : [];
    expect(Object.keys(first ?? {})).not.toContain('email');
  }
});

test('rejects a malformed learner row', async () => {
  for (const row of [
    { ...wireRow, courses: [] }, // a row must carry at least one course
    { ...wireRow, courses: [{ ...learnerCourse, progressPercent: 101 }] },
    { ...wireRow, userId: 'not-a-uuid' },
  ]) {
    server.use(
      http.get(`${apiOrigin()}/api/v1/teach/learners`, () =>
        HttpResponse.json({ entries: [row], nextCursor: null }),
      ),
    );
    expect(await createApiTeachRepository().listLearners('teach', {})).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  }
});

test('maps a forbidden course filter to learner_course_forbidden', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/teach/learners`, () =>
      HttpResponse.json(error('learner_course_forbidden'), { status: 403 }),
    ),
  );
  expect(await createApiTeachRepository().listLearners('teach', { courseId })).toEqual({
    ok: false,
    error: { code: 'learner_course_forbidden' },
  });
});

const grant = {
  id: reviewId,
  userId,
  courseId,
  source: 'moderation',
  state: 'blocked',
  grantedAt: '2026-10-01T09:00:00Z',
  grantedBy: null,
};

test('blocks and unblocks a comment or review author from the queue', async () => {
  for (const [kind, path] of [
    ['comment', 'comments'],
    ['review', 'reviews'],
  ] as const) {
    const id = '60000000-0000-4000-8000-000000000001';
    for (const action of ['block-author', 'unblock-author'] as const) {
      server.use(
        http.post(`${apiOrigin()}/api/v1/teach/${path}/${id}/${action}`, () =>
          HttpResponse.json(grant),
        ),
      );
      const result =
        action === 'block-author'
          ? await createApiTeachRepository().blockAuthor(kind, id)
          : await createApiTeachRepository().unblockAuthor(kind, id);
      expect(result).toEqual({ ok: true, value: undefined });
    }
  }
});

test('maps named author-block refusals', async () => {
  const id = '60000000-0000-4000-8000-000000000002';
  for (const [status, code] of [
    [409, 'grant_transition_invalid'],
    [403, 'enrollment_forbidden'],
    [404, 'comment_not_found'],
  ] as const) {
    server.use(
      http.post(`${apiOrigin()}/api/v1/teach/comments/${id}/block-author`, () =>
        HttpResponse.json(error(code), { status }),
      ),
    );
    expect(await createApiTeachRepository().blockAuthor('comment', id)).toEqual({
      ok: false,
      error: { code },
    });
  }
});
