import { createFileRoute, notFound } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
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
  const { catalog } = Route.useRouteContext();
  const { data } = useSuspenseQuery(catalog.detail(slug));
  return <CourseView course={data} />;
}
