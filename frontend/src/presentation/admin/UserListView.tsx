import { useState } from 'react';
import type { FormEvent } from 'react';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { AdminUserQueryError } from '../../app/composition/admin-user-queries';
import type { AdminUserListState } from '../../app/router/admin-search';
import type {
  AccountStatus,
  AdminUser,
  AdminUserPage,
  WritableAccountStatus,
} from '../../domain/identity/administration';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import {
  Button,
  EmptyState,
  ErrorState,
  Pagination,
  Skeleton,
  Table,
  Tabs,
  TextField,
} from '../ui';
import { administrationErrorMessage, ResetPasswordDialog } from './ResetPasswordDialog';

type UserListViewProps = {
  page: AdminUserPage;
  state: AdminUserListState;
  actor: AuthenticatedUser;
  onStatusChange: (status: AccountStatus | undefined) => void;
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
  onSearch,
  onPageChange,
  onSetStatus,
  onDelete,
  onResetPassword,
  onAuthenticationRequired,
}: UserListViewProps) {
  const { t } = useTranslation();
  const [pendingUser, setPendingUser] = useState<string>();
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const selectedStatus = state.status ?? 'all';

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

  const results = (
    <div className="admin-list-results">
      {page.items.length === 0 ? (
        <EmptyState title={t('adminUsers.empty')}>
          <p>{t('adminUsers.emptyHint')}</p>
        </EmptyState>
      ) : (
        <Table
          label={t('adminUsers.userTable')}
          columns={[
            { id: 'email', label: t('adminUsers.email') },
            { id: 'name', label: t('adminUsers.name') },
            { id: 'role', label: t('adminUsers.role') },
            { id: 'status', label: t('adminUsers.status') },
            { id: 'actions', label: t('adminUsers.actions') },
          ]}
          rows={page.items.map((user) => ({
            id: user.id,
            cells: [
              <span key="email">{user.email}</span>,
              <span key="name">{user.displayName}</span>,
              <span key="role">{t(`adminUsers.${user.role}`)}</span>,
              <span key="status" className="admin-user-status" data-status={user.status}>
                {t(`adminUsers.${user.status}`)}
              </span>,
              <UserActions
                key="actions"
                user={user}
                actor={actor}
                isPending={pendingUser === user.id}
                onSetStatus={(status) =>
                  runAction(
                    user,
                    () => onSetStatus(user.id, status),
                    t('adminUsers.actionComplete'),
                  )
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
      )}
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
  );
  const tabs = (['all', 'active', 'blocked', 'deleted'] as const).map((status) => ({
    id: status,
    label: t(status === 'all' ? 'adminUsers.allStatuses' : `adminUsers.${status}`),
    content: selectedStatus === status ? results : null,
  }));

  return (
    <section className="admin-user-list">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('adminUsers.eyebrow')}</p>
          <h1>{t('adminUsers.title')}</h1>
        </div>
        <span className="admin-list-count">
          {t('adminUsers.userCount', { count: page.pagination.total })}
        </span>
      </div>
      <UserSearch key={state.q} initialQuery={state.q} onSearch={onSearch} />
      {message && (
        <div className="admin-editor-message" role={messageKind}>
          {message}
        </div>
      )}
      <Tabs
        label={t('adminUsers.statusTabs')}
        selectedKey={selectedStatus}
        onSelectionChange={(key) => onStatusChange(statusFromKey(String(key)))}
        items={tabs}
      />
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
        onPress={() => void onSetStatus(user.status === 'active' ? 'blocked' : 'active')}
      >
        {t(user.status === 'active' ? 'adminUsers.block' : 'adminUsers.restore', {
          name: user.displayName,
        })}
      </Button>
      <Button type="button" variant="danger" isDisabled={isPending} onPress={() => void onDelete()}>
        {t('adminUsers.delete', { name: user.displayName })}
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
