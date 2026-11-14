import { afterEach, describe, expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import { createAudioTracks } from '../../modules/media';
import type { AudioTrackRepository } from '../../modules/media';
import type { DraftRepository, LessonDraft } from '../../modules/transcript';
import { createAdminAudioQueries } from './admin-audio-queries';
import { createDraftQueries, draftPollMs, draftTaskTimeoutMs } from './admin-draft-queries';

// Multi-audio S6 (D-S6-6): a default-track change reaches a mounted draft editor.
const lessonId = '20000000-0000-4000-8000-000000000001';
const courseId = '10000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const trackId = '60000000-0000-4000-8000-000000000001';

function draft(state: LessonDraft['state'], stale: boolean, updatedAt: string): LessonDraft {
  return {
    lessonId,
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
    getDraft: async () => {
      const value = responses[Math.min(reads, responses.length - 1)];
      reads++;
      if (value === undefined) throw new Error('no response');
      return { ok: true, value };
    },
    regenerate: async () => ({ ok: true, value: undefined }),
  };
  return { repository, reads: () => reads };
}

function audioQueries() {
  const unavailable = async () =>
    ({ ok: false, error: { code: 'dependency_unavailable' } }) as const;
  const repository: AudioTrackRepository = {
    list: async () => ({ ok: true, value: [] }),
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
    '%s re-reads the draft, which polls while stale and stops once replaced',
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
      expect(observer.getCurrentResult().data?.stale).toBe(false);

      await client
        .getMutationCache()
        .build(client, audioQueries()[mutation](client))
        .execute({ courseId, videoId, trackId });
      expect(observer.getCurrentResult().data?.stale).toBe(true);

      await rstest.advanceTimersByTimeAsync(draftPollMs);
      expect(observer.getCurrentResult().data?.stale).toBe(true);
      await rstest.advanceTimersByTimeAsync(draftPollMs);
      expect(observer.getCurrentResult().data?.state).toBe('running');
      await rstest.advanceTimersByTimeAsync(draftPollMs);
      expect(observer.getCurrentResult().data).toMatchObject({ state: 'ready', stale: false });
      const readsAtReplaced = source.reads();
      await rstest.advanceTimersByTimeAsync(draftPollMs * 3);
      expect(source.reads()).toBe(readsAtReplaced);
      unsubscribe();
    },
  );

  test('stale polling gives up after the regeneration window when the worker makes no progress', async () => {
    rstest.useFakeTimers();
    const source = drafts([draft('ready', true, '2026-11-14T10:00:00.000Z')]);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const observer = new QueryObserver(
      client,
      createDraftQueries(source.repository).lesson(lessonId, client),
    );
    const unsubscribe = observer.subscribe(() => {});
    await observer.refetch();
    await rstest.advanceTimersByTimeAsync(draftTaskTimeoutMs + draftPollMs);
    const readsAtLimit = source.reads();
    expect(readsAtLimit).toBeGreaterThan(2);
    await rstest.advanceTimersByTimeAsync(draftPollMs * 4);
    expect(source.reads()).toBe(readsAtLimit);
    expect(observer.getCurrentResult().data?.stale).toBe(true);
    unsubscribe();
  });
});
