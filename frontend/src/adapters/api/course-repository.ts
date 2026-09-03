import { context, SpanStatusCode, trace } from '@opentelemetry/api';
import { makePublishedCourse } from '../../domain/course/course';
import type { CatalogError, CoursePage, Result } from '../../domain/course/course';
import type { CourseRepository } from '../../ports/CourseRepository';
import { getPublishedCourse, listPublishedCourses } from './generated/courses';
import { courseSchema, errorSchema, pageSchema } from './course-schema';

type Options = {
  readonly cookie?: string;
  readonly requestSignal?: AbortSignal;
  readonly timeoutMs?: number;
};
type WireResponse = { readonly data: unknown; readonly status: number };
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
    }
  }
  return {
    code:
      response.status >= 500 || response.status === 429
        ? 'dependency_unavailable'
        : 'invalid_response',
  };
}

/** One repository per SSR request; no cookie-bearing singleton or HTTP data cache. */
export function createApiCourseRepository(options: Options = {}): CourseRepository {
  const tracer = trace.getTracer('courses.frontend.api');
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    map: (data: unknown) => Result<T>,
    caller?: AbortSignal,
  ): Promise<Result<T>> {
    const deadline = AbortSignal.timeout(options.timeoutMs ?? 5000);
    const controls = [caller, options.requestSignal].filter(
      (signal): signal is AbortSignal => signal !== undefined,
    );
    const signal = AbortSignal.any([...controls, deadline]);
    const headers = new Headers({ Accept: 'application/json' });
    if (options.cookie) headers.set('Cookie', options.cookie);
    return tracer.startActiveSpan(
      operation,
      {},
      context.active(),
      async (span): Promise<Result<T>> => {
        try {
          const response = await execute({
            headers,
            signal,
            credentials: 'same-origin',
            redirect: 'error',
            cache: 'no-store',
          });
          if (response.status !== 200) {
            span.setStatus({ code: SpanStatusCode.ERROR });
            return { ok: false, error: wireError(response) };
          }
          return map(response.data);
        } catch (error) {
          span.setStatus({ code: SpanStatusCode.ERROR });
          // Control signals remain distinct from malformed JSON and network errors.
          if (controls.some((control) => control.aborted))
            return { ok: false, error: { code: 'request_cancelled' } };
          if (deadline.aborted) return { ok: false, error: { code: 'request_timeout' } };
          if (error instanceof SyntaxError) return invalid;
          return { ok: false, error: { code: 'dependency_unavailable' } };
        } finally {
          span.end();
        }
      },
    );
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
      return request(
        'catalog.detail',
        (init) => getPublishedCourse(slug, init),
        (raw) => {
          const parsed = courseSchema.safeParse(raw);
          if (!parsed.success || parsed.data.slug !== slug) return invalid;
          return makePublishedCourse(parsed.data);
        },
        signal,
      );
    },
  };
}
