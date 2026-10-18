export type DraftState = 'pending' | 'running' | 'ready' | 'failed' | 'no_speech';

export type LessonDraft = {
  readonly lessonId: string;
  readonly state: DraftState | null;
  readonly language: string | null;
  readonly failure: string | null;
  readonly stale: boolean;
  readonly updatedAt: string | null;
  readonly assetId: string | null;
  readonly generation: number | null;
  readonly segments: readonly DraftSegment[];
};

export type DraftSegment = {
  readonly id: string;
  readonly startMs: number;
  readonly endMs: number;
  readonly text: string;
  readonly speaker?: string;
  readonly confidence?: number;
};

export type DraftErrorCode =
  | 'authentication_required'
  | 'origin_forbidden'
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'asset_not_found'
  | 'dependency_unavailable'
  | 'already_running'
  | 'invalid_request'
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response';

export type DraftResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: DraftErrorCode } };
