import { useTranslation } from 'react-i18next';
import type { Rating } from '../../contracts/course';
import { formatRating } from '../format-rating';

// The visible «★ 4,7 (3)» is hidden from assistive tech and paired with a full sentence: aria-label on a
// generic element (span/p) is prohibited by ARIA and ignored by some screen readers.
export function RatingSummary({
  rating,
  className,
}: {
  rating: Rating | null;
  className?: string;
}) {
  const { t, i18n } = useTranslation();
  if (rating === null) return <span className={className}>{t('ratings.new')}</span>;
  const average = formatRating(rating.average, i18n.resolvedLanguage ?? 'uk');
  return (
    <span className={className}>
      <span aria-hidden="true">{`★ ${average} (${rating.count})`}</span>
      <span className="sr-only">{t('ratings.summary', { average, count: rating.count })}</span>
    </span>
  );
}
