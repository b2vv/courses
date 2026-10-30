import { useMemo, useState } from 'react';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, notFound, type ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { TeachQueryError } from '../app/composition/teach-queries';
import { EnrollmentQueryError } from '../app/composition/enrollment-queries';
import { asBrandedId, validEntityId } from '../contracts/course';
import type { ModerationQueueItem, TeachComment } from '../modules/teach';
import { formatTimestamp } from '../presentation/format-timestamp';
import { ReviewModerationRow } from '../modules/teach';
import { CourseRoster, TeachLessonList } from '../modules/teach';
import { Button } from '../presentation/ui';
import { ErrorState, Skeleton } from '../presentation/ui/states';

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
      context.queryClient.prefetchQuery(context.teach.detail(asBrandedId(id, 'course'))),
      context.queryClient.prefetchInfiniteQuery(context.teach.courseComments(id)),
      context.queryClient.prefetchInfiniteQuery(context.teach.moderationQueue(id)),
      context.queryClient.prefetchInfiniteQuery(context.enrollment.roster(id)),
    ]);
    return { courseId: id };
  },
  component: TeachCourseRoute,
  errorComponent: TeachCourseFailure,
});

function TeachCourseFailure({ error }: ErrorComponentProps) {
  const message = error instanceof Error ? error.message : String(error);
  return (
    <section role="alert" className="notice">
      <h1>Не вдалося завантажити робоче місце</h1>
      <p>{message}</p>
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
function authorErrorKey(code: string): string {
  switch (code) {
    case 'enrollment_forbidden':
    case 'grant_transition_invalid':
    case 'membership_conflict':
    case 'comment_not_found':
    case 'review_not_found':
      return code;
    default:
      return 'generic';
  }
}
function commentLabel(comment: TeachComment): string {
  return `${comment.author.displayName ?? 'Former participant'}: ${comment.body ?? 'Comment removed'}`;
}

function TeachCourseRoute() {
  const { t } = useTranslation();
  const { teach } = Route.useRouteContext();
  const { enrollment } = Route.useRouteContext();
  const { identity } = Route.useRouteContext();
  const { courseId: id } = Route.useLoaderData();
  const queryClient = useQueryClient();
  const tree = useQuery(teach.detail(asBrandedId(id, 'course')));
  const comments = useInfiniteQuery(teach.courseComments(id));
  const queue = useInfiniteQuery(teach.moderationQueue(id));
  const roster = useInfiniteQuery(enrollment.roster(id));
  const dismiss = useMutation(teach.dismiss(queryClient));
  const remove = useMutation(teach.remove(queryClient));
  const dismissReview = useMutation(teach.dismissReview(queryClient));
  const removeReview = useMutation(teach.removeReview(queryClient));
  const blockAuthor = useMutation(teach.blockAuthor(queryClient));
  const unblockAuthor = useMutation(teach.unblockAuthor(queryClient));
  const block = useMutation(enrollment.block(queryClient));
  const unblock = useMutation(enrollment.unblock(queryClient));
  const reportStudent = useMutation(enrollment.reportStudent());
  const [actionError, setActionError] = useState<string>();
  const [rosterActionError, setRosterActionError] = useState<string>();
  const [authorError, setAuthorError] = useState<string>();
  // Per-row pending: authorActing only SCOPES the spinner to the acted row (keyed by the item id —
  // the comment or review id, never the author's id, Lo14). The busy state itself comes from the
  // mutation's isPending (see authorPending below), which falls back to false when the request
  // settles — so authorActing needs no reset, exactly as CourseRoster's actedUserId. Dropping that
  // multiplier left the flipped button stuck pending and unclickable (Courses 2 review, LR5).
  const [authorActing, setAuthorActing] = useState<string>();
  const runAuthor = async (kind: 'comment' | 'review', itemId: string, blocked: boolean) => {
    setAuthorError(undefined);
    setAuthorActing(itemId);
    try {
      await (blocked ? unblockAuthor : blockAuthor).mutateAsync({ kind, id: itemId });
    } catch (error) {
      setAuthorError(error instanceof TeachQueryError ? error.code : 'dependency_unavailable');
    }
  };
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
      <section aria-labelledby="teach-lessons-heading" className="teach-section">
        <h2 id="teach-lessons-heading">{t('administration.program')}</h2>
        {tree.isPending ? (
          <Skeleton label={t('administration.document.loading')} lines={3} />
        ) : tree.isError ? (
          <ErrorState title={t('teach.lessons.loadError')}>
            <Button onPress={() => void tree.refetch()}>{t('common.retry')}</Button>
          </ErrorState>
        ) : (
          <TeachLessonList courseId={id} title={tree.data.title} modules={tree.data.modules} />
        )}
      </section>
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
        {authorError ? (
          <p role="alert">{t(`teach.authorBlock.errors.${authorErrorKey(authorError)}`)}</p>
        ) : null}
        {queueItems.length > 0 ? (
          <ul aria-label="Черга модерації">
            {queueItems.map((item) =>
              isCommentQueueItem(item) ? (
                <ModerationRow
                  key={item.comment.id}
                  item={item}
                  busy={dismiss.isPending || remove.isPending}
                  authorPending={
                    (blockAuthor.isPending || unblockAuthor.isPending) &&
                    authorActing === item.comment.id
                  }
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
                  onToggleAuthorBlock={() =>
                    void runAuthor('comment', item.comment.id, item.authorBlocked)
                  }
                />
              ) : isReviewQueueItem(item) ? (
                <ReviewModerationRow
                  key={item.review.id}
                  item={item}
                  busy={dismissReview.isPending || removeReview.isPending}
                  admin={identity.status === 'signed-in' && identity.user.role === 'admin'}
                  authorPending={
                    (blockAuthor.isPending || unblockAuthor.isPending) &&
                    authorActing === item.review.id
                  }
                  onToggleAuthorBlock={() =>
                    void runAuthor('review', item.review.id, item.authorBlocked)
                  }
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
        onBlock={(userId) =>
          void runRosterAction(() => block.mutateAsync({ courseId: id, userId }))
        }
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
  authorPending,
  onDismiss,
  onRemove,
  onToggleAuthorBlock,
}: {
  readonly item: ModerationQueueItem & { readonly kind: 'comment'; readonly comment: TeachComment };
  readonly busy: boolean;
  readonly authorPending: boolean;
  readonly onDismiss: (reason: string) => void;
  readonly onRemove: () => void;
  readonly onToggleAuthorBlock: () => void;
}) {
  const { t } = useTranslation();
  return (
    <li>
      <article>
        <p>{commentLabel(item.comment)}</p>
        <p>Скарг: {item.openReports}</p>
        <div className="discussion-actions">
          {/* isPending (not isDisabled) for the in-flight state: a native disabled drops focus to
              <body> while the request runs (B-L7b.1), as on the author-block button below. */}
          <Button variant="secondary" isPending={busy} onPress={() => onDismiss(moderationReason)}>
            Відхилити скарги
          </Button>
          <Button
            variant="danger"
            isPending={busy}
            isDisabled={item.comment.removed}
            onPress={onRemove}
          >
            Прибрати коментар
          </Button>
          {/* isPending (per row), not isDisabled: a native disabled drops focus to <body> while the
              membership transition runs (B-L7b.1). */}
          <Button
            variant={item.authorBlocked ? 'secondary' : 'danger'}
            isPending={authorPending}
            onPress={onToggleAuthorBlock}
          >
            {t(item.authorBlocked ? 'teach.unblockAuthor' : 'teach.blockAuthor')}
          </Button>
        </div>
      </article>
    </li>
  );
}
