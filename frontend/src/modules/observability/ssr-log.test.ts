import { afterEach, expect, test } from '@rstest/core';
// The SSR log builders live beside server.mjs (a copied-verbatim .mjs, not
// bundled), so the test reaches up out of src to the real file the server uses.
import { ssrRenderFailedLine, ssrRequestLine } from '../../../server-log.mjs';

const originalRelease = process.env.APP_RELEASE;

afterEach(() => {
  if (originalRelease === undefined) {
    delete process.env.APP_RELEASE;
  } else {
    process.env.APP_RELEASE = originalRelease;
  }
});

test('ssrRequestLine carries source=ssr, release and route', () => {
  process.env.APP_RELEASE = 'abc1234';
  const line = ssrRequestLine({ status: 200, durationMs: 12, route: '/learn/x' });
  expect(line.endsWith('\n')).toBe(true);
  const parsed = JSON.parse(line);
  expect(parsed).toMatchObject({
    level: 'info',
    service: 'frontend',
    source: 'ssr',
    event: 'http_request',
    status: 200,
    durationMs: 12,
    route: '/learn/x',
    release: 'abc1234',
  });
});

test('ssrRenderFailedLine is an error tagged source=ssr with route and release', () => {
  process.env.APP_RELEASE = 'abc1234';
  const line = ssrRenderFailedLine({ route: '/learn/x' });
  expect(line.endsWith('\n')).toBe(true);
  const parsed = JSON.parse(line);
  expect(parsed).toMatchObject({
    level: 'error',
    service: 'frontend',
    source: 'ssr',
    event: 'render_failed',
    route: '/learn/x',
    release: 'abc1234',
  });
});

test('release defaults to "dev" when APP_RELEASE is unset', () => {
  delete process.env.APP_RELEASE;
  const parsed = JSON.parse(ssrRequestLine({ status: 200, durationMs: 1, route: '/' }));
  expect(parsed.release).toBe('dev');
});
