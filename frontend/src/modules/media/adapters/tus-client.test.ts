import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import type { VideoTransferTarget, VideoUploadTicket } from '../domain/video';
import { videoChunkBytes } from '../domain/video';
import { apiOrigin } from '../../../adapters/api/api-origin';
import {
  buildTusHeaders,
  createTusClient,
  createTusTransport,
  nextTusChunk,
  parseTusHead,
  parseTusOffset,
  sameUploadFingerprint,
  tusStatusCode,
} from './tus-client';
import type { FingerprintStorage } from './tus-client';
import type { TusTransport } from './tus-client';

const videoId = '50000000-0000-4000-8000-000000000001';
const url = `${apiOrigin()}/uploads/${videoId}`;
const ticket: VideoUploadTicket = {
  videoId,
  uploadUrl: `/uploads/${videoId}`,
  chunkBytes: videoChunkBytes,
  maxBytes: 2 * 1024 * 1024 * 1024,
  expiresAt: '2026-09-07T12:00:00.123456Z',
};
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function memoryStorage(): FingerprintStorage & { readonly entries: Map<string, string> } {
  const entries = new Map<string, string>();
  return {
    entries,
    getItem: (key) => entries.get(key) ?? null,
    setItem: (key, value) => {
      entries.set(key, value);
    },
    removeItem: (key) => {
      entries.delete(key);
    },
  };
}

describe('tus header and chunk contract', () => {
  test('builds the exact request headers as pure values', () => {
    expect(Object.fromEntries(buildTusHeaders('HEAD'))).toEqual({ 'tus-resumable': '1.0.0' });
    expect(Object.fromEntries(buildTusHeaders('PATCH', 3))).toEqual({
      'content-type': 'application/offset+octet-stream',
      'tus-resumable': '1.0.0',
      'upload-offset': '3',
    });
    expect(Object.fromEntries(buildTusHeaders('DELETE'))).toEqual({
      'tus-resumable': '1.0.0',
    });
  });

  test('parses offset and declared length only from valid tus response headers', () => {
    expect(
      parseTusHead(
        new Headers({
          'Tus-Resumable': '1.0.0',
          'Upload-Offset': '3',
          'Upload-Length': '6',
        }),
      ),
    ).toEqual({ offset: 3, declaredBytes: 6 });
    expect(
      parseTusHead(
        new Headers({
          'Tus-Resumable': '1.0.0',
          'Upload-Offset': '7',
          'Upload-Length': '6',
        }),
      ),
    ).toBeNull();
    expect(parseTusOffset(new Headers({ 'Tus-Resumable': '1.0.0' }))).toBeNull();
    expect(
      parseTusOffset(new Headers({ 'Tus-Resumable': '0.2.2', 'Upload-Offset': '6' })),
    ).toBeNull();
  });

  test.each([
    [400, 'unsupported_media'],
    [404, 'video_not_found'],
    [413, 'upload_too_large'],
    [503, 'dependency_unavailable'],
    [507, 'dependency_unavailable'],
    [409, 'invalid_response'],
  ] as const)('maps tus status %i to %s', (status, code) => {
    expect(tusStatusCode(status)).toBe(code);
  });

  test('computes bounded sequential chunks and stops exactly at the declared length', () => {
    expect(nextTusChunk(0, videoChunkBytes + 3)).toEqual({ start: 0, end: videoChunkBytes });
    expect(nextTusChunk(videoChunkBytes, videoChunkBytes + 3)).toEqual({
      start: videoChunkBytes,
      end: videoChunkBytes + 3,
    });
    expect(nextTusChunk(videoChunkBytes + 3, videoChunkBytes + 3)).toBeNull();
    expect(nextTusChunk(videoChunkBytes + 4, videoChunkBytes + 3)).toBeNull();
  });
});

describe('tus wire contract', () => {
  test('sends the exact PATCH and HEAD headers through the transport', async () => {
    server.use(
      http.head(url, ({ request }) => {
        expect(request.headers.get('Tus-Resumable')).toBe('1.0.0');
        expect(request.headers.has('Content-Type')).toBe(false);
        expect(request.headers.has('Upload-Offset')).toBe(false);
        return new HttpResponse(null, {
          status: 200,
          headers: {
            'Tus-Resumable': '1.0.0',
            'Upload-Offset': '3',
            'Upload-Length': '6',
          },
        });
      }),
      http.patch(url, async ({ request }) => {
        expect(request.headers.get('Tus-Resumable')).toBe('1.0.0');
        expect(request.headers.get('Content-Type')).toBe('application/offset+octet-stream');
        expect(request.headers.get('Upload-Offset')).toBe('3');
        expect(await request.text()).toBe('def');
        return new HttpResponse(null, {
          status: 204,
          headers: { 'Tus-Resumable': '1.0.0', 'Upload-Offset': '6' },
        });
      }),
    );

    const transport = createTusTransport();
    expect(await transport.send('HEAD', videoId)).toMatchObject({ ok: true });
    expect(await transport.send('PATCH', videoId, new Blob(['def']), 3)).toMatchObject({
      ok: true,
    });
  });

  test('reads each PATCH offset and advances to the next chunk', async () => {
    const declaredBytes = videoChunkBytes + 3;
    const offsets: number[] = [];
    const bodySizes: number[] = [];
    server.use(
      http.head(
        url,
        () =>
          new HttpResponse(null, {
            status: 200,
            headers: {
              'Tus-Resumable': '1.0.0',
              'Upload-Offset': '0',
              'Upload-Length': String(declaredBytes),
            },
          }),
      ),
      http.patch(url, async ({ request }) => {
        const offset = Number(request.headers.get('Upload-Offset'));
        const bodySize = (await request.arrayBuffer()).byteLength;
        offsets.push(offset);
        bodySizes.push(bodySize);
        return new HttpResponse(null, {
          status: 204,
          headers: {
            'Tus-Resumable': '1.0.0',
            'Upload-Offset': String(offset + bodySize),
          },
        });
      }),
    );

    const progress: number[] = [];
    const result = await createTusClient({}, memoryStorage()).transfer(
      { kind: 'new', ticket },
      new File([new ArrayBuffer(declaredBytes)], 'lesson.mp4', { lastModified: 1234 }),
      ({ receivedBytes }) => progress.push(receivedBytes),
    );

    expect(result).toEqual({ ok: true, value: null });
    expect(offsets).toEqual([0, videoChunkBytes]);
    expect(bodySizes).toEqual([videoChunkBytes, 3]);
    expect(progress).toEqual([0, videoChunkBytes, declaredBytes]);
  });

  test.each([
    ['missing', { 'Upload-Offset': '6' }],
    ['different', { 'Tus-Resumable': '0.2.2', 'Upload-Offset': '6' }],
  ])('rejects a PATCH response with a %s tus version', async (_case, responseHeaders) => {
    server.use(
      http.head(
        url,
        () =>
          new HttpResponse(null, {
            status: 200,
            headers: {
              'Tus-Resumable': '1.0.0',
              'Upload-Offset': '0',
              'Upload-Length': '6',
            },
          }),
      ),
      http.patch(url, () => new HttpResponse(null, { status: 204, headers: responseHeaders })),
    );

    expect(
      await createTusClient({}, memoryStorage()).transfer(
        { kind: 'new', ticket },
        new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 }),
        () => {},
      ),
    ).toEqual({ ok: false, error: { code: 'invalid_response' } });
  });
});

describe('tus client', () => {
  test('maps PATCH status and sends DELETE through MSW', async () => {
    server.use(
      http.patch(url, async ({ request }) => {
        expect(await request.text()).toBe('bad');
        return HttpResponse.json({}, { status: 400 });
      }),
      http.delete(url, () => new HttpResponse(null, { status: 204 })),
    );
    expect(await createTusTransport().send('PATCH', videoId, new Blob(['bad']), 0)).toEqual({
      ok: false,
      code: 'unsupported_media',
    });
    expect(await createTusClient({}, memoryStorage()).cancel(videoId)).toEqual({
      ok: true,
      value: null,
    });
  });

  test('persists the selected File fingerprint when a fresh transfer starts', async () => {
    const storage = memoryStorage();
    const file = new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 });
    server.use(http.head(url, () => HttpResponse.json({}, { status: 503 })));
    expect(
      await createTusClient({}, storage).transfer({ kind: 'new', ticket }, file, () => {}),
    ).toEqual({ ok: false, error: { code: 'dependency_unavailable' } });
    expect([...storage.entries.values()]).toEqual([
      JSON.stringify({ originalName: 'lesson.mp4', size: 6, lastModified: 1234 }),
    ]);
  });

  test('R44 refuses mismatched or absent fingerprints before HEAD can splice bytes', async () => {
    const storage = memoryStorage();
    const original = new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 });
    const different = new File(['ghijkl'], 'lesson.mp4', { lastModified: 5678 });
    server.use(http.head(url, () => HttpResponse.json({}, { status: 503 })));
    await createTusClient({}, storage).transfer({ kind: 'new', ticket }, original, () => {});
    server.resetHandlers();

    const resume: VideoTransferTarget = {
      kind: 'resume',
      videoId,
      declaredBytes: 6,
    };
    expect(sameUploadFingerprint(original, original)).toBe(true);
    expect(sameUploadFingerprint(original, different)).toBe(false);
    expect(await createTusClient({}, storage).transfer(resume, different, () => {})).toEqual({
      ok: false,
      error: { code: 'upload_resume_unverified' },
    });
    expect(await createTusClient({}, memoryStorage()).transfer(resume, original, () => {})).toEqual(
      {
        ok: false,
        error: { code: 'upload_resume_unverified' },
      },
    );
  });

  test('a matching fingerprint resumes from the parsed stored offset', async () => {
    const storage = memoryStorage();
    const file = new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 });
    server.use(http.head(url, () => new HttpResponse(null, { status: 503 })));
    await createTusClient({}, storage).transfer({ kind: 'new', ticket }, file, () => {});

    server.resetHandlers();
    server.use(
      http.head(url, ({ request }) => {
        expect(request.headers.get('Tus-Resumable')).toBe('1.0.0');
        return new HttpResponse(null, {
          status: 200,
          headers: {
            'Tus-Resumable': '1.0.0',
            'Upload-Offset': '3',
            'Upload-Length': '6',
          },
        });
      }),
      http.patch(url, async ({ request }) => {
        expect(request.headers.get('Upload-Offset')).toBe('3');
        expect(await request.text()).toBe('def');
        return new HttpResponse(null, {
          status: 204,
          headers: { 'Tus-Resumable': '1.0.0', 'Upload-Offset': '6' },
        });
      }),
    );
    const progress: number[] = [];
    expect(
      await createTusClient({}, storage).transfer(
        { kind: 'resume', videoId, declaredBytes: 6 },
        file,
        ({ receivedBytes }) => progress.push(receivedBytes),
      ),
    ).toEqual({ ok: true, value: null });
    expect(progress).toEqual([3, 6]);
    expect(storage.entries.size).toBe(0);
  });

  test('a verified resume still refuses a server length that differs from declaredBytes', async () => {
    const storage = memoryStorage();
    const file = new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 });
    server.use(http.head(url, () => new HttpResponse(null, { status: 503 })));
    await createTusClient({}, storage).transfer({ kind: 'new', ticket }, file, () => {});
    const transport: TusTransport = {
      async send(method) {
        if (method !== 'HEAD') throw new Error('PATCH must not run after a length mismatch');
        return {
          ok: true,
          response: new Response(null, {
            status: 200,
            headers: {
              'Tus-Resumable': '1.0.0',
              'Upload-Offset': '3',
              'Upload-Length': '7',
            },
          }),
        };
      },
    };
    expect(
      await createTusClient({}, storage, transport).transfer(
        { kind: 'resume', videoId, declaredBytes: 6 },
        file,
        () => {},
      ),
    ).toEqual({ ok: false, error: { code: 'invalid_request' } });
  });

  test('a fresh target without prior attestation refuses an already-advanced ticket', async () => {
    const storage = memoryStorage();
    const transport: TusTransport = {
      async send() {
        return {
          ok: true,
          response: new Response(null, {
            status: 200,
            headers: {
              'Tus-Resumable': '1.0.0',
              'Upload-Offset': '3',
              'Upload-Length': '6',
            },
          }),
        };
      },
    };
    expect(
      await createTusClient({}, storage, transport).transfer(
        { kind: 'new', ticket },
        new File(['abcdef'], 'lesson.mp4', { lastModified: 1234 }),
        () => {},
      ),
    ).toEqual({ ok: false, error: { code: 'upload_resume_unverified' } });
    expect(storage.entries.size).toBe(0);
  });
});
