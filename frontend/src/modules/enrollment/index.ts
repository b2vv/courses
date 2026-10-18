// Public surface of the `enrollment` feature module. Everything another module, the composition
// roots or the routes may use is re-exported here; nothing outside this folder imports the module's
// internals directly (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: membership, roster, participant-report and refund-queue types and codes.
export * from './domain/enrollment';

// Ports: the I/O contract this feature depends on.
export type { EnrollmentRepository } from './ports/EnrollmentRepository';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiEnrollmentRepository } from './adapters/enrollment-repository';

// UI.
export { ReportDialog } from './ui/ReportDialog';
