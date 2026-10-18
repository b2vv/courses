import { mutationOptions } from '@tanstack/react-query';
import { grantReason, type GrantErrorCode, type GrantResult } from '../../modules/purchase';
import type { CourseGrantRepository } from '../../modules/purchase';

export class GrantQueryError extends Error {
  readonly code: GrantErrorCode;
  constructor(code: GrantErrorCode) {
    super(code);
    this.name = 'GrantQueryError';
    this.code = code;
  }
}

function unwrap<T>(result: GrantResult<T>): T {
  if (!result.ok) throw new GrantQueryError(result.error.code);
  return result.value;
}

export type GrantCourseVariables = {
  readonly courseId: string;
  readonly userId: string;
  readonly reason: string;
};

export function createAdminGrantQueries(grants: CourseGrantRepository) {
  return {
    grant: () =>
      mutationOptions({
        mutationFn: async ({ courseId, userId, reason }: GrantCourseVariables) => {
          const trimmed = grantReason(reason);
          if (trimmed === null) throw new GrantQueryError('invalid_request');
          return unwrap(await grants.grant(courseId, userId, trimmed));
        },
      }),
  };
}
