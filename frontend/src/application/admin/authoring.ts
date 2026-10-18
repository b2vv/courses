import {
  makeAdminCourseQuery,
  validCourseDraft,
  validCourseTree,
  validCourseTreeEdit,
  validSlugChange,
  validTransition,
  unchangedEdit,
  validCoursePrice,
} from '../../contracts/course';
import type {
  AdminCourse,
  AdminCourseDraft,
  AdminCourseQuery,
  AuthoringResult,
  CourseId,
  CoursePrice,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
} from '../../contracts/course';
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
    /**
     * The price is saved on its own, over the stored tree: a first price is refused by the
     * server together with any other change, and keeping every price change on this one path
     * means the course form never carries money.
     */
    setPrice(
      current: CourseTree,
      price: CoursePrice | null,
      signal?: AbortSignal,
    ): Promise<AuthoringResult<CourseTree>> {
      if (!validCourseTree(current) || (price !== null && !validCoursePrice(price)))
        return Promise.resolve(invalid);
      return repository.save(
        current.id,
        current.updatedAt,
        { ...unchangedEdit(current), price },
        signal,
      );
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
    /**
     * Delete straight from a list row, without first loading the whole tree. The row carries
     * the id and the optimistic-concurrency token (updatedAt); only a draft is deletable, the
     * same rule the editor and the server enforce, so a non-draft never reaches the seam.
     */
    deleteListed(course: AdminCourse, signal?: AbortSignal): Promise<AuthoringResult<null>> {
      if (course.status !== 'draft')
        return Promise.resolve({ ok: false, error: { code: 'course_not_deletable' } });
      return repository.delete(course.id, course.updatedAt, signal);
    },
  };
}
