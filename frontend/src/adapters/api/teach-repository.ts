import {
  dismissCommentReports,
  listAdminLearners,
  listTeachCourseComments,
  listTeachLearners,
  listTeachModerationQueue,
  removeLessonComment,
  dismissCourseReviewReports,
  removeCourseReview,
} from './generated/courses';
import { reviewSchema } from './course-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';
import type { TeachRepository } from '../../ports/TeachRepository';
import type {
  DismissReportsInput,
  LearnerCourse,
  LearnerPage,
  LearnerRow,
  ModerationQueuePage,
  TeachComment,
  TeachCommentPage,
  TeachErrorCode,
  TeachResult,
} from '../../domain/teach/teach';

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
function uuid(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}
function date(value: unknown): value is string {
  return typeof value === 'string' && !Number.isNaN(Date.parse(value));
}
function comment(value: unknown): value is TeachComment {
  if (
    !record(value) ||
    !uuid(value.id) ||
    !uuid(value.lessonId) ||
    typeof value.lessonTitle !== 'string' ||
    value.lessonTitle.length === 0 ||
    value.lessonTitle.length > 200 ||
    (value.parentId !== null && !uuid(value.parentId)) ||
    !record(value.author) ||
    (value.author.displayName !== null && typeof value.author.displayName !== 'string') ||
    (value.body !== null && typeof value.body !== 'string') ||
    (value.positionMs !== null &&
      (typeof value.positionMs !== 'number' ||
        !Number.isInteger(value.positionMs) ||
        value.positionMs < 0)) ||
    !date(value.createdAt) ||
    (value.editedAt !== null && !date(value.editedAt)) ||
    typeof value.removed !== 'boolean'
  )
    return false;
  return true;
}
function page(value: unknown): value is TeachCommentPage {
  return (
    record(value) &&
    Array.isArray(value.comments) &&
    value.comments.length <= 50 &&
    typeof value.nextCursor === 'string' &&
    value.comments.every(comment)
  );
}
function queuePage(value: unknown): value is ModerationQueuePage {
  return (
    record(value) &&
    Array.isArray(value.items) &&
    value.items.length <= 50 &&
    typeof value.nextCursor === 'string' &&
    value.items.every((item) => {
      if (
        !record(item) ||
        typeof item.openReports !== 'number' ||
        !Number.isInteger(item.openReports) ||
        item.openReports <= 0
      )
        return false;
      // A review row is validated as strictly as a comment row: its fields reach the queue UI directly
      // ('★'.repeat(stars) throws on a negative or infinite count).
      return (
        (item.kind === 'comment' && comment(item.comment) && item.review === null) ||
        (item.kind === 'review' &&
          item.comment === null &&
          reviewSchema.safeParse(item.review).success)
      );
    })
  );
}
function learnerCourse(value: unknown): value is LearnerCourse {
  return (
    record(value) &&
    uuid(value.courseId) &&
    typeof value.slug === 'string' &&
    /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value.slug) &&
    typeof value.title === 'string' &&
    value.title.length >= 1 &&
    value.title.length <= 200 &&
    (value.state === 'active' || value.state === 'blocked') &&
    typeof value.progressPercent === 'number' &&
    Number.isInteger(value.progressPercent) &&
    value.progressPercent >= 0 &&
    value.progressPercent <= 100 &&
    date(value.enrolledAt)
  );
}
function learnerRow(value: unknown): value is LearnerRow {
  return (
    record(value) &&
    uuid(value.userId) &&
    // "" is the anonymised deleted account — a valid, expected value, not a parse failure.
    typeof value.displayName === 'string' &&
    value.displayName.length <= 120 &&
    (value.lastActivityAt === null || date(value.lastActivityAt)) &&
    Array.isArray(value.courses) &&
    value.courses.length >= 1 &&
    value.courses.every(learnerCourse)
  );
}
function learnerPage(value: unknown): value is LearnerPage {
  return (
    record(value) &&
    Array.isArray(value.entries) &&
    value.entries.length <= 100 &&
    value.entries.every(learnerRow) &&
    (value.nextCursor === null || typeof value.nextCursor === 'string')
  );
}
function failure(code: TeachErrorCode): TeachResult<never> {
  return { ok: false, error: { code } };
}
function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly TeachErrorCode[]>>,
): TeachErrorCode {
  const data = response.data;
  const error = record(data) && record(data.error) ? data.error.code : undefined;
  const named = allowed[response.status]?.find((candidate) => candidate === error);
  return named ?? (unnamedStatusCode(response.status) as TeachErrorCode);
}
const readErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden'],
  404: ['course_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;
const learnerErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden', 'learner_course_forbidden'],
  404: ['course_not_found'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;
const dismissErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden', 'comment_forbidden'],
  404: ['comment_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;
const removeErrors = {
  401: ['authentication_required'],
  403: ['comment_forbidden'],
  404: ['comment_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;

// Review moderation: only an admin removes (G9) — the course owner gets review_moderation_forbidden; owner or
// admin dismisses — anyone else gets course_forbidden.
const reviewDismissErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['course_forbidden'],
  404: ['review_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;
const reviewRemoveErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['review_moderation_forbidden'],
  404: ['review_not_found'],
} as const satisfies Readonly<Record<number, readonly TeachErrorCode[]>>;

export function createApiTeachRepository(options: TransportOptions = {}): TeachRepository {
  const send = createApiTransport(options);
  return {
    async listLearners(scope, filters, cursor, signal) {
      const params = {
        ...(filters.courseId === undefined ? {} : { courseId: filters.courseId }),
        ...(filters.state === undefined ? {} : { state: filters.state }),
        ...(filters.q === undefined || filters.q === '' ? {} : { q: filters.q }),
        ...(cursor === undefined ? {} : { cursor }),
        limit: 20,
      };
      const outcome = await send(
        scope === 'admin' ? 'teach.learners.admin' : 'teach.learners.teach',
        (init) =>
          scope === 'admin' ? listAdminLearners(params, init) : listTeachLearners(params, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, learnerErrors));
      if (!learnerPage(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          // Explicit per-field mapping, not a pass-through: the frontend itself enforces Lo07
          // (email never reaches the domain) rather than leaning on the contract alone. Any wire
          // field the schema does not name — email above all — is dropped here.
          entries: outcome.data.entries.map((entry) => ({
            userId: entry.userId,
            displayName: entry.displayName,
            lastActivityAt: entry.lastActivityAt,
            courses: entry.courses.map((course) => ({
              courseId: course.courseId,
              slug: course.slug,
              title: course.title,
              state: course.state,
              progressPercent: course.progressPercent,
              enrolledAt: course.enrolledAt,
            })),
          })),
          // Normalise an empty-string cursor to null: a Go handler emits "" for "no more pages",
          // not null (the last-page sentinel bug class), and "Показати ще" keys off null.
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },
    async listCourseComments(courseId, cursor, signal) {
      const outcome = await send(
        'teach.comments.list',
        (init) =>
          listTeachCourseComments(courseId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      if (!page(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          ...outcome.data,
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },
    async listModerationQueue(courseId, cursor, signal) {
      const outcome = await send(
        'teach.moderation.list',
        (init) =>
          listTeachModerationQueue(courseId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      if (!queuePage(outcome.data)) return failure('invalid_response');
      return {
        ok: true,
        value: {
          items: outcome.data.items,
          nextCursor: outcome.data.nextCursor === '' ? null : outcome.data.nextCursor,
        },
      };
    },
    async dismissReport(commentId, input: DismissReportsInput, signal) {
      const outcome = await send(
        'teach.moderation.dismiss',
        (init) => dismissCommentReports(commentId, input, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 204) return failure(namedError(outcome, dismissErrors));
      return { ok: true, value: undefined };
    },
    async removeComment(commentId, signal) {
      const outcome = await send(
        'teach.moderation.remove',
        (init) => removeLessonComment(commentId, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      if (outcome.status !== 204) return failure(namedError(outcome, removeErrors));
      return { ok: true, value: undefined };
    },
    async dismissReviewReport(reviewId, input, signal) {
      const outcome = await send(
        'teach.reviews.dismiss',
        (init) => dismissCourseReviewReports(reviewId, input, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      return outcome.status === 204
        ? { ok: true, value: undefined }
        : failure(namedError(outcome, reviewDismissErrors));
    },
    async removeReview(reviewId, input, signal) {
      const outcome = await send(
        'teach.reviews.remove',
        (init) => removeCourseReview(reviewId, input, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TeachErrorCode);
      return outcome.status === 204
        ? { ok: true, value: undefined }
        : failure(namedError(outcome, reviewRemoveErrors));
    },
  };
}
