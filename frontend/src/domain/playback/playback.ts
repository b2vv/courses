export type LearningVideoState = 'processing' | 'ready' | 'failed' | null;

export type LearningLesson = {
  readonly id: string;
  readonly title: string;
  readonly videoState: LearningVideoState;
  readonly position: number;
  readonly completed: boolean;
  readonly clientSeq: number;
};

export type LearningModule = {
  readonly id: string;
  readonly title: string;
  readonly lessons: readonly LearningLesson[];
};

export type LearningCourse = {
  readonly id: string;
  readonly slug: string;
  readonly title: string;
  readonly summary: string;
  readonly description: string;
  readonly status: 'draft' | 'published' | 'archived';
  readonly modules: readonly LearningModule[];
  readonly continueLessonId: string | null;
};

export type PlaybackSession = {
  readonly playbackId: string;
  readonly mediaPath: string;
  readonly clientSeq: number;
  readonly heartbeatSeconds: number;
  readonly ownershipTtlSeconds: number;
};

export type PlaybackHeartbeatReason = 'tick' | 'pause' | 'seek' | 'hide' | 'nav' | 'ended';

export type PlaybackHeartbeat = {
  readonly playbackId: string;
  readonly lessonId: string;
  readonly positionSeconds: number;
  readonly clientSeq: number;
  readonly reason: PlaybackHeartbeatReason;
};

export type PlaybackErrorCode =
  | 'authentication_required'
  | 'course_forbidden'
  | 'course_not_found'
  | 'invalid_request'
  | 'invalid_response'
  | 'invalid_slug'
  | 'lesson_not_found'
  | 'playback_forbidden'
  | 'playback_session_expired'
  | 'playback_taken_over'
  | 'request_cancelled'
  | 'request_timeout'
  | 'dependency_unavailable'
  | 'video_not_ready';

export type PlaybackResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: PlaybackErrorCode } };

export function lessonInCourse(
  course: LearningCourse,
  lessonId: string,
): PlaybackResult<LearningLesson> {
  for (const module of course.modules) {
    const lesson = module.lessons.find((candidate) => candidate.id === lessonId);
    if (lesson !== undefined) return { ok: true, value: lesson };
  }
  return { ok: false, error: { code: 'lesson_not_found' } };
}
