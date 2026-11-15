import type { DraftResult, LessonDraft } from '../domain/draft';

export interface DraftRepository {
  getDraft(lessonId: string, signal?: AbortSignal): Promise<DraftResult<LessonDraft>>;
  regenerate(lessonId: string, trackId: string, signal?: AbortSignal): Promise<DraftResult<void>>;
}
