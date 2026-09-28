import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiDiscussionRepository } from './discussion-repository';

const lessonId = '30000000-0000-4000-8000-000000000001';
const commentId = '40000000-0000-4000-8000-000000000001';
const route = `${apiOrigin()}/api/v1/learn/lessons/${lessonId}/comments`;
const comment = {
  id: commentId,
  author: { displayName: 'Ada' },
  body: 'A useful note',
  positionMs: 12_345,
  createdAt: '2026-09-28T10:00:00Z',
  editedAt: null,
  removed: false,
  canEdit: true,
  canRemove: true,
};
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('maps a generated discussion page into the frontend domain', async () => {
  server.use(
    http.get(route, () =>
      HttpResponse.json({ threads: [{ root: comment, replies: [] }], nextCursor: '' }),
    ),
  );
  const result = await createApiDiscussionRepository().list(lessonId);
  expect(result).toEqual({
    ok: true,
    value: { threads: [{ root: comment, replies: [] }], nextCursor: null },
  });
});

test('refuses an invalid wire comment instead of rendering it', async () => {
  server.use(
    http.get(route, () =>
      HttpResponse.json({
        threads: [{ root: { ...comment, positionMs: -1 }, replies: [] }],
        nextCursor: '',
      }),
    ),
  );
  const result = await createApiDiscussionRepository().list(lessonId);
  expect(result).toEqual({ ok: false, error: { code: 'invalid_response' } });
});

test('keeps named mutation refusals', async () => {
  server.use(
    http.post(route, () =>
      HttpResponse.json({ error: { code: 'comment_thread_full' } }, { status: 409 }),
    ),
  );
  const result = await createApiDiscussionRepository().publish(lessonId, { body: 'reply' });
  expect(result).toEqual({ ok: false, error: { code: 'comment_thread_full' } });
});
