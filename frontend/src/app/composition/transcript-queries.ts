import { queryOptions } from '@tanstack/react-query';
import type { TranscriptErrorCode } from '../../domain/transcript/transcript';
import type { TranscriptRepository } from '../../ports/TranscriptRepository';

export class TranscriptQueryError extends Error {
  readonly code: TranscriptErrorCode;
  constructor(code: TranscriptErrorCode) {
    super(code);
    this.name = 'TranscriptQueryError';
    this.code = code;
  }
}

export function lessonTranscriptQueryKey(lessonId: string, language?: string) {
  return ['learn', 'transcript', lessonId, language ?? null] as const;
}

export function createTranscriptQueries(transcript: TranscriptRepository) {
  return {
    transcript,
    lesson: (lessonId: string, language?: string) =>
      queryOptions({
        queryKey: lessonTranscriptQueryKey(lessonId, language),
        queryFn: async ({ signal }) => {
          const result = await transcript.getLessonTranscript(lessonId, language, signal);
          if (!result.ok) throw new TranscriptQueryError(result.error.code);
          return result.value;
        },
      }),
  };
}
