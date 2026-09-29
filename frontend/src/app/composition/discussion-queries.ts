import { infiniteQueryOptions, mutationOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type {
  DiscussionErrorCode,
  DiscussionResult,
  PublishDiscussionInput,
  UpdateDiscussionInput,
} from '../../domain/discussion/discussion';
import type { DiscussionRepository } from '../../ports/DiscussionRepository';

export function lessonDiscussionQueryKey(lessonId: string) {
  return ['learn', 'discussion', lessonId] as const;
}

export class DiscussionQueryError extends Error {
  readonly code: DiscussionErrorCode;
  constructor(code: DiscussionErrorCode) {
    super(code);
    this.name = 'DiscussionQueryError';
    this.code = code;
  }
}

function unwrap<T>(result: DiscussionResult<T>): T {
  if (!result.ok) throw new DiscussionQueryError(result.error.code);
  return result.value;
}

async function invalidate(client: QueryClient, lessonId: string) {
  await client.invalidateQueries({ queryKey: lessonDiscussionQueryKey(lessonId) });
}

export function createDiscussionQueries(discussion: DiscussionRepository) {
  return {
    discussion,
    lesson: (lessonId: string) =>
      infiniteQueryOptions({
        queryKey: lessonDiscussionQueryKey(lessonId),
        initialPageParam: undefined as string | undefined,
        queryFn: async ({ signal, pageParam }) =>
          unwrap(await discussion.list(lessonId, pageParam, signal)),
        getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
      }),
    publish: (_client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          input,
        }: {
          lessonId: string;
          input: PublishDiscussionInput;
        }) => unwrap(await discussion.publish(lessonId, input)),
        onSuccess: (_comment, { lessonId }) => invalidate(_client, lessonId),
      }),
    update: (_client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ input }: { lessonId: string; input: UpdateDiscussionInput }) =>
          unwrap(await discussion.update(input)),
        onSuccess: (_comment, { lessonId }) => invalidate(_client, lessonId),
      }),
    remove: (_client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ commentId }: { lessonId: string; commentId: string }) =>
          unwrap(await discussion.remove(commentId)),
        onSuccess: (_value, { lessonId }) => invalidate(_client, lessonId),
      }),
    report: (_client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ commentId }: { lessonId: string; commentId: string }) =>
          unwrap(
            discussion.report === undefined
              ? ({ ok: false, error: { code: 'dependency_unavailable' as const } } as const)
              : await discussion.report(commentId),
          ),
        onSuccess: (_value, { lessonId }) => invalidate(_client, lessonId),
      }),
  };
}
