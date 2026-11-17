import type { DraftResult, LessonDraft, TranscriptInventoryItem } from '../domain/draft';

export interface DraftRepository {
  listDrafts(lessonId: string, signal?: AbortSignal): Promise<DraftResult<LessonDraft[]>>;
  listTranscripts(
    lessonId: string,
    signal?: AbortSignal,
  ): Promise<DraftResult<TranscriptInventoryItem[]>>;
  regenerate(lessonId: string, trackId: string, signal?: AbortSignal): Promise<DraftResult<void>>;
}
