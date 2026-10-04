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
function invalidateReviews(client: QueryClient, courseId: string) {
  // A review changes the own review, the learning list and the rating summary shown on the learning course,
  // the catalog card and the course page. The mutation knows only the course id, so the slug-keyed
  // summaries are refreshed by prefix.
  return Promise.all([
    client.invalidateQueries({ queryKey: ['learn', 'review', courseId] }),
    client.invalidateQueries({ queryKey: learningReviewsQueryKey(courseId) }),
    client.invalidateQueries({ queryKey: ['learn', 'course'] }),
    client.invalidateQueries({ queryKey: ['catalog'] }),
    client.invalidateQueries({ queryKey: ['course'] }),
  ]);
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
          if (!result.ok) throw new LearnQueryError(result.error.code as PlaybackErrorCode);
          return result.value;
        },
      }),
    reviews: (courseId: string) =>
      infiniteQueryOptions({
        queryKey: learningReviewsQueryKey(courseId),
        initialPageParam: undefined as string | undefined,
        queryFn: async ({ signal, pageParam }) => {
          const result = await courses.listLearningReviews(courseId, pageParam, signal);
          if (!result.ok) throw new LearnQueryError(result.error.code as PlaybackErrorCode);
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
          if (!result.ok) throw new LearnQueryError(result.error.code as PlaybackErrorCode);
          return result.value;
        },
        // Both the own review and the learning-scoped list change: invalidating only the own review left
        // the list showing the old or deleted review.
        onSuccess: (_, variables) => invalidateReviews(client, variables.courseId),
      }),
    deleteReview: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async (courseId: string) => {
          const result = await courses.deleteReview(courseId);
          if (!result.ok) throw new LearnQueryError(result.error.code as PlaybackErrorCode);
        },
        onSuccess: (_, courseId) => invalidateReviews(client, courseId),
      }),
  };
}
