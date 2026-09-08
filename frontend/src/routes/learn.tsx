import { createFileRoute, Outlet, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import {
  ProtectedRouteFailure,
  ProtectedRoutePending,
} from '../presentation/identity/ProtectedRouteStates';
import { EmptyState, Link } from '../presentation/ui';

export const Route = createFileRoute('/learn')({
  ssr: false,
  beforeLoad: async ({ context, location }) => {
    await requireIdentity(context, location);
  },
  component: LearnRoute,
  pendingComponent: ProtectedRoutePending,
  errorComponent: ProtectedRouteFailure,
});

function LearnRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  if (pathname !== '/learn' && pathname !== '/learn/') return <Outlet />;
  return <LearnEmptyState />;
}

export function LearnEmptyState() {
  const { t } = useTranslation();
  return (
    <EmptyState title={t('learn.title')}>
      <p>{t('learn.empty')}</p>
      <Link href="/courses">{t('catalog.goToCatalog')}</Link>
    </EmptyState>
  );
}
