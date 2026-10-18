import { useMutation, useQueryClient, useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute, notFound } from '@tanstack/react-router';
import { ensureRouteQuery } from '../app/composition/catalog-queries';
import { sessionQueryKey } from '../app/composition/identity-queries';
import {
  adminUserListState,
  adminUserQuery,
  parseAdminUserSearch,
} from '../app/router/admin-search';
import type {
  AccountStatus,
  AdminUserRole,
  AdminUserSort,
  WritableAccountStatus,
} from '../modules/admin';
import { canAdministerAccounts } from '../modules/identity';
import { UserListFailure, UserListPending, UserListView } from '../modules/admin';

export const Route = createFileRoute('/admin/users/')({
  ssr: false,
  beforeLoad: ({ context }) => {
    if (!canAdministerAccounts(context.adminActor)) throw notFound();
  },
  validateSearch: parseAdminUserSearch,
  loaderDeps: ({ search }) => adminUserListState(search),
  loader: ({ context, deps, abortController }) =>
    ensureRouteQuery(
      context.queryClient,
      context.admin.users.list(adminUserQuery(deps)),
      abortController.signal,
    ),
  head: () => ({ meta: [{ title: 'Користувачі — CourseForge' }] }),
  component: AdminUserListRoute,
  pendingComponent: UserListPending,
  errorComponent: UserListFailure,
});

function AdminUserListRoute() {
  const state = adminUserListState(Route.useSearch());
  const { admin, adminActor } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const { data } = useSuspenseQuery(admin.users.list(adminUserQuery(state)));
  const navigate = Route.useNavigate();
  const setStatus = useMutation(admin.users.setStatus(queryClient));
  const remove = useMutation(admin.users.delete(queryClient));
  const resetPassword = useMutation(admin.users.resetPassword(queryClient));

  function go(next: {
    status?: AccountStatus | undefined;
    role?: AdminUserRole | undefined;
    sort?: AdminUserSort;
    q?: string;
    page?: number;
  }) {
    const status = 'status' in next ? next.status : state.status;
    const role = 'role' in next ? next.role : state.role;
    const sort = next.sort ?? state.sort;
    const q = next.q ?? state.q;
    const page = next.page ?? 1;
    void navigate({
      search: {
        ...(status === undefined ? {} : { status }),
        ...(role === undefined ? {} : { role }),
        ...(sort === 'name' ? {} : { sort }),
        ...(q === '' ? {} : { q }),
        ...(page === 1 ? {} : { page }),
      },
    });
  }

  return (
    <UserListView
      page={data}
      state={state}
      actor={adminActor}
      onStatusChange={(status) => go({ status })}
      onRoleChange={(role) => go({ role })}
      onSortChange={(sort) => go({ sort })}
      onSearch={(q) => go({ q: q.trim() })}
      onPageChange={(page) => go({ page })}
      onClearFilters={() => go({ status: undefined, role: undefined, q: '' })}
      onSetStatus={async (id: string, status: WritableAccountStatus) => {
        await setStatus.mutateAsync({ id, status });
      }}
      onDelete={async (id: string) => {
        await remove.mutateAsync(id);
      }}
      onResetPassword={async (id: string, password: string) => {
        await resetPassword.mutateAsync({ id, password });
      }}
      onAuthenticationRequired={() => queryClient.setQueryData(sessionQueryKey, null)}
    />
  );
}
