# A12 — degradation and in-flight guards

## Delivered

- R37: the admin list primary-read boundary renders the kit `ErrorState`; its retry button
  invalidates the current route and re-issues the list query. The component test drives a
  `503 dependency_unavailable` followed by a successful MSW response and observes two GETs.
- R38: the editor route now uses the same primary-read `ErrorState` boundary. Its component
  test proves that a 503 shows no course form, then retry re-issues the detail GET and reveals
  the editor only after a successful response.
- R40: no production change was needed. The existing React Hook Form `isSubmitting` guard
  disables Save while its promise is pending. The component test holds a real adapter PUT in
  flight, presses Save again, and observes exactly one PUT.
- The UI kit's `ErrorState` accepts either its existing retry link or a required-label retry
  button callback. No new screen or dependency was added.

## Review

- Specification axis: no missing, partial, extra, or incorrect A12 behavior found.
- Standards axis: the duplicated editor failure implementation was removed in favor of the
  existing admin read boundary. The two small MSW route harnesses remain local to their screen
  tests; a shared test-helper API for two cases was rejected under ponytail-full.

## Verification and denials

No check was represented as passed.

- **Denied by the brief: oxfmt.** `make format-check` / `npm run format:check` were not run.
- **Denied by the brief: Rstest.** The R37, R38, and R40 component tests were not run.
- **Denied by the brief: Docker.** Docker-backed Go checks, integration, generated checks,
  builds, Compose startup, seeds, smoke, and e2e were not run.
- `make lint`, `make typecheck`, and the full ordered `make check` were not run; the repository
  verification chain was left to the reviewer after its first denied gate.
- Git was not invoked. No file in a `generated/` directory was written.

## Limitation and next task

The new evidence is recorded as test code but has no local execution result because the
required runner is denied. The next planned slice is A13, shared browser sessions and the e2e
login budget.

## Reviewer verification (measured in the reviewing session)

One review fix: the new R38 test shadowed the module-scope `admin` fixture with a local
queries object, breaking typecheck — local renamed. Then `make format` applied and
format-check, lint, typecheck and the full frontend suite pass (0 failures): R37/R38
prove ErrorState-with-retry on 503 for both screens, R40 proves the single in-flight
PUT. Provenance: implemented with Codex (gpt-5.6-sol); reviewed, corrected and
verified here.
