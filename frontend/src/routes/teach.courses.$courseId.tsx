import { useMemo, useState } from 'react';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { TeachQueryError } from '../app/composition/teach-queries';
import { EnrollmentQueryError } from '../app/composition/enrollment-queries';
import { validEntityId } from '../domain/course/course';
import type { ModerationQueueItem, TeachComment } from '../domain/teach/teach';
import { formatTimestamp } from '../presentation/format-timestamp';
import { ReviewModerationRow } from '../presentation/teach/ReviewModerationRow';
import { CourseRoster } from '../presentation/teach/CourseRoster';
import { Button } from '../presentation/ui';

const moderationReason = 'moderation_reviewed';

function courseId(value: string): string {
  if (!validEntityId(value)) throw notFound();
  return value;
}

function isCommentQueueItem(
  item: ModerationQueueItem,
): item is ModerationQueueItem & { readonly kind: 'comment'; readonly comment: TeachComment } {
  return item.kind === 'comment' && item.comment !== null;
}
function isReviewQueueItem(item: ModerationQueueItem): item is ModerationQueueItem & {
  readonly kind: 'review';
  readonly review: NonNullable<ModerationQueueItem['review']>;
} {
  return item.kind === 'review' && item.review !== null;
}

export const Route = createFileRoute('/teach/courses/$courseId')({
  ssr: false,
  loader: async ({ context, params }) => {
    const id = courseId(params.courseId);
    await Promise.all([
      context.queryClient.prefetchInfiniteQuery(context.teach.courseComments(id)),
      context.queryClient.prefetchInfiniteQuery(context.teach.moderationQueue(id)),
      context.queryClient.prefetchInfiniteQuery(context.enrollment.roster(id)),
    ]);
    return { courseId: id };
  },
  component: TeachCourseRoute,
  errorComponent: TeachCourseFailure,
});

function TeachCourseFailure({ error }: { readonly error: Error }) {
  return (
    <section role="alert" className="notice">
      <h1>Не вдалося завантажити робоче місце</h1>
      <p>{error.message}</p>
    </section>
  );
}

function errorMessage(error: unknown): string | undefined {
  return error instanceof TeachQueryError
    ? error.code
    : error === undefined
      ? undefined
      : 'dependency_unavailable';
}
function commentLabel(comment: TeachComment): string {
  return `${comment.author.displayName ?? 'Former participant'}: ${comment.body ?? 'Comment removed'}`;
}

function TeachCourseRoute() {
  const { teach } = Route.useRouteContext();
  const { enrollment } = Route.useRouteContext();
  const { identity } = Route.useRouteContext();
  const { courseId: id } = Route.useLoaderData();
  const queryClient = useQueryClient();
  const comments = useInfiniteQuery(teach.courseComments(id));
  const queue = useInfiniteQuery(teach.moderationQueue(id));
  const roster = useInfiniteQuery(enrollment.roster(id));
  const dismiss = useMutation(teach.dismiss(queryClient));
  const remove = useMutation(teach.remove(queryClient));
  const dismissReview = useMutation(teach.dismissReview(queryClient));
  const removeReview = useMutation(teach.removeReview(queryClient));
  const block = useMutation(enrollment.block(queryClient));
  const unblock = useMutation(enrollment.unblock(queryClient));
  const reportStudent = useMutation(enrollment.reportStudent());
  const [actionError, setActionError] = useState<string>();
  const [rosterActionError, setRosterActionError] = useState<string>();
  const commentItems = useMemo(
    () => comments.data?.pages.flatMap((page) => page.comments) ?? [],
    [comments.data],
  );
  const queueItems = useMemo(
    () => queue.data?.pages.flatMap((page) => page.items) ?? [],
    [queue.data],
  );
  const rosterEntries = useMemo(
    () => roster.data?.pages.flatMap((page) => page.entries) ?? [],
    [roster.data],
  );
  const run = async (action: () => Promise<unknown>) => {
    setActionError(undefined);
    try {
      await action();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : 'dependency_unavailable');
    }
  };
  const runRosterAction = async (action: () => Promise<unknown>) => {
    setRosterActionError(undefined);
    try {
      await action();
    } catch (error) {
      setRosterActionError(
        error instanceof EnrollmentQueryError ? error.code : 'dependency_unavailable',
      );
    }
  };
  return (
    <div className="teach-workspace">
      <h1>Робоче місце інструктора</h1>
      <section aria-labelledby="teach-comments-heading" className="teach-section">
        <h2 id="teach-comments-heading">Коментарі курсу</h2>
        {comments.isLoading ? <output>Завантажуємо коментарі…</output> : null}
        {!comments.isLoading && comments.error ? (
          <div role="alert">
            <p>{errorMessage(comments.error)}</p>
            <Button onPress={() => void comments.refetch()}>Спробувати ще</Button>
          </div>
        ) : null}
        {!comments.isLoading && comments.error === null && commentItems.length === 0 ? (
          <p>Коментарів у курсі ще немає.</p>
        ) : null}
        {commentItems.length > 0 ? (
          <ul aria-label="Коментарі курсу">
            {commentItems.map((comment) => (
              <li key={comment.id}>
                <article>
                  <p>{comment.lessonTitle}</p>
                  <p>{commentLabel(comment)}</p>
                  {comment.positionMs === null ? null : (
                    <time dateTime={`PT${comment.positionMs / 1000}S`}>
                      {formatTimestamp(comment.positionMs)}
                    </time>
                  )}
                  <time dateTime={comment.createdAt}>
                    {new Date(comment.createdAt).toLocaleString()}
                  </time>
                </article>
              </li>
            ))}
          </ul>
        ) : null}
        {comments.hasNextPage ? (
          <Button
            variant="secondary"
            isDisabled={comments.isFetchingNextPage}
            onPress={() => void comments.fetchNextPage()}
          >
            Показати ще
          </Button>
        ) : null}
      </section>
      <section aria-labelledby="teach-moderation-heading" className="teach-section">
        <h2 id="teach-moderation-heading">Черга модерації</h2>
        {queue.isLoading ? <output>Завантажуємо чергу…</output> : null}
        {queue.error ? (
          <div role="alert">
            <p>{errorMessage(queue.error)}</p>
            <Button onPress={() => void queue.refetch()}>Спробувати ще</Button>
          </div>
        ) : null}
        {!queue.isLoading && queue.error === null && queueItems.length === 0 ? (
          <p>Відкритих скарг немає.</p>
        ) : null}
        {actionError ? <p role="alert">{actionError}</p> : null}
        {queueItems.length > 0 ? (
          <ul aria-label="Черга модерації">
            {queueItems.map((item) =>
              isCommentQueueItem(item) ? (
                <ModerationRow
                  key={item.comment.id}
                  item={item}
                  busy={dismiss.isPending || remove.isPending}
                  onDismiss={(reason) =>
                    void run(() =>
                      dismiss.mutateAsync({ commentId: item.comment.id, input: { reason } }),
                    )
                  }
                  onRemove={() =>
                    void run(() =>
                      remove.mutateAsync({
                        lessonId: item.comment.lessonId,
                        commentId: item.comment.id,
                      }),
                    )
                  }
                />
              ) : isReviewQueueItem(item) ? (
                <ReviewModerationRow
                  key={item.review.id}
                  item={item}
                  busy={dismissReview.isPending || removeReview.isPending}
                  admin={identity.status === 'signed-in' && identity.user.role === 'admin'}
                  onDismiss={() =>
                    void run(() =>
                      dismissReview.mutateAsync({
                        reviewId: item.review.id,
                        input: { reason: moderationReason },
                      }),
                    )
                  }
                  onRemove={() =>
                    void run(() =>
                      removeReview.mutateAsync({
                        reviewId: item.review.id,
                        input: { reason: moderationReason },
                      }),
                    )
                  }
                />
              ) : null,
            )}
          </ul>
        ) : null}
        {queue.hasNextPage ? (
          <Button
            variant="secondary"
            isDisabled={queue.isFetchingNextPage}
            onPress={() => void queue.fetchNextPage()}
          >
            Показати ще
          </Button>
        ) : null}
      </section>
      <CourseRoster
        entries={rosterEntries}
        isLoading={roster.isLoading}
        error={roster.error}
        hasNextPage={roster.hasNextPage}
        isFetchingNextPage={roster.isFetchingNextPage}
        isPending={block.isPending || unblock.isPending}
        actionError={rosterActionError}
        onRetry={() => void roster.refetch()}
        onLoadMore={() => void roster.fetchNextPage()}
        onBlock={(userId) => void runRosterAction(() => block.mutateAsync({ courseId: id, userId }))}
        onUnblock={(userId) =>
          void runRosterAction(() => unblock.mutateAsync({ courseId: id, userId }))
        }
        onReportStudent={(userId, reason) =>
          reportStudent.mutateAsync({ courseId: id, userId, reason })
        }
      />
    </div>
  );
}

function ModerationRow({
  item,
  busy,
  onDismiss,
  onRemove,
}: {
  readonly item: ModerationQueueItem & { readonly kind: 'comment'; readonly comment: TeachComment };
  readonly busy: boolean;
  readonly onDismiss: (reason: string) => void;
  readonly onRemove: () => void;
}) {
  return (
    <li>
      <article>
        <p>{commentLabel(item.comment)}</p>
        <p>Скарг: {item.openReports}</p>
        <div className="discussion-actions">
          <Button variant="secondary" isDisabled={busy} onPress={() => onDismiss(moderationReason)}>
            Відхилити скарги
          </Button>
          <Button variant="danger" isDisabled={busy || item.comment.removed} onPress={onRemove}>
            Прибрати коментар
          </Button>
        </div>
      </article>
    </li>
  );
}
