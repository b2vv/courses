import { expect, it } from '@rstest/core';
import { QueryClient, QueryClientProvider, queryOptions } from '@tanstack/react-query';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import { sessionQueryKey } from '../app/composition/identity-queries';
import { createI18n, I18nProvider } from '../app/providers/i18n';
import { Route } from './__root';

const student = {
  id: '10000000-0000-4000-8000-000000000001',
  displayName: 'Студентка Курсу',
  role: 'student',
  emailVerified: true,
} as const;

it.each([
  {
    pathname: '/admin',
    locale: 'en',
    signedIn: false,
    links: [['Back to administration', '/admin']],
  },
  {
    pathname: '/admin/missing',
    locale: 'uk',
    signedIn: true,
    links: [['Повернутися до адміністрування', '/admin']],
  },
  {
    pathname: '/missing',
    locale: 'en',
    signedIn: false,
    links: [
      ['Go to the catalog', '/courses'],
      ['Go to the home page', '/'],
    ],
  },
  {
    pathname: '/administrator',
    locale: 'uk',
    signedIn: false,
    links: [
      ['Перейти до каталогу', '/courses'],
      ['На головну', '/'],
    ],
  },
  // A signed-in viewer is sent back to their own courses first (G0 #7).
  {
    pathname: '/missing',
    locale: 'uk',
    signedIn: true,
    links: [
      ['Моє навчання', '/learn'],
      ['Каталог', '/courses'],
    ],
  },
] as const)(
  'Ux03 G0-6 root 404 at $pathname (signed in: $signedIn, $locale) offers its ways back',
  async ({ pathname, locale, signedIn, links }) => {
    const notFoundComponent = Route.options.notFoundComponent;
    if (notFoundComponent === undefined) throw new Error('Root NotFound component is required');
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    queryClient.setQueryData(sessionQueryKey, signedIn ? student : null);
    const context = {
      identityDependencies: {
        current: () =>
          queryOptions({
            queryKey: sessionQueryKey,
            queryFn: async () => null,
            staleTime: Infinity,
          }),
      },
    };
    const root = createRootRoute({ notFoundComponent });
    const index = createRoute({ getParentRoute: () => root, path: '/', component: () => null });
    const router = createRouter({
      routeTree: root.addChildren([index]),
      history: createMemoryHistory({ initialEntries: [pathname] }),
      context,
    });
    await router.load();
    render(
      <QueryClientProvider client={queryClient}>
        <I18nProvider instance={createI18n(locale)}>
          <RouterProvider router={router} />
        </I18nProvider>
      </QueryClientProvider>,
    );
    await screen.findByRole('heading', { level: 1 });
    const rendered = screen.getAllByRole('link');
    expect(rendered.map((link) => [link.textContent, link.getAttribute('href')])).toEqual(
      links.map(([label, href]) => [label, href]),
    );
    // Every way back is drawn as a button, the first as the primary one.
    expect(rendered[0]?.className).toContain('ui-button--primary');
  },
);
