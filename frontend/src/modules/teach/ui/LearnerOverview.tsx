import { useEffect, useRef, useState } from 'react';
import type { Key } from 'react-aria-components';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { LearnerFilters, LearnerRow } from '../domain/teach';
import { Button, PageHeader, ProgressBar, Select } from '../../../presentation/ui';

type LearnerOverviewProps = {
  readonly entries: readonly LearnerRow[];
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly hasNextPage: boolean;
  readonly isFetchingNextPage: boolean;
  readonly filters: LearnerFilters;
  readonly onRetry: () => void;
  readonly onLoadMore: () => void;
  readonly onFiltersChange: (next: LearnerFilters) => void;
};

function learnerErrorKey(error: unknown): string {
  const code =
    error instanceof Error ? error.message : typeof error === 'string' ? error : undefined;
  switch (code) {
    case 'learner_course_forbidden':
    case 'course_forbidden':
    case 'course_not_found':
    case 'invalid_request':
    case 'authentication_required':
    case 'dependency_unavailable':
      return code;
    default:
      return 'generic';
  }
}

/** One or two uppercase letters for the avatar monogram, from the resolved display name. */
function initialsOf(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '—';
  const first = parts[0] ?? '';
  const last = parts.length > 1 ? (parts[parts.length - 1] ?? '') : '';
  return ((first.charAt(0) ?? '') + (last.charAt(0) ?? '')).toUpperCase() || '—';
}

/** Decorative glyphs — the project ships no icon component, so views inline their SVGs. */
function ArrowGlyph() {
  return (
    <svg
      className="learner-enrollment-arrow"
      viewBox="0 0 16 16"
      width="14"
      height="14"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="m6 3 5 5-5 5" />
    </svg>
  );
}

function ClockGlyph() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="13"
      height="13"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <circle cx="8" cy="8" r="6.25" />
      <path d="M8 4.75V8l2.25 1.5" />
    </svg>
  );
}

export function LearnerOverview({
  entries,
  isLoading,
  error,
  hasNextPage,
  isFetchingNextPage,
  filters,
  onRetry,
  onLoadMore,
  onFiltersChange,
}: LearnerOverviewProps) {
  const { t, i18n } = useTranslation();

  // Load-more focus: after the final page arrives, move focus to the first newly-loaded row so
  // keyboard users are not dropped to the top (mirrors CourseRoster).
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

  // exactOptionalPropertyTypes: a dropped filter is an absent key, never `key: undefined`.
  const commitQuery = (raw: string) => {
    const q = raw.trim();
    onFiltersChange({
      ...(filters.courseId === undefined ? {} : { courseId: filters.courseId }),
      ...(filters.state === undefined ? {} : { state: filters.state }),
      ...(q === '' ? {} : { q }),
    });
  };
  const changeState = (key: Key | null) => {
    const state = key === 'active' || key === 'blocked' ? key : undefined;
    onFiltersChange({
      ...(filters.courseId === undefined ? {} : { courseId: filters.courseId }),
      ...(state === undefined ? {} : { state }),
      ...(filters.q === undefined ? {} : { q: filters.q }),
    });
  };
  const clearCourse = () => {
    onFiltersChange({
      ...(filters.state === undefined ? {} : { state: filters.state }),
      ...(filters.q === undefined ? {} : { q: filters.q }),
    });
  };

  const stateOptions = [
    { id: 'all', label: t('teach.learners.stateAll') },
    { id: 'active', label: t('enrollment.roster.state.active') },
    { id: 'blocked', label: t('enrollment.roster.state.blocked') },
  ];

  return (
    <section aria-labelledby="teach-learners-heading" className="teach-section learners-overview">
      <PageHeader
        eyebrow={t('teach.learners.eyebrow')}
        title={t('teach.learners.title')}
        titleId="teach-learners-heading"
        description={t('teach.learners.lede')}
      />

      <search className="admin-filter-bar">
        <form
          className="admin-user-search"
          onSubmit={(event) => {
            event.preventDefault();
            commitQuery(String(new FormData(event.currentTarget).get('q') ?? ''));
          }}
        >
          <label className="ui-field">
            <span>{t('teach.learners.searchLabel')}</span>
            {/* Uncontrolled: `key` remounts the box with the committed value when filters.q changes
                (e.g. browser back), so no prop→state effect is needed. */}
            <input
              key={filters.q ?? ''}
              className="ui-input"
              type="search"
              name="q"
              defaultValue={filters.q ?? ''}
              placeholder={t('teach.learners.searchPlaceholder')}
            />
          </label>
          <Button type="submit" variant="secondary">
            {t('teach.learners.searchSubmit')}
          </Button>
        </form>
        <Select
          label={t('teach.learners.stateLabel')}
          options={stateOptions}
          selectedKey={filters.state ?? 'all'}
          onSelectionChange={changeState}
        />
        {filters.courseId === undefined ? null : (
          <Button variant="secondary" onPress={clearCourse}>
            {t('teach.learners.clearCourse')}
          </Button>
        )}
      </search>

      {isLoading ? <output>{t('common.loading')}</output> : null}
      {error ? (
        <div role="alert">
          <p>{t(`teach.learners.errors.${learnerErrorKey(error)}`)}</p>
          <Button onPress={onRetry}>{t('common.retry')}</Button>
        </div>
      ) : null}
      {!isLoading && error === null && entries.length === 0 ? (
        <p>{t('teach.learners.empty')}</p>
      ) : null}

      {entries.length > 0 ? (
        <ul aria-label={t('teach.learners.title')} className="learner-grid">
          {entries.map((entry, index) => (
            <li
              key={entry.userId}
              {...(index === firstAddedIndex ? { ref: addedRow, tabIndex: -1 } : {})}
            >
              <article className="ui-card learner-card">
                <header className="learner-header">
                  <div className="person-cell">
                    <span className="admin-avatar" aria-hidden="true">
                      {initialsOf(entry.displayName || t('ratings.formerAuthor'))}
                    </span>
                    <h2>{entry.displayName || t('ratings.formerAuthor')}</h2>
                  </div>
                </header>

                <div className="learner-enrollment-zone">
                  <p className="eyebrow">{t('teach.learners.coursesLabel')}</p>
                  <ul aria-label={t('teach.learners.coursesLabel')} className="learners-courses">
                    {entry.courses.map((course) => (
                      <li key={course.courseId}>
                        <div className="learner-enrollment">
                          <Link
                            to="/teach/courses/$courseId"
                            params={{ courseId: course.courseId }}
                          >
                            <span>{course.title}</span>
                            <ArrowGlyph />
                          </Link>
                          <span
                            className={`ui-pill ui-pill--${course.state === 'active' ? 'success' : 'danger'}`}
                          >
                            {t(`enrollment.roster.state.${course.state}`)}
                          </span>
                        </div>
                        <ProgressBar
                          label={t('teach.learners.progressLabel')}
                          value={course.progressPercent}
                          minValue={0}
                          maxValue={100}
                        />
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="learner-meta">
                  <span>
                    <ClockGlyph />
                    {t('teach.learners.lastActivity')}
                  </span>
                  {entry.lastActivityAt === null ? (
                    <strong>{t('teach.learners.neverActive')}</strong>
                  ) : (
                    <strong>
                      <time dateTime={entry.lastActivityAt}>
                        {new Date(entry.lastActivityAt).toLocaleString(i18n.resolvedLanguage)}
                      </time>
                    </strong>
                  )}
                </div>
              </article>
            </li>
          ))}
        </ul>
      ) : null}

      {hasNextPage ? (
        <Button
          variant="secondary"
          // isPending, not isDisabled: a native disabled on the focused button drops focus to
          // <body> while the next page loads (B-L7b.1).
          isPending={isFetchingNextPage}
          onPress={onLoadMore}
        >
          {t('teach.learners.more')}
        </Button>
      ) : null}
    </section>
  );
}
