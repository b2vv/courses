import {
  makeAdminCourseQuery,
  validCourseDraft,
  validCourseTree,
  validCourseTreeEdit,
  validSlugChange,
  validTransition,
} from '../../domain/course/course';
import type {
  AdminCourseDraft,
  AdminCourseQuery,
  AuthoringResult,
  CourseId,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
} from '../../domain/course/course';
import type { AdminCourseRepository } from '../../ports/AdminCourseRepository';

const invalid = { ok: false, error: { code: 'invalid_request' } } as const;

export function createAuthoring(repository: AdminCourseRepository) {
  return {
    list(input: Partial<AdminCourseQuery>, signal?: AbortSignal) {
      const query = makeAdminCourseQuery(input);
      return query.ok ? repository.list(query.value, signal) : Promise.resolve(query);
    },
    load(id: CourseId, signal?: AbortSignal) {
      return repository.load(id, signal);
    },
    listInstructors(signal?: AbortSignal) {
      return repository.listInstructors(signal);
    },
    create(draft: AdminCourseDraft, signal?: AbortSignal): Promise<AuthoringResult<CourseTree>> {
      return validCourseDraft(draft) ? repository.create(draft, signal) : Promise.resolve(invalid);
    },
    save(
      current: CourseTree,
      edit: CourseTreeEdit,
      signal?: AbortSignal,
    ): Promise<AuthoringResult<CourseTree>> {
      if (!validCourseTree(current) || !validCourseTreeEdit(edit)) return Promise.resolve(invalid);
      if (!validSlugChange(current.slug, current.publishedAt, edit.slug))
        return Promise.resolve({ ok: false, error: { code: 'slug_immutable' } });
      return repository.save(current.id, current.updatedAt, edit, signal);
    },
    transition(
      current: CourseTree,
      status: CourseStatus,
      signal?: AbortSignal,
    ): Promise<AuthoringResult<CourseTree>> {
      if (!validCourseTree(current)) return Promise.resolve(invalid);
      if (!validTransition(current.status, status))
        return Promise.resolve({ ok: false, error: { code: 'invalid_status_transition' } });
      return repository.transition(current.id, status, current.updatedAt, signal);
    },
    delete(current: CourseTree, signal?: AbortSignal): Promise<AuthoringResult<null>> {
      if (!validCourseTree(current)) return Promise.resolve(invalid);
      if (current.status !== 'draft')
        return Promise.resolve({ ok: false, error: { code: 'course_not_deletable' } });
      return repository.delete(current.id, current.updatedAt, signal);
    },
  };
}
