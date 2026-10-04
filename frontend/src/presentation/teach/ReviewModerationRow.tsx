import { useTranslation } from 'react-i18next';
import type { ModerationQueueItem } from '../../domain/teach/teach';
import { Button } from '../ui';

// A reported review in the instructor queue. Only an admin can remove it (G9); the course owner can only
// dismiss the reports.
export function ReviewModerationRow({
  item,
  busy,
  admin,
  onDismiss,
  onRemove,
}: {
  readonly item: ModerationQueueItem & {
    readonly kind: 'review';
    readonly review: NonNullable<ModerationQueueItem['review']>;
  };
  readonly busy: boolean;
  readonly admin: boolean;
  readonly onDismiss: () => void;
  readonly onRemove: () => void;
}) {
  const { t } = useTranslation();
  return (
    <li>
      <article>
        <p>
          {item.review.authorDisplayName ?? t('ratings.formerAuthor')}:{' '}
          <span aria-hidden="true">{'★'.repeat(item.review.stars)}</span>
          <span className="sr-only">{t('ratings.stars', { count: item.review.stars })}</span>
        </p>
        <p>{item.review.body}</p>
        <p>{t('teach.reviewReports', { count: item.openReports })}</p>
        <div className="discussion-actions">
          <Button variant="secondary" isDisabled={busy} onPress={onDismiss}>
            {t('teach.dismissReviewReport')}
          </Button>
          {admin ? (
            <Button variant="danger" isDisabled={busy} onPress={onRemove}>
              {t('teach.removeReview')}
            </Button>
          ) : null}
        </div>
      </article>
    </li>
  );
}
