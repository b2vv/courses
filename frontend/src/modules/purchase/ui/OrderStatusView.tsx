import { useTranslation } from 'react-i18next';
import type { CheckoutForm, PurchaseErrorCode, PurchaseOrder } from '../domain/purchase';
import { formatPrice } from '../../../presentation/format-price';
import { Button, Link } from '../../../presentation/ui';
import { purchaseErrorKey } from './LockedLessonPanel';

type OrderStatusViewProps = {
  readonly order: PurchaseOrder | undefined;
  readonly error?: PurchaseErrorCode;
  readonly courseHref: string;
  readonly isRetrying: boolean;
  readonly retryError?: PurchaseErrorCode;
  readonly onRetry: () => void;
  readonly isCheckingOut: boolean;
  readonly checkoutError?: PurchaseErrorCode;
  readonly onCheckout: () => void;
};

export function submitCheckoutForm(checkout: CheckoutForm): void {
  const form = document.createElement('form');
  form.method = 'post';
  form.action = checkout.action;
  form.style.display = 'none';
  for (const field of checkout.fields) {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = field.name;
    input.value = field.value;
    form.append(input);
  }
  document.body.append(form);
  if (typeof form.requestSubmit === 'function') form.requestSubmit();
  else form.submit();
}

export function OrderStatusView({
  order,
  error,
  courseHref,
  isRetrying,
  retryError,
  onRetry,
  isCheckingOut,
  checkoutError,
  onCheckout,
}: OrderStatusViewProps) {
  const { t, i18n } = useTranslation();
  if (order === undefined) {
    if (error === undefined)
      return (
        <section className="order-status" aria-busy="true">
          <h1>{t('learn.purchase.title')}</h1>
          <p>{t('learn.purchase.loading')}</p>
        </section>
      );
    return (
      <section className="order-status" role="alert">
        <h1>{t('learn.purchase.title')}</h1>
        <p>{t(`learn.purchase.errors.${purchaseErrorKey(error)}`)}</p>
        <Link href={courseHref}>{t('learn.purchase.backToCourse')}</Link>
      </section>
    );
  }
  const amount = formatPrice(order.amount, i18n.resolvedLanguage);
  return (
    <section className="order-status" data-status={order.status}>
      <h1>{t('learn.purchase.title')}</h1>
      <p className="order-status-amount ui-mono">{amount}</p>
      {/* One live region whose text changes, so the outcome is announced once. */}
      <output className="order-status-state">{t(`learn.purchase.status.${order.status}`)}</output>
      {order.status === 'paid' && <Link href={courseHref}>{t('learn.purchase.openCourse')}</Link>}
      {order.status === 'failed' && (
        <div className="order-status-actions">
          <Button type="button" onPress={onRetry} isDisabled={isRetrying}>
            {isRetrying ? t('learn.purchase.starting') : t('learn.purchase.retry')}
          </Button>
          <Link href={courseHref}>{t('learn.purchase.backToCourse')}</Link>
        </div>
      )}
      {order.status === 'pending' && (
        <div className="order-status-actions">
          <Button type="button" onPress={onCheckout} isDisabled={isCheckingOut}>
            {isCheckingOut ? t('learn.purchase.redirecting') : t('learn.purchase.checkout')}
          </Button>
          <Link href={courseHref}>{t('learn.purchase.backToCourse')}</Link>
        </div>
      )}
      {checkoutError !== undefined && (
        <p className="locked-lesson-error" role="alert">
          {t(`learn.purchase.errors.${purchaseErrorKey(checkoutError)}`)}
        </p>
      )}
      {retryError !== undefined && (
        <p className="locked-lesson-error" role="alert">
          {t(`learn.purchase.errors.${purchaseErrorKey(retryError)}`)}
        </p>
      )}
    </section>
  );
}
