# Testing

Ordered gates: format-check → lint → typecheck → unit → integration → build → E2E. Makefile provides each target. Rstest runs frontend domain/application/component/MSW adapter tests. Playwright uses the production Compose stack, including no-JavaScript SSR checks. Go table-driven tests exercise public domain/use-case/HTTP seams.

Testcontainers creates a disposable PostgreSQL instance for migration up/down/up, constraints and repository tests. Integration checks fail rather than skip when Docker is unavailable. No migration-down test runs against the developer DB. `make check-generated` regenerates sqlc and Orval and fails on a tracked diff.

The recovery Playwright project controls only the local Compose stack: it stops API/PostgreSQL, verifies safe failure and retry, then restores services in `finally`. It also verifies retained seed data and sanitized Nginx logs. Accessibility tests cover keyboard/focus, locale, dark mode and zoom; automated checks are not a full accessibility certification.
