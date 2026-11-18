import { expect, it } from '@rstest/core';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router';
import { render, screen } from '@testing-library/react';
import { createI18n, I18nProvider } from '../app/providers/i18n';
import { Route } from './__root';

it.each([
  { pathname: '/admin', locale: 'en', href: '/admin', label: 'Back to administration' },
  {
    pathname: '/admin/missing',
    locale: 'uk',
    href: '/admin',
    label: 'Повернутися до адміністрування',
  },
  { pathname: '/missing', locale: 'en', href: '/courses', label: 'Go to the catalog' },
  { pathname: '/administrator', locale: 'uk', href: '/courses', label: 'Перейти до каталогу' },
] as const)(
  'Ux03 root 404 at $pathname links to $href ($locale)',
  async ({ pathname, locale, href, label }) => {
    const notFoundComponent = Route.options.notFoundComponent;
    if (notFoundComponent === undefined) throw new Error('Root NotFound component is required');
    const root = createRootRoute({ notFoundComponent });
    const index = createRoute({ getParentRoute: () => root, path: '/', component: () => null });
    const router = createRouter({
      routeTree: root.addChildren([index]),
      history: createMemoryHistory({ initialEntries: [pathname] }),
    });
    await router.load();
    render(
      <I18nProvider instance={createI18n(locale)}>
        <RouterProvider router={router} />
      </I18nProvider>,
    );
    const link = await screen.findByRole('link', { name: label });
    expect(link.getAttribute('href')).toBe(href);
    expect(screen.getAllByRole('link')).toHaveLength(1);
  },
);
