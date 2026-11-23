import { useRouterState } from '@tanstack/react-router';
import type { CatalogQuery } from '../../../contracts/course';
import { ErrorState, Skeleton, StatusPage } from '../../../presentation/ui';
import { useTranslation } from 'react-i18next';
import { CatalogCardsSkeleton, CatalogFrame } from './CatalogView';

export function CatalogLoading() {
  const { t } = useTranslation();
  return <Skeleton label={t('catalog.loading')} lines={5} />;
}
export function CatalogFailure() {
  const { t } = useTranslation();
  const href = useRouterState({ select: (state) => state.location.href });
  return (
    <ErrorState title={t('catalog.failure')} retryHref={href} retryLabel={t('common.retry')}>
      <p>{t('catalog.failureHint')}</p>
    </ErrorState>
  );
}

/** The catalog list while it loads: heading and search stay, cards are placeholders (C2 #10). */
export function CatalogListLoading({ query }: { query: CatalogQuery }) {
  const { t } = useTranslation();
  return (
    <CatalogFrame query={query}>
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
      <a className="ui-button ui-button--primary" href="/courses">
        {t('catalog.goToCatalog')}
      </a>
    </StatusPage>
  );
}
