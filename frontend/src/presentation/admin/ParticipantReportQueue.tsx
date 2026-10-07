import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EnrollmentQueryError } from '../../app/composition/enrollment-queries';
import type { ParticipantReportQueueItem } from '../../domain/enrollment/enrollment';
import { Button, Dialog, EmptyState } from '../ui';

type ParticipantReportQueueProps = {
  readonly items: readonly ParticipantReportQueueItem[];
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly hasNextPage: boolean;
  readonly isFetchingNextPage: boolean;
  readonly resolvePending: boolean;
  readonly revokePending: boolean;
  readonly onRetry: () => void;
  readonly onLoadMore: () => void;
  readonly onResolve: (reportId: string) => Promise<void>;
  readonly onRevoke: (courseId: string, userId: string) => Promise<void>;
};

function errorKey(error: unknown): string {
  const code = error instanceof EnrollmentQueryError ? error.code : undefined;
  switch (code) {
    case 'authentication_required':
    case 'report_forbidden':
    case 'comment_cursor_invalid':
    case 'invalid_request':
    case 'membership_conflict':
    case 'enrollment_forbidden':
    case 'dependency_unavailable':
    case 'internal_error':
      return code;
    default:
      return 'generic';
  }
}

export function ParticipantReportQueue({
  items,
  isLoading,
  error,
  hasNextPage,
  isFetchingNextPage,
  resolvePending,
  revokePending,
  onRetry,
  onLoadMore,
  onResolve,
  onRevoke,
}: ParticipantReportQueueProps) {
  const { t, i18n } = useTranslation();
  const [actedReportId, setActedReportId] = useState<string>();
  const [actionError, setActionError] = useState<string>();
  const priorCount = useRef(items.length);
  const firstAdded = useRef<HTMLLIElement>(null);
  const [firstAddedIndex, setFirstAddedIndex] = useState<number>();

  useEffect(() => {
    if (items.length > priorCount.current) setFirstAddedIndex(priorCount.current);
    priorCount.current = items.length;
  }, [items.length]);
  useEffect(() => {
    if (!isFetchingNextPage && firstAddedIndex !== undefined) firstAdded.current?.focus();
  }, [firstAddedIndex, isFetchingNextPage]);

  async function resolve(item: ParticipantReportQueueItem) {
    setActedReportId(item.id);
    setActionError(undefined);
    try {
      await onResolve(item.id);
    } catch (actionFailure) {
      setActionError(errorKey(actionFailure));
    }
  }
  async function revoke(item: ParticipantReportQueueItem, close: () => void) {
    setActedReportId(item.id);
    setActionError(undefined);
    try {
      await onRevoke(item.courseId, item.targetId);
      close();
    } catch (actionFailure) {
      setActionError(errorKey(actionFailure));
    }
  }

  return (
    <section aria-labelledby="participant-reports-heading" className="admin-user-list">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('enrollment.reports.eyebrow')}</p>
          <h1 id="participant-reports-heading">{t('enrollment.reports.title')}</h1>
        </div>
      </div>
      {isLoading ? <output>{t('common.loading')}</output> : null}
      {error ? (
        <div role="alert">
          <p>{t(`enrollment.errors.${errorKey(error)}`)}</p>
          <Button onPress={onRetry}>{t('common.retry')}</Button>
        </div>
      ) : null}
      {actionError ? <p role="alert">{t(`enrollment.errors.${actionError}`)}</p> : null}
      {!isLoading && error === null && items.length === 0 ? (
        <EmptyState title={t('enrollment.reports.empty')} />
      ) : null}
      {items.length > 0 ? (
        <ul aria-label={t('enrollment.reports.title')} className="teach-section">
          {items.map((item, index) => {
            const isPending = actedReportId === item.id && (resolvePending || revokePending);
            return (
              <li
                key={item.id}
                {...(index === firstAddedIndex ? { ref: firstAdded, tabIndex: -1 } : {})}
              >
                <article>
                  <h2>{t(`enrollment.reports.kind.${item.kind}`)}</h2>
                  <p>
                    {t('enrollment.reports.reporter')}:{' '}
                    {item.reporterName ?? t('ratings.formerAuthor')}
                  </p>
                  <p>
                    {t('enrollment.reports.target')}: {item.targetName ?? t('ratings.formerAuthor')}
                  </p>
                  <p>
                    {t('enrollment.reports.course')}: {item.courseTitle}
                  </p>
                  <p>
                    {t('enrollment.reports.reason')}: {item.reason}
                  </p>
                  <p>{t(`enrollment.reports.state.${item.state}`)}</p>
                  <time dateTime={item.createdAt}>
                    {new Date(item.createdAt).toLocaleString(i18n.resolvedLanguage)}
                  </time>
                  <div className="admin-user-actions">
                    <Button isPending={isPending} onPress={() => void resolve(item)}>
                      {t('enrollment.reports.resolve')}
                    </Button>
                    {/* Revoke only applies to a reported student; on a lecturer report the target is
                        the course owner, who has no membership — revoke would just fail grant_invalid,
                        a dead action. So the control is gated to on_student reports. */}
                    {item.kind === 'on_student' ? (
                      <Dialog
                        triggerLabel={t('enrollment.reports.revoke')}
                        title={t('enrollment.reports.revoke')}
                        closeLabel={t('common.close')}
                      >
                        {(close) => (
                          <>
                            <p>{t('enrollment.reports.revokeConfirm')}</p>
                            <Button
                              variant="danger"
                              isPending={isPending}
                              onPress={() => void revoke(item, close)}
                            >
                              {t('enrollment.reports.revoke')}
                            </Button>
                          </>
                        )}
                      </Dialog>
                    ) : null}
                  </div>
                </article>
              </li>
            );
          })}
        </ul>
      ) : null}
      {hasNextPage ? (
        <Button variant="secondary" isPending={isFetchingNextPage} onPress={onLoadMore}>
          {t('enrollment.reports.more')}
        </Button>
      ) : null}
    </section>
  );
}
