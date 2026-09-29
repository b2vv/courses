import type {
  DiscussionPage,
  DiscussionResult,
  DiscussionComment,
  PublishDiscussionInput,
  UpdateDiscussionInput,
} from '../domain/discussion/discussion';

export interface DiscussionRepository {
  list(
    lessonId: string,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<DiscussionResult<DiscussionPage>>;

  publish(
    lessonId: string,
    input: PublishDiscussionInput,
    signal?: AbortSignal,
  ): Promise<DiscussionResult<DiscussionComment>>;

  update(
    input: UpdateDiscussionInput,
    signal?: AbortSignal,
  ): Promise<DiscussionResult<DiscussionComment>>;

  remove(commentId: string, signal?: AbortSignal): Promise<DiscussionResult<void>>;
  report?(commentId: string, signal?: AbortSignal): Promise<DiscussionResult<void>>;
}
