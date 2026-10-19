ДАЛІ: Read the enrollment plan, current frontend contracts, generated endpoint status unions, and sibling patterns before making the L7c2 slice.

## Step 1 — briefing and scope read

Read `work/codex-working-model.md`, `work/brief-codex-L7c2.md`, the repository briefing, and `docs/enrollment/plan.md` through T7/En07. Constraints recorded: no git, `work/tasks.md`, phase report, backend, or hand edits in `generated/`; build/e2e stay with the reviewing session. Next I will verify every required frontend contract and the exact generated response statuses before implementation.

## Step 2 — contract inspection stopped

Inspected the existing enrollment domain/port/adapter/query layers, admin route and shell, roster infinite-list/focus pattern, confirm Dialog, locales, and both generated endpoints. The L7c2 ports and generated functions exist, but an API-contract inconsistency below would make the required exact error mapping incorrect, so no implementation changes were made.

## Discussion

**What is wrong.** The generated client declares the participant-report queue `400` as `BadRequestResponse` / `ErrorResponse` (`frontend/src/adapters/api/generated/courses.ts:956-959`; `models/badRequestResponse.ts:12`), whose generated `ErrorResponseErrorCode` enum does **not** include `invalid_cursor` (`models/errorResponseErrorCode.ts:12-99`). But the actual queue handler returns `domain.ErrInvalidCursor` for malformed, duplicate, or unknown cursor query input (`backend/internal/discussion/adapters/http/handler.go:470-473`), and the cursor decoder also returns that error for malformed values (`backend/internal/discussion/adapters/postgres/participant_report.go:113-129`). The error writer serializes `err.Error()` as the public code (`backend/internal/discussion/adapters/http/handler.go:735+`), so runtime emits `invalid_cursor`.

**Why this would cause a bug.** The brief requires the adapter to map error codes *exactly* as the generated client declares, without inventing or omitting any. Implementing against generated means the `400` map can only accept a generated code such as `invalid_request`; it will reject the backend's actual `invalid_cursor` and degrade it to `invalid_response`. Adding `invalid_cursor` to the enrollment union/adaptor would instead violate the explicit generated-only rule. Either behavior makes a valid documented endpoint failure misclassified.

**Proposed correction.** The reviewing session should reconcile the OpenAPI/generated schema with the backend: either add `invalid_cursor` to the generated error-code contract for `listParticipantReports`, or change the backend handler to emit an already-generated `invalid_request` code. Regenerate the checked-in client after any OpenAPI change. Then amend this brief to name the accepted code.

**What I did in the meantime.** Stopped before code changes. I did not touch backend, `generated/`, git, `work/tasks.md`, or the phase report.

## Step 3 — discussion resolved by reviewer

Reviewer verified `domain.ErrInvalidCursor` is `errors.New("comment_cursor_invalid")` in `backend/internal/discussion/domain/comment.go:20`, not `invalid_cursor`. `comment_cursor_invalid` is present in generated `ErrorResponseErrorCode` and OpenAPI. I will add that exact generated code to `EnrollmentErrorCode` and map queue `400` to it, with no backend/OpenAPI/generated change. Continuing L7c2.

## Step 4 — data layer, route, and admin queue implemented

Added participant-report queue domain types and repository methods; adapter guards for every queue item field and page limit; and the three generated transports. Exact generated status mappings are queue `400→comment_cursor_invalid`, `401→authentication_required`, `403→report_forbidden`, `503→dependency_unavailable`; resolve `401/403/503` respectively; revoke `400→invalid_request`, `401→authentication_required`, `403→enrollment_forbidden`, `409→membership_conflict`, `500→internal_error`, `503→dependency_unavailable`. Added infinite queue query options, resolve invalidation, and revoke invalidation plus its 409 refresh behavior.

Added `/admin/participant-reports`, loader prefetch, admin nav item, and an accessible labeled queue with no email rendering, former-author fallbacks, kind/state text, retry/empty/error states, per-row pending actions, keyset load-more pending/focus, and a React Aria confirm dialog before revoke. Added English/Ukrainian strings and query-cache coverage for resolve/revoke. Regenerated `src/routeTree.gen.ts` through TanStack's generator (not hand-edited). `npm run typecheck` passed.

## Step 5 — verification

The first `npm test` run executed 630 tests: 629 passed, including the new query-cache test; one route-tree allowlist assertion failed only because it needed the newly added `/admin/participant-reports` route. I updated that expected route list. Then `cd frontend && npm run typecheck` passed and `npm run lint` passed (architecture boundaries: 174 files). The final `npm test` attempt was **not run** because the sandbox denied Rstest's bind to `::1:3000` with `listen EPERM`; no green full-test result is claimed. Build and e2e were not run, per brief: they remain for the reviewing session.

ДАЛІ: Reviewing session should rerun `npm test` outside the sandbox port restriction, then its prescribed build/e2e/full gate before review; no backend, OpenAPI, hand-generated client, git, `work/tasks.md`, or phase-report changes were made.
