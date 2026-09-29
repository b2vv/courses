import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type {
  DiscussionComment,
  DiscussionErrorCode,
  DiscussionPage,
  PublishDiscussionInput,
} from '../../domain/discussion/discussion';
import { Button, TextArea } from '../ui';
import { formatTimestamp } from '../format-timestamp';

type DiscussionPanelProps = {
  readonly page: DiscussionPage | undefined;
  readonly error?: DiscussionErrorCode;
  readonly isLoading: boolean;
  readonly isLoadingMore?: boolean;
  readonly onRetry: () => void;
  readonly onLoadMore?: () => void;
  readonly positionMs: number;
  readonly onSeek?: (positionMs: number) => void;
  readonly onPublish: (input: PublishDiscussionInput) => Promise<void>;
  readonly onUpdate: (commentId: string, body: string) => Promise<void>;
  readonly onRemove: (commentId: string) => Promise<void>;
  readonly onReport?: (commentId: string) => Promise<void>;
};

function commentText(comment: DiscussionComment, t: (key: string) => string): string {
  if (comment.removed) return t('learn.discussion.removed');
  return comment.body ?? '';
}

function errorText(t: (key: string) => string, code: DiscussionErrorCode | undefined): string {
  if (code === undefined) return t('learn.discussion.errors.generic');
  return t(`learn.discussion.errors.${code}`);
}

function CommentCard({
  comment,
  replies,
  canReply,
  onSeek,
  onReply,
  onUpdate,
  onRemove,
  onReport,
}: {
  readonly comment: DiscussionComment;
  readonly replies: readonly DiscussionComment[];
  readonly canReply: boolean;
  readonly onSeek?: (positionMs: number) => void;
  readonly onReply: (parentId: string, body: string) => Promise<void>;
  readonly onUpdate: (commentId: string, body: string) => Promise<void>;
  readonly onRemove: (commentId: string) => Promise<void>;
  readonly onReport?: (commentId: string) => Promise<void>;
}) {
  const { t } = useTranslation();
  const [replying, setReplying] = useState(false);
  const [editing, setEditing] = useState(false);
  const [body, setBody] = useState(comment.body ?? '');
  const [replyBody, setReplyBody] = useState('');
  const [busy, setBusy] = useState(false);
  const [reported, setReported] = useState(false);
  const [reportError, setReportError] = useState(false);
  const author = comment.author.displayName ?? t('learn.discussion.formerParticipant');
  const position = comment.positionMs;
  const run = async (action: () => Promise<void>) => {
    setBusy(true);
    try {
      await action();
      setEditing(false);
      setReplying(false);
      setReplyBody('');
    } finally {
      setBusy(false);
    }
  };
  return (
    <li className="discussion-thread">
      <article className="discussion-comment">
        <header className="discussion-comment-header">
          <strong>{author}</strong>
          <time dateTime={comment.createdAt}>
            {new Date(comment.createdAt).toLocaleDateString()}
          </time>
          {comment.editedAt === null || comment.removed ? null : (
            <span>{t('learn.discussion.edited')}</span>
          )}
        </header>
        {position === null ? null : onSeek === undefined ? (
          <time dateTime={`PT${position / 1000}S`} aria-label={formatTimestamp(position)}>
            {formatTimestamp(position)}
          </time>
        ) : (
          <Button variant="secondary" onPress={() => onSeek(position)}>
            {formatTimestamp(position)}
          </Button>
        )}
        {editing ? (
          <div className="discussion-edit-form">
            <TextArea label={t('learn.discussion.form.body')} value={body} onChange={setBody} />
            <div className="discussion-actions">
              <Button
                isDisabled={busy || body.trim() === ''}
                onPress={() => void run(() => onUpdate(comment.id, body))}
              >
                {t('learn.discussion.save')}
              </Button>
              <Button variant="secondary" onPress={() => setEditing(false)}>
                {t('common.close')}
              </Button>
            </div>
          </div>
        ) : (
          <p
            className={
              comment.removed ? 'discussion-body discussion-body-removed' : 'discussion-body'
            }
          >
            {commentText(comment, t)}
          </p>
        )}
        <div className="discussion-actions">
          {canReply && !comment.removed ? (
            <Button variant="secondary" onPress={() => setReplying((value) => !value)}>
              {t('learn.discussion.reply')}
            </Button>
          ) : null}
          {comment.canEdit && !comment.removed ? (
            <Button variant="secondary" onPress={() => setEditing(true)}>
              {t('learn.discussion.edit')}
            </Button>
          ) : null}
          {comment.canRemove && !comment.removed ? (
            <Button variant="danger" onPress={() => void run(() => onRemove(comment.id))}>
              {t('learn.discussion.remove')}
            </Button>
          ) : null}
          {onReport !== undefined && !comment.canEdit && !comment.removed ? (
            <Button
              variant="secondary"
              isDisabled={busy || reported}
              onPress={() => {
                setReportError(false);
                void run(() => onReport(comment.id))
                  .then(() => setReported(true))
                  .catch(() => setReportError(true));
              }}
            >
              {reported ? t('learn.discussion.reported') : t('learn.discussion.report')}
            </Button>
          ) : null}
        </div>
        {reportError ? <p role="alert">{t('learn.discussion.reportError')}</p> : null}
        {canReply && replying ? (
          <form
            className="discussion-reply-form"
            onSubmit={(event) => {
              event.preventDefault();
              void run(() => onReply(comment.id, replyBody));
            }}
          >
            <TextArea
              label={t('learn.discussion.form.reply')}
              value={replyBody}
              onChange={setReplyBody}
            />
            <Button isDisabled={busy || replyBody.trim() === ''} type="submit">
              {t('learn.discussion.sendReply')}
            </Button>
          </form>
        ) : null}
      </article>
      {replies.length === 0 ? null : (
        <ol className="discussion-replies">
          {replies.map((reply) => (
            <CommentCard
              key={reply.id}
              comment={reply}
              replies={[]}
              canReply={false}
              {...(onSeek === undefined ? {} : { onSeek })}
              onReply={onReply}
              onUpdate={onUpdate}
              onRemove={onRemove}
              {...(onReport === undefined ? {} : { onReport })}
            />
          ))}
        </ol>
      )}
    </li>
  );
}

export function DiscussionPanel({
  page,
  error,
  isLoading,
  isLoadingMore = false,
  onRetry,
  onLoadMore,
  positionMs,
  onSeek,
  onPublish,
  onUpdate,
  onRemove,
  onReport,
}: DiscussionPanelProps) {
  const { t } = useTranslation();
  const [body, setBody] = useState('');
  const [position, setPosition] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const publish = async () => {
    setBusy(true);
    try {
      await onPublish({ body, positionMs: position });
      setBody('');
      setPosition(null);
    } finally {
      setBusy(false);
    }
  };
  return (
    <section className="discussion-panel" aria-labelledby="discussion-heading">
      <h2 id="discussion-heading">{t('learn.discussion.heading')}</h2>
      {isLoading ? <output>{t('learn.discussion.loading')}</output> : null}
      {!isLoading && error !== undefined ? (
        <div role="alert">
          <p>{errorText(t, error)}</p>
          <Button onPress={onRetry}>{t('common.retry')}</Button>
        </div>
      ) : null}
      {!isLoading && error === undefined && page?.threads.length === 0 ? (
        <p>{t('learn.discussion.empty')}</p>
      ) : null}
      {page === undefined || error !== undefined ? null : (
        <ol className="discussion-threads" aria-label={t('learn.discussion.heading')}>
          {page.threads.map((thread) => (
            <CommentCard
              key={thread.root.id}
              comment={thread.root}
              replies={thread.replies}
              canReply
              {...(onSeek === undefined ? {} : { onSeek })}
              onReply={(parentId, replyBody) => onPublish({ body: replyBody, parentId })}
              onUpdate={onUpdate}
              onRemove={onRemove}
              {...(onReport === undefined ? {} : { onReport })}
            />
          ))}
        </ol>
      )}
      {page?.nextCursor === null || onLoadMore === undefined ? null : (
        <Button variant="secondary" isDisabled={isLoadingMore} onPress={onLoadMore}>
          {t('learn.discussion.showMore')}
        </Button>
      )}
      <form
        className="discussion-compose"
        onSubmit={(event) => {
          event.preventDefault();
          void publish();
        }}
      >
        <TextArea label={t('learn.discussion.form.body')} value={body} onChange={setBody} />
        {position === null ? (
          <Button variant="secondary" onPress={() => setPosition(positionMs)}>
            {t('learn.discussion.form.currentMoment')}
          </Button>
        ) : (
          <p>{t('learn.discussion.form.attached', { time: formatTimestamp(position) })}</p>
        )}
        <Button isDisabled={busy || body.trim() === ''} type="submit">
          {t('learn.discussion.publish')}
        </Button>
      </form>
    </section>
  );
}
