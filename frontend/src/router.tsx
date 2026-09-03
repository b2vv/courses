import { createRouter } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { setupRouterSsrQueryIntegration } from '@tanstack/react-router-ssr-query';
import { createI18n, I18nProvider } from './app/providers/i18n';
import { createCatalogDependencies } from './app/composition/catalog';
import { detectLocale } from './app/composition/locale';
import { parseUrlSearch, stringifyUrlSearch } from './app/router/url-search';
import { routeTree } from './routeTree.gen';

export function getRouter() {
  const i18n = createI18n(detectLocale());
  const queryClient = new QueryClient({
    defaultOptions: { queries: { staleTime: 60_000, retry: false } },
  });
  const router = createRouter({
    routeTree,
    context: { queryClient, catalog: createCatalogDependencies() },
    defaultPreload: 'intent',
    parseSearch: parseUrlSearch,
    stringifySearch: stringifyUrlSearch,
    scrollRestoration: true,
    Wrap: ({ children }) => <I18nProvider instance={i18n}>{children}</I18nProvider>,
  });
  setupRouterSsrQueryIntegration({ router, queryClient });
  return router;
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }
}
