import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { OwnReview, PublicReviewPage } from '../../contracts/course';
import { formatDate } from '../format-timestamp';
import { Button, Dialog, RadioGroup, TextArea } from '../ui';
import { Reviews } from '../../modules/catalog';

const maximumRunes = 2000;
const minimumRunes = 20;

// The server counts runes after trimming (discussion/domain/review.go); UTF-16 `.length` would let ten
// emoji pass here and fail there.
function runeCount(text: string): number {
  return [...text.trim()].length;
}

export function ReviewBlock({
  own,
  ownError = false,
  onRetryOwn = () => {},
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
  ownError?: boolean | undefined;
  onRetryOwn?: (() => void) | undefined;
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
  const { t, i18n } = useTranslation();
  const locale = i18n.resolvedLanguage ?? 'uk';
  const [editing, setEditing] = useState(false);
  const [stars, setStars] = useState<string>();
  const [body, setBody] = useState('');
  const [error, setError] = useState<string>();
  const [deleteError, setDeleteError] = useState<string>();
  const choices = [1, 2, 3, 4, 5].map((value) => ({
    id: String(value),
    label: t('ratings.stars', { count: value }),
  }));
  const reviewList = (
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
  );
  if (own === undefined)
    return (
      <>
        <section aria-labelledby="your-review-heading">
          <h2 id="your-review-heading">{t('ratings.yourHeading')}</h2>
          {ownError ? (
            <div role="alert">
              <p>{t('ratings.ownFailure')}</p>
              <Button variant="secondary" onPress={onRetryOwn}>
                {t('ratings.retry')}
              </Button>
            </div>
          ) : null}
        </section>
        {reviewList}
      </>
    );
  const eligibility = own.eligibility;
  const refusal =
    eligibility === 'ok'
      ? undefined
      : t(`ratings.eligibility.${eligibility}`, {
          defaultValue: t('ratings.eligibility.review_not_available'),
        });
  const openEdit = () => {
    setStars(String(own.review?.stars ?? ''));
    setBody(own.review?.body ?? '');
    setError(undefined);
    setEditing(true);
  };
  const length = runeCount(body);
  const valid =
    stars !== undefined && stars !== '' && length >= minimumRunes && length <= maximumRunes;
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
          : failure instanceof Error &&
              (failure.message === 'enrollment_required' ||
                failure.message === 'enrollment_blocked')
            ? t(`ratings.errors.${failure.message}`)
            : t('ratings.writeFailure'),
      );
    }
  };
  const remove = async () => {
    setDeleteError(undefined);
    try {
      await onDelete();
    } catch {
      setDeleteError(t('ratings.deleteFailure'));
    }
  };
  return (
    <>
      <section aria-labelledby="your-review-heading">
        <h2 id="your-review-heading">{t('ratings.yourHeading')}</h2>
        {own.review !== null && !editing ? (
          <article>
            <p>
              <span aria-hidden="true">{'★'.repeat(own.review.stars)}</span>
              <span className="sr-only">{t('ratings.stars', { count: own.review.stars })}</span>
            </p>
            <p style={{ whiteSpace: 'pre-line' }}>{own.review.body}</p>
            <p>{t('ratings.progress', { percent: own.review.percent })}</p>
            <time dateTime={own.review.createdAt}>{formatDate(own.review.createdAt, locale)}</time>
            {own.review.editedAt ? (
              <span> {t('ratings.edited', { date: formatDate(own.review.editedAt, locale) })}</span>
            ) : null}
            {/* Editing needs a current right (G4); without it the reason is shown instead of a dead button. */}
            {refusal === undefined ? (
              <Button variant="secondary" onPress={openEdit}>
                {t('ratings.edit')}
              </Button>
            ) : (
              <p>{refusal}</p>
            )}
            <Dialog
              triggerLabel={t('ratings.delete')}
              title={t('ratings.deleteTitle')}
              closeLabel={t('common.close')}
            >
              <p>{t('ratings.deletePrompt')}</p>
              {/* Inside the modal: an alert outside an open modal dialog is not heard by assistive tech. */}
              {deleteError ? <p role="alert">{deleteError}</p> : null}
              <Button variant="danger" onPress={() => void remove()}>
                {t('ratings.delete')}
              </Button>
            </Dialog>
            {error ? <p role="alert">{error}</p> : null}
          </article>
        ) : null}
        {own.review === null && refusal !== undefined ? <p>{refusal}</p> : null}
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
            {/* The visible counter is not a live region (it would speak on every keystroke); only crossing
                the limit is announced. */}
            <span>
              {length} / {maximumRunes}
            </span>
            <output aria-live="polite" className="sr-only">
              {length > maximumRunes ? t('ratings.tooLong') : ''}
            </output>
            {error ? <p role="alert">{error}</p> : null}
            <Button type="submit" isDisabled={!valid}>
              {t('ratings.save')}
            </Button>
          </form>
        ) : null}
      </section>
      {reviewList}
    </>
  );
}
