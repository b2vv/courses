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
  server.use(
    http.get(`${apiOrigin()}/api/v1/courses/go-basics`, () => HttpResponse.json(detail)),
  );
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
