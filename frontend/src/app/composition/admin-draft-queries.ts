import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { DraftErrorCode, LessonDraft } from '../../modules/transcript';
import type { DraftRepository } from '../../modules/transcript';

export const draftPollMs = 5_000;
// Bound queue waiting before Claim, not recognition; protect against a task lost, e.g. in Redis.
export const draftTaskTimeoutMs = 15 * 60 * 1_000;

type DraftPendingMarker = { readonly startedAt: number; readonly updatedAt: string | null };

export class DraftQueryError extends Error {
  readonly code: DraftErrorCode;
  constructor(code: DraftErrorCode) {
    super(code);
    this.name = 'DraftQueryError';
    this.code = code;
  }
}

const lessonDraftQueryPrefix = ['admin', 'draft'] as const;

export function lessonDraftQueryKey(lessonId: string) {
  return [...lessonDraftQueryPrefix, lessonId] as const;
}

/**
 * Re-reads every lesson draft in the cache; only mounted ones refetch. A default audio-track change
 * keeps the video generation, so the generation-based invalidation cannot see it (multi-audio S6).
 */
export async function invalidateLessonDrafts(client: QueryClient) {
  await client.invalidateQueries({ queryKey: lessonDraftQueryPrefix });
}

/**
 * A stale draft is a pending recognition, so the server replaces it without the author asking
 * (multi-audio S6). Wait for that within the same window as a requested regeneration: if the
 * worker makes no progress, polling stops and the stale banner and Regenerate remain. The start
 * of the wait lives with the query factory, not in the query cache, where an unobserved entry is
 * garbage-collected after five minutes and would restart the window.
 */
function awaitingReplacement(
  staleSince: Map<string, number>,
  lessonId: string,
  stale: boolean,
): boolean {
  if (!stale) {
    staleSince.delete(lessonId);
    return false;
  }
  const since = staleSince.get(lessonId);
  if (since === undefined) {
    staleSince.set(lessonId, Date.now());
    return true;
  }
  return Date.now() - since < draftTaskTimeoutMs;
}

function lessonDraftPendingKey(lessonId: string) {
  return ['admin', 'draft-pending', lessonId] as const;
}

function pendingSince(client: QueryClient, lessonId: string): number | undefined {
  const marker = client.getQueryData<DraftPendingMarker>(lessonDraftPendingKey(lessonId));
  if (marker === undefined) return undefined;
  if (Date.now() - marker.startedAt >= draftTaskTimeoutMs) {
    client.removeQueries({ queryKey: lessonDraftPendingKey(lessonId), exact: true });
    return undefined;
  }
  return marker.startedAt;
}

function pendingRegeneration(client: QueryClient, lessonId: string): boolean {
  const marker = client.getQueryData<DraftPendingMarker>(lessonDraftPendingKey(lessonId));
  if (marker === undefined) return false;
  if (Date.now() - marker.startedAt >= draftTaskTimeoutMs) {
    client.removeQueries({ queryKey: lessonDraftPendingKey(lessonId), exact: true });
    return false;
  }
  return true;
}

function pendingDraft(lessonId: string): LessonDraft {
  return {
    lessonId,
    state: 'pending',
    language: null,
    failure: null,
    stale: false,
    updatedAt: null,
    assetId: null,
    generation: null,
    segments: [],
  };
}

function markPending(client: QueryClient, lessonId: string) {
  const cachedDraft = client.getQueryData<LessonDraft>(lessonDraftQueryKey(lessonId));
  client.setQueryData<DraftPendingMarker>(lessonDraftPendingKey(lessonId), {
    startedAt: Date.now(),
    updatedAt: cachedDraft?.updatedAt ?? null,
  });
}

function clearPending(client: QueryClient, lessonId: string) {
  client.removeQueries({ queryKey: lessonDraftPendingKey(lessonId), exact: true });
}

export function createDraftQueries(drafts: DraftRepository) {
  const staleSince = new Map<string, number>();
  return {
    drafts,
    lesson: (lessonId: string, client: QueryClient) =>
      queryOptions({
        queryKey: lessonDraftQueryKey(lessonId),
        queryFn: async ({ signal }) => {
          const result = await drafts.getDraft(lessonId, signal);
          if (!result.ok) throw new DraftQueryError(result.error.code);
          const marker = client.getQueryData<DraftPendingMarker>(lessonDraftPendingKey(lessonId));
          if (pendingRegeneration(client, lessonId)) {
            const oldDraftStillVisible =
              marker?.updatedAt !== null && marker?.updatedAt === result.value.updatedAt;
            if (result.value.state === null || oldDraftStillVisible) return pendingDraft(lessonId);
          }
          if (result.value.state === null && pendingSince(client, lessonId) !== undefined) {
            return pendingDraft(lessonId);
          }
          if (
            result.value.state === 'ready' ||
            result.value.state === 'no_speech' ||
            result.value.state === 'failed'
          ) {
            clearPending(client, lessonId);
          }
          return result.value;
        },
        refetchInterval: (query) => {
          const waitingForTask = pendingSince(client, lessonId) !== undefined;
          if (
            !waitingForTask &&
            query.state.data?.state === 'pending' &&
            query.state.data.updatedAt === null
          ) {
            client.setQueryData<LessonDraft>(lessonDraftQueryKey(lessonId), {
              ...query.state.data,
              state: null,
            });
          }
          const serverIsRunning =
            query.state.data?.state === 'running' ||
            (query.state.data?.state === 'pending' && query.state.data.updatedAt !== null);
          const replacing = awaitingReplacement(
            staleSince,
            lessonId,
            query.state.data?.stale === true,
          );
          return waitingForTask || serverIsRunning || replacing ? draftPollMs : false;
        },
      }),
    regenerate: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          language,
        }: {
          readonly lessonId: string;
          readonly language: string | undefined;
        }) => {
          const result = await drafts.regenerate(lessonId, language);
          if (!result.ok) throw new DraftQueryError(result.error.code);
        },
        onSuccess: (_value, { lessonId }) => {
          markPending(client, lessonId);
          return client.invalidateQueries({ queryKey: lessonDraftQueryKey(lessonId) });
        },
        onError: (error, { lessonId }) => {
          if (error instanceof DraftQueryError && error.code === 'already_running') {
            markPending(client, lessonId);
            return client.invalidateQueries({ queryKey: lessonDraftQueryKey(lessonId) });
          }
        },
      }),
  };
}
