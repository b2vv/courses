# C04 component and Query verification

2026-09-03. This bounded task changed tests and this evidence file only. Catalog implementation, generated client, API adapter, routes and backend were left to their owning tasks.

## Result

Rstest passed **15 tests in 2 files**, with **0 skipped**. TypeScript typecheck and targeted Oxlint/Oxfmt checks passed.

[CatalogView.test.tsx](../../frontend/src/presentation/storefront/CatalogView.test.tsx) covers eight public presentation behaviors:

- Published course links have meaningful accessible names, correct URLs and summaries.
- Course detail includes its description and the accurate foundation availability message.
- GET form submission keeps edited search/sort and the selected limit while omitting the old offset, which restores the catalog's first-page default.
- Next-page URLs preserve search, sort and limit and advance offset.
- A zero-result catalog exposes an empty state and recovery link.
- A request beyond the result range retains the known total and offers first-page recovery.
- Loading is announced; a missing course offers catalog navigation.
- Error recovery uses the complete current URL from a real in-memory TanStack Router.

[catalog-queries.test.tsx](../../frontend/src/app/composition/catalog-queries.test.tsx) covers seven Query behaviors using real QueryClient/useQuery:

- Request B becomes active while A is pending. A's AbortSignal is aborted. B resolves first; even when the fake external repository deliberately ignores A's cancellation and resolves it later, B remains visible.
- Search, sort, limit and offset each independently select a different cached result; returning to the original parameters still retrieves its original data.
- Two QueryClient instances requesting an identical URL retain different request results; clearing A does not remove B's data.
- Domain failure becomes a typed Query error and is not cached as successful result data. Missing-course detail retains its error code.

Fake implementations exist only at the CourseRepository port. React, Query, Router, domain/application and components are real. Tests interact through rendered controls, DOM forms, returned hook state and QueryClient's public API; they do not mock implementation internals. Query clients and pending requests are created inside each test and cleaned up.

## Commands

Working directory: `frontend/`.

```sh
npm test -- src/presentation/storefront/CatalogView.test.tsx src/app/composition/catalog-queries.test.tsx
npm run typecheck
node node_modules/oxlint/bin/oxlint src/presentation/storefront/CatalogView.test.tsx src/app/composition/catalog-queries.test.tsx --deny-warnings
node node_modules/oxfmt/bin/oxfmt --check src/presentation/storefront/CatalogView.test.tsx src/app/composition/catalog-queries.test.tsx
```

## Limits and next gate

These are component/Query seam checks. They do not prove network transport correctness, cookie forwarding, composition-root request scoping, live PostgreSQL behavior, initial SSR HTML, hydration, browser layout or production container operation. API-adapter tests and the parent production Compose/Playwright gate provide those separate checks. C04/C05 completion still waits for live backend and full runtime evidence. No implementation bug remained in the tested component/Query behaviors and no user decision was needed.
