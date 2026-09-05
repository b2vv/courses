import { createFileRoute, notFound, Outlet } from '@tanstack/react-router';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { canAccessAdministration } from '../domain/identity/identity';
import {
  ProtectedRouteFailure,
  ProtectedRoutePending,
} from '../presentation/identity/ProtectedRouteStates';

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
  return <Outlet />;
}
