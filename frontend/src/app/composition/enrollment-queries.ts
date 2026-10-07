import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { EnrollmentErrorCode, EnrollmentResult } from '../../domain/enrollment/enrollment';
import type { EnrollmentRepository } from '../../ports/EnrollmentRepository';

export class EnrollmentQueryError extends Error {
  readonly code: EnrollmentErrorCode;
  constructor(code: EnrollmentErrorCode) {
    super(code);
    this.name = 'EnrollmentQueryError';
    this.code = code;
  }
}
function unwrap<T>(result: EnrollmentResult<T>): T {
  if (!result.ok) throw new EnrollmentQueryError(result.error.code);
  return result.value;
}
export function enrollmentMembershipQueryKey(courseId: string) {
  return ['enrollment', 'membership', courseId] as const;
}
export const enrollmentMyMembershipsQueryKey = ['enrollment', 'myMemberships'] as const;

export function createEnrollmentQueries(enrollment: EnrollmentRepository) {
  return {
    membership: (courseId: string) =>
      queryOptions({
        queryKey: enrollmentMembershipQueryKey(courseId),
        queryFn: ({ signal }) => enrollment.membership(courseId, signal).then(unwrap),
      }),
    myMemberships: () =>
      queryOptions({
        queryKey: enrollmentMyMembershipsQueryKey,
        queryFn: ({ signal }) => enrollment.myMemberships(signal).then(unwrap),
      }),
    enroll: (client: QueryClient) =>
      mutationOptions({
        mutationFn: (courseId: string) => enrollment.enroll(courseId).then(unwrap),
        onSuccess: async (_, courseId) => {
          await client.invalidateQueries({ queryKey: enrollmentMembershipQueryKey(courseId) });
          await client.invalidateQueries({ queryKey: enrollmentMyMembershipsQueryKey });
          await client.invalidateQueries({ queryKey: ['course'] });
        },
        onError: (error, courseId) => {
          if (error instanceof EnrollmentQueryError && error.code === 'membership_conflict') {
            return client.invalidateQueries({ queryKey: enrollmentMembershipQueryKey(courseId) });
          }
          return undefined;
        },
      }),
  };
}
