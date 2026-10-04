import { expect, test } from '@rstest/core';
import { createCatalog } from './catalog';
import type { CourseRepository } from '../../ports/CourseRepository';

const unavailable: CourseRepository = {
  listPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  getPublished: async () => ({ ok: false, error: { code: 'course_not_found' } }),
  listReviews: async () => ({ ok: true, value: { rating: null, items: [], nextCursor: null } }),
  reportReview: async () => ({ ok: true, value: undefined }),
  ownReview: async () => ({ ok: true, value: { eligibility: 'review_not_started', review: null } }),
  putReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  deleteReview: async () => ({ ok: true, value: undefined }),
  listLearningReviews: async () => ({
    ok: true,
    value: { rating: null, items: [], nextCursor: null },
  }),
};
test('rejects invalid input before an unavailable dependency and preserves dependency failure', async () => {
  const catalog = createCatalog(unavailable);
  expect(await catalog.list({ limit: 0 })).toEqual({ ok: false, error: { code: 'invalid_query' } });
  expect(await catalog.list({})).toEqual({ ok: false, error: { code: 'dependency_unavailable' } });
  expect(await catalog.get('../draft')).toEqual({ ok: false, error: { code: 'invalid_slug' } });
  expect(await catalog.get('not-found')).toEqual({
    ok: false,
    error: { code: 'course_not_found' },
  });
});
