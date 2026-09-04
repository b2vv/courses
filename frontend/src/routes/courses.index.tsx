import { createFileRoute } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { catalogQuery, parseCatalogSearch } from '../app/router/catalog-search';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { CatalogView } from '../presentation/storefront/CatalogView';
import { CatalogFailure, CatalogLoading } from '../presentation/storefront/CatalogStates';

export const Route = createFileRoute('/courses/')({
  validateSearch: parseCatalogSearch,
  // Refusal happens inside the route lane, which answers a not-found with 404.
  // validateSearch cannot do it: the router turns a throw from there into 500.
  beforeLoad: ({ search }) => {
    catalogQuery(search);
  },
  loaderDeps: ({ search }) => search,
  loader: ({ context, deps, abortController }) =>
    ensureRouteQuery(
      context.queryClient,
      context.catalog.list(catalogQuery(deps)),
      abortController.signal,
    ),
  head: () => ({ meta: [{ title: 'Каталог курсів — CourseForge' }] }),
  component: CatalogRoute,
  pendingComponent: CatalogLoading,
  errorComponent: CatalogFailure,
});
function CatalogRoute() {
  const query = catalogQuery(Route.useSearch());
  const { catalog } = Route.useRouteContext();
  const { data } = useSuspenseQuery(catalog.list(query));
  return <CatalogView page={data} query={query} />;
}
