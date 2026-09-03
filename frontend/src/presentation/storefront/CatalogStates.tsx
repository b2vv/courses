import { useRouterState } from '@tanstack/react-router';
import { EmptyState, ErrorState, Link, Skeleton } from '../ui';
import { useTranslation } from 'react-i18next';

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
export function CourseNotFound() {
  const { t } = useTranslation();
  return (
    <EmptyState title={t('catalog.notFound')}>
      <p>{t('catalog.notFoundHint')}</p>
      <Link href="/courses">{t('catalog.goToCatalog')}</Link>
    </EmptyState>
  );
}
