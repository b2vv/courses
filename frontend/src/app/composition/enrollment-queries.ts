import { infiniteQueryOptions, mutationOptions, queryOptions } from '@tanstack/react-query';
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
export function enrollmentRosterQueryKey(courseId: string) {
  return ['enrollment', 'roster', courseId] as const;
}

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
    roster: (courseId: string) =>
      infiniteQueryOptions({
        queryKey: enrollmentRosterQueryKey(courseId),
        initialPageParam: undefined as string | undefined,
        queryFn: ({ signal, pageParam }) => enrollment.roster(courseId, pageParam, signal).then(unwrap),
        getNextPageParam: (page) => page.nextCursor ?? undefined,
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
    block: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ courseId, userId }: { courseId: string; userId: string }) =>
          enrollment.block(courseId, userId).then(unwrap),
        onSuccess: async (_, { courseId }) => {
          await client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
        },
      }),
    unblock: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ courseId, userId }: { courseId: string; userId: string }) =>
          enrollment.unblock(courseId, userId).then(unwrap),
        onSuccess: async (_, { courseId }) => {
          await client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
        },
      }),
  };
}
