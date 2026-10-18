// Public surface of the `catalog` feature module (the public storefront: catalog listing, course
// detail, home and ratings summary). Everything another module, the composition roots or the routes
// may use is re-exported here; nothing outside this folder imports the module's internals directly
// (enforced by scripts/check-architecture.mjs — no deep cross-module import). The catalog has no
// domain of its own: its vocabulary is the shared contracts/course kernel.

// Application: the catalog query factory.
export { createCatalog } from './application/catalog';

// Ports: the I/O contract this feature depends on.
export type { CourseRepository } from './ports/CourseRepository';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiCourseRepository } from './adapters/course-repository';

// UI.
export { CatalogView } from './ui/CatalogView';
export { CourseView, Reviews, enrollErrorKey } from './ui/CourseView';
export { HomeView } from './ui/HomeView';
export { RatingSummary } from './ui/RatingSummary';
export { ValueStrip } from './ui/ValueStrip';
export { CatalogFailure, CatalogLoading, CourseNotFound } from './ui/CatalogStates';
