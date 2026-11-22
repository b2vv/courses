import { createFileRoute, notFound, Outlet } from '@tanstack/react-router';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { canAccessAdministration } from '../modules/identity';
import { ProtectedRouteFailure, ProtectedRoutePending } from '../modules/identity';
import { pageTitleHead } from '../app/composition/page-title';

export const Route = createFileRoute('/teach')({
  ssr: false,
  beforeLoad: async ({ context, location }) => {
    const user = await requireIdentity(context, location);
    if (!canAccessAdministration(user)) throw notFound();
    return { teachActor: user };
  },
  head: ({ match }) => pageTitleHead(match.context.i18n, 'teach'),
  component: TeachRoute,
  pendingComponent: ProtectedRoutePending,
  errorComponent: ProtectedRouteFailure,
});

function TeachRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  return <Outlet />;
}
