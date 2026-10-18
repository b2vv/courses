// Public surface of the `discussion` feature module. Everything another module, the composition
// roots or the routes may use is re-exported here; nothing outside this folder imports the module's
// internals directly (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: comment-thread types, codes and result shapes.
export * from './domain/discussion';

// Ports: the I/O contract this feature depends on.
export type { DiscussionRepository } from './ports/DiscussionRepository';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiDiscussionRepository } from './adapters/discussion-repository';

// UI.
export { DiscussionPanel } from './ui/DiscussionPanel';
