import { expect, test } from '@rstest/core';
import { createCatalog } from './catalog';
import type { CourseRepository } from '../../ports/CourseRepository';

const unavailable: CourseRepository = {
  listPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  getPublished: async () => ({ ok: false, error: { code: 'course_not_found' } }),
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
