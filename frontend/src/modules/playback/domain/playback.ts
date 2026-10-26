export type LearningVideoState = 'processing' | 'ready' | 'failed' | null;

export type LearningLesson = {
  readonly id: string;
  readonly title: string;
  readonly videoState: LearningVideoState;
  readonly position: number;
  readonly completed: boolean;
  readonly clientSeq: number;
  /** Open to anyone who can see the published course, without a purchase. */
  readonly isPreview: boolean;
  /** For this viewer the lesson needs a purchase — decided by the server, never re-derived here. */
  readonly locked: boolean;
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
  /** Null for a free course; the amount a locked lesson asks for. */
  readonly price: { readonly minor: number; readonly currency: 'UAH' } | null;
  readonly modules: readonly LearningModule[];
  readonly continueLessonId: string | null;
  readonly canPurchase?: boolean;
};

export type PlaybackSession = {
  readonly playbackId: string;
  readonly mediaPath: string;
  readonly clientSeq: number;
  readonly heartbeatSeconds: number;
  readonly ownershipTtlSeconds: number;
};

export type PlaybackHeartbeatReason = 'tick' | 'pause' | 'seek' | 'hide' | 'nav' | 'ended';

export type Av1ProbeResult = 'supported' | 'unsupported' | 'probe_failed';

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
  | 'course_purchase_required'
  | 'email_not_verified'
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

/**
 * The lesson immediately after `lessonId` in the outline, or null when it is the last one (or the id
 * is not in the course). Order follows the module/lesson arrays exactly as the outline renders them,
 * so "next" is always the row the student sees below the current one — never a reordering of their
 * own. Pure and total; the caller decides whether that neighbour is a sensible autoplay target (an
 * unpurchased or not-yet-ready lesson is not), because `locked`/`videoState` are read there already.
 */
export function nextLesson(course: LearningCourse, lessonId: string): LearningLesson | null {
  const ordered = course.modules.flatMap((module) => module.lessons);
  const index = ordered.findIndex((candidate) => candidate.id === lessonId);
  if (index === -1 || index + 1 >= ordered.length) return null;
  return ordered[index + 1] ?? null;
}
