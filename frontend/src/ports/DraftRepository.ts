import type { DraftResult, LessonDraft } from '../domain/transcript/draft';

export interface DraftRepository {
  getDraft(lessonId: string, signal?: AbortSignal): Promise<DraftResult<LessonDraft>>;
  regenerate(
    lessonId: string,
    language: string | undefined,
    signal?: AbortSignal,
  ): Promise<DraftResult<void>>;
}
