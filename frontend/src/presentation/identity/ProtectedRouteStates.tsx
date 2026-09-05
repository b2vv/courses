import { useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ErrorState, Skeleton } from '../ui';

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
