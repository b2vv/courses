import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { EnrollmentQueryError } from '../../app/composition/enrollment-queries';
import type { RefundQueueItem } from '../../domain/enrollment/enrollment';
import { Button, Dialog, EmptyState } from '../ui';

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
        <EmptyState title={t('enrollment.refunds.empty')} />
      ) : null}
      {items.length > 0 ? (
        <ul aria-label={t('enrollment.refunds.title')} className="teach-section">
          {items.map((item) => {
            const isPending = actedOrderId === item.orderId && (resolvePending || revokePending);
            return (
              <li key={item.orderId}>
                <article>
                  <h2>{t(`enrollment.refunds.reason.${item.reason}`)}</h2>
                  <p>
                    {t('enrollment.refunds.course')}: {item.courseId}
                  </p>
                  <p>
                    {t('enrollment.refunds.user')}: {item.userId}
                  </p>
                  <p>
                    {t('enrollment.refunds.ordered')}:{' '}
                    {hryvnia(item.orderAmountMinor, i18n.resolvedLanguage)}
                  </p>
                  <p>
                    {t('enrollment.refunds.charged')}:{' '}
                    {item.chargedAmountMinor === null
                      ? t('enrollment.refunds.unknownAmount')
                      : hryvnia(item.chargedAmountMinor, i18n.resolvedLanguage)}
                  </p>
                  {item.receivedAt !== null ? (
                    <time dateTime={item.receivedAt}>
                      {new Date(item.receivedAt).toLocaleString(i18n.resolvedLanguage)}
                    </time>
                  ) : null}
                  <div className="admin-user-actions">
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
                            isPending={isPending}
                            onPress={() => void revoke(item, close)}
                          >
                            {t('enrollment.refunds.revoke')}
                          </Button>
                        </>
                      )}
                    </Dialog>
                    <Button
                      variant="secondary"
                      isPending={isPending}
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
