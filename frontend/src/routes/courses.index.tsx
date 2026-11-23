import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { catalogQuery, parseCatalogSearch } from '../app/router/catalog-search';
import type { CatalogSearch } from '../app/router/catalog-search';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { CatalogView } from '../modules/catalog';
import { CatalogListFailure, CatalogListLoading } from '../modules/catalog';
import type { CatalogSortChange } from '../modules/catalog';
import type { CatalogQuery } from '../contracts/course';
import { pageTitleHead } from '../app/composition/page-title';

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
  head: ({ match }) => pageTitleHead(match.context.i18n, 'catalog'),
  component: CatalogRoute,
  pendingComponent: CatalogPending,
  errorComponent: CatalogError,
});

/** The query for the frame around a pending or failed list; an unusable address falls back to defaults. */
function frameQuery(search: CatalogSearch): CatalogQuery {
  try {
    return catalogQuery(search);
  } catch {
    return catalogQuery({});
  }
}

function CatalogPending() {
  return <CatalogListLoading query={frameQuery(Route.useSearch())} />;
}

function CatalogError() {
  return <CatalogListFailure query={frameQuery(Route.useSearch())} />;
}

function CatalogRoute() {
  const query = catalogQuery(Route.useSearch());
  const { catalog } = Route.useRouteContext();
  const { data } = useSuspenseQuery(catalog.list(query));
  const navigate = useNavigate({ from: Route.fullPath });
  // A new order applies at once, with the text typed so far, from the first page (C2 #11).
  // Client navigation keeps the page mounted, so focus stays on the sort control.
  const changeSort = ({ q, sort }: CatalogSortChange) => {
    void navigate({ search: { q, sort, limit: query.limit, offset: 0 } });
  };
  return <CatalogView page={data} query={query} onSortChange={changeSort} />;
}
