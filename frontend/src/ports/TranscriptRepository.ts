import type { LessonTranscript, TranscriptResult } from '../domain/transcript/transcript';

export interface TranscriptRepository {
  getLessonTranscript(
    lessonId: string,
    language?: string,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<LessonTranscript>>;
}
