export type DiscussionErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'comment_not_found'
  | 'comment_forbidden'
  | 'comment_removed'
  | 'comment_body_invalid'
  | 'comment_position_invalid'
  | 'comment_parent_invalid'
  | 'comment_cursor_invalid'
  | 'comment_thread_full';

export type DiscussionResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: DiscussionErrorCode } };

export type DiscussionComment = {
  readonly id: string;
  readonly author: { readonly displayName: string | null };
  readonly body: string | null;
  readonly positionMs: number | null;
  readonly createdAt: string;
  readonly editedAt: string | null;
  readonly removed: boolean;
  readonly canEdit: boolean;
  readonly canRemove: boolean;
};

export type DiscussionThread = {
  readonly root: DiscussionComment;
  readonly replies: readonly DiscussionComment[];
};

export type DiscussionPage = {
  readonly threads: readonly DiscussionThread[];
  readonly nextCursor: string | null;
};

export type PublishDiscussionInput = {
  readonly body: string;
  readonly positionMs?: number | null;
  readonly parentId?: string | null;
};

export type UpdateDiscussionInput = { readonly commentId: string; readonly body: string };
