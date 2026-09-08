import { queryOptions } from '@tanstack/react-query';
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

export function createLearnQueries(playback: PlaybackRepository) {
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
  };
}
