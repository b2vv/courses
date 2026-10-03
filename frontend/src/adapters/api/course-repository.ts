import { makePublishedCourse, reviewPercent } from '../../domain/course/course';
import type {
  CatalogError,
  CourseDetail,
  CoursePage,
  OwnReview,
  Result,
} from '../../domain/course/course';
import type { CourseRepository } from '../../ports/CourseRepository';
import {
  getPublishedCourse,
  getOwnCourseReview,
  listCourseReviews,
  listPublishedCourses,
  reportCourseReview,
} from './generated/courses';
import {
  courseDetailSchema,
  errorSchema,
  pageSchema,
  publicReviewPageSchema,
  ownReviewSchema,
} from './course-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

const invalid = { ok: false, error: { code: 'invalid_response' } } satisfies Result<never>;

function wireError(response: WireResponse): CatalogError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'course_not_found':
        if (response.status === 404) return { code: 'course_not_found' };
        break;
      case 'invalid_slug':
        if (response.status === 400) return { code: 'invalid_slug' };
        break;
      case 'invalid_query':
        if (response.status === 400) return { code: 'invalid_query' };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'report_own_review':
        if (response.status === 403) return { code: 'report_own_review' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

/** One repository per SSR request; no cookie-bearing singleton or HTTP data cache. */
export function createApiCourseRepository(options: TransportOptions = {}): CourseRepository {
  const send = createApiTransport(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    map: (data: unknown) => Result<T>,
    caller?: AbortSignal,
  ): Promise<Result<T>> {
    const outcome = await send(operation, execute, 200, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status !== 200) return { ok: false, error: wireError(outcome) };
    return map(outcome.data);
  }
  return {
    listPublished(query, signal) {
      return request<CoursePage>(
        'catalog.list',
        (init) => listPublishedCourses(query, init),
        (raw) => {
          const parsed = pageSchema.safeParse(raw);
          if (
            !parsed.success ||
            parsed.data.pagination.limit !== query.limit ||
            parsed.data.pagination.offset !== query.offset
          )
            return invalid;
          const items = [];
          for (const dto of parsed.data.items) {
            const result = makePublishedCourse(dto);
            if (!result.ok) return result;
            items.push(result.value);
          }
          if (
            items.length > query.limit ||
            items.length > Math.max(0, parsed.data.pagination.total - query.offset)
          )
            return invalid;
          return { ok: true, value: { items, pagination: { ...parsed.data.pagination } } };
        },
        signal,
      );
    },
    getPublished(slug, signal) {
      return request<CourseDetail>(
        'catalog.detail',
        // The generated builder interpolates the path segment verbatim, so the
        // slug is encoded here rather than relying only on server validation.
        (init) => getPublishedCourse(encodeURIComponent(slug), init),
        (raw) => {
          const parsed = courseDetailSchema.safeParse(raw);
          if (!parsed.success || parsed.data.slug !== slug) return invalid;
          const course = makePublishedCourse(parsed.data);
          if (!course.ok) return course;
          return {
            ok: true,
            value: {
              ...course.value,
              program: parsed.data.program.map((module) => ({
                title: module.title,
                lessons: module.lessons.map((lesson) => ({ title: lesson.title })),
              })),
            },
          };
        },
        signal,
      );
    },
    listReviews(slug, cursor, signal) {
      return request(
        'reviews.list',
        (init) =>
          listCourseReviews(
            encodeURIComponent(slug),
            cursor === undefined ? undefined : { cursor },
            init,
          ),
        (raw) => {
          const parsed = publicReviewPageSchema.safeParse(raw);
          if (!parsed.success) return invalid;
          return {
            ok: true,
            value: {
              rating: parsed.data.rating,
              nextCursor: parsed.data.nextCursor,
              items: parsed.data.items.map((review) => ({
                ...review,
                percent: reviewPercent(review.completedLessons, review.totalLessons),
              })),
            },
          };
        },
        signal,
      );
    },
    async reportReview(reviewId, signal) {
      const outcome = await send(
        'reviews.report',
        (init) => reportCourseReview(encodeURIComponent(reviewId), init),
        204,
        signal,
      );
      if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
      return outcome.status === 204
        ? { ok: true, value: undefined }
        : { ok: false, error: wireError(outcome) };
    },
    ownReview(courseId, signal) {
      return request<OwnReview>(
        'reviews.own',
        (init) => getOwnCourseReview(encodeURIComponent(courseId), init),
        (raw) => {
          const parsed = ownReviewSchema.safeParse(raw);
          if (!parsed.success) return invalid;
          const review = parsed.data.review;
          return {
            ok: true,
            value: {
              eligibility: parsed.data.eligibility,
              review:
                review === null
                  ? null
                  : {
                      ...review,
                      percent: reviewPercent(review.completedLessons, review.totalLessons),
                    },
            },
          };
        },
        signal,
      );
    },
  };
}
