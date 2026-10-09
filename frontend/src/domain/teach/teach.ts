export type TeachErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'course_forbidden'
  | 'course_not_found'
  | 'learner_course_forbidden'
  | 'comment_not_found'
  | 'comment_forbidden'
  | 'review_not_found'
  | 'review_moderation_forbidden'
  // Author block/unblock from the queue (shared with the roster's transition, Lo17).
  | 'enrollment_forbidden'
  | 'grant_transition_invalid'
  | 'membership_conflict';

export type TeachResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: TeachErrorCode } };

export type TeachComment = {
  readonly id: string;
  readonly lessonId: string;
  readonly lessonTitle: string;
  readonly parentId: string | null;
  readonly author: { readonly displayName: string | null };
  readonly body: string | null;
  readonly positionMs: number | null;
  readonly createdAt: string;
  readonly editedAt: string | null;
  readonly removed: boolean;
};

export type TeachCommentPage = {
  readonly comments: readonly TeachComment[];
  readonly nextCursor: string | null;
};

export type ModerationQueueItem = {
  readonly kind: 'comment' | 'review';
  readonly comment: TeachComment | null;
  readonly review: {
    readonly id: string;
    readonly authorDisplayName: string | null;
    readonly stars: number;
    readonly body: string;
    readonly completedLessons: number;
    readonly totalLessons: number;
    readonly createdAt: string;
    readonly editedAt: string | null;
  } | null;
  readonly openReports: number;
  // Whether the item's author is blocked on the course, so the control reads Block / Unblock
  // (Lo14: state only — the author's id is never in the response).
  readonly authorBlocked: boolean;
};

export type ModerationQueuePage = {
  readonly items: readonly ModerationQueueItem[];
  readonly nextCursor: string | null;
};

export type DismissReportsInput = { readonly reason: string };

// Learner overview (teach + admin). No `email`: the field is never carried into the domain, so the
// UI cannot leak it (Lo07). A deleted account arrives with `displayName === ''` and is shown
// anonymously, exactly as the roster handles it.
export type LearnerCourseState = 'active' | 'blocked';

export type LearnerCourse = {
  readonly courseId: string;
  readonly slug: string;
  readonly title: string;
  readonly state: LearnerCourseState;
  readonly progressPercent: number;
  readonly enrolledAt: string;
};

export type LearnerRow = {
  readonly userId: string;
  readonly displayName: string;
  readonly lastActivityAt: string | null;
  readonly courses: readonly LearnerCourse[];
};

export type LearnerPage = {
  readonly entries: readonly LearnerRow[];
  readonly nextCursor: string | null;
};

// `teach` lists only the actor's own courses; `admin` lists every course. The scope picks the
// endpoint, not a client-side filter — the server decides visibility.
export type LearnerScope = 'teach' | 'admin';

export type LearnerFilters = {
  readonly courseId?: string;
  readonly state?: LearnerCourseState;
  readonly q?: string;
};
