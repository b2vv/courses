import type {
  DismissReportsInput,
  LearnerFilters,
  LearnerPage,
  LearnerScope,
  ModerationQueuePage,
  TeachCommentPage,
  TeachResult,
} from '../domain/teach';

export interface TeachRepository {
  listLearners(
    scope: LearnerScope,
    filters: LearnerFilters,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<TeachResult<LearnerPage>>;
  listCourseComments(
    courseId: string,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<TeachResult<TeachCommentPage>>;
  listModerationQueue(
    courseId: string,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<TeachResult<ModerationQueuePage>>;
  blockAuthor(
    kind: 'comment' | 'review',
    id: string,
    signal?: AbortSignal,
  ): Promise<TeachResult<void>>;
  unblockAuthor(
    kind: 'comment' | 'review',
    id: string,
    signal?: AbortSignal,
  ): Promise<TeachResult<void>>;
  dismissReport(
    commentId: string,
    input: DismissReportsInput,
    signal?: AbortSignal,
  ): Promise<TeachResult<void>>;
  removeComment(commentId: string, signal?: AbortSignal): Promise<TeachResult<void>>;
  dismissReviewReport(
    reviewId: string,
    input: DismissReportsInput,
    signal?: AbortSignal,
  ): Promise<TeachResult<void>>;
  removeReview(
    reviewId: string,
    input: DismissReportsInput,
    signal?: AbortSignal,
  ): Promise<TeachResult<void>>;
}
