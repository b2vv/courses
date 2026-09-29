import type { CatalogQuery, CoursePage, PublishedCourse } from '../../domain/course/course';
import { Button, EmptyState, Link, TextField } from '../ui';
import { useTranslation } from 'react-i18next';

export function CatalogView({ page, query }: { page: CoursePage; query: CatalogQuery }) {
  const { t } = useTranslation();
  const last = query.offset + page.items.length;
  return (
    <section className="catalog-view">
      <div className="catalog-heading">
        <div>
          <p className="eyebrow">{t('catalog.label')}</p>
          <h1>{t('catalog.title')}</h1>
          <p className="lede">{t('catalog.description')}</p>
        </div>
        <span className="catalog-count ui-pill ui-pill--neutral">
          {t('catalog.count', { count: page.pagination.total })}
        </span>
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
            <select name="sort" defaultValue={query.sort}>
              <option value="recent">{t('catalog.recent')}</option>
              <option value="title">{t('catalog.byTitle')}</option>
            </select>
          </label>
          <input type="hidden" name="limit" value={query.limit} />
          <Button type="submit">{t('catalog.submit')}</Button>
        </form>
      </search>
      {page.items.length === 0 ? (
        <EmptyState title={query.offset > 0 ? t('catalog.emptyPage') : t('catalog.empty')}>
          <p>{t('catalog.emptyHint')}</p>
          <Link href="/courses">{t('catalog.firstPage')}</Link>
        </EmptyState>
      ) : (
        <ul className="course-list">
          {page.items.map((course) => (
            <CourseCard key={course.id} course={course} />
          ))}
        </ul>
      )}
      <nav className="catalog-pagination" aria-label={t('catalog.pages')}>
        <span>
          {page.items.length
            ? t('catalog.range', { from: query.offset + 1, to: last, total: page.pagination.total })
            : t('catalog.zeroRange', { total: page.pagination.total })}
        </span>
        <div>
          {query.offset > 0 && (
            <Link href={catalogHref({ ...query, offset: Math.max(0, query.offset - query.limit) })}>
              {t('catalog.previous')}
            </Link>
          )}
          {query.offset + query.limit < page.pagination.total &&
            query.offset + query.limit <= 10000 && (
              <Link href={catalogHref({ ...query, offset: query.offset + query.limit })}>
                {t('catalog.next')}
              </Link>
            )}
        </div>
      </nav>
    </section>
  );
}

function catalogHref(query: CatalogQuery) {
  return `/courses?${new URLSearchParams({ q: query.q, sort: query.sort, limit: String(query.limit), offset: String(query.offset) })}`;
}
function CourseCard({ course }: { course: PublishedCourse }) {
  const { t, i18n } = useTranslation();
  const action = course.paid ? t('catalog.buy') : t('catalog.open');
  return (
    <li>
      <article className="course-card ui-card">
        <div className="course-cover" aria-hidden="true">
          <span className="ui-icon-tile course-glyph">{'</>'}</span>
          <span className="course-cover-orbit" />
        </div>
        <div className="course-card-copy">
          <p className="eyebrow">{t('catalog.kind')}</p>
          <h2>
            <Link href={`/courses/${course.slug}`}>{course.title}</Link>
          </h2>
          <p>{course.summary}</p>
        </div>
        <footer className="course-card-footer">
          <span className="course-price ui-mono">
            {course.price === null
              ? t('catalog.free')
              : new Intl.NumberFormat(i18n.resolvedLanguage === 'en' ? 'en-US' : 'uk-UA', {
                  style: 'currency',
                  currency: course.price.currency,
                  minimumFractionDigits: 0,
                  maximumFractionDigits: 2,
                }).format(course.price.minor / 100)}
          </span>
          <Link
            href={`/courses/${course.slug}`}
            aria-label={t('catalog.details', { title: course.title })}
            className="course-action"
          >
            {action} <span aria-hidden="true">↗</span>
          </Link>
        </footer>
      </article>
    </li>
  );
}
