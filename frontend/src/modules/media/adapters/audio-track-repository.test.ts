import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { createAudioTracks } from '../application/audio-tracks';
import type { AudioTrackErrorCode, AudioUploadTicket } from '../domain/audio-track';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { audioTusResource, createApiAudioTrackRepository } from './audio-track-repository';
import type { FingerprintStorage } from './tus-client';

const videoId = '50000000-0000-4000-8000-000000000001';
const trackId = '60000000-0000-4000-8000-000000000001';
const api = `${apiOrigin()}/api/v1/admin/videos/${videoId}/audio-tracks`;
const tusUrl = `${apiOrigin()}/uploads/audio/${trackId}`;
const fingerprintKey = `courses.media.upload-fingerprint.v1.audio.${trackId}`;
const envelope = (code: string) => ({
  error: { code, message: 'x', requestId: 'a'.repeat(32) },
});
const wireTrack = {
  trackId,
  videoId,
  kind: 'uploaded',
  language: 'en',
  label: 'English',
  isDefault: false,
  state: 'ready',
  updatedAt: '2026-11-13T12:00:00.123456Z',
};
const ticket: AudioUploadTicket = {
  trackId,
  uploadUrl: `/uploads/audio/${trackId}`,
  chunkBytes: 8 * 1024 * 1024,
  maxBytes: 500 * 1024 * 1024,
  expiresAt: '2026-11-14T12:00:00.123456Z',
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

function tusHead(offset: number, length: number) {
  return new HttpResponse(null, {
    status: 200,
    headers: {
      'Tus-Resumable': '1.0.0',
      'Upload-Offset': String(offset),
      'Upload-Length': String(length),
    },
  });
}

function tusPatchEcho() {
  return http.patch(tusUrl, async ({ request }) => {
    const offset = Number(request.headers.get('Upload-Offset'));
    const body = await request.arrayBuffer();
    return new HttpResponse(null, {
      status: 204,
      headers: { 'Tus-Resumable': '1.0.0', 'Upload-Offset': String(offset + body.byteLength) },
    });
  });
}

const file = new File(['abcdef'], 'dub.m4a', { lastModified: 7 });

describe('audio track API repository', () => {
  test('reads the track list and refuses a track of another asset', async () => {
    const tracks = createAudioTracks(createApiAudioTrackRepository());
    server.use(http.get(api, () => HttpResponse.json({ items: [wireTrack] })));
    expect(await tracks.list(videoId)).toEqual({ ok: true, value: [wireTrack] });
    server.use(
      http.get(api, () =>
        HttpResponse.json({
          items: [{ ...wireTrack, videoId: '50000000-0000-4000-8000-000000000002' }],
        }),
      ),
    );
    expect(await tracks.list(videoId)).toEqual({ ok: false, error: { code: 'invalid_response' } });
  });

  test('mints a ticket only for the derived audio upload path', async () => {
    const tracks = createAudioTracks(createApiAudioTrackRepository());
    const input = { declaredBytes: 6, originalName: 'dub.m4a', language: 'en', label: 'English' };
    server.use(http.post(api, () => HttpResponse.json(ticket, { status: 201 })));
    expect(await tracks.mint(videoId, input)).toEqual({ ok: true, value: ticket });
    server.use(
      http.post(api, () =>
        HttpResponse.json({ ...ticket, uploadUrl: `/uploads/${trackId}` }, { status: 201 }),
      ),
    );
    expect(await tracks.mint(videoId, input)).toEqual({
      ok: false,
      error: { code: 'invalid_response' },
    });
  });

  test.each<{ name: string; status: number; code: string; expected: AudioTrackErrorCode }>([
    {
      name: 'duplicate language',
      status: 409,
      code: 'duplicate_language',
      expected: 'duplicate_language',
    },
    { name: 'last Ready track', status: 400, code: 'invalid_request', expected: 'invalid_request' },
    { name: 'foreign course', status: 403, code: 'course_forbidden', expected: 'course_forbidden' },
    {
      name: 'code on the wrong status',
      status: 500,
      code: 'duplicate_language',
      expected: 'dependency_unavailable',
    },
  ])('maps the $name wire error', async ({ status, code, expected }) => {
    const tracks = createAudioTracks(createApiAudioTrackRepository());
    server.use(
      http.patch(`${api}/${trackId}`, () => HttpResponse.json(envelope(code), { status })),
    );
    expect(await tracks.rename(videoId, trackId, 'en', 'English')).toEqual({
      ok: false,
      error: { code: expected },
    });
  });

  test('set-default and delete succeed on 204 with no body', async () => {
    const tracks = createAudioTracks(createApiAudioTrackRepository());
    server.use(
      http.post(`${api}/${trackId}/default`, () => new HttpResponse(null, { status: 204 })),
      http.delete(`${api}/${trackId}`, () => new HttpResponse(null, { status: 204 })),
    );
    expect(await tracks.setDefault(videoId, trackId)).toEqual({ ok: true, value: null });
    expect(await tracks.remove(videoId, trackId)).toEqual({ ok: true, value: null });
  });

  test('refuses invalid input before any request (msw would fail an unhandled call)', async () => {
    const tracks = createAudioTracks(createApiAudioTrackRepository());
    const invalid = { ok: false, error: { code: 'invalid_request' } };
    expect(await tracks.rename(videoId, trackId, 'en us', 'English')).toEqual(invalid);
    expect(await tracks.rename(videoId, trackId, 'en', 'Say "hi"')).toEqual(invalid);
    expect(
      await tracks.mint(videoId, {
        declaredBytes: 500 * 1024 * 1024 + 1,
        originalName: 'dub.m4a',
        language: 'en',
        label: 'English',
      }),
    ).toEqual(invalid);
    expect(
      await tracks.transfer({ kind: 'new', ticket: { ...ticket, maxBytes: 1 } }, file, () => {}),
    ).toEqual(invalid);
  });

  test('a new transfer uploads to the audio path and clears its fingerprint', async () => {
    const storage = memoryStorage();
    const tracks = createAudioTracks(createApiAudioTrackRepository({}, storage));
    server.use(
      http.head(tusUrl, () => tusHead(0, 6)),
      tusPatchEcho(),
    );
    const progress: number[] = [];
    expect(
      await tracks.transfer({ kind: 'new', ticket }, file, (p) => progress.push(p.receivedBytes)),
    ).toEqual({ ok: true, value: null });
    expect(progress).toEqual([0, 6]);
    expect(storage.entries.size).toBe(0);
  });

  test('a resume after reload starts from the track id and the stored fingerprint alone', async () => {
    const storage = memoryStorage();
    storage.setItem(
      fingerprintKey,
      JSON.stringify({ originalName: 'dub.m4a', size: 6, lastModified: 7 }),
    );
    const tracks = createAudioTracks(createApiAudioTrackRepository({}, storage));
    server.use(
      http.head(tusUrl, () => tusHead(3, 6)),
      tusPatchEcho(),
    );
    const progress: number[] = [];
    expect(
      await tracks.transfer({ kind: 'resume', trackId }, file, (p) =>
        progress.push(p.receivedBytes),
      ),
    ).toEqual({ ok: true, value: null });
    expect(progress).toEqual([3, 6]);
    expect(storage.entries.has(fingerprintKey)).toBe(false);
  });

  test.each([
    { name: 'no stored fingerprint', stored: null },
    {
      name: 'a different file',
      stored: { originalName: 'other.m4a', size: 6, lastModified: 7 },
    },
  ])('refuses a resume with $name before any byte is sent', async ({ stored }) => {
    const storage = memoryStorage();
    if (stored !== null) storage.setItem(fingerprintKey, JSON.stringify(stored));
    const tracks = createAudioTracks(createApiAudioTrackRepository({}, storage));
    expect(await tracks.transfer({ kind: 'resume', trackId }, file, () => {})).toEqual({
      ok: false,
      error: { code: 'upload_resume_unverified' },
    });
  });

  test('refuses a resume whose server length disagrees with the re-selected file', async () => {
    const storage = memoryStorage();
    storage.setItem(
      fingerprintKey,
      JSON.stringify({ originalName: 'dub.m4a', size: 6, lastModified: 7 }),
    );
    const tracks = createAudioTracks(createApiAudioTrackRepository({}, storage));
    server.use(http.head(tusUrl, () => tusHead(3, 9)));
    expect(await tracks.transfer({ kind: 'resume', trackId }, file, () => {})).toEqual({
      ok: false,
      error: { code: 'invalid_request' },
    });
  });

  test('cancel needs no ticket: DELETE on the derived path, then the fingerprint goes', async () => {
    const storage = memoryStorage();
    storage.setItem(fingerprintKey, '{}');
    const tracks = createAudioTracks(createApiAudioTrackRepository({}, storage));
    server.use(http.delete(tusUrl, () => new HttpResponse(null, { status: 204 })));
    expect(await tracks.cancelUpload(trackId)).toEqual({ ok: true, value: null });
    expect(storage.entries.size).toBe(0);
  });

  test('the audio tus resource is keyed apart from the video resource', () => {
    expect(audioTusResource.path(trackId)).toBe(`/uploads/audio/${trackId}`);
    expect(audioTusResource.fingerprintKey(trackId)).toBe(fingerprintKey);
    expect(audioTusResource.fingerprintKey(trackId)).not.toBe(
      `courses.media.upload-fingerprint.v1.${trackId}`,
    );
  });
});
