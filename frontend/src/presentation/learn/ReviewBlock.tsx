import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { OwnReview, PublicReviewPage } from '../../domain/course/course';
import { Button, Dialog, RadioGroup, TextArea } from '../ui';
import { Reviews } from '../storefront/CourseView';

const choices = [1, 2, 3, 4, 5].map((value) => ({ id: String(value), label: `${value} з 5` }));

export function ReviewBlock({
  own,
  reviews,
  reviewsError,
  hasMore,
  loadingMore,
  nextPageError,
  onMore,
  onRetry,
  onRetryMore,
  onSave,
  onDelete,
}: {
  own: OwnReview | undefined;
  reviews: PublicReviewPage | undefined;
  reviewsError: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  nextPageError: boolean;
  onMore: () => void;
  onRetry: () => void;
  onRetryMore: () => void;
  onSave: (stars: number, body: string) => Promise<void>;
  onDelete: () => Promise<void>;
}) {
  const { t } = useTranslation();
  const [editing, setEditing] = useState(false);
  const [stars, setStars] = useState<string>();
  const [body, setBody] = useState('');
  const [error, setError] = useState<string>();
  if (own === undefined)
    return (
      <section>
        <h2>{t('ratings.yourHeading')}</h2>
      </section>
    );
  const eligibility = own.eligibility;
  const openEdit = () => {
    setStars(String(own.review?.stars ?? ''));
    setBody(own.review?.body ?? '');
    setEditing(true);
  };
  const valid = stars !== undefined && body.trim().length >= 20 && body.length <= 2000;
  const save = async () => {
    if (!valid) return;
    setError(undefined);
    try {
      await onSave(Number(stars), body);
      setEditing(false);
    } catch (failure) {
      setError(
        failure instanceof Error && failure.message === 'review_removed'
          ? t('ratings.removed')
          : t('ratings.writeFailure'),
      );
    }
  };
  return (
    <>
      <section aria-labelledby="your-review-heading">
        <h2 id="your-review-heading">{t('ratings.yourHeading')}</h2>
        {own.review !== null && !editing ? (
          <article>
            <p>
              {'★'.repeat(own.review.stars)}{' '}
              <span className="sr-only">{t('ratings.stars', { count: own.review.stars })}</span>
            </p>
            <p>{own.review.body}</p>
            <p>{t('ratings.progress', { percent: own.review.percent })}</p>
            <time dateTime={own.review.createdAt}>{own.review.createdAt}</time>
            {own.review.editedAt ? <span>{t('ratings.edited')}</span> : null}
            <Button variant="secondary" onPress={openEdit}>
              {t('ratings.edit')}
            </Button>
            <Dialog
              triggerLabel={t('ratings.delete')}
              title={t('ratings.deleteTitle')}
              closeLabel={t('common.close')}
            >
              <p>{t('ratings.deletePrompt')}</p>
              <Button variant="danger" onPress={() => void onDelete()}>
                {t('ratings.delete')}
              </Button>
            </Dialog>
          </article>
        ) : null}
        {own.review === null && eligibility !== 'ok' ? (
          <p>
            {t(`ratings.eligibility.${eligibility}`, {
              defaultValue: t('ratings.eligibility.review_not_available'),
            })}
          </p>
        ) : null}
        {eligibility === 'ok' && (own.review === null || editing) ? (
          <form
            onSubmit={(event) => {
              event.preventDefault();
              void save();
            }}
          >
            <RadioGroup
              label={t('ratings.starsLabel')}
              value={stars}
              onChange={setStars}
              options={choices}
            />
            <TextArea
              label={t('ratings.body')}
              value={body}
              onChange={setBody}
              description={t('ratings.minimum')}
            />
            <output aria-live="polite">{body.length} / 2000</output>
            {error ? <p role="alert">{error}</p> : null}
            <Button type="submit" isDisabled={!valid}>
              {t('ratings.save')}
            </Button>
          </form>
        ) : null}
      </section>
      <Reviews
        reviews={reviews}
        error={reviewsError}
        hasMore={hasMore}
        isLoadingMore={loadingMore}
        isNextPageError={nextPageError}
        onLoadMore={onMore}
        onRetry={onRetry}
        onRetryMore={onRetryMore}
        signedIn={false}
        ownReviewId={undefined}
        onReport={async () => {}}
      />
    </>
  );
}
