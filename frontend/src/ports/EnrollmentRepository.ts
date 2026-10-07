import type {
  EnrollmentResult,
  Membership,
  MembershipCard,
  ParticipantReportQueuePage,
  RosterPage,
} from '../domain/enrollment/enrollment';

export interface EnrollmentRepository {
  membership(courseId: string, signal?: AbortSignal): Promise<EnrollmentResult<Membership>>;
  enroll(courseId: string, signal?: AbortSignal): Promise<EnrollmentResult<Membership>>;
  myMemberships(signal?: AbortSignal): Promise<EnrollmentResult<readonly MembershipCard[]>>;
  roster(
    courseId: string,
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<EnrollmentResult<RosterPage>>;
  block(courseId: string, userId: string, signal?: AbortSignal): Promise<EnrollmentResult<void>>;
  unblock(courseId: string, userId: string, signal?: AbortSignal): Promise<EnrollmentResult<void>>;
  reportLecturer(
    courseId: string,
    reason: string,
    signal?: AbortSignal,
  ): Promise<EnrollmentResult<void>>;
  reportStudent(
    courseId: string,
    userId: string,
    reason: string,
    signal?: AbortSignal,
  ): Promise<EnrollmentResult<void>>;
  participantReports(
    cursor?: string,
    signal?: AbortSignal,
  ): Promise<EnrollmentResult<ParticipantReportQueuePage>>;
  resolveReport(reportId: string, signal?: AbortSignal): Promise<EnrollmentResult<void>>;
  revoke(courseId: string, userId: string, signal?: AbortSignal): Promise<EnrollmentResult<void>>;
}
