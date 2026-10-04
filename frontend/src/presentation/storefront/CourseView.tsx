import type { CourseDetail } from '../../domain/course/course';
import type { PublicReviewPage } from '../../domain/course/course';
import { Button, Link } from '../ui';
import { useTranslation } from 'react-i18next';
import { useEffect, useRef, useState } from 'react';
import { RatingSummary } from './RatingSummary';
import { formatDate } from '../format-timestamp';

export function CourseView({
  course,
  reviews,
  reviewsError = false,
  hasMore = false,
  isLoadingMore = false,
  isNextPageError = false,
  onLoadMore = () => {},
  onRetry = () => {},
  onRetryMore = () => {},
  signedIn = false,
  ownReviewId,
  onReport = async () => {},
}: {
  course: CourseDetail;
  reviews?: PublicReviewPage | undefined;
  reviewsError?: boolean | undefined;
  hasMore?: boolean | undefined;
  isLoadingMore?: boolean | undefined;
  isNextPageError?: boolean | undefined;
  onLoadMore?: (() => void) | undefined;
  onRetry?: (() => void) | undefined;
  onRetryMore?: (() => void) | undefined;
  signedIn?: boolean | undefined;
  ownReviewId?: string | undefined;
  onReport?: ((id: string) => Promise<void>) | undefined;
}) {
  const { t } = useTranslation();
  return (
    <article className="course-detail">
      <div className="detail-heading">
        <Link href="/courses">{t('course.all')}</Link>
        <p className="eyebrow">{t('catalog.kind')}</p>
        <h1>{course.title}</h1>
        <p>
          <RatingSummary rating={course.rating ?? null} />
        </p>
        <p className="lede">{course.summary}</p>
      </div>
      <div className="detail-body">
        <div>
          <section>
            <h2>{t('course.about')}</h2>
            <p className="description">{course.description}</p>
          </section>
          {course.program.length > 0 && (
            <section className="program">
              <h2>{t('course.program')}</h2>
              <ol className="course-program">
                {course.program.map((module, moduleIndex) => (
                  <li key={`${module.title}-${moduleIndex}`}>
                    <h3>{module.title}</h3>
                    {module.lessons.length > 0 && (
                      <ol>
                        {module.lessons.map((lesson, lessonIndex) => (
                          <li key={`${lesson.title}-${lessonIndex}`}>{lesson.title}</li>
                        ))}
                      </ol>
                    )}
                  </li>
                ))}
              </ol>
            </section>
          )}
          <Reviews
            reviews={reviews}
            error={reviewsError}
            hasMore={hasMore}
            isLoadingMore={isLoadingMore}
            isNextPageError={isNextPageError}
            onLoadMore={onLoadMore}
            onRetry={onRetry}
            onRetryMore={onRetryMore}
            signedIn={signedIn}
            ownReviewId={ownReviewId}
            onReport={onReport}
          />
        </div>
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          <p>{t('course.learningHint')}</p>
          <Link href={`/learn/${course.slug}`}>{t('course.startLearning')}</Link>
        </aside>
      </div>
    </article>
  );
}

export function Reviews({
  reviews,
  error,
  hasMore,
  isLoadingMore,
  isNextPageError,
  onLoadMore,
  onRetry,
  onRetryMore,
  signedIn,
  ownReviewId,
  onReport,
}: {
  reviews: PublicReviewPage | undefined;
  error: boolean;
  hasMore: boolean;
  isLoadingMore: boolean;
  isNextPageError: boolean;
  onLoadMore: () => void;
  onRetry: () => void;
  onRetryMore: () => void;
  signedIn: boolean;
  ownReviewId: string | undefined;
  onReport: (id: string) => Promise<void>;
}) {
  const { t, i18n } = useTranslation();
  const priorCount = useRef(reviews?.items.length ?? 0);
  const addedReview = useRef<HTMLLIElement>(null);
  const [firstAddedIndex, setFirstAddedIndex] = useState<number>();
  const [announcement, setAnnouncement] = useState('');
  const count = reviews?.items.length ?? 0;
  useEffect(() => {
    const added = count - priorCount.current;
    if (added > 0) {
      setFirstAddedIndex(priorCount.current);
      setAnnouncement(hasMore ? t('ratings.loadedMore', { count: added }) : t('ratings.allShown'));
    }
    priorCount.current = count;
  }, [count, hasMore, t]);
  useEffect(() => {
    if (!hasMore && firstAddedIndex !== undefined) addedReview.current?.focus();
  }, [firstAddedIndex, hasMore]);
  if (error && reviews === undefined)
    return (
      <section aria-labelledby="reviews-heading">
        <h2 id="reviews-heading">{t('ratings.heading')}</h2>
        <div role="alert">
          <p>{t('ratings.failure')}</p>
          <Button onPress={onRetry}>{t('ratings.retry')}</Button>
        </div>
      </section>
    );
  if (reviews === undefined)
    return (
      <section aria-labelledby="reviews-heading">
        <h2 id="reviews-heading">{t('ratings.heading')}</h2>
      </section>
    );
  return (
    <section aria-labelledby="reviews-heading">
      <h2 id="reviews-heading">{t('ratings.heading')}</h2>
      {reviews.items.length === 0 ? (
        <p>{t('ratings.empty')}</p>
      ) : (
        <ol aria-label={t('ratings.list')}>
          {reviews.items.map((review, index) => (
            <ReviewCard
              key={review.id}
              {...(index === firstAddedIndex ? { itemRef: addedReview } : {})}
              review={review}
              signedIn={signedIn}
              ownReviewId={ownReviewId}
              onReport={onReport}
              locale={i18n.resolvedLanguage ?? 'uk'}
            />
          ))}
        </ol>
      )}
      {hasMore ? (
        <>
          {/* isPending, not isDisabled: a native `disabled` on the focused button sends focus to <body>
              (HTML focus fixup) while the next page loads. */}
          <Button variant="secondary" isPending={isLoadingMore} onPress={onLoadMore}>
            {t('ratings.more')}
          </Button>
          {isNextPageError ? (
            <div role="alert">
              <p>{t('ratings.failure')}</p>
              <Button variant="secondary" onPress={onRetryMore}>
                {t('ratings.retry')}
              </Button>
            </div>
          ) : null}
        </>
      ) : null}
      <output aria-live="polite" className="sr-only">
        {isLoadingMore ? t('ratings.loading') : announcement}
      </output>
    </section>
  );
}

function ReviewCard({
  review,
  signedIn,
  ownReviewId,
  onReport,
  locale,
  itemRef,
}: {
  review: PublicReviewPage['items'][number];
  signedIn: boolean;
  ownReviewId: string | undefined;
  onReport: (id: string) => Promise<void>;
  locale: string;
  itemRef?: React.RefObject<HTMLLIElement | null>;
}) {
  const { t } = useTranslation();
  const [reported, setReported] = useState(false);
  const [reportError, setReportError] = useState<string>();
  const [isReporting, setIsReporting] = useState(false);
  return (
    <li ref={itemRef} tabIndex={itemRef === undefined ? undefined : -1}>
      <article>
        <header>
          <strong>{review.authorDisplayName ?? t('ratings.formerAuthor')}</strong>
          <span aria-hidden="true">{'★'.repeat(review.stars)}</span>
          <span className="sr-only">{t('ratings.stars', { count: review.stars })}</span>
          <time dateTime={review.createdAt}>{formatDate(review.createdAt, locale)}</time>
          {review.editedAt ? (
            <span>{t('ratings.edited', { date: formatDate(review.editedAt, locale) })}</span>
          ) : null}
        </header>
        <p style={{ whiteSpace: 'pre-line' }}>{review.body}</p>
        <p>{t('ratings.progress', { percent: review.percent })}</p>
        {signedIn && review.id !== ownReviewId ? (
          <Button
            variant="secondary"
            // Never natively disabled while focused (focus would drop to <body>): pending while sending, then
            // the press is a no-op once the report is sent.
            isPending={isReporting}
            onPress={() =>
              reported
                ? undefined
                : void (setIsReporting(true), onReport(review.id))
                    .then(() => {
                      setReported(true);
                      setReportError(undefined);
                    })
                    .catch((error: unknown) =>
                      setReportError(
                        error instanceof Error && error.message === 'report_own_review'
                          ? t('ratings.ownReportError')
                          : t('ratings.reportError'),
                      ),
                    )
                    .finally(() => setIsReporting(false))
            }
          >
            {isReporting
              ? t('ratings.reporting')
              : reported
                ? t('ratings.reported')
                : t('ratings.report')}
          </Button>
        ) : null}
        {reportError === undefined ? null : <p role="alert">{reportError}</p>}
      </article>
    </li>
  );
}
