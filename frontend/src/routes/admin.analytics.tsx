import { useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { canAdministerAccounts } from '../domain/identity/identity';
import {
  AnalyticsFailure,
  AnalyticsPending,
  AnalyticsView,
} from '../presentation/admin/AnalyticsView';

export const Route = createFileRoute('/admin/analytics')({
  ssr: false,
  beforeLoad: ({ context }) => {
    if (!canAdministerAccounts(context.adminActor)) throw notFound();
  },
  loader: ({ context, abortController }) =>
    ensureRouteQuery(context.queryClient, context.admin.stats.snapshot(), abortController.signal),
  head: () => ({ meta: [{ title: 'Analytics — CourseForge' }] }),
  component: AdminAnalyticsRoute,
  pendingComponent: AnalyticsPending,
  errorComponent: AnalyticsFailure,
});

function AdminAnalyticsRoute() {
  const { admin } = Route.useRouteContext();
  const { data } = useSuspenseQuery(admin.stats.snapshot());
  return <AnalyticsView snapshot={data} stats={admin.stats} />;
}
