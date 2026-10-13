import type {
  AdminCourseDraft,
  AdminCoursePage,
  AdminCourseQuery,
  AuthoringError,
  AuthoringResult,
  CourseId,
  CourseTree,
  CourseTreeEdit,
  InstructorOption,
} from '../../domain/course/course';
import type { AdminCourseRepository } from '../../ports/AdminCourseRepository';
import {
  createAdminCourse,
  deleteAdminCourse,
  getAdminCourse,
  listAdminCourses,
  listInstructors,
  saveAdminCourse,
  transitionAdminCourse,
} from './generated/courses';
import { errorSchema } from './course-schema';
import {
  adminCoursePageSchema,
  adminCourseTreeSchema,
  instructorOptionsSchema,
} from './admin-course-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

const invalid = {
  ok: false,
  error: { code: 'invalid_response' },
} satisfies AuthoringResult<never>;

function wireError(response: WireResponse): AuthoringError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'invalid_request':
      case 'invalid_owner':
      case 'lesson_limit':
        if (response.status === 400) return { code: parsed.data.error.code };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'origin_forbidden':
      case 'forbidden':
      case 'course_forbidden':
        if (response.status === 403) return { code: parsed.data.error.code };
        break;
      case 'course_not_found':
        if (response.status === 404) return { code: 'course_not_found' };
        break;
      case 'slug_taken':
      case 'slug_immutable':
      case 'invalid_status_transition':
      case 'course_not_deletable':
      case 'course_modified':
        if (response.status === 409) return { code: parsed.data.error.code };
        break;
      case 'dependency_unavailable':
        if (response.status === 503) return { code: 'dependency_unavailable' };
        break;
      case 'internal_error':
        if (response.status === 500) return { code: 'dependency_unavailable' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

function courseTree(raw: unknown, expectedId?: CourseId): AuthoringResult<CourseTree> {
  const parsed = adminCourseTreeSchema.safeParse(raw);
  if (!parsed.success || (expectedId !== undefined && parsed.data.id !== expectedId))
    return invalid;
  return { ok: true, value: parsed.data };
}

function coursePage(raw: unknown, query: AdminCourseQuery): AuthoringResult<AdminCoursePage> {
  const parsed = adminCoursePageSchema.safeParse(raw);
  if (
    !parsed.success ||
    parsed.data.pagination.limit !== query.limit ||
    parsed.data.pagination.offset !== query.offset ||
    parsed.data.items.length > query.limit ||
    parsed.data.items.length > Math.max(0, parsed.data.pagination.total - query.offset) ||
    (query.status !== undefined && parsed.data.items.some((item) => item.status !== query.status))
  )
    return invalid;
  return { ok: true, value: parsed.data };
}

function courseWrite(draft: AdminCourseDraft) {
  const fields = {
    slug: draft.slug,
    title: draft.title,
    summary: draft.summary,
    description: draft.description,
  };
  return draft.instructorId === undefined
    ? fields
    : { ...fields, instructorId: draft.instructorId };
}

function courseSave(edit: CourseTreeEdit, expectedUpdatedAt: string) {
  return {
    expectedUpdatedAt,
    course: {
      slug: edit.slug,
      title: edit.title,
      summary: edit.summary,
      description: edit.description,
      instructorId: edit.instructorId,
      // Absent means unchanged on the server; null removes the price.
      ...(edit.price === undefined ? {} : { price: edit.price }),
    },
    modules: edit.modules.map((module) => ({
      id: module.id,
      title: module.title,
      lessons: module.lessons.map((lesson) => ({
        id: lesson.id,
        title: lesson.title,
        isPreview: lesson.isPreview,
      })),
    })),
  };
}

/** One repository per router request; generated transport types never leave this adapter. */
export function createApiAdminCourseRepository(
  options: TransportOptions = {},
): AdminCourseRepository {
  const send = createApiTransport(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => AuthoringResult<T>,
    caller?: AbortSignal,
  ): Promise<AuthoringResult<T>> {
    const outcome = await send(operation, execute, expected, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === expected) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }

  return {
    list(query, signal) {
      const params = {
        limit: query.limit,
        offset: query.offset,
        ...(query.status === undefined ? {} : { status: query.status }),
      };
      return request<AdminCoursePage>(
        'admin.courses.list',
        (init) => listAdminCourses(params, init),
        200,
        (raw) => coursePage(raw, query),
        signal,
      );
    },
    load(id, signal) {
      return request<CourseTree>(
        'admin.courses.load',
        (init) => getAdminCourse(id, init),
        200,
        (raw) => courseTree(raw, id),
        signal,
      );
    },
    listInstructors(signal) {
      return request<readonly InstructorOption[]>(
        'admin.instructors.list',
        (init) => listInstructors(init),
        200,
        (raw) => {
          const parsed = instructorOptionsSchema.safeParse(raw);
          return parsed.success ? { ok: true, value: parsed.data } : invalid;
        },
        signal,
      );
    },
    create(draft, signal) {
      return request<CourseTree>(
        'admin.courses.create',
        (init) => createAdminCourse(courseWrite(draft), init),
        201,
        (raw) => {
          const result = courseTree(raw);
          return result.ok && (result.value.status !== 'draft' || result.value.modules.length > 0)
            ? invalid
            : result;
        },
        signal,
      );
    },
    save(id, expectedUpdatedAt, edit, signal) {
      return request<CourseTree>(
        'admin.courses.save',
        (init) => saveAdminCourse(id, courseSave(edit, expectedUpdatedAt), init),
        200,
        (raw) => courseTree(raw, id),
        signal,
      );
    },
    transition(id, status, expectedUpdatedAt, signal) {
      return request<CourseTree>(
        'admin.courses.transition',
        (init) => transitionAdminCourse(id, { status, expectedUpdatedAt }, init),
        200,
        (raw) => {
          const result = courseTree(raw, id);
          return result.ok && result.value.status !== status ? invalid : result;
        },
        signal,
      );
    },
    delete(id, expectedUpdatedAt, signal) {
      return request<null>(
        'admin.courses.delete',
        (init) => deleteAdminCourse(id, { expectedUpdatedAt }, init),
        204,
        () => ({ ok: true, value: null }),
        signal,
      );
    },
  };
}
