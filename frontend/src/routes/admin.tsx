import { createFileRoute, notFound } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { canAccessAdministration } from '../domain/identity/identity';
import {
  ProtectedRouteFailure,
  ProtectedRoutePending,
} from '../presentation/identity/ProtectedRouteStates';
import { EmptyState, Link } from '../presentation/ui';

export const Route = createFileRoute('/admin')({
  ssr: false,
  beforeLoad: async ({ context, location }) => {
    const user = await requireIdentity(context, location);
    if (!canAccessAdministration(user)) throw notFound();
  },
  component: AdminRoute,
  pendingComponent: ProtectedRoutePending,
  errorComponent: ProtectedRouteFailure,
});

function AdminRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const { t } = useTranslation();
  return (
    <EmptyState title={t('administration.title')}>
      <p>{t('administration.empty')}</p>
      <Link href="/courses">{t('catalog.goToCatalog')}</Link>
    </EmptyState>
  );
}
