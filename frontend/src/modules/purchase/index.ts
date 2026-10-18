// Public surface of the `purchase` feature module (course-access grants, orders and checkout).
// Everything another module, the composition roots or the routes may use is re-exported here;
// nothing outside this folder imports the module's internals directly (enforced by
// scripts/check-architecture.mjs — no deep cross-module import).

// Domain: grant and order/checkout types, codes and pure helpers.
export * from './domain/grant';
export * from './domain/purchase';

// Ports: the I/O contracts this feature depends on.
export type { CourseGrantRepository } from './ports/CourseGrantRepository';
export type { PurchaseRepository } from './ports/PurchaseRepository';

// Adapters: the HTTP implementations wired by the composition roots.
export { createApiCourseGrantRepository } from './adapters/course-grant-repository';
export { createApiPurchaseRepository } from './adapters/purchase-repository';

// UI.
export { OrderStatusView, submitCheckoutForm } from './ui/OrderStatusView';
export { LockedLessonPanel, purchaseErrorKey } from './ui/LockedLessonPanel';
