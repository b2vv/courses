import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { DraftErrorCode, LessonDraft, DraftRepository } from '../../modules/transcript';

export const draftPollMs = 5_000;
// Bound queue waiting before Claim, not recognition; protect against a task lost, e.g. in Redis.
export const draftTaskTimeoutMs = 15 * 60 * 1_000;

type DraftPendingMarker = {
  readonly startedAt: number;
  readonly previous: LessonDraft | undefined;
};

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
export function lessonTranscriptInventoryQueryKey(lessonId: string) {
  return ['admin', 'transcript-inventory', lessonId] as const;
}

export async function invalidateLessonDrafts(client: QueryClient, lessonId?: string) {
  await client.invalidateQueries({
    queryKey: lessonId === undefined ? lessonDraftQueryPrefix : lessonDraftQueryKey(lessonId),
    ...(lessonId === undefined ? {} : { exact: true }),
  });
}

export function draftOfTrack(drafts: readonly LessonDraft[] | undefined, trackId: string) {
  return drafts?.find((draft) => draft.trackId === trackId);
}

function pendingDraft(lessonId: string, trackId: string, previous?: LessonDraft): LessonDraft {
  return {
    lessonId,
    trackId,
    trackLanguage: previous?.trackLanguage ?? 'und',
    trackMultilingual: previous?.trackMultilingual ?? false,
    language: null,
    failure: null,
    stale: false,
    assetId: null,
    generation: null,
    segments: [],
    ...previous,
    state: 'pending',
    updatedAt: null,
  };
}

export function createDraftQueries(drafts: DraftRepository) {
  // Factory lifetime, not query-cache lifetime: unobserved cache entries are GC'd after 5 min,
  // but a requested regeneration must keep its original 15-minute queue window across remounts.
  const pending = new Map<string, DraftPendingMarker>();
  const markerKey = (lessonId: string, trackId: string) => `${lessonId}:${trackId}`;

  function reconcile(lessonId: string, rows: readonly LessonDraft[]): LessonDraft[] {
    const next = [...rows];
    for (const [key, marker] of pending) {
      if (!key.startsWith(`${lessonId}:`)) continue;
      const trackId = key.slice(lessonId.length + 1);
      const index = next.findIndex((draft) => draft.trackId === trackId);
      const draft = next[index];
      if (Date.now() - marker.startedAt >= draftTaskTimeoutMs) {
        pending.delete(key);
        // Remove the local pending overlay; keep the last server snapshot, if any.
        if (draft?.state === 'pending' && draft.updatedAt === null) {
          if (marker.previous === undefined) next.splice(index, 1);
          else next[index] = marker.previous;
        }
      } else if (
        draft !== undefined &&
        draft.updatedAt !== null &&
        (draft.state === 'running' || draft.updatedAt !== marker.previous?.updatedAt)
      ) {
        // Only this track's claim/completion acknowledges this marker; sibling changes do not.
        pending.delete(key);
      } else {
        const overlay = pendingDraft(lessonId, trackId, marker.previous);
        if (index < 0) next.push(overlay);
        else next[index] = overlay;
      }
    }
    return next;
  }

  function markPending(client: QueryClient, lessonId: string, trackId: string) {
    const key = markerKey(lessonId, trackId);
    if (!pending.has(key)) {
      pending.set(key, {
        startedAt: Date.now(),
        previous: draftOfTrack(
          client.getQueryData<LessonDraft[]>(lessonDraftQueryKey(lessonId)),
          trackId,
        ),
      });
    }
  }

  return {
    drafts,
    inventory: (lessonId: string) =>
      queryOptions({
        queryKey: lessonTranscriptInventoryQueryKey(lessonId),
        queryFn: async ({ signal }) => {
          const result = await drafts.listTranscripts(lessonId, signal);
          if (!result.ok) throw new DraftQueryError(result.error.code);
          return result.value;
        },
      }),
    lesson: (lessonId: string, client: QueryClient) =>
      queryOptions({
        queryKey: lessonDraftQueryKey(lessonId),
        queryFn: async ({ signal }) => {
          const result = await drafts.listDrafts(lessonId, signal);
          if (!result.ok) throw new DraftQueryError(result.error.code);
          return reconcile(lessonId, result.value);
        },
        refetchInterval: (query) => {
          const rows = query.state.data;
          const next = reconcile(lessonId, rows ?? []);
          if (rows !== undefined) {
            if (next.some((draft, index) => draft !== rows[index]) || next.length !== rows.length) {
              // Expiry must also remove the local pending UI when no further fetch is scheduled.
              const expiredOverlay = rows.some(
                (draft) =>
                  draft.updatedAt === null &&
                  draft.trackId !== null &&
                  !pending.has(markerKey(lessonId, draft.trackId)),
              );
              if (expiredOverlay) client.setQueryData(lessonDraftQueryKey(lessonId), next);
            }
          }
          const waiting = [...pending.keys()].some((key) => key.startsWith(`${lessonId}:`));
          const running =
            query.state.data?.some(
              (draft) =>
                draft.state === 'running' ||
                (draft.state === 'pending' && draft.updatedAt !== null),
            ) === true;
          return waiting || running ? draftPollMs : false;
        },
      }),
    regenerate: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          trackId,
        }: {
          readonly lessonId: string;
          readonly trackId: string;
        }) => {
          const result = await drafts.regenerate(lessonId, trackId);
          if (!result.ok) throw new DraftQueryError(result.error.code);
        },
        onSuccess: (_value, { lessonId, trackId }) => {
          markPending(client, lessonId, trackId);
          return invalidateLessonDrafts(client, lessonId);
        },
        onError: (error, { lessonId, trackId }) => {
          if (error instanceof DraftQueryError && error.code === 'already_running') {
            markPending(client, lessonId, trackId);
            return invalidateLessonDrafts(client, lessonId);
          }
        },
      }),
  };
}
