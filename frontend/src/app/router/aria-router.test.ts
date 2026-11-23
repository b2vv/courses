import { expect, test } from '@rstest/core';
import { ariaNavigation, surfaceOf } from './aria-router';

const at = (path: string) => `https://courses.test${path}`;

test.each([
  ['/admin', 'workspace'],
  ['/admin/users', 'workspace'],
  ['/teach', 'workspace'],
  ['/teach/courses/1', 'workspace'],
  ['/administrator', 'client'],
  ['/teacher', 'client'],
  ['/courses', 'client'],
  ['/', 'client'],
] as const)('C3b R2 %s belongs to the %s surface', (path, surface) => {
  expect(surfaceOf(path)).toBe(surface);
});

test.each([
  // Inside one surface: a move inside the app.
  ['/courses', '/courses/go-basics', false],
  ['/courses/go-basics', '/courses?offset=12&q=go', false],
  ['/admin', '/admin/users', false],
  ['/teach', '/teach/learners', false],
  ['/', '#values', false],
  // Across surfaces: a full load (R2).
  ['/courses', '/admin', true],
  ['/courses/go', '/teach', true],
  ['/admin/courses/1', '/courses/go', true],
  ['/teach', '/', true],
] as const)('C3b R2 a press on %s → %s reloads the document: %s', (from, href, reload) => {
  const next = ariaNavigation(at(from), href, undefined);
  expect(next.reloadDocument === true).toBe(reload);
});

test('C3b R1 keeps the search and hash of the target and resolves a relative href', () => {
  expect(ariaNavigation(at('/courses?q=go'), '/courses?q=go&offset=12', undefined).href).toBe(
    '/courses?q=go&offset=12',
  );
  expect(ariaNavigation(at('/'), '#values', undefined).href).toBe('/#values');
});

test('C3b R3 a retry asks for a full load inside one surface', () => {
  expect(
    ariaNavigation(at('/courses/go'), '/courses/go', { reloadDocument: true }).reloadDocument,
  ).toBe(true);
});

test('passes replace through', () => {
  expect(ariaNavigation(at('/courses'), '/courses/go', { replace: true }).replace).toBe(true);
  expect('replace' in ariaNavigation(at('/courses'), '/courses/go', undefined)).toBe(false);
});
