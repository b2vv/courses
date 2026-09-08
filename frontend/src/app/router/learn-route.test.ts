import { expect, test } from '@rstest/core';
import { parseCourseRoute, parseLessonRoute } from './learn-route';

const lessonId = '30000000-0000-4000-8000-000000000001';

test('parses exactly a course slug and lesson UUID from the protected learn splat', () => {
  expect(parseLessonRoute(`go-basics/${lessonId}`)).toEqual({ slug: 'go-basics', lessonId });
});

test('P03 parses a course-only learning entry without inventing a lesson id', () => {
  expect(parseCourseRoute('go-basics')).toBe('go-basics');
  expect(parseCourseRoute(`go-basics/${lessonId}`)).toBeNull();
});

test.each([
  undefined,
  '',
  'go-basics',
  `go-basics/${lessonId}/extra`,
  `Go-Basics/${lessonId}`,
  'go-basics/not-a-uuid',
])('rejects a non-lesson learn path: %s', (path) => {
  expect(parseLessonRoute(path)).toBeNull();
});
