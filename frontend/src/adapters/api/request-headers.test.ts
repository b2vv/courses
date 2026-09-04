import { expect, test } from '@rstest/core';
import { apiRequestHeaders } from './request-headers';

test('forwards the request cookie verbatim when one was given', () => {
  const headers = apiRequestHeaders('courseforge_session=abc; courseforge_locale=uk');
  expect(headers.get('Cookie')).toBe('courseforge_session=abc; courseforge_locale=uk');
  expect(headers.get('Accept')).toBe('application/json');
});

test.each([undefined, ''])('sends no Cookie header at all for %p', (cookie) => {
  const headers = apiRequestHeaders(cookie);
  expect(headers.has('Cookie')).toBe(false);
  expect(headers.get('Accept')).toBe('application/json');
});

test('builds an independent header set per request', () => {
  const first = apiRequestHeaders('courseforge_session=first');
  const second = apiRequestHeaders('courseforge_session=second');
  expect(first.get('Cookie')).toBe('courseforge_session=first');
  expect(second.get('Cookie')).toBe('courseforge_session=second');
});
