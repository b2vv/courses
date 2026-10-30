// Public surface of the `admin` feature module (course authoring/editor, course & user lists,
// analytics, course-grant panel, participant-report and refund queues, document library and the
// password-reset dialog). Everything the composition roots, the router or the routes may use is
// re-exported here; nothing outside this folder imports the module's internals directly
// (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: administration (users/courses) and analytics stats types, codes and pure helpers.
export * from './domain/administration';
export * from './domain/stats';

// Application: authoring and user-administration policies.
export * from './application/authoring';
export * from './application/users';

// Ports: the I/O contracts this feature depends on.
export type { AdminCourseRepository } from './ports/AdminCourseRepository';
export type { AdminStatsRepository } from './ports/AdminStatsRepository';
export type { AdminUserRepository } from './ports/AdminUserRepository';

// Adapters: the HTTP implementations wired by the composition roots.
export {
  createApiAdminCourseRepository,
  mapAuthoringCourseTree,
  mapAuthoringError,
} from './adapters/admin-course-repository';
export { createApiAdminStatsRepository } from './adapters/admin-stats-repository';
export { createApiAdminUserRepository } from './adapters/admin-user-repository';

// UI (the surfaces mounted by routes; internal pieces stay module-private).
export { AdminShell } from './ui/AdminShell';
export type { AdminNavItem, AdminNavSection } from './ui/AdminShell';
export { AdminNavIcon } from './ui/AdminNavIcon';
export { CourseCreateDialog, CourseEditorPending, CourseEditorView } from './ui/CourseEditorView';
export { CourseListFailure, CourseListPending, CourseListView } from './ui/CourseListView';
export { UserListFailure, UserListPending, UserListView } from './ui/UserListView';
export { AnalyticsFailure, AnalyticsPending, AnalyticsView } from './ui/AnalyticsView';
export { CourseGrantPanel } from './ui/CourseGrantPanel';
export { ParticipantReportQueue } from './ui/ParticipantReportQueue';
export { RefundQueue } from './ui/RefundQueue';
