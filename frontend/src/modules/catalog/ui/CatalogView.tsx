import type { ChangeEvent, ReactNode } from 'react';
import type { CatalogQuery, CoursePage, PublishedCourse } from '../../../contracts/course';
import { ValueStrip } from './ValueStrip';
import { Button, EmptyState, Link, TextField } from '../../../presentation/ui';
import { useTranslation } from 'react-i18next';
import { formatPrice } from '../../../presentation/format-price';
import { RatingSummary } from './RatingSummary';

/** The search a sort change applies at once: the typed text and the new order, from offset 0. */
export type CatalogSortChange = { readonly q: string; readonly sort: CatalogQuery['sort'] };

type CatalogFrameProps = {
  query: CatalogQuery;
  /** Total number of courses for the query; omitted while it is unknown (loading, error). */
  total?: number;
  onSortChange?: ((next: CatalogSortChange) => void) | undefined;
  children: ReactNode;
};

/**
 * The parts of the catalog that do not depend on the result: the heading, the one course
 * counter and the search form. Loading and failure render inside the same frame, so the page
 * keeps its heading and the query stays editable whatever the list is doing (C2 #10).
 */
export function CatalogFrame({ query, total, onSortChange, children }: CatalogFrameProps) {
  const { t } = useTranslation();
  const changeSort = (event: ChangeEvent<HTMLSelectElement>) => {
    if (!onSortChange) return;
    const value = event.currentTarget.value === 'title' ? 'title' : 'recent';
    const field = event.currentTarget.form?.elements.namedItem('q');
    const q = field instanceof HTMLInputElement ? field.value.trim() : query.q;
    onSortChange({ q, sort: value });
  };
  return (
    <section className="catalog-view">
      <div className="catalog-heading">
        <p className="eyebrow">{t('catalog.label')}</p>
        <h1>{t('catalog.title')}</h1>
        <p className="lede">{t('catalog.description')}</p>
        {total === undefined ? null : (
          <p className="catalog-count">{t('catalog.count', { count: total })}</p>
        )}
      </div>
      <search>
        <form className="catalog-search" action="/courses" method="get">
          <TextField
            label={t('catalog.search')}
            name="q"
            defaultValue={query.q}
            placeholder={t('catalog.placeholder')}
          />
          <label className="sort-field">
            {t('catalog.sort')}
            <select name="sort" defaultValue={query.sort} onChange={changeSort}>
              <option value="recent">{t('catalog.recent')}</option>
              <option value="title">{t('catalog.byTitle')}</option>
            </select>
          </label>
          <input type="hidden" name="limit" value={query.limit} />
          <Button type="submit">{t('catalog.submit')}</Button>
        </form>
      </search>
      {children}
      <ValueStrip />
    </section>
  );
}

export function CatalogView({
  page,
  query,
  onSortChange,
}: {
  page: CoursePage;
  query: CatalogQuery;
  onSortChange?: (next: CatalogSortChange) => void;
}) {
  const { t } = useTranslation();
  const total = page.pagination.total;
  const last = query.offset + page.items.length;
  // Pages exist only when there is something on them and more than one page of it (C2 #7).
  const paged = page.items.length > 0 && total > query.limit;
  const lastPageOffset = Math.max(0, Math.floor((total - 1) / query.limit) * query.limit);
  return (
    <CatalogFrame query={query} total={total} onSortChange={onSortChange}>
      {page.items.length === 0 ? (
        total === 0 ? (
          <EmptyState title={t('catalog.empty')}>
            <p>{t('catalog.emptyHint')}</p>
            <Link href="/courses" className="ui-button ui-button--primary">
              {t('catalog.resetSearch')}
            </Link>
          </EmptyState>
        ) : (
          // Past the end of a non-empty result the one way out is the last page that has
          // courses, with the same query; "previous" from here would land on another empty
          // page (C2 #1).
          <EmptyState title={t('catalog.emptyPage')}>
            <p>{t('catalog.emptyPageHint')}</p>
            <Link
              href={catalogHref({ ...query, offset: lastPageOffset })}
              className="ui-button ui-button--primary"
            >
              {t('catalog.lastPage')}
            </Link>
          </EmptyState>
        )
      ) : (
        <ul className="course-list">
          {page.items.map((course) => (
            <CourseCard key={course.id} course={course} />
          ))}
        </ul>
      )}
      {paged && (
        <nav className="catalog-pagination" aria-label={t('catalog.pages')}>
          <span>{t('catalog.range', { from: query.offset + 1, to: last, total })}</span>
          <div>
            {query.offset > 0 && (
              <Link
                href={catalogHref({ ...query, offset: Math.max(0, query.offset - query.limit) })}
              >
                {t('catalog.previous')}
              </Link>
            )}
            {query.offset + query.limit < total && query.offset + query.limit <= 10000 && (
              <Link href={catalogHref({ ...query, offset: query.offset + query.limit })}>
                {t('catalog.next')}
              </Link>
            )}
          </div>
        </nav>
      )}
    </CatalogFrame>
  );
}

function catalogHref(query: CatalogQuery) {
  return `/courses?${new URLSearchParams({ q: query.q, sort: query.sort, limit: String(query.limit), offset: String(query.offset) })}`;
}

function CourseCard({ course }: { course: PublishedCourse }) {
  const { t, i18n } = useTranslation();
  const action = course.paid ? t('catalog.buy') : t('catalog.open');
  // A published course with no lessons yet is announced as coming, not as "0 lessons";
  // a rating placeholder would only repeat that it is new (C2 #6).
  const comingSoon = (course.lessonCount ?? 0) === 0;
  return (
    <li>
      <article className="course-card ui-card">
        <div className="course-cover">
          <span className="ui-icon-tile course-glyph" aria-hidden="true">
            {'</>'}
          </span>
          <h2 className="course-cover-title">
            <Link href={`/courses/${course.slug}`}>{course.title}</Link>
          </h2>
        </div>
        <div className="course-card-copy">
          <p>{course.summary}</p>
        </div>
        <footer className="course-card-footer">
          <p className="course-meta">
            <span className="course-price">
              {course.price === null
                ? t('catalog.free')
                : formatPrice(course.price, i18n.resolvedLanguage)}
            </span>
            {comingSoon ? (
              <span className="ui-pill ui-pill--neutral course-soon">
                {t('catalog.comingSoon')}
              </span>
            ) : (
              <>
                <RatingSummary rating={course.rating ?? null} className="course-rating" />
                <span>{t('ratings.lessons', { count: course.lessonCount ?? 0 })}</span>
              </>
            )}
          </p>
          <Link
            href={`/courses/${course.slug}`}
            // The name starts with the visible word, so "click Open" works (WCAG 2.5.3).
            aria-label={t('catalog.actionLabel', { action, title: course.title })}
            className="course-action"
          >
            {action} <span aria-hidden="true">↗</span>
          </Link>
        </footer>
      </article>
    </li>
  );
}

/** Card-shaped placeholders for the list area while the catalog loads (C2 #10). */
export function CatalogCardsSkeleton({ label }: { label: string }) {
  return (
    <output className="catalog-skeleton">
      <span className="ui-visually-hidden">{label}</span>
      <ul className="course-list" aria-hidden="true">
        {[0, 1, 2].map((index) => (
          <li key={index} className="ui-card catalog-skeleton-card">
            <span className="catalog-skeleton-cover" />
            <span className="ui-skeleton-line" />
            <span className="ui-skeleton-line" />
            <span className="ui-skeleton-line catalog-skeleton-short" />
          </li>
        ))}
      </ul>
    </output>
  );
}
