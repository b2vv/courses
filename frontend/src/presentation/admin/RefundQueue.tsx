import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EnrollmentQueryError } from '../../app/composition/enrollment-queries';
import type {
  RefundOutcome,
  RefundQueueItem,
  RefundStateFilter,
} from '../../domain/enrollment/enrollment';
import { Button, Dialog, PageHeader, SegmentedControl } from '../ui';

const stateFilters: readonly RefundStateFilter[] = ['all', 'pending', 'refunded', 'declined'];

const outcomeVariants = {
  refunded: 'success',
  declined: 'neutral',
} satisfies Record<RefundOutcome, string>;

// Label for the membership-state context line. `blocked` reads danger across admin, but here the
// grant state is shown as plain context (it gates the revoke control, not a status pill), so only
// the label matters. `revoked` reuses the refund-specific "access revoked" wording.
const membershipLabels = {
  active: 'enrollment.roster.state.active',
  blocked: 'enrollment.roster.state.blocked',
  revoked: 'enrollment.refunds.revoked',
} satisfies Record<NonNullable<RefundQueueItem['membershipState']>, string>;

type RefundQueueProps = {
  readonly items: readonly RefundQueueItem[];
  readonly state: RefundStateFilter;
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly refundPending: boolean;
  readonly declinePending: boolean;
  readonly revokePending: boolean;
  readonly onFilterChange: (state: RefundStateFilter) => void;
  readonly onRetry: () => void;
  readonly onRefund: (orderId: string) => Promise<void>;
  readonly onDecline: (orderId: string) => Promise<void>;
  readonly onRevoke: (orderId: string) => Promise<void>;
};

function errorKey(error: unknown): string {
  const code = error instanceof EnrollmentQueryError ? error.code : undefined;
  switch (code) {
    case 'authentication_required':
    case 'grant_forbidden':
    case 'order_not_found':
    case 'invalid_request':
    case 'membership_conflict':
    case 'grant_transition_invalid':
    case 'dependency_unavailable':
    case 'internal_error':
      return code;
    default:
      return 'generic';
  }
}

function hryvnia(minor: number, language: string | undefined): string {
  return new Intl.NumberFormat(language, { style: 'currency', currency: 'UAH' }).format(
    minor / 100,
  );
}

const refundIcon = <path d="M4 6h16v14H4zM4 10h16M15 15h2M7 3l-3 3 3 3M4 6h10" />;

export function RefundQueue({
  items,
  state,
  isLoading,
  error,
  refundPending,
  declinePending,
  revokePending,
  onFilterChange,
  onRetry,
  onRefund,
  onDecline,
  onRevoke,
}: RefundQueueProps) {
  const { t, i18n } = useTranslation();
  const [actedOrderId, setActedOrderId] = useState<string>();
  const [actionError, setActionError] = useState<string>();

  async function settle(item: RefundQueueItem, action: (orderId: string) => Promise<void>) {
    setActedOrderId(item.orderId);
    setActionError(undefined);
    try {
      await action(item.orderId);
    } catch (actionFailure) {
      setActionError(errorKey(actionFailure));
    }
  }
  async function revoke(item: RefundQueueItem, close: () => void) {
    setActedOrderId(item.orderId);
    setActionError(undefined);
    try {
      await onRevoke(item.orderId);
      close();
    } catch (actionFailure) {
      setActionError(errorKey(actionFailure));
    }
  }

  const pendingCount = items.length;

  return (
    <section aria-labelledby="refunds-heading" className="admin-user-list" data-surface="admin">
      <PageHeader
        eyebrow={t('enrollment.refunds.eyebrow')}
        title={t('enrollment.refunds.title')}
        titleId="refunds-heading"
        description={t('enrollment.refunds.lede')}
        actions={
          state === 'pending' && pendingCount > 0 ? (
            <span className="ui-pill ui-pill--info">
              {t('enrollment.refunds.pending', { count: pendingCount })}
            </span>
          ) : undefined
        }
      />
      <div className="admin-queue-filter">
        <SegmentedControl
          label={t('enrollment.refunds.filterLabel')}
          options={stateFilters.map((value) => ({
            id: value,
            label: t(`enrollment.refunds.filter.${value}`),
          }))}
          selectedKeys={new Set([state])}
          onSelectionChange={(keys) =>
            onFilterChange(filterFromKey(String([...keys][0] ?? 'pending')))
          }
        />
      </div>
      {isLoading ? <output>{t('common.loading')}</output> : null}
      {error ? (
        <div role="alert">
          <p>{t(`enrollment.errors.${errorKey(error)}`)}</p>
          <Button onPress={onRetry}>{t('common.retry')}</Button>
        </div>
      ) : null}
      {actionError ? <p role="alert">{t(`enrollment.errors.${actionError}`)}</p> : null}
      {!isLoading && error === null && items.length === 0 ? (
        <div className="ui-card admin-queue-empty">
          <span className="ui-icon-tile" aria-hidden="true">
            <svg
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              focusable="false"
            >
              {refundIcon}
            </svg>
          </span>
          <h2>
            {t(
              state === 'pending' ? 'enrollment.refunds.empty' : 'enrollment.refunds.emptyFiltered',
            )}
          </h2>
          <p>
            {t(
              state === 'pending'
                ? 'enrollment.refunds.emptyHint'
                : 'enrollment.refunds.emptyFilteredHint',
            )}
          </p>
          {state === 'pending' ? null : (
            <Button variant="secondary" onPress={() => onFilterChange('pending')}>
              {t('enrollment.refunds.showPending')}
            </Button>
          )}
        </div>
      ) : null}
      {items.length > 0 ? (
        <ul aria-label={t('enrollment.refunds.title')} className="admin-queue-list">
          {items.map((item) => {
            const acted = actedOrderId === item.orderId;
            return (
              <li key={item.orderId}>
                <article className="ui-card admin-queue-card">
                  <div className="admin-queue-heading">
                    <span className="ui-icon-tile" aria-hidden="true">
                      <svg
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        focusable="false"
                      >
                        {refundIcon}
                      </svg>
                    </span>
                    <div className="admin-queue-title">
                      <p className="eyebrow">
                        {item.orderId}
                        {item.receivedAt !== null ? (
                          <>
                            {' '}
                            ·{' '}
                            <time dateTime={item.receivedAt}>
                              {new Date(item.receivedAt).toLocaleString(i18n.resolvedLanguage)}
                            </time>
                          </>
                        ) : null}
                      </p>
                      <h2>{t(`enrollment.refunds.reason.${item.reason}`)}</h2>
                    </div>
                    {item.outcome === null ? (
                      <span className="ui-pill ui-pill--warning">
                        {t('enrollment.refunds.state.pending')}
                      </span>
                    ) : (
                      <span className={`ui-pill ui-pill--${outcomeVariants[item.outcome]}`}>
                        {t(`enrollment.refunds.state.${item.outcome}`)}
                      </span>
                    )}
                  </div>
                  <div className="admin-queue-details admin-refund-summary">
                    <div>
                      <span>{t('enrollment.refunds.course')}</span>
                      <strong>{item.courseId}</strong>
                    </div>
                    <div>
                      <span>{t('enrollment.refunds.user')}</span>
                      <strong>{item.userId}</strong>
                    </div>
                    <div>
                      <span>{t('enrollment.refunds.access')}</span>
                      <strong>
                        {item.membershipState === null
                          ? t('enrollment.refunds.noGrant')
                          : t(membershipLabels[item.membershipState])}
                      </strong>
                    </div>
                    <div>
                      <span>{t('enrollment.refunds.ordered')}</span>
                      <strong className="admin-refund-amount">
                        {hryvnia(item.orderAmountMinor, i18n.resolvedLanguage)}
                      </strong>
                    </div>
                    <div>
                      <span>{t('enrollment.refunds.charged')}</span>
                      <strong className="admin-refund-amount">
                        {item.chargedAmountMinor === null
                          ? t('enrollment.refunds.unknownAmount')
                          : hryvnia(item.chargedAmountMinor, i18n.resolvedLanguage)}
                      </strong>
                    </div>
                  </div>
                  {/* A settled order (visible under the Refunded/Declined/All filters) carries no
                      live action — the outcome pill is the whole story. Only a pending order can be
                      refunded, declined, or have its grant revoked. The three are independent:
                      revoking access does not settle the refund, and settling leaves the grant. */}
                  {item.outcome === null ? (
                    <div className="admin-queue-actions">
                      <Button
                        isPending={acted && refundPending}
                        onPress={() => void settle(item, onRefund)}
                      >
                        {t('enrollment.refunds.refund')}
                      </Button>
                      <Button
                        variant="secondary"
                        isPending={acted && declinePending}
                        onPress={() => void settle(item, onDecline)}
                      >
                        {t('enrollment.refunds.decline')}
                      </Button>
                      {/* Revoke only for a live grant (active/blocked). When access is gone or
                          already revoked, the backend would refuse — so the control is hidden. */}
                      {item.membershipState === 'active' || item.membershipState === 'blocked' ? (
                        <Dialog
                          triggerLabel={t('enrollment.refunds.revoke')}
                          title={t('enrollment.refunds.revoke')}
                          closeLabel={t('common.close')}
                        >
                          {(close) => (
                            <>
                              <p>{t('enrollment.refunds.revokeConfirm')}</p>
                              <Button
                                variant="danger"
                                isPending={acted && revokePending}
                                onPress={() => void revoke(item, close)}
                              >
                                {t('enrollment.refunds.revoke')}
                              </Button>
                            </>
                          )}
                        </Dialog>
                      ) : null}
                    </div>
                  ) : null}
                </article>
              </li>
            );
          })}
        </ul>
      ) : null}
    </section>
  );
}

function filterFromKey(key: string): RefundStateFilter {
  switch (key) {
    case 'refunded':
      return 'refunded';
    case 'declined':
      return 'declined';
    case 'all':
      return 'all';
    default:
      return 'pending';
  }
}
