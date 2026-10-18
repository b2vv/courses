// Public surface of the `teach` feature module (instructor course roster, cross-course learner
// overview and review moderation). Everything the composition roots, the router or the routes may
// use is re-exported here; nothing outside this folder imports the module's internals directly
// (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: teach types, error codes and the learner/moderation page shapes.
export * from './domain/teach';

// Ports: the I/O contract this feature depends on.
export type { TeachRepository } from './ports/TeachRepository';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiTeachRepository } from './adapters/teach-repository';

// UI.
export { CourseRoster } from './ui/CourseRoster';
export { LearnerOverview } from './ui/LearnerOverview';
export { ReviewModerationRow } from './ui/ReviewModerationRow';
