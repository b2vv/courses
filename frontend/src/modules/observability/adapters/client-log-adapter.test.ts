import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { createClientLogReporter, describeUnknown } from './client-log-adapter';

const ingestUrl = `${apiOrigin()}/api/v1/client-logs`;

const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const flush = () => new Promise((resolve) => setTimeout(resolve, 60));

describe('describeUnknown', () => {
  test('extracts name, message and stack from an Error', () => {
    const described = describeUnknown(new TypeError('boom'));
    expect(described.name).toBe('TypeError');
    expect(described.message).toBe('boom');
    expect(described.stack).toBeTruthy();
  });

  test('keeps a string reason as the message', () => {
    expect(describeUnknown('plain rejection').message).toBe('plain rejection');
  });

  test('gives bare undefined and null distinct, non-empty messages', () => {
    expect(describeUnknown(undefined).message).toBe('undefined rejection');
    expect(describeUnknown(null).message).toBe('null rejection');
  });

  test('distinct non-error objects do not collapse to one message', () => {
    const a = describeUnknown({ code: 1 }).message;
    const b = describeUnknown({ code: 2 }).message;
    expect(a).not.toBe(b);
  });
});

describe('createClientLogReporter.capture', () => {
  test('posts a well-formed event with the required context fields', async () => {
    const bodies: Array<Record<string, unknown>> = [];
    server.use(
      http.post(ingestUrl, async ({ request }) => {
        bodies.push((await request.json()) as Record<string, unknown>);
        return new HttpResponse(null, { status: 204 });
      }),
    );

    createClientLogReporter().capture(new Error('render failed'));
    await flush();

    expect(bodies).toHaveLength(1);
    const body = bodies[0];
    if (!body) {
      throw new Error('expected one ingested body');
    }
    expect(body.level).toBe('error');
    expect(body.message).toBe('render failed');
    expect(body.errorName).toBe('Error');
    expect(typeof body.route).toBe('string');
    expect(typeof body.release).toBe('string');
    expect((body.release as string).length).toBeGreaterThan(0);
    expect(typeof body.userAgent).toBe('string');
  });

  test('swallows a 5xx ingest response without throwing', async () => {
    server.use(http.post(ingestUrl, () => new HttpResponse(null, { status: 503 })));
    const reporter = createClientLogReporter();
    expect(() => reporter.capture(new Error('x'))).not.toThrow();
    await flush();
  });

  test('swallows a transport/network failure without throwing', async () => {
    server.use(http.post(ingestUrl, () => HttpResponse.error()));
    const reporter = createClientLogReporter();
    expect(() => reporter.capture('offline')).not.toThrow();
    await flush();
  });

  test('caps the number of events at the per-session ceiling', async () => {
    let hits = 0;
    server.use(
      http.post(ingestUrl, () => {
        hits += 1;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const reporter = createClientLogReporter();
    for (let i = 0; i < 60; i += 1) {
      reporter.capture(new Error(`e${i}`));
    }
    await flush();
    expect(hits).toBe(50);
  });
});
