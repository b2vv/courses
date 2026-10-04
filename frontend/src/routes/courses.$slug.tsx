import { createFileRoute, notFound } from '@tanstack/react-router';
import { useInfiniteQuery, useMutation, useQuery, useSuspenseQuery } from '@tanstack/react-query';
import { CatalogQueryError, ensureRouteQuery } from '../app/composition/catalog-queries';
import { CourseView } from '../presentation/storefront/CourseView';
import {
  CatalogFailure,
  CatalogLoading,
  CourseNotFound,
} from '../presentation/storefront/CatalogStates';

export const Route = createFileRoute('/courses/$slug')({
  loader: async ({ context, params, abortController }) => {
    try {
      return await ensureRouteQuery(
        context.queryClient,
        context.catalog.detail(params.slug),
        abortController.signal,
      );
    } catch (error) {
      if (
        error instanceof CatalogQueryError &&
        ['course_not_found', 'invalid_slug'].includes(error.code)
      )
        throw notFound();
      throw error;
    }
  },
  head: ({ loaderData }) => ({
    meta: [
      { title: loaderData ? `${loaderData.title} — CourseForge` : 'Курс — CourseForge' },
      { name: 'description', content: loaderData?.summary ?? 'Практичні курси для розробників' },
    ],
  }),
  component: CourseRoute,
  pendingComponent: CatalogLoading,
  errorComponent: CatalogFailure,
  notFoundComponent: CourseNotFound,
});
function CourseRoute() {
  const { slug } = Route.useParams();
  const { catalog, identityDependencies } = Route.useRouteContext();
  const { data } = useSuspenseQuery(catalog.detail(slug));
  const reviews = useInfiniteQuery(catalog.reviews(slug));
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const ownReview = useQuery({
    ...catalog.ownReview(data.id, session.data?.id ?? 'anonymous'),
    enabled: session.data !== null && session.data !== undefined,
  });
  const report = useMutation(catalog.report());
  const page =
    reviews.data === undefined
      ? undefined
      : {
          rating: reviews.data.pages[0]?.rating ?? null,
          items: reviews.data.pages.flatMap((value) => value.items),
          nextCursor: reviews.data.pages.at(-1)?.nextCursor ?? null,
        };
  return (
    <CourseView
      course={data}
      reviews={page}
      reviewsError={reviews.isError}
      isNextPageError={reviews.isFetchNextPageError}
      hasMore={reviews.hasNextPage}
      isLoadingMore={reviews.isFetchingNextPage}
      onLoadMore={() => void reviews.fetchNextPage()}
      onRetry={() => void reviews.refetch()}
      onRetryMore={() => void reviews.fetchNextPage()}
      signedIn={session.data !== null && session.data !== undefined}
      ownReviewId={ownReview.data?.review?.id}
      onReport={(reviewId) => report.mutateAsync({ reviewId })}
    />
  );
}
