import type { ErrorComponentProps } from '@tanstack/react-router';
import { useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ErrorState, Link, EmptyState, Skeleton } from '../../../presentation/ui';
import { LearnQueryError } from '../../../app/composition/learn-queries';

export function ProtectedRoutePending() {
  const { t } = useTranslation();
  return <Skeleton label={t('login.checking')} lines={3} />;
}

export function ProtectedRouteFailure() {
  const { t } = useTranslation();
  const href = useRouterState({ select: (state) => state.location.href });
  return (
    <ErrorState title={t('login.failure')} retryHref={href} retryLabel={t('common.retry')}>
      <p>{t('login.failureHint')}</p>
    </ErrorState>
  );
}

export function LearnRouteFailure({ error }: ErrorComponentProps<unknown>) {
  if (error instanceof LearnQueryError && error.code === 'course_forbidden') {
    return <LearnAccessDenied />;
  }
  return <ProtectedRouteFailure />;
}

export function LearnAccessDenied() {
  const { t } = useTranslation();
  return (
    <EmptyState title={t('learn.accessDenied')}>
      <p>{t('learn.accessDeniedHint')}</p>
      <Link href="/courses">{t('catalog.goToCatalog')}</Link>
    </EmptyState>
  );
}
