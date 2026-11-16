import { describe, expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import { createAudioTracks } from '../../modules/media';
import type { AudioTrack, AudioTrackRepository } from '../../modules/media';
import {
  adminAudioTracksQueryKey,
  audioTrackRefetchInterval,
  createAdminAudioQueries,
} from './admin-audio-queries';
import { lessonDraftQueryKey } from './admin-draft-queries';
import { adminCourseVideosQueryKey, videoPollMs } from './admin-video-queries';

const courseId = '10000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const trackId = '60000000-0000-4000-8000-000000000001';

function track(state: AudioTrack['state']): AudioTrack {
  return {
    trackId,
    videoId,
    kind: 'uploaded',
    language: 'en',
    label: 'English',
    isDefault: false,
    multilingual: false,
    state,
    updatedAt: '2026-11-13T12:00:00.000000Z',
  };
}

function setup(responses: readonly (readonly AudioTrack[])[]) {
  let read = 0;
  const unavailable = async () =>
    ({ ok: false, error: { code: 'dependency_unavailable' } }) as const;
  const repository: AudioTrackRepository = {
    list: async () => {
      const value = responses[Math.min(read, responses.length - 1)] ?? [];
      read += 1;
      return { ok: true, value: { items: value, multilingualDefault: false } };
    },
    mint: unavailable,
    rename: unavailable,
    setDefault: unavailable,
    remove: unavailable,
    transfer: unavailable,
    cancelUpload: unavailable,
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const queries = createAdminAudioQueries(createAudioTracks(repository));
  let videoInvalidations = 0;
  // Seed the course video list so an invalidation is observable on its state.
  client.setQueryData(adminCourseVideosQueryKey(courseId), []);
  const unsubscribe = client.getQueryCache().subscribe((event) => {
    if (
      event.type === 'updated' &&
      event.action.type === 'invalidate' &&
      JSON.stringify(event.query.queryKey) === JSON.stringify(adminCourseVideosQueryKey(courseId))
    )
      videoInvalidations += 1;
  });
  const read_ = () => client.fetchQuery({ ...queries.list(courseId, videoId), staleTime: 0 });
  return { read: read_, invalidations: () => videoInvalidations, unsubscribe };
}

describe('admin audio query boundary', () => {
  test('a track reaching Ready re-reads the course video list', async () => {
    const harness = setup([[track('processing')], [track('ready')]]);
    await harness.read();
    expect(harness.invalidations()).toBe(0);
    await harness.read();
    expect(harness.invalidations()).toBe(1);
    harness.unsubscribe();
  });

  test.each(['ready', 'failed'] as const)(
    'the first read already %s re-reads the course video list, with no snapshot',
    async (state) => {
      // After a reload the worker may have finished before the first audio read: there is no
      // earlier snapshot to compare with, and the asset may still be cached as awaiting_audio.
      const harness = setup([[track(state)]]);
      await harness.read();
      expect(harness.invalidations()).toBe(1);
      harness.unsubscribe();
    },
  );

  test('an identical response does not re-read the course video list again', async () => {
    const harness = setup([[track('ready')], [track('ready')], [track('ready')]]);
    await harness.read();
    await harness.read();
    await harness.read();
    expect(harness.invalidations()).toBe(1);
    harness.unsubscribe();
  });

  test('a list with no terminal track never touches the course video list', async () => {
    const harness = setup([[], [track('uploading')]]);
    await harness.read();
    await harness.read();
    expect(harness.invalidations()).toBe(0);
    harness.unsubscribe();
  });

  test('polls while a track is queued or processing, or a transfer runs in this tab', () => {
    expect(audioTrackRefetchInterval([track('processing')], false)).toBe(videoPollMs);
    expect(audioTrackRefetchInterval([track('queued')], false)).toBe(videoPollMs);
    expect(audioTrackRefetchInterval([track('ready')], true)).toBe(videoPollMs);
    expect(audioTrackRefetchInterval([track('ready')], false)).toBe(false);
    expect(audioTrackRefetchInterval([track('uploading')], false)).toBe(false);
    expect(audioTrackRefetchInterval(undefined, false)).toBe(false);
  });
});

test.each(['rename', 'remove', 'setDefault'] as const)(
  '%s invalidates audio and only the affected lesson drafts',
  async (method) => {
    const lessonId = 'lesson';
    const otherLessonId = 'other';
    const unavailable = async () =>
      ({ ok: false, error: { code: 'dependency_unavailable' } }) as const;
    const repository: AudioTrackRepository = {
      list: unavailable,
      mint: unavailable,
      transfer: unavailable,
      cancelUpload: unavailable,
      rename: async () => ({ ok: true, value: track('ready') }),
      remove: async () => ({ ok: true, value: null }),
      setDefault: async () => ({ ok: true, value: null }),
    };
    const client = new QueryClient();
    client.setQueryData(adminCourseVideosQueryKey(courseId), [{ videoId, lessonId }]);
    client.setQueryData(adminAudioTracksQueryKey(videoId), {
      items: [],
      multilingualDefault: false,
    });
    client.setQueryData(lessonDraftQueryKey(lessonId), []);
    client.setQueryData(lessonDraftQueryKey(otherLessonId), []);
    const queries = createAdminAudioQueries(createAudioTracks(repository));
    const variables = {
      courseId,
      videoId,
      trackId,
      language: 'en',
      label: 'English',
      multilingual: true,
    };
    if (method === 'rename')
      await client.getMutationCache().build(client, queries.rename(client)).execute(variables);
    else await client.getMutationCache().build(client, queries[method](client)).execute(variables);
    expect(client.getQueryState(adminAudioTracksQueryKey(videoId))?.isInvalidated).toBe(true);
    expect(client.getQueryState(lessonDraftQueryKey(lessonId))?.isInvalidated).toBe(true);
    expect(client.getQueryState(lessonDraftQueryKey(otherLessonId))?.isInvalidated).toBe(false);
    client.clear();
  },
);

test('mint forwards multilingual to the audio port', async () => {
  const unavailable = async () =>
    ({ ok: false, error: { code: 'dependency_unavailable' } }) as const;
  const calls: unknown[] = [];
  const repository: AudioTrackRepository = {
    list: unavailable,
    transfer: unavailable,
    cancelUpload: unavailable,
    rename: unavailable,
    remove: unavailable,
    setDefault: unavailable,
    mint: async (_video, input) => {
      calls.push(input);
      return {
        ok: true,
        value: {
          trackId,
          uploadUrl: `/uploads/audio/${trackId}`,
          chunkBytes: 8388608,
          maxBytes: 524288000,
          expiresAt: '2026-11-16T12:00:00Z',
        },
      };
    },
  };
  const client = new QueryClient();
  const queries = createAdminAudioQueries(createAudioTracks(repository));
  await client.getMutationCache().build(client, queries.mint(client)).execute({
    courseId,
    videoId,
    declaredBytes: 6,
    originalName: 'dub.m4a',
    language: 'en',
    label: 'Dub',
    multilingual: true,
  });
  expect(calls).toEqual([
    { declaredBytes: 6, originalName: 'dub.m4a', language: 'en', label: 'Dub', multilingual: true },
  ]);
  client.clear();
});
