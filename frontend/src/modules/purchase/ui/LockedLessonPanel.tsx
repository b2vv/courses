import { useTranslation } from 'react-i18next';
import type { PurchaseErrorCode } from '../domain/purchase';
import { formatPrice, type DisplayPrice } from '../../../presentation/format-price';
import { Button } from '../../../presentation/ui';

type LockedLessonPanelProps = {
  readonly courseTitle: string;
  readonly lessonTitle: string;
  readonly price: DisplayPrice | null;
  readonly isStarting: boolean;
  readonly error?: PurchaseErrorCode;
  readonly onBuy: () => void;
};

/**
 * A locked lesson shows its place in the course and the way to buy it — never its content:
 * the route renders this instead of the player, so no playback, transcript or discussion
 * request is made for a lesson the viewer cannot open.
 */
export function LockedLessonPanel({
  courseTitle,
  lessonTitle,
  price,
  isStarting,
  error,
  onBuy,
}: LockedLessonPanelProps) {
  const { t, i18n } = useTranslation();
  return (
    <section className="lesson-player locked-lesson" aria-labelledby="locked-lesson-title">
      <p className="eyebrow">{courseTitle}</p>
      <h1 id="locked-lesson-title">{lessonTitle}</h1>
      <div className="locked-lesson-offer">
        <p>{t('learn.purchase.locked')}</p>
        {price !== null && (
          <p className="locked-lesson-price ui-mono">{formatPrice(price, i18n.resolvedLanguage)}</p>
        )}
        <Button type="button" onPress={onBuy} isDisabled={isStarting || price === null}>
          {isStarting ? t('learn.purchase.starting') : t('learn.purchase.buy')}
        </Button>
        {error !== undefined && (
          <p className="locked-lesson-error" role="alert">
            {t(`learn.purchase.errors.${purchaseErrorKey(error)}`)}
          </p>
        )}
      </div>
    </section>
  );
}

export function purchaseErrorKey(
  code: PurchaseErrorCode,
): 'unavailable' | 'session' | 'notPurchasable' | 'notFound' | 'checkoutAlreadyOpened' {
  switch (code) {
    case 'authentication_required':
    case 'origin_forbidden':
      return 'session';
    case 'invalid_request':
      return 'notPurchasable';
    case 'order_not_found':
      return 'notFound';
    case 'checkout_already_opened':
      return 'checkoutAlreadyOpened';
    default:
      return 'unavailable';
  }
}
