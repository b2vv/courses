import type { CourseDetail } from '../../../contracts/course';
import type { PublicReviewPage } from '../../../contracts/course';
import type { EnrollmentErrorCode, Membership } from '../../enrollment';
import { Button, Link } from '../../../presentation/ui';
import { useTranslation } from 'react-i18next';
import { useEffect, useRef, useState } from 'react';
import { RatingSummary } from './RatingSummary';
import { formatDate } from '../../../presentation/format-timestamp';
import { ReportDialog } from '../../enrollment';

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
  viewerIsAdmin = false,
  membership,
  enrolling = false,
  enrollError,
  onEnroll = () => {},
  onReportLecturer = async () => {},
  ownReviewId,
  onReport = async () => {},
}: {
  course: CourseDetail;
  viewerIsAdmin?: boolean | undefined;
  reviews?: PublicReviewPage | undefined;
  reviewsError?: boolean | undefined;
  hasMore?: boolean | undefined;
  isLoadingMore?: boolean | undefined;
  isNextPageError?: boolean | undefined;
  onLoadMore?: (() => void) | undefined;
  onRetry?: (() => void) | undefined;
  onRetryMore?: (() => void) | undefined;
  signedIn?: boolean | undefined;
  membership?: Membership | undefined;
  enrolling?: boolean | undefined;
  enrollError?: EnrollmentErrorCode | undefined;
  onEnroll?: (() => void) | undefined;
  onReportLecturer?: ((reason: string) => Promise<void>) | undefined;
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
        <EnrollmentCallToAction
          course={course}
          viewerIsAdmin={viewerIsAdmin}
          membership={membership}
          enrolling={enrolling}
          enrollError={enrollError}
          onEnroll={onEnroll}
          onReportLecturer={onReportLecturer}
        />
      </div>
    </article>
  );
}

function EnrollmentCallToAction({
  course,
  viewerIsAdmin,
  membership,
  enrolling,
  enrollError,
  onEnroll,
  onReportLecturer,
}: {
  course: CourseDetail;
  viewerIsAdmin: boolean;
  membership: Membership | undefined;
  enrolling: boolean;
  enrollError: EnrollmentErrorCode | undefined;
  onEnroll: () => void;
  onReportLecturer: (reason: string) => Promise<void>;
}) {
  const { t } = useTranslation();
  const [reportSent, setReportSent] = useState(false);
  // The backend's `member` flag means "a membership row exists", which is true for revoked too;
  // access follows the STATE. active|blocked are current members; revoked|none are not and may
  // re-enroll (free) — so branch on state, not the member boolean.
  const activeMember = membership?.state === 'active' || membership?.state === 'blocked';
  const blocked = membership?.state === 'blocked';
  // G8 (Lo19): an administrator is not a learner — they view the course as exempt and start
  // learning, but never see the enrol call to action.
  const canEnroll = !activeMember && course.price === null && !viewerIsAdmin;
  return (
    <aside className="availability">
      <h2>{activeMember ? t('enrollment.member') : t('course.learningTitle')}</h2>
      {activeMember ? (
        <>
          {blocked && <output>{t('enrollment.blockedNote')}</output>}
          <Link href={`/learn/${course.slug}`}>{t('enrollment.startLearning')}</Link>
          <ReportDialog
            triggerLabel={t('enrollment.report.lecturer')}
            onReport={onReportLecturer}
            onSent={() => setReportSent(true)}
          />
          {reportSent ? <output>{t('enrollment.report.sent')}</output> : null}
        </>
      ) : canEnroll ? (
        <>
          <p>{t('course.learningHint')}</p>
          {/* Keep focus on this control while the request is in flight. */}
          <Button type="button" onPress={onEnroll} isPending={enrolling}>
            {t('enrollment.enroll')}
          </Button>
          {enrollError !== undefined && (
            <p role="alert">{t(`enrollment.errors.${enrollErrorKey(enrollError)}`)}</p>
          )}
        </>
      ) : (
        <>
          <p>{t('course.learningHint')}</p>
          <Link href={`/learn/${course.slug}`}>{t('course.startLearning')}</Link>
        </>
      )}
    </aside>
  );
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
