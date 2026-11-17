import { afterEach, expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import {
  createDraftQueries,
  draftOfTrack,
  draftPollMs,
  draftTaskTimeoutMs,
  DraftQueryError,
  lessonDraftQueryKey,
} from './admin-draft-queries';
import type { DraftRepository, LessonDraft } from '../../modules/transcript';

const lessonId = 'lesson';
const trackT = 'T';
const trackU = 'U';
function draft(trackId: string | null, overrides: Partial<LessonDraft> = {}): LessonDraft {
  return {
    lessonId,
    trackId,
    trackLanguage: 'en',
    trackMultilingual: false,
    state: 'ready',
    language: 'en',
    failure: null,
    stale: false,
    updatedAt: '2026-11-16T12:00:00Z',
    assetId: 'asset',
    generation: 1,
    segments: [],
    ...overrides,
  };
}
function harness(initial: LessonDraft[] = [], rejected = false) {
  let rows = initial;
  let reads = 0;
  const calls: string[][] = [];
  const repository: DraftRepository = {
    listDrafts: async () => {
      reads++;
      return { ok: true, value: rows };
    },
    listTranscripts: async () => ({ ok: true, value: [] }),
    regenerate: async (lesson, track) => {
      calls.push([lesson, track]);
      return rejected
        ? { ok: false, error: { code: 'already_running' } }
        : { ok: true, value: undefined };
    },
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const queries = createDraftQueries(repository);
  const options = queries.lesson(lessonId, client);
  const observer = new QueryObserver(client, options);
  const unsubscribe = observer.subscribe(() => {});
  const regenerate = (trackId: string) =>
    client
      .getMutationCache()
      .build(client, queries.regenerate(client))
      .execute({ lessonId, trackId });
  return {
    client,
    queries,
    observer,
    unsubscribe,
    regenerate,
    calls,
    reads: () => reads,
    set: (next: LessonDraft[]) => {
      rows = next;
    },
    track: (id: string) => draftOfTrack(observer.getCurrentResult().data, id),
  };
}
afterEach(() => {
  rstest.useRealTimers();
});

test('one list per lesson, selectors return the requested track', async () => {
  const h = harness([draft(trackT), draft(trackU)]);
  await h.observer.refetch();
  expect(h.observer.getCurrentResult().data).toHaveLength(2);
  expect(h.track(trackU)?.trackId).toBe(trackU);
  expect(h.track('absent')).toBeUndefined();
  h.unsubscribe();
  h.client.clear();
});

test('two tracks keep separate markers; changes to U do not acknowledge T', async () => {
  rstest.useFakeTimers();
  const h = harness([draft(trackT), draft(trackU)]);
  await h.observer.refetch();
  await h.regenerate(trackT);
  await h.regenerate(trackU);
  expect(h.calls).toEqual([
    [lessonId, trackT],
    [lessonId, trackU],
  ]);
  expect(h.track(trackT)?.state).toBe('pending');
  expect(h.track(trackU)?.state).toBe('pending');
  h.set([draft(trackT), draft(trackU, { updatedAt: '2026-11-16T12:01:00Z' })]);
  await rstest.advanceTimersByTimeAsync(draftPollMs);
  expect(h.track(trackU)?.state).toBe('ready');
  expect(h.track(trackT)?.state).toBe('pending');
  const before = h.reads();
  await rstest.advanceTimersByTimeAsync(draftPollMs);
  expect(h.reads()).toBeGreaterThan(before);
  h.set([draft(trackT, { state: 'running', updatedAt: '2026-11-16T12:02:00Z' }), draft(trackU)]);
  await rstest.advanceTimersByTimeAsync(draftPollMs);
  expect(h.track(trackT)?.state).toBe('running');
  h.set([draft(trackT, { updatedAt: '2026-11-16T12:03:00Z' }), draft(trackU)]);
  await rstest.advanceTimersByTimeAsync(draftPollMs);
  expect(h.track(trackT)?.state).toBe('ready');
  const done = h.reads();
  await rstest.advanceTimersByTimeAsync(draftPollMs * 2);
  expect(h.reads()).toBe(done);
  h.unsubscribe();
  h.client.clear();
});

test('a stale ready draft has no refetch interval (O7)', async () => {
  rstest.useFakeTimers();
  const h = harness([draft(trackT, { stale: true })]);
  await h.observer.refetch();
  const before = h.reads();
  await rstest.advanceTimersByTimeAsync(draftTaskTimeoutMs + draftPollMs);
  expect(h.reads()).toBe(before);
  expect(h.track(trackT)?.stale).toBe(true);
  h.unsubscribe();
  h.client.clear();
});

test.each(['running', 'pending'] as const)(
  'server %s state polls without a local marker',
  async (state) => {
    rstest.useFakeTimers();
    const h = harness([draft(trackT, { state })]);
    await h.observer.refetch();
    const before = h.reads();
    await rstest.advanceTimersByTimeAsync(draftPollMs);
    expect(h.reads()).toBeGreaterThan(before);
    h.unsubscribe();
    h.client.clear();
  },
);

test('already_running starts only the requested track marker', async () => {
  rstest.useFakeTimers();
  const h = harness([draft(trackT), draft(trackU)], true);
  await h.observer.refetch();
  await expect(h.regenerate(trackT)).rejects.toBeInstanceOf(DraftQueryError);
  expect(h.track(trackT)?.state).toBe('pending');
  expect(h.track(trackU)?.state).toBe('ready');
  h.unsubscribe();
  h.client.clear();
});

test.each([false, true])('queue marker expires at 15 min; existing snapshot=%s', async (exists) => {
  rstest.useFakeTimers();
  const h = harness(exists ? [draft(trackT)] : []);
  await h.observer.refetch();
  await h.regenerate(trackT);
  expect(h.track(trackT)?.state).toBe('pending');
  await rstest.advanceTimersByTimeAsync(draftTaskTimeoutMs);
  expect(h.track(trackT)?.state).toBe(exists ? 'ready' : undefined);
  const done = h.reads();
  await rstest.advanceTimersByTimeAsync(draftPollMs * 2);
  expect(h.reads()).toBe(done);
  h.unsubscribe();
  h.client.clear();
});

test.each([false, true])(
  'marker survives real unobserved cache GC; existing snapshot=%s',
  async (exists) => {
    rstest.useFakeTimers();
    const h = harness(exists ? [draft(trackT)] : []);
    await h.observer.refetch();
    await h.regenerate(trackT);
    h.unsubscribe();
    await rstest.advanceTimersByTimeAsync(10 * 60 * 1000);
    expect(h.client.getQueryState(h.queries.lesson(lessonId, h.client).queryKey)).toBeUndefined();
    const observer = new QueryObserver(h.client, h.queries.lesson(lessonId, h.client));
    const stop = observer.subscribe(() => {});
    await observer.refetch();
    expect(draftOfTrack(observer.getCurrentResult().data, trackT)?.state).toBe('pending');
    await rstest.advanceTimersByTimeAsync(5 * 60 * 1000);
    expect(draftOfTrack(observer.getCurrentResult().data, trackT)?.state).toBe(
      exists ? 'ready' : undefined,
    );
    const done = h.reads();
    await rstest.advanceTimersByTimeAsync(draftPollMs * 2);
    expect(h.reads()).toBe(done);
    stop();
    h.client.clear();
  },
);

test('inventory is per lesson and does not poll', async () => {
  rstest.useFakeTimers();
  const h = harness();
  const options = h.queries.inventory(lessonId);
  const observer = new QueryObserver(h.client, options);
  const stop = observer.subscribe(() => {});
  await observer.refetch();
  expect(observer.getCurrentResult().data).toEqual([]);
  expect(options).not.toHaveProperty('refetchInterval');
  expect(h.queries.inventory('other').queryKey).not.toEqual(options.queryKey);
  stop();
  h.unsubscribe();
  h.client.clear();
});

function isolatedQueries(initial: LessonDraft[] = []) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const repository: DraftRepository = {
    listDrafts: async () => ({ ok: true, value: initial }),
    listTranscripts: async () => ({ ok: true, value: [] }),
    regenerate: async () => ({ ok: true, value: undefined }),
  };
  const queries = createDraftQueries(repository);
  const options = queries.lesson(lessonId, client);
  const read = () => client.fetchQuery(options);
  const regenerate = () =>
    client
      .getMutationCache()
      .build(client, queries.regenerate(client))
      .execute({ lessonId, trackId: trackT });
  const interval = (id = lessonId) => {
    const opts = queries.lesson(id, client);
    const query = client
      .getQueryCache()
      .build<LessonDraft[], Error, LessonDraft[], ReturnType<typeof lessonDraftQueryKey>>(
        client,
        opts,
      );
    if (typeof opts.refetchInterval !== 'function') throw new Error('Missing query');
    return opts.refetchInterval(query);
  };
  return { client, repository, queries, read, regenerate, interval };
}

test('regeneration retains the complete prior snapshot but clears its server timestamp', async () => {
  const previous = draft(trackT, {
    language: 'fr',
    failure: 'recognition_failed',
    stale: true,
    generation: 7,
    segments: [{ id: 'line', startMs: 10, endMs: 300, text: 'Bonjour' }],
  });
  const h = isolatedQueries([previous]);
  await h.read();
  await h.regenerate();
  expect(await h.read()).toEqual([{ ...previous, state: 'pending', updatedAt: null }]);
  h.client.clear();
});

test.each([false, true])(
  'interval expiry restores cache without a fetch; prior snapshot=%s',
  async (exists) => {
    rstest.useFakeTimers();
    rstest.setSystemTime(1000);
    const previous = draft(trackT);
    const h = isolatedQueries(exists ? [previous] : []);
    await h.read();
    await h.regenerate();
    await h.read();
    expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toMatchObject([
      { state: 'pending', updatedAt: null },
    ]);
    rstest.setSystemTime(1000 + draftTaskTimeoutMs);
    expect(h.interval()).toBe(false);
    expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual(
      exists ? [previous] : [],
    );
    h.client.clear();
  },
);

test('markers never leak rows or polling into a sibling lesson', async () => {
  const h = isolatedQueries();
  await h.read();
  await h.regenerate();
  const sibling = h.queries.lesson('sibling', h.client);
  expect(await h.client.fetchQuery(sibling)).toEqual([]);
  expect(h.interval('sibling')).toBe(false);
  h.client.clear();
});

test('running with the same timestamp acknowledges the claim and terminal data stops polling', async () => {
  const h = isolatedQueries([draft(trackT)]);
  await h.read();
  await h.regenerate();
  h.repository.listDrafts = async () => ({
    ok: true,
    value: [draft(trackT, { state: 'running' })],
  });
  expect(await h.read()).toMatchObject([{ state: 'running' }]);
  h.repository.listDrafts = async () => ({ ok: true, value: [draft(trackT)] });
  expect(await h.read()).toEqual([draft(trackT)]);
  expect(h.interval()).toBe(false);
  h.client.clear();
});

test('repeated regeneration keeps the original fifteen-minute queue deadline', async () => {
  rstest.useFakeTimers();
  rstest.setSystemTime(1000);
  const h = isolatedQueries();
  await h.read();
  await h.regenerate();
  await h.read();
  rstest.setSystemTime(1000 + 10 * 60 * 1000);
  await h.regenerate();
  rstest.setSystemTime(1000 + draftTaskTimeoutMs);
  expect(h.interval()).toBe(false);
  expect(h.client.getQueryData(h.queries.lesson(lessonId, h.client).queryKey)).toEqual([]);
  h.client.clear();
});

test.each(['inventory', 'lesson'] as const)('%s errors preserve their named code', async (kind) => {
  const h = isolatedQueries();
  h.repository.listTranscripts = async () => ({
    ok: false,
    error: { code: 'dependency_unavailable' },
  });
  h.repository.listDrafts = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
  const read =
    kind === 'inventory'
      ? h.client.fetchQuery(h.queries.inventory(lessonId))
      : h.client.fetchQuery(h.queries.lesson(lessonId, h.client));
  await expect(read).rejects.toMatchObject({
    name: 'DraftQueryError',
    code: 'dependency_unavailable',
  });
  h.client.clear();
});

test('failed regeneration never creates a pending row or starts polling', async () => {
  const h = isolatedQueries();
  await h.read();
  h.repository.regenerate = async () => ({ ok: false, error: { code: 'dependency_unavailable' } });
  await expect(h.regenerate()).rejects.toMatchObject({ code: 'dependency_unavailable' });
  expect(await h.read()).toEqual([]);
  expect(h.interval()).toBe(false);
  h.client.clear();
});

test('null-track placeholder does not rewrite the cache while a requested marker waits', async () => {
  rstest.useFakeTimers();
  rstest.setSystemTime(1000);
  const h = isolatedQueries([draft(null, { updatedAt: null, assetId: null, generation: null })]);
  await h.read();
  await h.regenerate();
  const key = h.queries.lesson(lessonId, h.client).queryKey;
  const before = h.client.getQueryState(key)?.dataUpdatedAt;
  rstest.setSystemTime(2000);
  expect(h.interval()).toBe(draftPollMs);
  expect(h.client.getQueryState(key)?.dataUpdatedAt).toBe(before);
  expect(h.client.getQueryData(key)).toHaveLength(1);
  h.client.clear();
});
