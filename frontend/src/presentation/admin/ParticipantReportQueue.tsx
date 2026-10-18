import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EnrollmentQueryError } from '../../app/composition/enrollment-queries';
import type {
  ParticipantReportQueueItem,
  ParticipantReportStateFilter,
} from '../../modules/enrollment';
import { Button, Dialog, PageHeader, SegmentedControl } from '../ui';

const stateFilters: readonly ParticipantReportStateFilter[] = ['all', 'open', 'resolved'];

const reportStateVariants = {
  open: 'warning',
  resolved: 'success',
} satisfies Record<ParticipantReportQueueItem['state'], string>;

type ParticipantReportQueueProps = {
  readonly items: readonly ParticipantReportQueueItem[];
  readonly state: ParticipantReportStateFilter;
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly hasNextPage: boolean;
  readonly isFetchingNextPage: boolean;
  readonly resolvePending: boolean;
  readonly revokePending: boolean;
  readonly onFilterChange: (state: ParticipantReportStateFilter) => void;
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
  state,
  isLoading,
  error,
  hasNextPage,
  isFetchingNextPage,
  resolvePending,
  revokePending,
  onFilterChange,
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
    <section className="admin-user-list" data-surface="admin">
      <PageHeader
        eyebrow={t('enrollment.reports.eyebrow')}
        title={t('enrollment.reports.title')}
        description={t('enrollment.reports.lede')}
      />
      <div className="admin-queue-filter">
        <SegmentedControl
          label={t('enrollment.reports.filterLabel')}
          options={stateFilters.map((value) => ({
            id: value,
            label: t(`enrollment.reports.filter.${value}`),
          }))}
          selectedKeys={new Set([state])}
          onSelectionChange={(keys) =>
            onFilterChange(filterFromKey(String([...keys][0] ?? 'open')))
          }
        />
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
        <div className="ui-card admin-queue-empty">
          <span className="ui-icon-tile" aria-hidden="true">
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              focusable="false"
            >
              <path d="M12 3 3 7v5c0 5 9 9 9 9s9-4 9-9V7zM12 8v5m0 4h.01" />
            </svg>
          </span>
          <h2>
            {t(state === 'open' ? 'enrollment.reports.empty' : 'enrollment.reports.emptyFiltered')}
          </h2>
          <p>
            {t(
              state === 'open'
                ? 'enrollment.reports.emptyHint'
                : 'enrollment.reports.emptyFilteredHint',
            )}
          </p>
          {state === 'open' ? null : (
            <Button variant="secondary" onPress={() => onFilterChange('open')}>
              {t('enrollment.reports.showOpen')}
            </Button>
          )}
        </div>
      ) : null}
      {items.length > 0 ? (
        <ul aria-label={t('enrollment.reports.title')} className="admin-queue-list">
          {items.map((item, index) => {
            const isPending = actedReportId === item.id && (resolvePending || revokePending);
            return (
              <li
                key={item.id}
                {...(index === firstAddedIndex ? { ref: firstAdded, tabIndex: -1 } : {})}
              >
                <article className="ui-card admin-queue-card">
                  <div className="admin-queue-heading">
                    <span className="ui-icon-tile" aria-hidden="true">
                      <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        focusable="false"
                      >
                        <path d="M12 3 3 7v5c0 5 9 9 9 9s9-4 9-9V7zM12 8v5m0 4h.01" />
                      </svg>
                    </span>
                    <div className="admin-queue-title">
                      <p className="eyebrow">
                        {item.id} ·{' '}
                        <time dateTime={item.createdAt}>
                          {new Date(item.createdAt).toLocaleString(i18n.resolvedLanguage)}
                        </time>
                      </p>
                      <h2>{t(`enrollment.reports.kind.${item.kind}`)}</h2>
                    </div>
                    <span className={`ui-pill ui-pill--${reportStateVariants[item.state]}`}>
                      {t(`enrollment.reports.state.${item.state}`)}
                    </span>
                  </div>
                  <div className="admin-queue-details">
                    <div>
                      <span>{t('enrollment.reports.reporter')}</span>
                      <strong>{item.reporterName ?? t('ratings.formerAuthor')}</strong>
                    </div>
                    <div>
                      <span>{t('enrollment.reports.course')}</span>
                      <strong>{item.courseTitle}</strong>
                    </div>
                    <div>
                      <span>{t('enrollment.reports.target')}</span>
                      <strong>{item.targetName ?? t('ratings.formerAuthor')}</strong>
                    </div>
                  </div>
                  <div className="admin-queue-reason">
                    <span>{t('enrollment.reports.reason')}</span>
                    <p>{item.reason}</p>
                  </div>
                  {/* A resolved report (visible under the Resolved/All filters) carries no
                      live action — the status pill is the whole story. Only open reports
                      can be resolved, and only an open on_student report can be revoked. */}
                  {item.state === 'open' ? (
                    <div className="admin-queue-actions">
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
                  ) : null}
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

function filterFromKey(key: string): ParticipantReportStateFilter {
  switch (key) {
    case 'resolved':
      return 'resolved';
    case 'all':
      return 'all';
    default:
      return 'open';
  }
}
