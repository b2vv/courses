import type { CourseGrant, GrantResult } from '../domain/grant';

export interface CourseGrantRepository {
  grant(
    courseId: string,
    userId: string,
    reason: string,
    signal?: AbortSignal,
  ): Promise<GrantResult<CourseGrant>>;
}
