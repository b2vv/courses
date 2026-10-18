/**
 * Admin-statistics view model. These are the shapes the analytics screen
 * consumes; the API adapter validates the wire against them. Counts are plain
 * numbers; the two averages are nullable because the server reports null (not a
 * misleading 0) when there is nothing to average.
 */

export type RoleStat = {
  readonly student: number;
  readonly instructor: number;
  readonly admin: number;
};

export type UsersStat = {
  readonly active: number;
  readonly blocked: number;
  readonly deleted: number;
  readonly byRole: RoleStat;
  readonly verified: number;
  readonly twoFactor: number;
};

export type CoursesStat = {
  readonly draft: number;
  readonly published: number;
  readonly archived: number;
};

export type GrantsStat = {
  readonly active: number;
  readonly blocked: number;
  readonly revoked: number;
};

export type OrdersStat = {
  readonly pending: number;
  readonly paid: number;
  readonly failed: number;
  readonly revenueMinorUAH: number;
};

export type ReportsStat = { readonly open: number; readonly resolved: number };

export type MediaStat = {
  readonly inProgress: number;
  readonly ready: number;
  readonly failed: number;
  readonly totalDurationSeconds: number;
  readonly avgDurationSeconds: number | null;
  readonly withDurationCount: number;
};

export type EngagementStat = {
  readonly activeLearners30d: number;
  readonly avgProgress: number | null;
  readonly completions: number;
};

export type InstructorStat = {
  readonly instructorId: string;
  readonly displayName: string;
  readonly publishedCourses: number;
  readonly activeEnrollments: number;
  readonly revenueMinorUAH: number;
};

export type Platform = { readonly singleInstructor: boolean };

export type Snapshot = {
  readonly users: UsersStat;
  readonly courses: CoursesStat;
  readonly grants: GrantsStat;
  readonly orders: OrdersStat;
  readonly reports: ReportsStat;
  readonly media: MediaStat;
  readonly engagement: EngagementStat;
  readonly instructors: readonly InstructorStat[];
  readonly platform: Platform;
};

export type Metric = 'signups' | 'enrollments' | 'orders' | 'revenue';
export type Bucket = 'day' | 'week';

export type TimeseriesPoint = { readonly bucketStart: string; readonly value: number };

export type TimeseriesRange = {
  readonly metric: Metric;
  readonly bucket: Bucket;
  readonly from?: string;
  readonly to?: string;
};

/**
 * Every outcome of a stats read: the four transport failures every adapter can
 * raise, the admin gate (401/403), and the timeseries input rejections.
 */
export type StatsErrorCode =
  | 'invalid_range'
  | 'range_too_wide'
  | 'unknown_metric'
  | 'unknown_bucket'
  | 'authentication_required'
  | 'forbidden'
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable';

export type StatsError = { readonly code: StatsErrorCode };

export type StatsResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: StatsError };
