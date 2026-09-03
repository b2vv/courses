import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { parseCatalogSearch } from '../app/router/catalog-search';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { CatalogView } from '../presentation/storefront/CatalogView';
import { CatalogFailure, CatalogLoading } from '../presentation/storefront/CatalogStates';

export const Route = createFileRoute('/courses/')({
  validateSearch: parseCatalogSearch,
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps, abortController }) =>
    ensureRouteQuery(context.queryClient, context.catalog.list(deps), abortController.signal),
  head: () => ({ meta: [{ title: 'Каталог курсів — CourseForge' }] }),
  component: CatalogRoute,
  pendingComponent: CatalogLoading,
  errorComponent: CatalogFailure,
});
function CatalogRoute() {
  const query = Route.useSearch();
  const { catalog } = Route.useRouteContext();
  const { data } = useSuspenseQuery(catalog.list(query));
  return <CatalogView page={data} query={query} />;
}
