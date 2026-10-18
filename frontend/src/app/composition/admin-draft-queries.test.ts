import { afterEach, describe, expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import {
  DraftQueryError,
  createDraftQueries,
  draftPollMs,
  draftTaskTimeoutMs,
} from './admin-draft-queries';
import type { LessonDraft } from '../../modules/transcript';
import type { DraftRepository } from '../../modules/transcript';

const lessonId = '20000000-0000-4000-8000-000000000001';
const otherLessonId = '20000000-0000-4000-8000-000000000002';

function draft(lesson: string, state: LessonDraft['state']): LessonDraft {
  return {
    lessonId: lesson,
    state,
    language: null,
    failure: null,
    stale: false,
    updatedAt: state === null ? null : '2026-09-24T12:00:00.000Z',
    assetId: null,
    generation: null,
    segments: [],
  };
}

function draftWithUpdatedAt(
  lesson: string,
  state: LessonDraft['state'],
  updatedAt: string,
): LessonDraft {
  return { ...draft(lesson, state), updatedAt };
}

function repository(
  getDraft: DraftRepository['getDraft'],
  regenerate: DraftRepository['regenerate'] = async () => ({ ok: true, value: undefined }),
): DraftRepository {
  return { getDraft, regenerate };
}

async function observeInitial(
  client: QueryClient,
  queries: ReturnType<typeof createDraftQueries>,
  id = lessonId,
) {
  const observer = new QueryObserver(client, queries.lesson(id, client));
  const unsubscribe = observer.subscribe(() => {});
  await observer.refetch();
  return { observer, unsubscribe };
}

afterEach(() => {
  rstest.useRealTimers();
});

describe('admin draft query boundary', () => {
  test('keeps polling and exposes pending after a successful regenerate until ready', async () => {
    rstest.useFakeTimers();
    let reads = 0;
    const queries = createDraftQueries(
      repository(async (id) => {
        reads++;
        return {
          ok: true,
          value:
            reads === 1
              ? draftWithUpdatedAt(id, 'ready', '2026-09-24T12:00:00.000Z')
              : reads === 2
                ? draftWithUpdatedAt(id, 'ready', '2026-09-24T12:00:00.000Z')
                : reads === 3
                  ? draftWithUpdatedAt(id, 'running', '2026-09-24T12:00:00.500Z')
                  : draftWithUpdatedAt(id, 'ready', '2026-09-24T12:00:01.000Z'),
        };
      }),
    );
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { observer, unsubscribe } = await observeInitial(client, queries);

    await client
      .getMutationCache()
      .build(client, queries.regenerate(client))
      .execute({ lessonId, language: undefined });

    expect(observer.getCurrentResult().data?.state).toBe('pending');
    expect(reads).toBe(2);

    await rstest.advanceTimersByTimeAsync(draftPollMs);
    expect(observer.getCurrentResult().data?.state).toBe('running');
    await rstest.advanceTimersByTimeAsync(draftPollMs);
    expect(observer.getCurrentResult().data?.state).toBe('ready');
    expect(observer.getCurrentResult().data?.updatedAt).toBe('2026-09-24T12:00:01.000Z');
    const readsAtReady = reads;
    await rstest.advanceTimersByTimeAsync(draftPollMs * 2);
    expect(reads).toBe(readsAtReady);
    unsubscribe();
  });

  test('already_running also starts polling for only that lesson', async () => {
    rstest.useFakeTimers();
    let reads = 0;
    const queries = createDraftQueries(
      repository(
        async (id) => {
          reads++;
          return { ok: true, value: draft(id, null) };
        },
        async () => ({ ok: false, error: { code: 'already_running' } }),
      ),
    );
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const first = await observeInitial(client, queries);
    const other = await observeInitial(client, queries, otherLessonId);
    await expect(
      client
        .getMutationCache()
        .build(client, queries.regenerate(client))
        .execute({ lessonId, language: undefined }),
    ).rejects.toBeInstanceOf(DraftQueryError);

    expect(first.observer.getCurrentResult().data?.state).toBe('pending');
    expect(other.observer.getCurrentResult().data?.state).toBeNull();
    const readsBeforePoll = reads;
    await rstest.advanceTimersByTimeAsync(draftPollMs);
    expect(reads).toBe(readsBeforePoll + 1);
    first.unsubscribe();
    other.unsubscribe();
  });

  test('stops polling and returns ordinary null after the 15-minute queue timeout', async () => {
    rstest.useFakeTimers();
    let reads = 0;
    const queries = createDraftQueries(
      repository(async (id) => {
        reads++;
        return { ok: true, value: draft(id, null) };
      }),
    );
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { observer, unsubscribe } = await observeInitial(client, queries);
    await client
      .getMutationCache()
      .build(client, queries.regenerate(client))
      .execute({ lessonId, language: undefined });
    expect(observer.getCurrentResult().data?.state).toBe('pending');

    await rstest.advanceTimersByTimeAsync(draftTaskTimeoutMs);
    expect(observer.getCurrentResult().data?.state).toBeNull();
    const readsAtTimeout = reads;
    await rstest.advanceTimersByTimeAsync(draftPollMs * 2);
    expect(reads).toBe(readsAtTimeout);
    unsubscribe();
  });
});
