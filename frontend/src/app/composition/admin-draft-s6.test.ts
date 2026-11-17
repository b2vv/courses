import { afterEach, describe, expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import { createAudioTracks } from '../../modules/media';
import type { AudioTrackRepository } from '../../modules/media';
import type { DraftRepository, LessonDraft } from '../../modules/transcript';
import { createAdminAudioQueries } from './admin-audio-queries';
import { createDraftQueries, draftPollMs } from './admin-draft-queries';

// MS3a: audio writes reach the mounted per-track list; stale never starts recognition (O7).
const lessonId = '20000000-0000-4000-8000-000000000001';
const courseId = '10000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const trackId = '60000000-0000-4000-8000-000000000001';

function draft(state: LessonDraft['state'], stale: boolean, updatedAt: string): LessonDraft {
  return {
    lessonId,
    trackLanguage: 'en',
    trackMultilingual: false,
    trackId: state === null ? null : '60000000-0000-4000-8000-000000000001',
    state,
    language: state === 'ready' ? 'en' : null,
    failure: null,
    stale,
    updatedAt,
    assetId: videoId,
    generation: 1,
    segments: [],
  };
}

function drafts(responses: readonly LessonDraft[]) {
  let reads = 0;
  const repository: DraftRepository = {
    listTranscripts: async () => ({ ok: true, value: [] }),
    listDrafts: async () => {
      const value = responses[Math.min(reads, responses.length - 1)];
      reads++;
      if (value === undefined) throw new Error('no response');
      return { ok: true, value: [value] };
    },
    regenerate: async () => ({ ok: true, value: undefined }),
  };
  return { repository, reads: () => reads };
}

function audioQueries() {
  const unavailable = async () =>
    ({ ok: false, error: { code: 'dependency_unavailable' } }) as const;
  const repository: AudioTrackRepository = {
    list: async () => ({ ok: true, value: { items: [], multilingualDefault: false } }),
    mint: unavailable,
    rename: unavailable,
    setDefault: async () => ({ ok: true, value: null }),
    remove: async () => ({ ok: true, value: null }),
    transfer: unavailable,
    cancelUpload: unavailable,
  };
  return createAdminAudioQueries(createAudioTracks(repository));
}

afterEach(() => {
  rstest.useRealTimers();
});

describe('a default audio-track change and the mounted draft editor', () => {
  test.each(['setDefault', 'remove'] as const)(
    '%s re-reads drafts without polling a stale draft',
    async (mutation) => {
      rstest.useFakeTimers();
      const source = drafts([
        draft('ready', false, '2026-11-14T10:00:00.000Z'),
        draft('ready', true, '2026-11-14T10:00:00.000Z'),
        draft('ready', true, '2026-11-14T10:00:00.000Z'),
        draft('running', false, '2026-11-14T10:01:00.000Z'),
        draft('ready', false, '2026-11-14T10:02:00.000Z'),
      ]);
      const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const observer = new QueryObserver(
        client,
        createDraftQueries(source.repository).lesson(lessonId, client),
      );
      const unsubscribe = observer.subscribe(() => {});
      await observer.refetch();
      expect(observer.getCurrentResult().data?.[0]?.stale).toBe(false);

      await client
        .getMutationCache()
        .build(client, audioQueries()[mutation](client))
        .execute({ courseId, videoId, trackId });
      expect(observer.getCurrentResult().data?.[0]?.stale).toBe(true);

      const readsAfterWrite = source.reads();
      await rstest.advanceTimersByTimeAsync(draftPollMs * 3);
      expect(source.reads()).toBe(readsAfterWrite);
      unsubscribe();
    },
  );
});
