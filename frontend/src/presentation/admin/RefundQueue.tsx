import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EnrollmentQueryError } from '../../app/composition/enrollment-queries';
import type { RefundQueueItem } from '../../domain/enrollment/enrollment';
import { Button, Dialog } from '../ui';

const membershipStates = {
  active: { variant: 'success', label: 'enrollment.roster.state.active' },
  blocked: { variant: 'warning', label: 'enrollment.roster.state.blocked' },
  revoked: { variant: 'neutral', label: 'enrollment.refunds.revoked' },
} satisfies Record<
  NonNullable<RefundQueueItem['membershipState']>,
  { variant: string; label: string }
>;

type RefundQueueProps = {
  readonly items: readonly RefundQueueItem[];
  readonly isLoading: boolean;
  readonly error: unknown;
  readonly resolvePending: boolean;
  readonly revokePending: boolean;
  readonly onRetry: () => void;
  readonly onResolve: (orderId: string) => Promise<void>;
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

export function RefundQueue({
  items,
  isLoading,
  error,
  resolvePending,
  revokePending,
  onRetry,
  onResolve,
  onRevoke,
}: RefundQueueProps) {
  const { t, i18n } = useTranslation();
  const [actedOrderId, setActedOrderId] = useState<string>();
  const [actionError, setActionError] = useState<string>();

  async function resolve(item: RefundQueueItem) {
    setActedOrderId(item.orderId);
    setActionError(undefined);
    try {
      await onResolve(item.orderId);
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

  return (
    <section aria-labelledby="refunds-heading" className="admin-user-list">
      <div className="admin-list-heading">
        <div>
          <p className="eyebrow">{t('enrollment.refunds.eyebrow')}</p>
          <h1 id="refunds-heading">{t('enrollment.refunds.title')}</h1>
        </div>
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
              <path d="M4 6h16v14H4zM4 10h16M15 15h2M7 3l-3 3 3 3M4 6h10" />
            </svg>
          </span>
          <h2>{t('enrollment.refunds.empty')}</h2>
          <p>{t('enrollment.refunds.emptyHint')}</p>
        </div>
      ) : null}
      {items.length > 0 ? (
        <ul aria-label={t('enrollment.refunds.title')} className="admin-queue-list">
          {items.map((item) => {
            const acted = actedOrderId === item.orderId;
            const resolveIsPending = acted && resolvePending;
            const revokeIsPending = acted && revokePending;
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
                        <path d="M4 6h16v14H4zM4 10h16M15 15h2M7 3l-3 3 3 3M4 6h10" />
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
                    {item.membershipState !== null ? (
                      <span
                        className={`ui-pill ui-pill--${membershipStates[item.membershipState].variant}`}
                      >
                        {t(membershipStates[item.membershipState].label)}
                      </span>
                    ) : null}
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
                  <div className="admin-queue-actions">
                    {/* Revoke only for a live grant (active/blocked). When the grant is gone or
                        already revoked, the backend would refuse — so the control is hidden and
                        only resolve remains (A14). */}
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
                              isPending={revokeIsPending}
                              onPress={() => void revoke(item, close)}
                            >
                              {t('enrollment.refunds.revoke')}
                            </Button>
                          </>
                        )}
                      </Dialog>
                    ) : null}
                    <Button
                      variant="secondary"
                      isPending={resolveIsPending}
                      onPress={() => void resolve(item)}
                    >
                      {t('enrollment.refunds.resolve')}
                    </Button>
                  </div>
                </article>
              </li>
            );
          })}
        </ul>
      ) : null}
    </section>
  );
}
