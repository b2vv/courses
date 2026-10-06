import { createRouter } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { setupRouterSsrQueryIntegration } from '@tanstack/react-router-ssr-query';
import { getGlobalStartContext } from '@tanstack/react-start';
import { createI18n, I18nProvider } from './app/providers/i18n';
import { createCatalogDependencies } from './app/composition/catalog';
import { createAdminDependencies } from './app/composition/admin';
import { createIdentityDependencies } from './app/composition/identity';
import { createLearnDependencies } from './app/composition/learn';
import { createTeachDependencies } from './app/composition/teach';
import { detectLocale } from './app/composition/locale';
import { createTranscriptHistory } from './app/composition/transcript-history';
import { parseUrlSearch, stringifyUrlSearch } from './app/router/url-search';
import { routeTree } from './routeTree.gen';

export function getRouter() {
  // getGlobalStartContext throws outside a Start request runtime (tests, client
  // hydration), so treat any failure as "no nonce" rather than letting it abort
  // router creation and the SSR route load.
  let requestContext: unknown;
  try {
    requestContext = getGlobalStartContext();
  } catch {
    requestContext = undefined;
  }
  const nonce =
    typeof requestContext === 'object' &&
    requestContext !== null &&
    'nonce' in requestContext &&
    typeof requestContext.nonce === 'string'
      ? requestContext.nonce
      : undefined;
  const i18n = createI18n(detectLocale());
  const queryClient = new QueryClient({
    defaultOptions: { queries: { staleTime: 60_000, retry: false } },
  });
  const router = createRouter({
    routeTree,
    context: {
      queryClient,
      admin: createAdminDependencies(),
      catalog: createCatalogDependencies(),
      identityDependencies: createIdentityDependencies(),
      learn: createLearnDependencies(queryClient),
      teach: createTeachDependencies(),
      // Top level rather than inside `admin`: the editor writes it, and the identity sweep in
      // `__root` clears it, so neither owns it.
      transcriptHistory: createTranscriptHistory(),
    },
    defaultPreload: 'intent',
    parseSearch: parseUrlSearch,
    stringifySearch: stringifyUrlSearch,
    scrollRestoration: true,
    ...(nonce === undefined ? {} : { ssr: { nonce } }),
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
