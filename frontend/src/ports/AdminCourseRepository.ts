import type {
  AdminCourseDraft,
  AdminCoursePage,
  AdminCourseQuery,
  AuthoringResult,
  CourseId,
  CourseStatus,
  CourseTree,
  CourseTreeEdit,
  InstructorOption,
} from '../contracts/course';

export interface AdminCourseRepository {
  list(query: AdminCourseQuery, signal?: AbortSignal): Promise<AuthoringResult<AdminCoursePage>>;
  load(id: CourseId, signal?: AbortSignal): Promise<AuthoringResult<CourseTree>>;
  listInstructors(signal?: AbortSignal): Promise<AuthoringResult<readonly InstructorOption[]>>;
  create(draft: AdminCourseDraft, signal?: AbortSignal): Promise<AuthoringResult<CourseTree>>;
  save(
    id: CourseId,
    expectedUpdatedAt: string,
    edit: CourseTreeEdit,
    signal?: AbortSignal,
  ): Promise<AuthoringResult<CourseTree>>;
  transition(
    id: CourseId,
    status: CourseStatus,
    expectedUpdatedAt: string,
    signal?: AbortSignal,
  ): Promise<AuthoringResult<CourseTree>>;
  delete(
    id: CourseId,
    expectedUpdatedAt: string,
    signal?: AbortSignal,
  ): Promise<AuthoringResult<null>>;
}
