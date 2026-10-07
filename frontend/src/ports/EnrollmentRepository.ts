import type {
  EnrollmentResult,
  Membership,
  MembershipCard,
} from '../domain/enrollment/enrollment';

export interface EnrollmentRepository {
  membership(courseId: string, signal?: AbortSignal): Promise<EnrollmentResult<Membership>>;
  enroll(courseId: string, signal?: AbortSignal): Promise<EnrollmentResult<Membership>>;
  myMemberships(signal?: AbortSignal): Promise<EnrollmentResult<readonly MembershipCard[]>>;
}
