# S0.1b-1 — Params Go-ward (resolve Q2b) — WORK FILE (Codex appends)

Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append). Перший запис — рядок `ДАЛІ:` до першого інструменту.
Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінальна відповідь = стислий переказ файлу.
Base: branch `s0.0-spike` @ `7216648`.

## Журнал

### 2026-10-27 — inventory / pre-code checkpoint (Codex)
- ДАЛІ: read required sources, implement only Params + metadata merge, prove fidelity and ordered backend gates, commit on `s0.0-spike`; no merge/push.
- ФАКТ: initial tree clean except this owner-supplied untracked work file; preserved its contents. Required plan §16.8, generator, course contribution/fidelity/handler tests and hand/overlay parameter blocks read. Fidelity already compares complete path items, including parameters, without normalization of parameter arrays.
- РІШЕННЯ: owner-selected §16.8 resolution; Small continuation, READY. Params are operation-scoped under each path/method, matching the hand contract. Metadata is a name-keyed `parameters` map with description/schema validation only; reject orphan names, structural overrides and replacement arrays. Preserve slug minLength/maxLength/pattern. Sort enum uses domain.SortRecent/SortTitle (catalog domain/query.go).
- Scope: generator model/emission/overlay/tests, course contribution, overlay, regenerated parallel artifact, this journal. No runtime/shared persisted state, dependencies, hand spec or build wiring changes.
- Review route: existing separate Courses 2 session, exact reference pending owner handoff; delivery owner Codex session `01a17e76-af95-76f2-bd7a-8c2c14633b6b`; no new reviewer or shared test resource. Author runs explicitly requested Go gates. Independent review is pending, not claimed PASS.
- ДАЛІ: implement model + overlay deletion in one logical commit, then regenerate twice and execute requested gates.

### Implementation
- ФАКТ: Param/RouteBinding.Params added; emission sorts a cloned slice by name and validates declarations. Catalog query/path structure and sort constants now live in Go. Overlay structural blocks deleted in the same pending changeset; metadata merge rejects unknown names, replacement arrays/deletions and structural keys (including enum/required).
- Tests added for sorted output without caller mutation, query/path metadata preservation, orphan metadata with/without declared params, structural override attempts and invalid Go declarations.
- ДАЛІ: regenerate parallel artifact, verify fidelity without weakening assertions, run ordered gates.

### Fidelity / initial gate results
- ФАКТ: `cd backend && UPDATE_OPENAPI_GENERATED=1 go test -count=1 -v -run TestCatalogOpenAPIFidelity ./internal/course/adapters/http` exited 0 after overlay structure deletion:
```text
=== RUN   TestCatalogOpenAPIFidelity
    openapi_contract_test.go:103: catalog paths + complete referenced component closure + standalone Rating equal hand spec; two generations byte-identical
--- PASS: TestCatalogOpenAPIFidelity (0.08s)
PASS
ok  courses/internal/course/adapters/http 1.654s
```
- Regenerated `openapi.generated.yaml` is byte-identical to the existing tracked artifact; `git diff --exit-code -- backend/openapi/openapi.yaml backend/openapi/openapi.generated.yaml` exited 0. No artificial artifact change introduced. Fidelity assertions unchanged.
- Ordered gates: `gofmt -l .` exit 0, empty stdout. First sandboxed `go vet ./...` exited 1 because Go cache access was denied (`operation not permitted`); rerun with approved cache access exited 0, empty stdout. Build/tests next.
- Native checkpoint first failed on sandbox receipt-directory creation, then approved retry succeeded: `cee23e8eb6f9a3adac9c28a6f663502f-1-718c9460fb28`. Current identity confirmed from host event; exact courses/codex policy row found. No transcript reads.

### Ordered gates / final diff review
- `go build ./...`: exit 0, empty stdout.
- `go test ./internal/platform/openapigen/... ./internal/course/adapters/http/...`: exit 0:
```text
ok  courses/internal/platform/openapigen 1.204s
ok  courses/internal/course/adapters/http 6.644s
```
- Final diff review found metadata interception should apply only to path/operation parameter sites, not an ordinary nested response field named `parameters`. Scoped the condition and added a regression test. ДАЛІ: repeat ordered gates for this final code delta, then commit.

### Final verification / delivery
Final code snapshot, ordered from `backend/` (all exit 0):
```text
$ gofmt -l .
$ go vet ./...
$ go build ./...
$ go test ./internal/platform/openapigen/... ./internal/course/adapters/http/...
ok  courses/internal/platform/openapigen 1.035s
ok  courses/internal/course/adapters/http 6.726s
```
First three commands produced no stdout/stderr. `git diff --check` exit 0.
- Fidelity stayed green in the final suite after deletion; it compares both complete paths including parameters and checks two fresh generations plus tracked-artifact bytes. No fidelity assertions changed.
- Generated artifact SHA-256: `b3323f7c68b77bd38b5fde544f2863f2381483eaf080b0cc6360cb3228866934`; regenerated file already equals tracked bytes, therefore no artifact diff exists to commit. Hand spec unchanged.
- Final scope: five implementation/test/overlay files plus this journal. No new dependencies, business-module imports in openapigen, request DTO/error-parser/other-module changes or cutover wiring. Full product gates/e2e not run: this task explicitly requests the scoped backend gates and changes only the parallel generator.
- ДАЛІ: one atomic commit includes Go Params and overlay structure deletion. Existing Courses 2 session receives the committed diff via owner; independent final review pending. No merge/push.

## ПИТАННЯ
None. No missing or contradictory contract stopped implementation. Artifact regeneration intentionally produces identical bytes; no synthetic diff was manufactured.

## ВЕРИФІКАЦІЯ Courses 13 (writer) — 2026-10-27, HEAD ac970b7
Прогнав сам: gofmt/vet/build чисто; openapigen+course http тести ok; `TestCatalogOpenAPIFidelity` PASS
явно (1.5s, не кеш) — зелений ПІСЛЯ видалення param-структури з overlay = доказ, що Params-з-Go
відтворюють ту саму поверхню (watch-point Courses 2 виконано). M8 architecture ok. Overlay-params тепер
лише metadata (min/max/default/description за name), без in:/type:/enum. sort-enum з domain.SortRecent/
SortTitle. RouteBinding.Params []Param + type Param присутні. Q2b РОЗВ'ЯЗАНО. Готово на рев'ю Courses 2.
