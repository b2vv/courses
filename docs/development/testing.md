# Testing

Ordered gates: format-check → lint → typecheck → unit → integration → build → E2E. Makefile provides each target. Rstest runs frontend domain/application/component/MSW adapter tests. Playwright uses the production Compose stack, including no-JavaScript SSR checks. Go table-driven tests exercise public domain/use-case/HTTP seams.

Testcontainers creates a disposable PostgreSQL instance for migration up/down/up, constraints and repository tests. Integration checks fail rather than skip when Docker is unavailable. No migration-down test runs against the developer DB. sqlc and Orval regeneration must leave generated files unchanged.

The foundation report distinguishes executed evidence from planned acceptance. Accessibility tests cover keyboard/focus, light/dark and zoom; automated checks are not a full accessibility certification.
