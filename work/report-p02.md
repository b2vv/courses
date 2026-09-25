# P02 report

## Done

- `LearningCourse.CanBeViewedBy`, `VideoCourse.CanBeViewedBy` and transcript `Course.CanBeViewedBy` now map their local actor/course shapes to `access.CanView` in one return statement.
- `access.CanView` now requires a non-empty `UserID` for the owner branch, with the payment-bearing transport comment.
- Added the empty subject × unpublished ownerless course test; expected result is `false`.
- Updated `shared_rules_test.go` to describe and verify mappings, including the corrected empty-owner/empty-actor row.
- Added `access_mapping_ast_test.go`: parses non-test Go files under `internal/**`, requires exactly three `CanBeViewedBy` methods, and requires each body to be a single `return access.CanView(...)`.

## Verification

- `gofmt -d` on all changed Go files: passed (no output).
- `go vet ./...`: passed.
- Targeted unit tests for architecture, access, course, playback and transcript: passed.
- `go test ./...`: not passed in this sandbox. All packages passed until `platform/httpserver`; `TestShutdownDrainsAnInFlightResponse` was denied by the environment at `listen tcp 127.0.0.1:0: bind: operation not permitted`. This is an environment denial, not reported as a passing check.

## AST red proof

Temporarily added an extra inline boolean branch to the course `CanBeViewedBy`, then ran:

`go test ./platform/architecture -run '^TestEveryCanBeViewedByDelegatesToAccessCanView$'`

The test failed at `internal/course/domain/learning.go` with `CanBeViewedBy must return access.CanView(...)`. The method was restored and the targeted package tests passed.

## Discussion

None. The three mappings agree with the contract truth table; the only intentional truth-table change is the empty-subject/empty-owner unpublished case required by P02.

## Правка після рев'ю

- У `access.CanView` дужками явно позначено порядок перевірки гілки власника; поведінка не змінюється.
