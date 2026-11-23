import { useRouterState } from '@tanstack/react-router';
import type { CatalogQuery } from '../../../contracts/course';
import { ErrorState, Link, Skeleton, StatusPage } from '../../../presentation/ui';
import { useTranslation } from 'react-i18next';
import { CatalogCardsSkeleton, CatalogFrame } from './CatalogView';

/** A failed route load; the retry reloads the same address as a full document (C3b R3). */
function RouteFailure({ title, hint }: { title: string; hint: string }) {
  const { t } = useTranslation();
  const href = useRouterState({ select: (state) => state.location.href });
  return (
    <ErrorState title={title} retryHref={href} retryLabel={t('common.retry')}>
      <p>{hint}</p>
    </ErrorState>
  );
}

export function CatalogFailure() {
  const { t } = useTranslation();
  return <RouteFailure title={t('catalog.failure')} hint={t('catalog.failureHint')} />;
}

/** The course page while its course loads, reachable by moving inside the app (C3b R5). */
export function CourseLoading() {
  const { t } = useTranslation();
  return <Skeleton label={t('course.loading')} lines={5} />;
}

/**
 * Any failure of the course load other than "no such course": the network, a server error, or
 * an answer that does not pass the adapter's validation (C3b R5).
 */
export function CourseFailure() {
  const { t } = useTranslation();
  return <RouteFailure title={t('course.failure')} hint={t('course.failureHint')} />;
}

/** The catalog list while it loads: heading and search stay, cards are placeholders (C2 #10). */
export function CatalogListLoading({ query }: { query: CatalogQuery }) {
  const { t } = useTranslation();
  return (
    <CatalogFrame query={query} controlsInert>
      <CatalogCardsSkeleton label={t('catalog.loading')} />
    </CatalogFrame>
  );
}

/** The catalog list after a failed load: the error sits where the list was (C2 #10). */
export function CatalogListFailure({ query }: { query: CatalogQuery }) {
  return (
    <CatalogFrame query={query}>
      <CatalogFailure />
    </CatalogFrame>
  );
}

export function CourseNotFound() {
  const { t } = useTranslation();
  return (
    // The same page as the site's own 404 (G0), so a missing course is not a second design (C3 #8).
    <StatusPage label="404" title={t('catalog.notFound')} hint={t('catalog.notFoundHint')}>
      {/* A link through the router like the rest of the course page (C3b R5); only the root
          404 and error pages keep plain anchors (G0). */}
      <Link href="/courses" className="ui-button ui-button--primary">
        {t('catalog.goToCatalog')}
      </Link>
    </StatusPage>
  );
}
