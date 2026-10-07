import { useTranslation } from 'react-i18next';
import type { RosterEntry } from '../../domain/enrollment/enrollment';
import { Button, EmptyState, ProgressBar } from '../ui';

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
}: CourseRosterProps) {
  const { t, i18n } = useTranslation();
  const errorCode = enrollmentErrorKey(error);
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
          {entries.map((entry) => (
            <li key={entry.userId}>
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
                  isDisabled={isPending}
                  onPress={() =>
                    entry.state === 'active' ? onBlock(entry.userId) : onUnblock(entry.userId)
                  }
                >
                  {t(`enrollment.roster.${entry.state === 'active' ? 'block' : 'unblock'}`)}
                </Button>
              </article>
            </li>
          ))}
        </ul>
      ) : null}
      {hasNextPage ? (
        <Button
          variant="secondary"
          isDisabled={isFetchingNextPage}
          onPress={onLoadMore}
        >
          {t('enrollment.roster.more')}
        </Button>
      ) : null}
    </section>
  );
}
