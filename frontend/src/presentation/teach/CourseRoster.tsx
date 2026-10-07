import { useTranslation } from 'react-i18next';
import type { RosterEntry } from '../../domain/enrollment/enrollment';
import { Button, EmptyState, ProgressBar } from '../ui';
import { ReportDialog } from '../enrollment/ReportDialog';
import { useEffect, useRef, useState } from 'react';

type CourseRosterProps = {
  readonly entries: readonly RosterEntry[];
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly hasNextPage: boolean;
  readonly isFetchingNextPage: boolean;
  readonly isPending: boolean;
  readonly actionError?: string | undefined;
  readonly onRetry: () => void;
  readonly onLoadMore: () => void;
  readonly onBlock: (userId: string) => void;
  readonly onUnblock: (userId: string) => void;
  readonly onReportStudent: (userId: string, reason: string) => Promise<void>;
};

function enrollmentErrorKey(error: unknown): string {
  const code = error instanceof Error ? error.message : typeof error === 'string' ? error : undefined;
  switch (code) {
    case 'membership_conflict':
    case 'enrollment_forbidden':
    case 'invalid_request':
    case 'authentication_required':
    case 'dependency_unavailable':
      return code;
    default:
      return 'generic';
  }
}

export function CourseRoster({
  entries,
  isLoading,
  error,
  hasNextPage,
  isFetchingNextPage,
  isPending,
  actionError,
  onRetry,
  onLoadMore,
  onBlock,
  onUnblock,
  onReportStudent,
}: CourseRosterProps) {
  const { t, i18n } = useTranslation();
  const [reportSent, setReportSent] = useState(false);
  const errorCode = enrollmentErrorKey(error);
  // Per-row pending: only the acted row's button goes busy, and it stays focusable (isPending,
  // not isDisabled — a native disabled on the focused button drops focus to <body>). A single
  // list-wide flag would disable every row and lose the user's place (B-L7b.1). No reset needed:
  // `isPending && actedUserId === entry.userId` falls to false on its own when the request settles.
  const [actedUserId, setActedUserId] = useState<string>();
  // Load-more focus: after the final page arrives, move focus to the first newly-loaded row
  // (mirrors the reviews list / Rt14) so keyboard users are not dropped to the top.
  const priorCount = useRef(entries.length);
  const addedRow = useRef<HTMLLIElement>(null);
  const [firstAddedIndex, setFirstAddedIndex] = useState<number>();
  useEffect(() => {
    if (entries.length - priorCount.current > 0) setFirstAddedIndex(priorCount.current);
    priorCount.current = entries.length;
  }, [entries.length]);
  useEffect(() => {
    if (!hasNextPage && firstAddedIndex !== undefined) addedRow.current?.focus();
  }, [firstAddedIndex, hasNextPage]);
  return (
    <section aria-labelledby="teach-roster-heading" className="teach-section">
      <h2 id="teach-roster-heading">{t('enrollment.roster.title')}</h2>
      {isLoading ? <output>{t('common.loading')}</output> : null}
      {error ? (
        <div role="alert">
          <p>{t(`enrollment.errors.${errorCode}`)}</p>
          <Button onPress={onRetry}>{t('common.retry')}</Button>
        </div>
      ) : null}
      {!isLoading && error === null && entries.length === 0 ? (
        <EmptyState title={t('enrollment.roster.empty')} />
      ) : null}
      {actionError ? <p role="alert">{t(`enrollment.errors.${enrollmentErrorKey(actionError)}`)}</p> : null}
      {entries.length > 0 ? (
        <ul aria-label={t('enrollment.roster.title')}>
          {entries.map((entry, index) => (
            <li
              key={entry.userId}
              {...(index === firstAddedIndex ? { ref: addedRow, tabIndex: -1 } : {})}
            >
              <article>
                <h3>{entry.displayName || t('ratings.formerAuthor')}</h3>
                <p>{t(`enrollment.roster.state.${entry.state}`)}</p>
                <p>
                  {t('enrollment.roster.enrolledAt')}{' '}
                  <time dateTime={entry.enrolledAt}>
                    {new Date(entry.enrolledAt).toLocaleString(i18n.resolvedLanguage)}
                  </time>
                </p>
                <ProgressBar
                  label={t('enrollment.progress', { percent: entry.progressPercent })}
                  value={entry.progressPercent}
                  minValue={0}
                  maxValue={100}
                />
                <Button
                  variant={entry.state === 'active' ? 'danger' : 'secondary'}
                  isPending={isPending && actedUserId === entry.userId}
                  onPress={() => {
                    setActedUserId(entry.userId);
                    if (entry.state === 'active') onBlock(entry.userId);
                    else onUnblock(entry.userId);
                  }}
                >
                  {t(`enrollment.roster.${entry.state === 'active' ? 'block' : 'unblock'}`)}
                </Button>
                <ReportDialog
                  triggerLabel={t('enrollment.report.student')}
                  onReport={(reason) => onReportStudent(entry.userId, reason)}
                  onSent={() => setReportSent(true)}
                />
              </article>
            </li>
          ))}
        </ul>
      ) : null}
      {reportSent ? <output>{t('enrollment.report.sent')}</output> : null}
      {hasNextPage ? (
        <Button
          variant="secondary"
          // isPending, not isDisabled: a native disabled on the focused button drops focus to <body>
          // while the next page loads (B-L7b.1).
          isPending={isFetchingNextPage}
          onPress={onLoadMore}
        >
          {t('enrollment.roster.more')}
        </Button>
      ) : null}
    </section>
  );
}
