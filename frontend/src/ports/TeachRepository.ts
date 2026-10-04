import type {
  DismissReportsInput,
  ModerationQueuePage,
  TeachCommentPage,
  TeachResult,
} from '../domain/teach/teach';

export interface TeachRepository {
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
