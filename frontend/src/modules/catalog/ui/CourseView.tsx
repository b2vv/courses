import type { CourseDetail } from '../../../contracts/course';
import type { PublicReviewPage } from '../../../contracts/course';
import type { EnrollmentErrorCode } from '../../enrollment';
import type { PurchaseErrorCode } from '../../purchase';
import { purchaseErrorKey } from '../../purchase';
import { Button, Link } from '../../../presentation/ui';
import { useTranslation } from 'react-i18next';
import { useEffect, useRef, useState } from 'react';
import { RatingSummary } from './RatingSummary';
import { formatDate } from '../../../presentation/format-timestamp';
import { formatPrice } from '../../../presentation/format-price';
import { ReportDialog } from '../../enrollment';
import { lessonTotal, type CourseAction } from './course-action';

export function CourseView({
  course,
  action = { kind: 'checking' },
  reviews,
  reviewsError = false,
  hasMore = false,
  isLoadingMore = false,
  isNextPageError = false,
  onLoadMore = () => {},
  onRetry = () => {},
  onRetryMore = () => {},
  signedIn = false,
  enrolling = false,
  enrollError,
  onEnroll = () => {},
  buying = false,
  buyError,
  onBuy = () => {},
  onRetryAccess = () => {},
  onReportLecturer = () => Promise.resolve(),
  ownReviewId,
  onReport = () => Promise.resolve(),
}: {
  course: CourseDetail;
  /** The action box's state, decided by `courseAction` (plan §16 table A1–A9). */
  action?: CourseAction | undefined;
  reviews?: PublicReviewPage | undefined;
  reviewsError?: boolean | undefined;
  hasMore?: boolean | undefined;
  isLoadingMore?: boolean | undefined;
  isNextPageError?: boolean | undefined;
  onLoadMore?: (() => void) | undefined;
  onRetry?: (() => void) | undefined;
  onRetryMore?: (() => void) | undefined;
  signedIn?: boolean | undefined;
  enrolling?: boolean | undefined;
  enrollError?: EnrollmentErrorCode | undefined;
  onEnroll?: (() => void) | undefined;
  buying?: boolean | undefined;
  buyError?: PurchaseErrorCode | undefined;
  onBuy?: (() => void) | undefined;
  onRetryAccess?: (() => void) | undefined;
  onReportLecturer?: ((reason: string) => Promise<void>) | undefined;
  ownReviewId?: string | undefined;
  onReport?: ((id: string) => Promise<void>) | undefined;
}) {
  const { t } = useTranslation();
  const comingSoon = lessonTotal(course) === 0;
  return (
    <article className="course-detail">
      <div className="detail-heading">
        <Link href="/courses">{t('course.all')}</Link>
        <h1>{course.title}</h1>
        {/* The same line the catalog card shows: coming soon without lessons, else the rating
            (C3 #7). */}
        <p className="detail-meta">
          {comingSoon ? (
            <span className="ui-pill ui-pill--neutral">{t('catalog.comingSoon')}</span>
          ) : (
            <RatingSummary rating={course.rating ?? null} />
          )}
        </p>
        <p className="lede">{course.summary}</p>
      </div>
      {/* The action box comes first in reading order, so on a phone it follows the header
          instead of ending the page; on a wide screen the grid puts it on the right (C3 #5). */}
      <div className="detail-body">
        <CourseActionBox
          course={course}
          action={action}
          enrolling={enrolling}
          enrollError={enrollError}
          onEnroll={onEnroll}
          buying={buying}
          buyError={buyError}
          onBuy={onBuy}
          onRetryAccess={onRetryAccess}
          onReportLecturer={onReportLecturer}
        />
        <div className="detail-main">
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
      </div>
    </article>
  );
}

function CourseActionBox({
  course,
  action,
  enrolling,
  enrollError,
  onEnroll,
  buying,
  buyError,
  onBuy,
  onRetryAccess,
  onReportLecturer,
}: {
  course: CourseDetail;
  action: CourseAction;
  enrolling: boolean;
  enrollError: EnrollmentErrorCode | undefined;
  onEnroll: () => void;
  buying: boolean;
  buyError: PurchaseErrorCode | undefined;
  onBuy: () => void;
  onRetryAccess: () => void;
  onReportLecturer: (reason: string) => Promise<void>;
}) {
  const { t, i18n } = useTranslation();
  const [reportSent, setReportSent] = useState(false);
  const price = (
    <p className="course-box-price">
      {course.price === null ? t('catalog.free') : formatPrice(course.price, i18n.resolvedLanguage)}
    </p>
  );
  const soon = (
    <>
      <span className="ui-pill ui-pill--neutral course-box-soon">{t('catalog.comingSoon')}</span>
      <p>{t('course.comingSoonHint')}</p>
    </>
  );
  const returnTo = encodeURIComponent(`/courses/${course.slug}`);
  switch (action.kind) {
    case 'checking':
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {price}
          <output>{t('course.checking')}</output>
        </aside>
      );
    case 'unavailable':
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {price}
          <div role="alert" className="course-box-alert">
            <p>{t('course.accessFailed')}</p>
            <Button variant="secondary" onPress={onRetryAccess}>
              {t('common.retry')}
            </Button>
          </div>
        </aside>
      );
    case 'owner':
      return (
        <aside className="availability">
          <h2>{t('course.ownerTitle')}</h2>
          {action.comingSoon ? soon : null}
          <a className="ui-button ui-button--primary" href={`/teach/courses/${course.id}`}>
            {t('course.edit')}
          </a>
        </aside>
      );
    case 'admin':
      // G8 (Lo19): an administrator is not a learner; they open the course without enrolling.
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {action.comingSoon ? (
            soon
          ) : (
            <Link href={`/learn/${course.slug}`} className="ui-button ui-button--primary">
              {t('course.startLearning')}
            </Link>
          )}
        </aside>
      );
    case 'member':
      return (
        <aside className="availability">
          <h2>{t('enrollment.member')}</h2>
          {action.blocked && <output>{t('enrollment.blockedNote')}</output>}
          {action.comingSoon ? (
            soon
          ) : (
            <Link href={`/learn/${course.slug}`} className="ui-button ui-button--primary">
              {t('enrollment.startLearning')}
            </Link>
          )}
          {/* A complaint is a rare, secondary action: a text button set apart below the main
              one, never heavier than it (C3 #3). */}
          <div className="course-box-secondary">
            <ReportDialog
              quiet
              triggerLabel={t('enrollment.report.lecturer')}
              onReport={onReportLecturer}
              onSent={() => setReportSent(true)}
            />
            {reportSent ? <output>{t('enrollment.report.sent')}</output> : null}
          </div>
        </aside>
      );
    case 'coming-soon':
      return (
        <aside className="availability">
          <h2>{t('catalog.comingSoon')}</h2>
          {price}
          <p>{t('course.comingSoonHint')}</p>
        </aside>
      );
    case 'sign-in':
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {price}
          <a className="ui-button ui-button--primary" href={`/login?returnTo=${returnTo}`}>
            {action.paid ? t('course.signInToBuy') : t('course.signInToEnroll')}
          </a>
        </aside>
      );
    case 'enroll':
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {price}
          <p>{t('course.learningHint')}</p>
          {/* Keep focus on this control while the request is in flight. */}
          <Button type="button" onPress={onEnroll} isPending={enrolling}>
            {t('enrollment.enroll')}
          </Button>
          {enrollError !== undefined && enrollError !== 'authentication_required' && (
            <p role="alert">{t(`enrollment.errors.${enrollErrorKey(enrollError)}`)}</p>
          )}
        </aside>
      );
    case 'buy':
      return (
        <aside className="availability">
          <h2>{t('course.learningTitle')}</h2>
          {price}
          <Button type="button" onPress={onBuy} isPending={buying}>
            {t('learn.purchase.buy')}
          </Button>
          {buyError !== undefined && buyError !== 'authentication_required' && (
            <p role="alert">
              {/* A paid course that cannot be bought again (revoked by an administrator) is
                  refused by design: say so, not "no connection". */}
              {buyError === 'course_already_owned'
                ? t('learn.purchase.errors.notPurchasable')
                : t(`learn.purchase.errors.${purchaseErrorKey(buyError)}`)}
            </p>
          )}
        </aside>
      );
  }
}

export function enrollErrorKey(
  code: EnrollmentErrorCode,
):
  | 'membership_conflict'
  | 'enrollment_required'
  | 'enrollment_blocked'
  | 'course_not_enrollable'
  | 'enrollment_forbidden'
  | 'course_already_owned'
  | 'email_not_verified'
  | 'admin_not_learner'
  | 'generic' {
  switch (code) {
    case 'membership_conflict':
    case 'enrollment_required':
    case 'enrollment_blocked':
    case 'course_not_enrollable':
    case 'enrollment_forbidden':
    case 'course_already_owned':
    case 'email_not_verified':
    case 'admin_not_learner':
      return code;
    default:
      return 'generic';
  }
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
                          : error instanceof Error &&
                              (error.message === 'enrollment_required' ||
                                error.message === 'enrollment_blocked')
                            ? t(`ratings.errors.${error.message}`)
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
