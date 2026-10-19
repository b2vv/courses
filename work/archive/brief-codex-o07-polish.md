# Codex brief — O07 polish: PR #3 review findings 1, 3, 4 (orders module)

Repository /Users/strelia/projects/courses, branch `feat/payment`. Read `AGENTS.md`. These are three
non-blocking cleanups from the PR #3 review (`work/review-pr3-patterns.md`, `work/backlog.md` B-24).
Finding #2 (in-memory throttle) is intentionally left as a named limitation — do NOT change it.
Log: `work/codex-o07-polish.log`. Do **not** commit. Backend only; keep behaviour identical except
where #1 adds a check.

## #1 — duplicate-JSON-key check: make it consistent, with a test
`validateUniqueJSONObject` (rejects duplicate JSON keys) is enforced in `decodeGrant`
(`internal/orders/adapters/http/handler.go`) but NOT in `decodeCheckout` (same file), which uses only
`DisallowUnknownFields`. The same helper is also copy-pasted in `internal/course/adapters/http/admin_handler.go`
and `internal/discussion/adapters/http/handler.go` (three copies).
- Apply the duplicate-key rejection to `decodeCheckout` too, so both orders write endpoints behave
  the same (a checkout body with a duplicate `courseSlug` → 400 invalid_request).
- Do NOT attempt a repo-wide shared-helper refactor across modules in this PR (that touches three
  modules and is out of scope) — only make the orders module internally consistent. If the orders
  file has two local copies after this, extract ONE unexported helper within that file.
- Add a contract/unit test: a checkout request with a duplicate JSON key is refused (400), mirroring
  however decodeGrant's dup-key refusal is tested (find that test; if none exists, add one for grant too).

## #3 — one explicit nilable parameter instead of variadic + two constructors
`service.New(...)` and `http.NewHandler(...)` take the payment provider / callback verifier as a
variadic, and there are two handler constructors (`NewHandler`, `NewHandlerWithConfig`). Replace the
variadic with a single explicit parameter that may be `nil`, and collapse to one constructor if that
removes the second cleanly. Update all call sites (`cmd/api/main.go`, tests). Behaviour identical:
nil provider → checkout 503, callback route not registered, as today.

## #4 — extract the repeated date-format literal
The RFC3339-ish date format string appears ~3× in `handler.go`. Extract a single `const`.

## Verify
Run with `DOCKER_CONFIG=/tmp/claude-501/dockercfg`: `make format-check lint typecheck test integration`.
Report each. If Docker is denied, say so; Courses 1 runs the gate. Do not claim a gate you did not run.

## Channels
If #1's "orders-only consistency" turns out to need the cross-module helper to avoid a third copy,
stop and say so rather than refactoring three modules silently. Final answer = changed files +
results.
