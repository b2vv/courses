import { createFileRoute, notFound, useNavigate } from '@tanstack/react-router';
import { useEffect, useRef } from 'react';
import { useInfiniteQuery, useMutation, useQuery, useSuspenseQuery } from '@tanstack/react-query';
import { CatalogQueryError, ensureRouteQuery } from '../app/composition/catalog-queries';
import { EnrollmentQueryError } from '../app/composition/enrollment-queries';
import { AuthoringQueryError } from '../app/composition/admin-queries';
import { PurchaseQueryError } from '../app/composition/purchase-queries';
import { CourseView, courseAction, type Fact } from '../modules/catalog';
import { CatalogFailure, CatalogLoading, CourseNotFound } from '../modules/catalog';
import { namedPageTitle, pageTitle } from '../app/composition/page-title';
import { asBrandedId } from '../contracts/course';

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
  head: ({ loaderData, match }) => ({
    meta: [
      {
        title: loaderData
          ? namedPageTitle(loaderData.title)
          : pageTitle(match.context.i18n, 'course'),
      },
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
  const { catalog, enrollment, identityDependencies, learn, teach, queryClient } =
    Route.useRouteContext();
  const navigate = useNavigate();
  const { data } = useSuspenseQuery(catalog.detail(slug));
  const reviews = useInfiniteQuery(catalog.reviews(slug));
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const user = session.data ?? null;
  const userId = user?.id ?? 'anonymous';
  // One bounded recovery for a 401 from any protected read (plan §16, the 401 rule): re-read the
  // session once; anonymous ends the cycle (the box becomes the guest's), a failed re-read is
  // "unavailable", and a confirmed session earns exactly one repeat of each read. The session
  // re-read is shared by every read of the same viewer and course (one budget per cycle), and the
  // cycle ends only on a manual retry, another viewer or another course. The reads below never
  // refetch on their own (focus, reconnect), so a settled failure stays settled until the user
  // retries.
  const cycle = `${userId}:${data.id}`;
  // Read only by the query functions, the retry handler and the effect below, never during
  // render. It holds the current cycle only: a new viewer or course starts from nothing, so
  // returning to an earlier viewer never reuses that viewer's old re-read.
  const sessionRechecks = useRef(new Map<string, Promise<unknown>>());
  useEffect(() => {
    sessionRechecks.current.clear();
  }, [cycle]);
  const withSessionRecheck = async <T,>(read: () => Promise<T>): Promise<T> => {
    try {
      return await read();
    } catch (error) {
      if (!isAuthenticationRequired(error)) throw error;
      let pending = sessionRechecks.current.get(cycle);
      if (pending === undefined) {
        pending = queryClient.fetchQuery(identityDependencies.current());
        sessionRechecks.current.set(cycle, pending);
      }
      const current = await pending;
      if (current === null) throw error;
      return read();
    }
  };
  const manualOnly = { retry: false, refetchOnWindowFocus: false, refetchOnReconnect: false };
  const membership = useQuery({
    ...enrollment.membership(data.id, userId),
    queryFn: ({ signal }) => withSessionRecheck(() => enrollment.readMembership(data.id, signal)),
    enabled: user !== null,
    ...manualOnly,
  });
  const authorship = useQuery({
    queryKey: ['course-authorship', userId, data.id] as const,
    queryFn: ({ signal }) =>
      withSessionRecheck(() => teach.readDetail(asBrandedId(data.id, 'course'), signal)).then(
        () => true,
        (error: unknown) => {
          // 403/404 from the owner-scoped editor read mean "not the author"; anything else is
          // a failure the box reports instead of guessing.
          if (
            error instanceof AuthoringQueryError &&
            ['course_forbidden', 'forbidden', 'course_not_found'].includes(error.code)
          )
            return false;
          throw error;
        },
      ),
    enabled: user?.role === 'instructor',
    ...manualOnly,
  });
  const ownReview = useQuery({
    ...catalog.ownReview(data.id, userId),
    enabled: user !== null,
  });
  const enroll = useMutation(enrollment.enroll(queryClient));
  const purchase = useMutation(learn.purchase.start(queryClient));
  const reportLecturer = useMutation(enrollment.reportLecturer());
  const report = useMutation(catalog.report());
  const recheckSession = () => void identityDependencies.refresh(queryClient);
  const action = courseAction({
    course: data,
    viewer: session.isPending
      ? { status: 'pending' }
      : session.isError
        ? { status: 'unavailable' }
        : user === null
          ? { status: 'anonymous' }
          : { status: 'signed-in', role: user.role },
    membership: fact(membership),
    owner: fact(authorship),
  });
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
      action={action}
      reviews={page}
      reviewsError={reviews.isError}
      isNextPageError={reviews.isFetchNextPageError}
      hasMore={reviews.hasNextPage}
      isLoadingMore={reviews.isFetchingNextPage}
      onLoadMore={() => void reviews.fetchNextPage()}
      onRetry={() => void reviews.refetch()}
      onRetryMore={() => void reviews.fetchNextPage()}
      signedIn={user !== null}
      enrolling={enroll.isPending}
      enrollError={enroll.error instanceof EnrollmentQueryError ? enroll.error.code : undefined}
      onEnroll={() =>
        enroll.mutate(data.id, {
          // A lost session turns the box into the guest's sign-in action; the write itself is
          // never repeated automatically.
          onError: (error) => {
            if (isAuthenticationRequired(error)) recheckSession();
          },
        })
      }
      buying={purchase.isPending}
      buyError={
        purchase.error instanceof PurchaseQueryError
          ? purchase.error.code
          : purchase.error === null
            ? undefined
            : 'dependency_unavailable'
      }
      onBuy={() =>
        purchase.mutate(
          { courseId: data.id },
          {
            // Per-call callbacks do not run after this page unmounts, so a late order never
            // navigates away from a page the buyer has already left.
            onSuccess: (order) =>
              void navigate({
                to: '/orders/$orderId',
                params: { orderId: order.id },
                search: { course: data.slug },
              }),
            onError: (error) => {
              if (isAuthenticationRequired(error)) recheckSession();
              if (error instanceof PurchaseQueryError && error.code === 'course_already_owned')
                void membership.refetch();
            },
          },
        )
      }
      onRetryAccess={() => {
        // A manual retry opens a new recovery cycle.
        sessionRechecks.current.delete(cycle);
        void session.refetch();
        if (user !== null) void membership.refetch();
        if (user?.role === 'instructor') void authorship.refetch();
      }}
      onReportLecturer={(reason) => reportLecturer.mutateAsync({ courseId: data.id, reason })}
      ownReviewId={ownReview.data?.review?.id}
      onReport={(reviewId) => report.mutateAsync({ reviewId })}
    />
  );
}

function isAuthenticationRequired(error: unknown): boolean {
  return (
    (error instanceof EnrollmentQueryError ||
      error instanceof AuthoringQueryError ||
      error instanceof PurchaseQueryError) &&
    error.code === 'authentication_required'
  );
}

function fact<T>(query: {
  readonly isPending: boolean;
  readonly isError: boolean;
  readonly data: T | undefined;
}): Fact<T> {
  if (query.isError) return { status: 'unavailable' };
  if (query.isPending || query.data === undefined) return { status: 'pending' };
  return { status: 'known', value: query.data };
}
