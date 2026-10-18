import { useEffect, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { formatLocale } from '../../../app/providers/format-locale';
import { AdminUserQueryError } from '../../../app/composition/admin-user-queries';
import type { AdminUserListState } from '../../../app/router/admin-search';
import type {
  AccountStatus,
  AdminUser,
  AdminUserPage,
  AdminUserRole,
  AdminUserSort,
  WritableAccountStatus,
} from '../domain/administration';
import type { AuthenticatedUser } from '../../identity';
import {
  Button,
  DataTable,
  EmptyState,
  ErrorState,
  PageHeader,
  Pagination,
  SegmentedControl,
  Select,
  Skeleton,
  TextField,
  Tooltip,
} from '../../../presentation/ui';
import type { DataColumn } from '../../../presentation/ui';
import { administrationErrorMessage, ResetPasswordDialog } from './ResetPasswordDialog';

const statusVariant: Record<AccountStatus, string> = {
  active: 'success',
  blocked: 'danger',
  deleted: 'neutral',
};

// Each sort has one natural direction; a header reflects it and never fakes a toggle.
// `email` stays mapped (a hand-typed `?sort=email` URL still resolves) although the UI
// no longer offers it — the account header sorts by the person's name, the prominent text.
const naturalDirection: Record<AdminUserSort, 'asc' | 'desc'> = {
  email: 'asc',
  name: 'asc',
  recent: 'desc',
};

/** One or two uppercase letters for the avatar monogram, from the resolved display name. */
function initialsOf(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '—';
  const first = parts[0] ?? '';
  const last = parts.length > 1 ? (parts[parts.length - 1] ?? '') : '';
  return ((first.charAt(0) ?? '') + (last.charAt(0) ?? '')).toUpperCase() || '—';
}

type UserListViewProps = {
  page: AdminUserPage;
  state: AdminUserListState;
  actor: AuthenticatedUser;
  onStatusChange: (status: AccountStatus | undefined) => void;
  onRoleChange: (role: AdminUserRole | undefined) => void;
  onSortChange: (sort: AdminUserSort) => void;
  onSearch: (query: string) => void;
  onPageChange: (page: number) => void;
  onClearFilters: () => void;
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
  onClearFilters,
  onSetStatus,
  onDelete,
  onResetPassword,
  onAuthenticationRequired,
}: UserListViewProps) {
  const { t, i18n } = useTranslation();
  const [pendingUser, setPendingUser] = useState<string>();
  // Guards against a double-press while an action is in flight. A ref (not the
  // pendingUser state) so a second press within the same render is caught
  // synchronously; the buttons stay enabled and focused (aria-disabled, never
  // native disabled, which would drop focus to <body>). While any action is in
  // flight every row's action buttons read aria-disabled, so the serialization
  // the ref enforces is visible rather than a silent no-op on the other rows.
  const inFlight = useRef(false);
  const [message, setMessage] = useState<string>();
  const [messageKind, setMessageKind] = useState<'status' | 'alert'>('status');
  const messageRef = useRef<HTMLDivElement>(null);
  // Delete soft-deletes the user, unmounting its action buttons; the browser
  // then drops focus to <body>. When an action leaves focus stranded there,
  // move it to the result message (a focusable live region) so keyboard and
  // screen-reader users are not silently sent to the top of the page. The
  // activeElement check means block/restore (trigger still mounted and focused)
  // are left untouched — only a vanished trigger moves focus.
  useEffect(() => {
    if (!message) return;
    const active = document.activeElement;
    if (active === null || active === document.body) {
      messageRef.current?.focus();
    }
  }, [message]);
  const currentPage = Math.floor(page.pagination.offset / page.pagination.limit) + 1;
  const totalPages = Math.ceil(page.pagination.total / page.pagination.limit);
  const selectedStatus = state.status ?? 'all';
  // An empty list under an active search, status or role filter reads as "no matches,
  // adjust the filters", not "no accounts exist" — the two states offer a different way out.
  const isFiltered = state.q !== '' || state.status !== undefined || state.role !== undefined;
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

  async function runAction(user: AdminUser, action: () => Promise<void>, success: string) {
    if (inFlight.current) return;
    inFlight.current = true;
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
      inFlight.current = false;
    }
  }

  const columns: readonly DataColumn<AdminUser>[] = [
    {
      key: 'name',
      label: t('adminUsers.account'),
      identifier: true,
      render: (user) => {
        const name = user.displayName || t('ratings.formerAuthor');
        return (
          <span className="person-cell">
            <span className="admin-avatar" aria-hidden="true">
              {initialsOf(name)}
            </span>
            <span className="admin-user-ident-text">
              <span className="admin-user-primary">{name}</span>
              <span className="admin-user-secondary">{user.email}</span>
            </span>
          </span>
        );
      },
    },
    {
      key: 'role',
      label: t('adminUsers.role'),
      sortable: false,
      render: (user) => (
        <span className="admin-role-tag" data-role={user.role}>
          {t(`adminUsers.${user.role}`)}
        </span>
      ),
    },
    {
      key: 'status',
      label: t('adminUsers.status'),
      sortable: false,
      render: (user) => (
        <span
          className={`ui-pill ui-pill--${statusVariant[user.status]}`}
          data-status={user.status}
        >
          {t(`adminUsers.${user.status}`)}
        </span>
      ),
    },
    {
      key: 'recent',
      label: t('adminUsers.updated'),
      render: (user) => (
        <time className="admin-user-updated" dateTime={user.updatedAt}>
          {dateFormatter.format(new Date(user.updatedAt))}
        </time>
      ),
    },
  ];

  return (
    <section className="admin-user-list" data-surface="admin">
      <PageHeader
        eyebrow={t('adminUsers.eyebrow')}
        title={t('adminUsers.title')}
        description={t('adminUsers.lede')}
      />
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
        <span className="admin-filter-count">
          {t('adminUsers.userCount', { count: page.pagination.total })}
        </span>
      </div>
      {message && (
        <div className="admin-editor-message" role={messageKind} ref={messageRef} tabIndex={-1}>
          {message}
        </div>
      )}
      <DataTable<AdminUser>
        caption={t('adminUsers.userTable')}
        columns={columns}
        rows={page.items}
        rowId={(user) => user.id}
        actionsLabel={t('adminUsers.actions')}
        actions={(user) => (
          <UserActions
            user={user}
            actor={actor}
            isBusy={pendingUser !== undefined}
            onSetStatus={(status) =>
              runAction(user, () => onSetStatus(user.id, status), t('adminUsers.actionComplete'))
            }
            onDelete={() =>
              runAction(user, () => onDelete(user.id), t('adminUsers.deletedComplete'))
            }
            onReset={(password) => onResetPassword(user.id, password)}
            onAuthenticationRequired={onAuthenticationRequired}
          />
        )}
        sort={{ key: state.sort, dir: naturalDirection[state.sort] }}
        onSortChange={(next) => onSortChange(sortFromKey(next.key))}
        mobileSortLabel={t('adminUsers.sortMobile')}
        empty={
          <EmptyState title={t(isFiltered ? 'adminUsers.emptyFiltered' : 'adminUsers.empty')}>
            <p>{t(isFiltered ? 'adminUsers.emptyFilteredHint' : 'adminUsers.emptyHint')}</p>
            {isFiltered && (
              <Button variant="secondary" onPress={onClearFilters}>
                {t('adminUsers.clearFilters')}
              </Button>
            )}
          </EmptyState>
        }
        footer={
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
        }
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
  isBusy,
  onSetStatus,
  onDelete,
  onReset,
  onAuthenticationRequired,
}: {
  user: AdminUser;
  actor: AuthenticatedUser;
  isBusy: boolean;
  onSetStatus: (status: WritableAccountStatus) => Promise<void>;
  onDelete: () => Promise<void>;
  onReset: (password: string) => Promise<void>;
  onAuthenticationRequired: () => void;
}) {
  const { t } = useTranslation();
  if (user.status === 'deleted') return <span>{t('adminUsers.noActions')}</span>;
  return (
    <>
      <Tooltip
        content={t(user.status === 'active' ? 'adminUsers.blockShort' : 'adminUsers.restoreShort')}
      >
        <Button
          type="button"
          variant="secondary"
          {...(isBusy ? { 'aria-disabled': true } : {})}
          aria-label={t(user.status === 'active' ? 'adminUsers.block' : 'adminUsers.restore', {
            name: user.displayName,
          })}
          onPress={() => void onSetStatus(user.status === 'active' ? 'blocked' : 'active')}
        >
          <UserActionIcon name={user.status === 'active' ? 'block' : 'restore'} />
        </Button>
      </Tooltip>
      <Tooltip content={t('adminUsers.deleteShort')}>
        <Button
          type="button"
          variant="danger"
          {...(isBusy ? { 'aria-disabled': true } : {})}
          aria-label={t('adminUsers.delete', { name: user.displayName })}
          onPress={() => void onDelete()}
        >
          <UserActionIcon name="delete" />
        </Button>
      </Tooltip>
      <ResetPasswordDialog
        trigger={
          <Tooltip content={t('adminUsers.resetPasswordShort')}>
            <Button
              type="button"
              variant="secondary"
              aria-label={t('adminUsers.resetPassword', { name: user.displayName })}
            >
              <UserActionIcon name="resetPassword" />
            </Button>
          </Tooltip>
        }
        user={user}
        isSelf={user.id === actor.id}
        onReset={onReset}
        onAuthenticationRequired={onAuthenticationRequired}
      />
    </>
  );
}

function UserActionIcon({ name }: { name: 'block' | 'restore' | 'delete' | 'resetPassword' }) {
  return (
    <svg
      viewBox="0 0 16 16"
      width="16"
      height="16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {name === 'block' && (
        <>
          <circle cx="8" cy="8" r="5.5" />
          <path d="m4 4 8 8" />
        </>
      )}
      {name === 'restore' && (
        <>
          <path d="M2.5 6A5.5 5.5 0 1 1 3 11M2.5 2.5V6H6" />
          <path d="m6 8 1.5 1.5L10 7" />
        </>
      )}
      {name === 'delete' && (
        <path d="M2.5 4.5h11M6 4.5V2.5h4v2M4 4.5l.5 9h7l.5-9M6.5 7v4M9.5 7v4" />
      )}
      {name === 'resetPassword' && (
        <>
          <circle cx="5" cy="6" r="3" />
          <path d="m7.2 8.2 5.3 5.3 1.5-1.5-1.5-1.5L11 12" />
        </>
      )}
    </svg>
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
    case 'recent':
      return 'recent';
    case 'email':
      return 'email';
    default:
      return 'name';
  }
}
