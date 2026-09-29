export type TeachErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'course_forbidden'
  | 'course_not_found'
  | 'comment_not_found'
  | 'comment_forbidden';

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
  readonly comment: TeachComment;
  readonly openReports: number;
};

export type ModerationQueuePage = {
  readonly items: readonly ModerationQueueItem[];
  readonly nextCursor: string | null;
};

export type DismissReportsInput = { readonly reason: string };
