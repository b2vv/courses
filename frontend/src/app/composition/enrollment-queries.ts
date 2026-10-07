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
export const enrollmentParticipantReportsQueryKey = ['enrollment', 'participantReports'] as const;

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
    participantReports: () =>
      infiniteQueryOptions({
        queryKey: enrollmentParticipantReportsQueryKey,
        initialPageParam: undefined as string | undefined,
        queryFn: ({ signal, pageParam }) => enrollment.participantReports(pageParam, signal).then(unwrap),
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
        // A concurrent change (e.g. admin revokes while the owner blocks) lands as a 409
        // membership_conflict; re-read the roster so the stale row doesn't mislead (B-L7b.2).
        onError: (error, { courseId }) => {
          if (error instanceof EnrollmentQueryError && error.code === 'membership_conflict') {
            return client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
          }
          return undefined;
        },
      }),
    unblock: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ courseId, userId }: { courseId: string; userId: string }) =>
          enrollment.unblock(courseId, userId).then(unwrap),
        onSuccess: async (_, { courseId }) => {
          await client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
        },
        onError: (error, { courseId }) => {
          if (error instanceof EnrollmentQueryError && error.code === 'membership_conflict') {
            return client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
          }
          return undefined;
        },
      }),
    reportLecturer: () =>
      mutationOptions({
        mutationFn: ({ courseId, reason }: { courseId: string; reason: string }) =>
          enrollment.reportLecturer(courseId, reason).then(unwrap),
      }),
    reportStudent: () =>
      mutationOptions({
        mutationFn: ({ courseId, userId, reason }: { courseId: string; userId: string; reason: string }) =>
          enrollment.reportStudent(courseId, userId, reason).then(unwrap),
      }),
    resolveReport: (client: QueryClient) =>
      mutationOptions({
        mutationFn: (reportId: string) => enrollment.resolveReport(reportId).then(unwrap),
        onSuccess: () => client.invalidateQueries({ queryKey: enrollmentParticipantReportsQueryKey }),
      }),
    revoke: (client: QueryClient) =>
      mutationOptions({
        mutationFn: ({ courseId, userId }: { courseId: string; userId: string }) =>
          enrollment.revoke(courseId, userId).then(unwrap),
        onSuccess: async (_, { courseId }) => {
          await client.invalidateQueries({ queryKey: enrollmentParticipantReportsQueryKey });
          await client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
        },
        onError: (error, { courseId }) => {
          if (error instanceof EnrollmentQueryError && error.code === 'membership_conflict') {
            void client.invalidateQueries({ queryKey: enrollmentParticipantReportsQueryKey });
            return client.invalidateQueries({ queryKey: enrollmentRosterQueryKey(courseId) });
          }
          return undefined;
        },
      }),
  };
}
