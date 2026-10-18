import { useMemo } from 'react';
import { useInfiniteQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import type { LearnerFilters } from '../modules/teach';
import { learnerFilters, parseLearnerSearch } from '../app/router/learner-search';
import { LearnerOverview } from '../modules/teach';

function toSearch(filters: LearnerFilters) {
  return {
    ...(filters.courseId === undefined ? {} : { courseId: filters.courseId }),
    ...(filters.state === undefined ? {} : { state: filters.state }),
    ...(filters.q === undefined ? {} : { q: filters.q }),
  };
}

export const Route = createFileRoute('/teach/learners')({
  ssr: false,
  validateSearch: parseLearnerSearch,
  loaderDeps: ({ search }) => learnerFilters(search),
  loader: ({ context, deps }) =>
    context.queryClient.prefetchInfiniteQuery(context.teach.learners('teach', deps)),
  component: TeachLearnersRoute,
});

function TeachLearnersRoute() {
  const { teach } = Route.useRouteContext();
  const navigate = Route.useNavigate();
  const filters = learnerFilters(Route.useSearch());
  const learners = useInfiniteQuery(teach.learners('teach', filters));
  const entries = useMemo(
    () => learners.data?.pages.flatMap((page) => page.entries) ?? [],
    [learners.data],
  );
  return (
    <LearnerOverview
      entries={entries}
      isLoading={learners.isLoading}
      error={learners.error}
      hasNextPage={learners.hasNextPage}
      isFetchingNextPage={learners.isFetchingNextPage}
      filters={filters}
      onRetry={() => void learners.refetch()}
      onLoadMore={() => void learners.fetchNextPage()}
      onFiltersChange={(next) => void navigate({ search: toSearch(next) })}
    />
  );
}
