import { infiniteQueryOptions, mutationOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { TeachRepository } from '../../ports/TeachRepository';
import type {
  DismissReportsInput,
  LearnerFilters,
  LearnerScope,
  TeachErrorCode,
  TeachResult,
} from '../../domain/teach/teach';

export function teachCourseCommentsQueryKey(courseId: string) {
  return ['teach', 'comments', courseId] as const;
}
export function teachModerationQueueQueryKey(courseId: string) {
  return ['teach', 'moderation', courseId] as const;
}
export function teachLearnersQueryKey(scope: LearnerScope, filters: LearnerFilters) {
  return ['teach', 'learners', scope, filters] as const;
}
export class TeachQueryError extends Error {
  readonly code: TeachErrorCode;
  constructor(code: TeachErrorCode) {
    super(code);
    this.name = 'TeachQueryError';
    this.code = code;
  }
}
function unwrap<T>(result: TeachResult<T>): T {
  if (!result.ok) throw new TeachQueryError(result.error.code);
  return result.value;
}
export function createTeachQueries(teach: TeachRepository) {
  const invalidateWorkspace = async (client: QueryClient) => {
    await client.invalidateQueries({ queryKey: ['teach', 'moderation'] });
    await client.invalidateQueries({ queryKey: ['teach', 'comments'] });
    await client.invalidateQueries({ queryKey: ['learn', 'discussion'] });
  };
  return {
    teach,
    learners: (scope: LearnerScope, filters: LearnerFilters) =>
      infiniteQueryOptions({
        queryKey: teachLearnersQueryKey(scope, filters),
        initialPageParam: undefined as string | undefined,
        queryFn: ({ signal, pageParam }) =>
          teach.listLearners(scope, filters, pageParam, signal).then(unwrap),
        getNextPageParam: (page) => page.nextCursor ?? undefined,
      }),
    courseComments: (courseId: string) =>
      infiniteQueryOptions({
        queryKey: teachCourseCommentsQueryKey(courseId),
        initialPageParam: undefined as string | undefined,
        queryFn: ({ signal, pageParam }) =>
          teach.listCourseComments(courseId, pageParam, signal).then(unwrap),
        getNextPageParam: (page) => page.nextCursor ?? undefined,
      }),
    moderationQueue: (courseId: string) =>
      infiniteQueryOptions({
        queryKey: teachModerationQueueQueryKey(courseId),
        initialPageParam: undefined as string | undefined,
        queryFn: ({ signal, pageParam }) =>
          teach.listModerationQueue(courseId, pageParam, signal).then(unwrap),
        getNextPageParam: (page) => page.nextCursor ?? undefined,
      }),
    dismiss: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ commentId, input }: { commentId: string; input: DismissReportsInput }) =>
          teach.dismissReport(commentId, input).then(unwrap),
        onSuccess: () => invalidateWorkspace(client),
      }),
    remove: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ commentId }: { lessonId: string; commentId: string }) =>
          teach.removeComment(commentId).then(unwrap),
        onSuccess: () => invalidateWorkspace(client),
      }),
    dismissReview: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          reviewId,
          input,
        }: {
          reviewId: string;
          input: DismissReportsInput;
        }) => {
          const result = await teach.dismissReviewReport(reviewId, input);
          if (result === undefined) throw new TeachQueryError('dependency_unavailable');
          return unwrap(result);
        },
        onSuccess: () => invalidateWorkspace(client),
      }),
    removeReview: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          reviewId,
          input,
        }: {
          reviewId: string;
          input: DismissReportsInput;
        }) => {
          const result = await teach.removeReview(reviewId, input);
          if (result === undefined) throw new TeachQueryError('dependency_unavailable');
          return unwrap(result);
        },
        onSuccess: () => invalidateWorkspace(client),
      }),
  };
}
