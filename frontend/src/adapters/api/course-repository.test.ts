import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse, delay } from 'msw';
import { setupServer } from 'msw/node';
import { createApiCourseRepository } from './course-repository';
import { apiOrigin } from './api-origin';

const course = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go',
  summary: 'Практика',
  description: 'Опис',
  status: 'published',
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
  publishedAt: '2026-09-01T00:00:00Z',
  price: { minor: 149900, currency: 'UAH' },
  paid: true,
  rating: { average: 4.7, count: 3 },
  lessonCount: 3,
};
const detail = {
  ...course,
  program: [
    { title: 'Модуль 1', lessons: [{ title: 'Урок 1' }, { title: 'Урок 2' }] },
    { title: 'Модуль 2', lessons: [{ title: 'Урок 3' }] },
  ],
};
const query = { q: '', sort: 'recent', limit: 20, offset: 0 } satisfies Parameters<
  ReturnType<typeof createApiCourseRepository>['listPublished']
>[0];
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('validates transport data and maps it to the published domain model', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses`, () =>
      HttpResponse.json({ items: [course], pagination: { limit: 20, offset: 0, total: 1 } }),
    ),
  );
  const result = await createApiCourseRepository().listPublished(query);
  expect(result.ok).toBe(true);
  if (result.ok) {
    expect(result.value.items[0]?.title).toBe('Go');
    expect(result.value.items[0]).not.toHaveProperty('createdAt');
  }
});
test('maps the ordered title-only program on a published course detail', async () => {
  server.use(http.get(`${apiOrigin()}/api/v1/courses/go-basics`, () => HttpResponse.json(detail)));
  const result = await createApiCourseRepository().getPublished('go-basics');
  expect(result).toEqual({
    ok: true,
    value: {
      id: course.id,
      slug: course.slug,
      title: course.title,
      summary: course.summary,
      description: course.description,
      publishedAt: course.publishedAt,
      price: course.price,
      paid: course.paid,
      rating: course.rating,
      lessonCount: course.lessonCount,
      program: detail.program,
    },
  });
});
test('rejects malformed success and hidden content', async () => {
  const repository = createApiCourseRepository();
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses/go-basics`, () =>
      HttpResponse.json({ ...detail, status: 'draft' }),
    ),
  );
  expect(await repository.getPublished('go-basics')).toEqual({
    ok: false,
    error: { code: 'invalid_response' },
  });
  server.use(
    http.get(
      `${apiOrigin()}/api/v1/courses/go-basics`,
      () => new HttpResponse('{broken', { headers: { 'Content-Type': 'application/json' } }),
    ),
  );
  expect(await repository.getPublished('go-basics')).toEqual({
    ok: false,
    error: { code: 'invalid_response' },
  });
});
test('preserves not-found and distinguishes deadline from caller cancellation', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses/go-basics`, () =>
      HttpResponse.json(
        { error: { code: 'course_not_found', message: 'Missing', requestId: 'test' } },
        { status: 404 },
      ),
    ),
  );
  expect(await createApiCourseRepository().getPublished('go-basics')).toEqual({
    ok: false,
    error: { code: 'course_not_found' },
  });
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses/go-basics`, async () => {
      await delay(100);
      return HttpResponse.json(course);
    }),
  );
  expect(await createApiCourseRepository({ timeoutMs: 20 }).getPublished('go-basics')).toEqual({
    ok: false,
    error: { code: 'request_timeout' },
  });
  const controller = new AbortController();
  controller.abort();
  expect(await createApiCourseRepository().getPublished('go-basics', controller.signal)).toEqual({
    ok: false,
    error: { code: 'request_cancelled' },
  });
});

const reviewItem = {
  id: '75000000-0000-4000-8000-000000000001',
  authorDisplayName: 'Автор',
  stars: 4,
  body: 'Корисний курс, усе по ділу.',
  completedLessons: 1,
  totalLessons: 8,
  createdAt: '2026-10-01T09:00:00Z',
  editedAt: null,
};
const errorBody = (code: string) => ({ error: { code, message: code, requestId: 'a'.repeat(32) } });

test('maps a public reviews page with its cursor and the frozen percentage', async () => {
  let cursor: string | null = 'unset';
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses/go-basics/reviews`, ({ request }) => {
      cursor = new URL(request.url).searchParams.get('cursor');
      return HttpResponse.json({
        rating: { average: 4.7, count: 3 },
        items: [reviewItem],
        nextCursor: 'next',
      });
    }),
  );
  const repository = createApiCourseRepository();
  const result = await repository.listReviews('go-basics', 'page-2');
  expect(cursor).toBe('page-2');
  expect(result).toEqual({
    ok: true,
    value: {
      rating: { average: 4.7, count: 3 },
      nextCursor: 'next',
      items: [{ ...reviewItem, percent: 13 }],
    },
  });
});

test('reads the viewer’s own review with its eligibility', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/courses/${course.id}/review`, () =>
      HttpResponse.json({ review: reviewItem, eligibility: 'ok' }),
    ),
  );
  const result = await createApiCourseRepository().ownReview(course.id);
  expect(result).toEqual({
    ok: true,
    value: { eligibility: 'ok', review: { ...reviewItem, percent: 13 } },
  });
});

test('maps review report outcomes to named codes', async () => {
  const outcomes: Array<[number, ReturnType<typeof errorBody> | null, unknown]> = [
    [204, null, { ok: true, value: undefined }],
    [403, errorBody('report_own_review'), { ok: false, error: { code: 'report_own_review' } }],
    [
      401,
      errorBody('authentication_required'),
      { ok: false, error: { code: 'authentication_required' } },
    ],
    [
      503,
      errorBody('dependency_unavailable'),
      { ok: false, error: { code: 'dependency_unavailable' } },
    ],
  ];
  for (const [status, body, expected] of outcomes) {
    server.use(
      http.post(`${apiOrigin()}/api/v1/learn/reviews/${reviewItem.id}/reports`, () =>
        status === 204 ? new HttpResponse(null, { status }) : HttpResponse.json(body, { status }),
      ),
    );
    expect(await createApiCourseRepository().reportReview(reviewItem.id)).toEqual(expected);
  }
});

test('maps created and updated own reviews', async () => {
  for (const status of [201, 200]) {
    server.use(
      http.put(`${apiOrigin()}/api/v1/learn/courses/${course.id}/review`, () =>
        HttpResponse.json(reviewItem, { status }),
      ),
    );
    const result = await createApiCourseRepository().putReview(course.id, {
      stars: 4,
      body: reviewItem.body,
    });
    expect(result).toEqual({ ok: true, value: { ...reviewItem, percent: 13 } });
  }
});
test('maps own-review refusal codes', async () => {
  for (const code of [
    'review_owner_forbidden',
    'review_purchase_required',
    'review_not_started',
    'review_not_available',
    'review_invalid',
    'review_removed',
  ]) {
    server.use(
      http.put(`${apiOrigin()}/api/v1/learn/courses/${course.id}/review`, () =>
        HttpResponse.json(errorBody(code), {
          status: code === 'review_invalid' ? 400 : code === 'review_removed' ? 409 : 403,
        }),
      ),
    );
    expect(
      await createApiCourseRepository().putReview(course.id, { stars: 4, body: reviewItem.body }),
    ).toEqual({ ok: false, error: { code } });
  }
});
test('maps own-review deletion and named refusal outcomes', async () => {
  for (const [status, code] of [
    [204, undefined],
    [409, 'review_removed'],
    [404, 'review_not_found'],
  ] as const) {
    server.use(
      http.delete(`${apiOrigin()}/api/v1/learn/courses/${course.id}/review`, () =>
        code === undefined
          ? new HttpResponse(null, { status })
          : HttpResponse.json(errorBody(code), { status }),
      ),
    );
    expect(await createApiCourseRepository().deleteReview(course.id)).toEqual(
      code === undefined ? { ok: true, value: undefined } : { ok: false, error: { code } },
    );
  }
});
test('maps a learning review page with cursor', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/courses/${course.id}/reviews`, () =>
      HttpResponse.json({ rating: null, items: [reviewItem], nextCursor: 'next' }),
    ),
  );
  expect(await createApiCourseRepository().listLearningReviews(course.id)).toEqual({
    ok: true,
    value: { rating: null, items: [{ ...reviewItem, percent: 13 }], nextCursor: 'next' },
  });
});
