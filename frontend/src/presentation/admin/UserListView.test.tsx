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
import type { AdminUser, AdminUserPage } from '../../domain/admin/administration';
import type { AuthenticatedUser } from '../../modules/identity';
import { ResetPasswordDialog } from './ResetPasswordDialog';
import { UserListFailure, UserListView } from './UserListView';

const actor: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000003',
  email: 'admin@courseforge.local',
  displayName: 'Administrator',
  emailVerified: true,
  twoFactorEnabled: false,
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
const secondUser: AdminUser = {
  id: '40000000-0000-4000-8000-000000000002',
  email: 'teacher@courseforge.local',
  displayName: 'Teacher',
  role: 'instructor',
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
    state: { q: '', page: 1, sort: 'name' },
    actor,
    onStatusChange: () => {},
    onRoleChange: () => {},
    onSortChange: () => {},
    onSearch: () => {},
    onPageChange: () => {},
    onClearFilters: () => {},
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
    expect(within(table).getByRole('columnheader', { name: 'Користувач' })).toBeTruthy();
    expect(within(table).getByText('Student')).toBeTruthy();
    expect(within(table).getByText('Студент')).toBeTruthy();
    expect(within(table).getByText('Активний')).toBeTruthy();

    await browser.click(screen.getByRole('button', { name: 'Заблокувати Student' }));
    await waitFor(() => expect(statuses).toEqual([`${listedUser.id}:blocked`]));

    rerender(
      <UserListView
        page={{ ...firstPage, items: [{ ...listedUser, status: 'blocked' }] }}
        state={{ q: '', page: 1, sort: 'name' }}
        actor={actor}
        onStatusChange={() => {}}
        onRoleChange={() => {}}
        onSortChange={() => {}}
        onSearch={() => {}}
        onPageChange={() => {}}
        onClearFilters={() => {}}
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

  it('P07b locks every row action while one action is in flight and stays axe-clean', async () => {
    const browser = userEvent.setup();
    let release!: () => void;
    const { container } = renderList({
      page: { items: [listedUser, secondUser], pagination: { limit: 20, offset: 0, total: 2 } },
      onSetStatus: () =>
        new Promise<void>((resolve) => {
          release = resolve;
        }),
    });

    await browser.click(screen.getByRole('button', { name: 'Заблокувати Student' }));

    // While the one action is in flight, every row's action buttons read
    // aria-disabled: the ref serialization is visible rather than a silent
    // no-op on the untouched row. aria-disabled (not native disabled) keeps
    // the pressed button focusable, and the dimmed state is axe-clean.
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Заблокувати Student' }).getAttribute('aria-disabled'),
      ).toBe('true'),
    );
    expect(
      screen.getByRole('button', { name: 'Видалити Teacher' }).getAttribute('aria-disabled'),
    ).toBe('true');
    expect((await axe.run(container)).violations).toEqual([]);

    release();
    await waitFor(() =>
      expect(
        screen.getByRole('button', { name: 'Заблокувати Student' }).hasAttribute('aria-disabled'),
      ).toBe(false),
    );
  });

  it('invokes the role filter and the column-header sort handlers', async () => {
    const browser = userEvent.setup();
    const roles: Array<string | undefined> = [];
    const sorts: string[] = [];
    renderList({
      onRoleChange: (role) => roles.push(role),
      onSortChange: (sort) => sorts.push(sort),
    });

    await browser.click(screen.getByRole('button', { name: /Фільтр за роллю/ }));
    await browser.click(await screen.findByRole('option', { name: 'Викладач' }));
    expect(roles).toEqual(['instructor']);

    // Sort moved from a filter-bar select to interactive column headers: the
    // "Updated" header sorts by the backend `recent` key, "Account" by `name`.
    await browser.click(screen.getByRole('button', { name: 'Оновлено' }));
    expect(sorts).toEqual(['recent']);
    await browser.click(screen.getByRole('button', { name: 'Користувач' }));
    expect(sorts).toEqual(['recent', 'name']);
  });

  it('B02 renders the filtered empty state with a working clear-filters action', async () => {
    const browser = userEvent.setup();
    let cleared = 0;
    renderList({
      page: { items: [], pagination: { limit: 20, offset: 0, total: 0 } },
      state: { q: 'nobody', page: 1, sort: 'name' },
      onClearFilters: () => {
        cleared += 1;
      },
    });

    expect(screen.getByRole('heading', { name: 'Нічого не знайдено' })).toBeTruthy();
    expect(screen.queryByRole('table')).toBeNull();
    // The shared DataTable replaces its footer (pagination) with the empty slot,
    // so an empty page shows no pagination; clear-filters is the way back.
    expect(screen.queryByRole('navigation', { name: 'Сторінки користувачів' })).toBeNull();
    await browser.click(screen.getByRole('button', { name: 'Скинути фільтри' }));
    expect(cleared).toBe(1);
  });

  it('renders the unfiltered empty state without a clear-filters action', () => {
    renderList({ page: { items: [], pagination: { limit: 20, offset: 0, total: 0 } } });
    expect(screen.getByRole('heading', { name: 'Користувачів не знайдено' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Скинути фільтри' })).toBeNull();
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
