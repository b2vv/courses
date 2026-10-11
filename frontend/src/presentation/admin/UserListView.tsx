import { useState } from 'react';
import type { FormEvent } from 'react';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../app/providers/format-locale';
import { AdminUserQueryError } from '../../app/composition/admin-user-queries';
import type { AdminUserListState } from '../../app/router/admin-search';
import type {
  AccountStatus,
  AdminUser,
  AdminUserPage,
  AdminUserRole,
  AdminUserSort,
  WritableAccountStatus,
} from '../../domain/identity/administration';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import {
  Button,
  EmptyState,
  ErrorState,
  Pagination,
  SegmentedControl,
  Select,
  Skeleton,
  Table,
  TextField,
} from '../ui';
import { administrationErrorMessage, ResetPasswordDialog } from './ResetPasswordDialog';

type UserListViewProps = {
  page: AdminUserPage;
  state: AdminUserListState;
  actor: AuthenticatedUser;
  onStatusChange: (status: AccountStatus | undefined) => void;
  onRoleChange: (role: AdminUserRole | undefined) => void;
  onSortChange: (sort: AdminUserSort) => void;
  onSearch: (query: string) => void;
  onPageChange: (page: number) => void;
  onSetStatus: (id: string, status: WritableAccountStatus) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onResetPassword: (id: string, password: string) => Promise<void>;
  onAuthenticationRequired: () => void;
};

export function UserListView({
  page,
  state,
  actor,
  onStatusChange,
  onRoleChange,
  onSortChange,
  onSearch,
  onPageChange,
  onSetStatus,
  onDelete,
  onResetPassword,
  onAuthenticationRequired,
}: UserListViewProps) {
  const { t, i18n } = useTranslation();
  const [pendingUser, setPendingUser] = useState<string>();
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const selectedStatus = state.status ?? 'all';
  const dateFormatter = new Intl.DateTimeFormat(formatLocale(i18n.resolvedLanguage), {
    dateStyle: 'short',
    timeStyle: 'short',
  });
  const statusOptions = (['all', 'active', 'blocked', 'deleted'] as const).map((status) => ({
    id: status,
    label: t(status === 'all' ? 'adminUsers.allStatuses' : `adminUsers.${status}`),
  }));
  const roleOptions = (['all', 'student', 'instructor', 'admin'] as const).map((role) => ({
    id: role,
    label: t(role === 'all' ? 'adminUsers.allRoles' : `adminUsers.${role}`),
  }));
  const sortOptions = (['email', 'name', 'recent'] as const).map((sort) => ({
    id: sort,
    label: t(`adminUsers.sort_${sort}`),
  }));

  async function runAction(user: AdminUser, action: () => Promise<void>, success: string) {
    setPendingUser(user.id);
    setMessage(undefined);
    try {
      await action();
      setMessageKind('status');
      setMessage(success);
    } catch (error) {
      const code = error instanceof AdminUserQueryError ? error.code : undefined;
      setMessageKind('alert');
      setMessage(administrationErrorMessage(code, t));
      if (code === 'authentication_required') onAuthenticationRequired();
    } finally {
      setPendingUser(undefined);
    }
  }

  const results =
    page.items.length === 0 ? (
      <EmptyState title={t('adminUsers.empty')}>
        <p>{t('adminUsers.emptyHint')}</p>
      </EmptyState>
    ) : (
      <Table
        label={t('adminUsers.userTable')}
        columns={[
          { id: 'account', label: t('adminUsers.email') },
          { id: 'role', label: t('adminUsers.role') },
          { id: 'status', label: t('adminUsers.status') },
          { id: 'updated', label: t('adminUsers.updated') },
          { id: 'actions', label: t('adminUsers.actions') },
        ]}
        rows={page.items.map((user) => ({
          id: user.id,
          cells: [
            <span key="account" className="admin-user-identity">
              <span className="admin-status-dot" data-status={user.status} aria-hidden="true" />
              <span className="admin-user-ident-text">
                <span className="admin-user-primary">{user.email}</span>
                <span className="admin-user-secondary">{user.displayName}</span>
              </span>
            </span>,
            <span key="role" className="admin-role-tag" data-role={user.role}>
              {t(`adminUsers.${user.role}`)}
            </span>,
            <span key="status" className="admin-user-status" data-status={user.status}>
              {t(`adminUsers.${user.status}`)}
            </span>,
            <time key="updated" className="admin-user-updated" dateTime={user.updatedAt}>
              {dateFormatter.format(new Date(user.updatedAt))}
            </time>,
            <UserActions
              key="actions"
              user={user}
              actor={actor}
              isPending={pendingUser === user.id}
              onSetStatus={(status) =>
                runAction(user, () => onSetStatus(user.id, status), t('adminUsers.actionComplete'))
              }
              onDelete={() =>
                runAction(user, () => onDelete(user.id), t('adminUsers.deletedComplete'))
              }
              onReset={(password) => onResetPassword(user.id, password)}
              onAuthenticationRequired={onAuthenticationRequired}
            />,
          ],
        }))}
        emptyMessage={t('ui.noRecords')}
      />
    );

  return (
    <section className="admin-user-list">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('adminUsers.eyebrow')}</p>
          <h1>{t('adminUsers.title')}</h1>
        </div>
      </div>
      <div className="admin-filter-bar">
        <UserSearch key={state.q} initialQuery={state.q} onSearch={onSearch} />
        <SegmentedControl
          label={t('adminUsers.statusTabs')}
          options={statusOptions}
          selectedKeys={new Set([selectedStatus])}
          onSelectionChange={(keys) => onStatusChange(statusFromKey(String([...keys][0] ?? 'all')))}
        />
        <Select
          label={t('adminUsers.roleFilter')}
          options={roleOptions}
          selectedKey={state.role ?? 'all'}
          onSelectionChange={(key) => onRoleChange(roleFromKey(String(key)))}
        />
        <Select
          label={t('adminUsers.sortBy')}
          options={sortOptions}
          selectedKey={state.sort}
          onSelectionChange={(key) => onSortChange(sortFromKey(String(key)))}
        />
        <span className="admin-filter-count">
          {t('adminUsers.userCount', { count: page.pagination.total })}
        </span>
      </div>
      {message && (
        <div className="admin-editor-message" role={messageKind}>
          {message}
        </div>
      )}
      <div className="admin-list-results">
        {results}
        <Pagination
          label={t('adminUsers.pages')}
          previousLabel={t('ui.back')}
          nextLabel={t('ui.forward')}
          pageLabel={t('ui.pageOf', { page: currentPage, total: totalPages })}
          emptyLabel={t('ui.noPages')}
          page={currentPage}
          totalPages={totalPages}
          onPageChange={onPageChange}
        />
      </div>
    </section>
  );
}

function UserSearch({
  initialQuery,
  onSearch,
}: {
  initialQuery: string;
  onSearch: (query: string) => void;
}) {
  const { t } = useTranslation();
  const [query, setQuery] = useState(initialQuery);
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSearch(query);
  }
  return (
    <search>
      <form className="admin-user-search" onSubmit={submit}>
        <TextField
          label={t('adminUsers.search')}
          placeholder={t('adminUsers.searchPlaceholder')}
          value={query}
          onChange={setQuery}
        />
        <Button type="submit">{t('adminUsers.searchAction')}</Button>
      </form>
    </search>
  );
}

function UserActions({
  user,
  actor,
  isPending,
  onSetStatus,
  onDelete,
  onReset,
  onAuthenticationRequired,
}: {
  user: AdminUser;
  actor: AuthenticatedUser;
  isPending: boolean;
  onSetStatus: (status: WritableAccountStatus) => Promise<void>;
  onDelete: () => Promise<void>;
  onReset: (password: string) => Promise<void>;
  onAuthenticationRequired: () => void;
}) {
  const { t } = useTranslation();
  if (user.status === 'deleted') return <span>{t('adminUsers.noActions')}</span>;
  return (
    <div className="admin-user-actions">
      <Button
        type="button"
        variant="secondary"
        isDisabled={isPending}
        aria-label={t(user.status === 'active' ? 'adminUsers.block' : 'adminUsers.restore', {
          name: user.displayName,
        })}
        onPress={() => void onSetStatus(user.status === 'active' ? 'blocked' : 'active')}
      >
        {t(user.status === 'active' ? 'adminUsers.blockShort' : 'adminUsers.restoreShort')}
      </Button>
      <Button
        type="button"
        variant="danger"
        isDisabled={isPending}
        aria-label={t('adminUsers.delete', { name: user.displayName })}
        onPress={() => void onDelete()}
      >
        {t('adminUsers.deleteShort')}
      </Button>
      <ResetPasswordDialog
        user={user}
        isSelf={user.id === actor.id}
        onReset={onReset}
        onAuthenticationRequired={onAuthenticationRequired}
      />
    </div>
  );
}

export function UserListPending() {
  const { t } = useTranslation();
  return <Skeleton label={t('adminUsers.loading')} lines={5} />;
}

export function UserListFailure({ error, reset }: ErrorComponentProps) {
  const { t } = useTranslation();
  const router = useRouter();
  const degraded = error instanceof AdminUserQueryError && error.code === 'dependency_unavailable';
  return (
    <ErrorState
      title={t(degraded ? 'adminUsers.degraded' : 'adminUsers.failure')}
      retryLabel={t('common.retry')}
      onRetry={() => void router.invalidate().then(reset)}
    >
      <p>{t(degraded ? 'adminUsers.degradedHint' : 'adminUsers.failureHint')}</p>
    </ErrorState>
  );
}

function statusFromKey(key: string): AccountStatus | undefined {
  switch (key) {
    case 'active':
    case 'blocked':
    case 'deleted':
      return key;
    default:
      return undefined;
  }
}

function roleFromKey(key: string): AdminUserRole | undefined {
  switch (key) {
    case 'student':
    case 'instructor':
    case 'admin':
      return key;
    default:
      return undefined;
  }
}

function sortFromKey(key: string): AdminUserSort {
  switch (key) {
    case 'name':
    case 'recent':
      return key;
    default:
      return 'email';
  }
}
