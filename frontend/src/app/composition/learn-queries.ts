import { infiniteQueryOptions, mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { CourseRepository } from '../../modules/catalog';
import type { PlaybackErrorCode } from '../../modules/playback';
import type { PlaybackRepository } from '../../modules/playback';

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
/**
 * One-use tokens for the «continue» redirect of `/learn/<slug>` (C3b). The redirect issues one
 * and puts it in history state; the lesson it leads to takes it, and only that first take lets
 * the lesson reuse the course read the redirect just made. History keeps the token, but a Back
 * or Forward to that entry finds it already taken, so it reads again (Codex delta review). Held
 * per router (request scoped), in memory only; at most a few are ever outstanding.
 */
export function createContinueReads() {
  const issued = new Set<string>();
  return {
    issue(): string {
      const token = crypto.randomUUID();
      issued.add(token);
      // A redirect whose lesson never loads leaves its token behind; keep the set small.
      if (issued.size > 8) {
        const oldest = issued.values().next().value;
        if (oldest !== undefined) issued.delete(oldest);
      }
      return token;
    },
    take(token: string | undefined): boolean {
      return token !== undefined && issued.delete(token);
    },
  };
}

/**
 * Leaving a lesson (`nav`) or finishing it (`ended`) is what moves completion, «continue» and
 * review eligibility. With lesson links moving inside the app (C3b) the next lesson is already
 * mounted and read its course before that last heartbeat landed, so both are refreshed; other
 * heartbeats change only the position, which the next entry reads afresh.
 */
export function refreshAfterHeartbeat(
  client: QueryClient,
  { slug, courseId, reason }: { slug: string; courseId: string; reason: string },
): Promise<unknown> {
  if (reason !== 'nav' && reason !== 'ended') return Promise.resolve();
  return Promise.all([
    client.invalidateQueries({ queryKey: learningCourseQueryKey(slug), exact: true }),
    client.invalidateQueries({ queryKey: ['learn', 'review', courseId] }),
  ]);
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
    continueReads: createContinueReads(),
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
