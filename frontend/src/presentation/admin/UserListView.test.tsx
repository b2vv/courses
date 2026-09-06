import { afterAll, afterEach, beforeAll, describe, expect, it } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import axe from 'axe-core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../adapters/api/api-origin';
import { createApiAdminUserRepository } from '../../adapters/api/admin-user-repository';
import { createUserAdministration } from '../../application/admin/users';
import { createAdminUserQueries } from '../../app/composition/admin-user-queries';
import { ensureRouteQuery } from '../../app/composition/catalog-queries';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { AdminUser, AdminUserPage } from '../../domain/identity/administration';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import { ResetPasswordDialog } from './ResetPasswordDialog';
import { UserListFailure, UserListView } from './UserListView';

const actor: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000003',
  email: 'admin@courseforge.local',
  displayName: 'Administrator',
  role: 'admin',
};
const listedActor: AdminUser = {
  ...actor,
  status: 'active',
  updatedAt: '2026-09-01T00:00:00.000001Z',
};
const listedUser: AdminUser = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Student',
  role: 'student',
  status: 'active',
  updatedAt: '2026-09-01T00:00:00.000001Z',
};
const firstPage: AdminUserPage = {
  items: [listedUser],
  pagination: { limit: 20, offset: 0, total: 21 },
};
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());

function renderList(overrides: Partial<React.ComponentProps<typeof UserListView>> = {}) {
  const props: React.ComponentProps<typeof UserListView> = {
    page: firstPage,
    state: { q: '', page: 1 },
    actor,
    onStatusChange: () => {},
    onSearch: () => {},
    onPageChange: () => {},
    onSetStatus: async () => {},
    onDelete: async () => {},
    onResetPassword: async () => {},
    onAuthenticationRequired: () => {},
    ...overrides,
  };
  return render(<UserListView {...props} />);
}

describe('admin user directory', () => {
  it('P07 renders the directory and completes block, restore, reset and delete actions', async () => {
    const browser = userEvent.setup();
    const statuses: string[] = [];
    const resets: Array<{ id: string; password: string }> = [];
    const deletes: string[] = [];
    const { rerender } = renderList({
      onSetStatus: async (id, status) => {
        statuses.push(`${id}:${status}`);
      },
      onDelete: async (id) => {
        deletes.push(id);
      },
      onResetPassword: async (id, password) => {
        resets.push({ id, password });
      },
    });

    const table = screen.getByRole('table', { name: 'Користувачі' });
    expect(within(table).getByRole('columnheader', { name: 'Електронна пошта' })).toBeTruthy();
    expect(within(table).getByText('Student')).toBeTruthy();
    expect(within(table).getByText('Студент')).toBeTruthy();
    expect(within(table).getByText('Активний')).toBeTruthy();

    await browser.click(screen.getByRole('button', { name: 'Заблокувати Student' }));
    await waitFor(() => expect(statuses).toEqual([`${listedUser.id}:blocked`]));

    rerender(
      <UserListView
        page={{ ...firstPage, items: [{ ...listedUser, status: 'blocked' }] }}
        state={{ q: '', page: 1 }}
        actor={actor}
        onStatusChange={() => {}}
        onSearch={() => {}}
        onPageChange={() => {}}
        onSetStatus={async (id, status) => {
          statuses.push(`${id}:${status}`);
        }}
        onDelete={async (id) => {
          deletes.push(id);
        }}
        onResetPassword={async (id, password) => {
          resets.push({ id, password });
        }}
        onAuthenticationRequired={() => {}}
      />,
    );
    await browser.click(screen.getByRole('button', { name: 'Відновити Student' }));
    expect(statuses.at(-1)).toBe(`${listedUser.id}:active`);

    await browser.click(screen.getByRole('button', { name: 'Скинути пароль Student' }));
    await browser.type(screen.getByLabelText('Новий пароль'), 'fresh password');
    await browser.click(screen.getByRole('button', { name: 'Зберегти новий пароль' }));
    await waitFor(() =>
      expect(resets).toEqual([{ id: listedUser.id, password: 'fresh password' }]),
    );

    await browser.keyboard('{Escape}');
    await browser.click(screen.getByRole('button', { name: 'Видалити Student' }));
    expect(deletes).toEqual([listedUser.id]);
  });

  it('B02 renders a localized empty state and preserves the pagination boundary', async () => {
    let page = 2;
    renderList({
      page: { items: [], pagination: { limit: 20, offset: 0, total: 0 } },
      onPageChange: (next) => {
        page = next;
      },
    });

    expect(screen.getByRole('heading', { name: 'Користувачів не знайдено' })).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
    const pages = screen.getByRole('navigation', { name: 'Сторінки користувачів' });
    expect(within(pages).getByRole('button', { name: 'Назад' }).hasAttribute('disabled')).toBe(
      true,
    );
    expect(within(pages).getByRole('button', { name: 'Далі' }).hasAttribute('disabled')).toBe(true);
    expect(page).toBe(2);
  });

  it('resets the caller password and then requests local sign-out', async () => {
    const browser = userEvent.setup();
    let resets = 0;
    let authenticationRequired = 0;
    render(
      <ResetPasswordDialog
        user={listedActor}
        isSelf
        onReset={async () => {
          resets += 1;
        }}
        onAuthenticationRequired={() => {
          authenticationRequired += 1;
        }}
      />,
    );

    await browser.click(screen.getByRole('button', { name: 'Скинути пароль Administrator' }));
    await browser.type(screen.getByLabelText('Новий пароль'), 'new admin password');
    await browser.click(screen.getByRole('button', { name: 'Зберегти новий пароль' }));

    await waitFor(() => expect(resets).toBe(1));
    expect(authenticationRequired).toBe(1);
    expect(screen.getByRole('status').textContent).toContain('Ви вийдете із системи');
  });

  it('keeps the populated directory and reset dialog free of axe violations', async () => {
    const { container } = renderList();
    expect((await axe.run(container)).violations).toEqual([]);
    await userEvent.click(screen.getByRole('button', { name: 'Скинути пароль Student' }));
    expect((await axe.run(document.body)).violations).toEqual([]);
  });

  it.each([
    { kind: 'degraded', heading: 'User directory is temporarily unavailable' },
    { kind: 'generic', heading: 'Could not load users' },
  ])('renders and retries the $kind load failure through MSW', async ({ kind, heading }) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const users = createAdminUserQueries(createUserAdministration(createApiAdminUserRepository()));
    const query = { q: '', limit: 20, offset: 0 };
    let reads = 0;
    server.use(
      http.get(`${apiOrigin()}/api/v1/admin/users`, () => {
        reads += 1;
        return reads === 1
          ? kind === 'degraded'
            ? HttpResponse.json(
                {
                  error: {
                    code: 'dependency_unavailable',
                    message: 'x',
                    requestId: 'a'.repeat(32),
                  },
                },
                { status: 503 },
              )
            : HttpResponse.json({ items: [{ status: 'unknown' }], pagination: {} })
          : HttpResponse.json({ items: [], pagination: { limit: 20, offset: 0, total: 0 } });
      }),
    );
    const root = createRootRoute({ component: Outlet });
    const list = createRoute({
      getParentRoute: () => root,
      path: '/admin/users',
      loader: ({ abortController }) =>
        ensureRouteQuery(queryClient, users.list(query), abortController.signal),
      component: () => <p>Directory recovered</p>,
      errorComponent: UserListFailure,
    });
    const router = createRouter({
      routeTree: root.addChildren([list]),
      history: createMemoryHistory({ initialEntries: ['/admin/users'] }),
      defaultPendingMs: 0,
      defaultPendingMinMs: 0,
      Wrap: ({ children }) => <I18nProvider instance={createI18n('en')}>{children}</I18nProvider>,
    });

    await router.load();
    render(<RouterProvider router={router} />);
    expect(screen.getByRole('heading', { name: heading })).toBeTruthy();
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByText('Directory recovered')).toBeTruthy();
    expect(reads).toBe(2);
  });
});
