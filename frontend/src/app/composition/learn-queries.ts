import { infiniteQueryOptions, mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { CourseRepository } from '../../ports/CourseRepository';
import type { PlaybackErrorCode } from '../../domain/playback/playback';
import type { PlaybackRepository } from '../../ports/PlaybackRepository';

export class LearnQueryError extends Error {
  readonly code: PlaybackErrorCode;
  constructor(code: PlaybackErrorCode) {
    super(code);
    this.name = 'LearnQueryError';
    this.code = code;
  }
}

export function learningCourseQueryKey(slug: string) {
  return ['learn', 'course', slug] as const;
}
export function ownReviewQueryKey(courseId: string, userId: string) {
  return ['learn', 'review', courseId, userId] as const;
}
export function learningReviewsQueryKey(courseId: string) {
  return ['learn', 'reviews', courseId] as const;
}

export function createLearnQueries(playback: PlaybackRepository, courses: CourseRepository) {
  return {
    playback,
    course: (slug: string) =>
      queryOptions({
        queryKey: learningCourseQueryKey(slug),
        queryFn: async ({ signal }) => {
          const result = await playback.getCourse(slug, signal);
          if (!result.ok) throw new LearnQueryError(result.error.code);
          return result.value;
        },
      }),
    ownReview: (courseId: string, userId: string) =>
      queryOptions({
        queryKey: ownReviewQueryKey(courseId, userId),
        queryFn: async ({ signal }) => {
          const result = await courses.ownReview(courseId, signal);
          if (result === undefined || !result.ok)
            throw new LearnQueryError(
              result === undefined
                ? 'dependency_unavailable'
                : (result.error.code as PlaybackErrorCode),
            );
          return result.value;
        },
      }),
    reviews: (courseId: string) =>
      infiniteQueryOptions({
        queryKey: learningReviewsQueryKey(courseId),
        initialPageParam: undefined as string | undefined,
        queryFn: async ({ signal, pageParam }) => {
          const result = await courses.listLearningReviews(courseId, pageParam, signal);
          if (result === undefined || !result.ok)
            throw new LearnQueryError(
              result === undefined
                ? 'dependency_unavailable'
                : (result.error.code as PlaybackErrorCode),
            );
          return result.value;
        },
        getNextPageParam: (page) => page.nextCursor ?? undefined,
      }),
    putReview: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          courseId,
          stars,
          body,
        }: {
          courseId: string;
          stars: number;
          body: string;
        }) => {
          const result = await courses.putReview(courseId, { stars, body });
          if (result === undefined || !result.ok)
            throw new LearnQueryError(
              result === undefined
                ? 'dependency_unavailable'
                : (result.error.code as PlaybackErrorCode),
            );
          return result.value;
        },
        onSuccess: (_, variables) =>
          client.invalidateQueries({ queryKey: ['learn', 'review', variables.courseId] }),
      }),
    deleteReview: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async (courseId: string) => {
          const result = await courses.deleteReview(courseId);
          if (result === undefined || !result.ok)
            throw new LearnQueryError(
              result === undefined
                ? 'dependency_unavailable'
                : (result.error.code as PlaybackErrorCode),
            );
        },
        onSuccess: (_, courseId) =>
          client.invalidateQueries({ queryKey: ['learn', 'review', courseId] }),
      }),
  };
}
