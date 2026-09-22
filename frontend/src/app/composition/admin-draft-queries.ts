import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { DraftErrorCode } from '../../domain/transcript/draft';
import type { DraftRepository } from '../../ports/DraftRepository';

export const draftPollMs = 5_000;

export class DraftQueryError extends Error {
  readonly code: DraftErrorCode;
  constructor(code: DraftErrorCode) {
    super(code);
    this.name = 'DraftQueryError';
    this.code = code;
  }
}

export function lessonDraftQueryKey(lessonId: string) {
  return ['admin', 'draft', lessonId] as const;
}

export function createDraftQueries(drafts: DraftRepository) {
  return {
    drafts,
    lesson: (lessonId: string) =>
      queryOptions({
        queryKey: lessonDraftQueryKey(lessonId),
        queryFn: async ({ signal }) => {
          const result = await drafts.getDraft(lessonId, signal);
          if (!result.ok) throw new DraftQueryError(result.error.code);
          return result.value;
        },
        refetchInterval: (query) =>
          query.state.data?.state === 'running' || query.state.data?.state === 'pending'
            ? draftPollMs
            : false,
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
        onSuccess: (_value, { lessonId }) =>
          client.invalidateQueries({ queryKey: lessonDraftQueryKey(lessonId) }),
      }),
  };
}
