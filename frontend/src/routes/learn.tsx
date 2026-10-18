import { createFileRoute, Outlet, useRouterState } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { ProtectedRouteFailure, ProtectedRoutePending } from '../modules/identity';
import { MyLearningList } from '../presentation/learn/MyLearningList';
import { EmptyState, Link } from '../presentation/ui';

export const Route = createFileRoute('/learn')({
  ssr: false,
  beforeLoad: async ({ context, location }) => {
    await requireIdentity(context, location);
  },
  loader: ({ context, location, abortController }) =>
    location.pathname === '/learn' || location.pathname === '/learn/'
      ? ensureRouteQuery(
          context.queryClient,
          context.enrollment.myMemberships(),
          abortController.signal,
        )
      : undefined,
  component: LearnRoute,
  pendingComponent: ProtectedRoutePending,
  errorComponent: ProtectedRouteFailure,
});

function LearnRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  if (pathname !== '/learn' && pathname !== '/learn/') return <Outlet />;
  return <LearnIndex />;
}

function LearnIndex() {
  const { enrollment } = Route.useRouteContext();
  const { data } = useSuspenseQuery(enrollment.myMemberships());
  return <MyLearningList items={data} />;
}

/** The splat route uses this for malformed learning URLs. */
export function LearnEmptyState() {
  const { t } = useTranslation();
  return (
    <EmptyState title={t('learn.title')}>
      <p>{t('learn.empty')}</p>
      <Link href="/courses">{t('catalog.goToCatalog')}</Link>
    </EmptyState>
  );
}
