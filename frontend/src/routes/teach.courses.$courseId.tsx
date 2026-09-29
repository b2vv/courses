import { useMemo, useState } from 'react';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { TeachQueryError } from '../app/composition/teach-queries';
import { validEntityId } from '../domain/course/course';
import type { ModerationQueueItem, TeachComment } from '../domain/teach/teach';
import { formatTimestamp } from '../presentation/format-timestamp';
import { Button } from '../presentation/ui';

function courseId(value: string): string {
  if (!validEntityId(value)) throw notFound();
  return value;
}

export const Route = createFileRoute('/teach/courses/$courseId')({
  ssr: false,
  loader: async ({ context, params }) => {
    const id = courseId(params.courseId);
    await Promise.all([
      context.queryClient.prefetchInfiniteQuery(context.teach.courseComments(id)),
      context.queryClient.prefetchInfiniteQuery(context.teach.moderationQueue(id)),
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
  const { courseId: id } = Route.useLoaderData();
  const queryClient = useQueryClient();
  const comments = useInfiniteQuery(teach.courseComments(id));
  const queue = useInfiniteQuery(teach.moderationQueue(id));
  const dismiss = useMutation(teach.dismiss(queryClient));
  const remove = useMutation(teach.remove(queryClient));
  const [actionError, setActionError] = useState<string>();
  const commentItems = useMemo(
    () => comments.data?.pages.flatMap((page) => page.comments) ?? [],
    [comments.data],
  );
  const queueItems = useMemo(
    () => queue.data?.pages.flatMap((page) => page.items) ?? [],
    [queue.data],
  );
  const run = async (action: () => Promise<unknown>) => {
    setActionError(undefined);
    try {
      await action();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : 'dependency_unavailable');
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
            {queueItems.map((item) => (
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
            ))}
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
    </div>
  );
}

function ModerationRow({
  item,
  busy,
  onDismiss,
  onRemove,
}: {
  readonly item: ModerationQueueItem;
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
          <Button
            variant="secondary"
            isDisabled={busy}
            onPress={() => onDismiss('moderation_reviewed')}
          >
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
