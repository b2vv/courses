# S0 — Go→openapi gate: generate openapi from Go (design, plan-first)

Owner decision 2026-10-27: close S0 (audit §S0) by **generating `openapi.yaml` from Go** and gating
drift in `check-generated`, not by contract tests. This doc is the plan-first artifact; no code until
it passes plan-critique / plan-defense and the owner approves. Root: `main @ 11bb03f`. Reviewer
(plan): Courses 2 [e6a45d].

## 1. Problem

`openapi.yaml` is hand-maintained (**85 paths, 120 schemas, 3694 lines**). `check-generated` gates
sqlc (DB→Go) and Orval (openapi→FE) but **not** the Go-handler→openapi boundary. So a handler whose
response shape drifts from the spec is caught only by e2e — which is how **H2** (the `Block`
bullet-list serialization) survived green gates. The fix must make Go the source of truth for the
response/request shapes and fail the build on drift.

## 2. Grounding facts (measured on `main @ 11bb03f`)

- **One** custom `MarshalJSON` in the whole backend: `course/domain/document.go:98` `Block` — the exact
  H2 type. Struct-tag reflection does **not** see it, so it is the one type a schema reflector would
  get wrong. Everything else is plain tagged structs.
- **One** response seam: `httpserver.WriteJSON(w, status, value)`; handlers pass a concrete DTO at the
  call site (e.g. `pageDTO`, `courseDTO`, `priceDTO`, `ratingDTO` in `course/adapters/http`).
- `openapi.yaml` holds **~249 hand-written prose fields** (174 `description:`, 75 `summary:`, 0
  `example:`) plus a hand-written `servers:` block. Prose is documentation value a naive regeneration
  would destroy.
- Only Go OpenAPI tooling already vendored: `github.com/getkin/kin-openapi` (v0.149.0) — used today for
  response validation in handler tests (`openapi3filter.ValidateResponse`). It ships `openapi3gen`
  (Go type → `*openapi3.Schema` by reflection) and `openapi3` (model + YAML marshal).
- Routes register on stdlib `http.ServeMux` as `"METHOD /path"` string patterns across
  `*/adapters/http/*.go`. AGENTS forbids a code-first framework / service locator, so the generator is
  in-repo, not a framework swap.

## 3. The two hard requirements (from the audit rule + Courses 2 plan guidance)

1. **H2-class detection is explicit and fail-closed.** The generator must *know* which types have a
   custom `MarshalJSON` and how to schema them, and must **fail the build** when it meets an
   unrecognized `MarshalJSON` rather than silently emitting the (wrong) struct-tag schema. A silent
   skip reintroduces exactly the H2 drift the gate exists to stop. Mechanism options:
   - (a) **marker + explicit override registry**: a `map[reflect.Type]*openapi3.Schema` of
     hand-written schemas for custom-marshal types; the generator, while reflecting, asserts that any
     type implementing `json.Marshaler` is present in the registry, else `log.Fatal` (build red). One
     entry today (`Block`); the assertion is what keeps it honest as types are added.
   - (b) marshal-sample inference: marshal representative values and infer a schema. Rejected —
     non-deterministic field presence (omitempty, nil maps) and can't express a closed enum/oneOf.
   - **Recommendation: (a).** Minimal (one override), deterministic, and the fail-closed assertion is
     the spec-conformance guarantee.
2. **Deterministic output → clean diff.** 120 schemas must emit in a **stable order** (sorted keys),
   stable required-field order, stable path order, fixed YAML marshaling, or `check-generated` diffs
   on noise. The generator sorts every map-derived list and pins the YAML encoder.

## 4. Approach options for the route/type surface

The generator needs, per route: method, path, path/query params, request body type, and
response type(s) per status. `WriteJSON(w, status, dto)` gives the response type only at the call
site, not in an enumerable registry. Options:

- **A — central typed route registry.** One Go file lists all 85 routes as values binding
  (method, path, params, reqType, respType-per-status) to the *existing* handler DTO types. The
  generator reflects those types (with the §3.1 override) and emits paths+schemas. A cross-check test
  asserts the registry's (method,path) set equals the mux's actually-registered set (parsed like
  `boundaries_test`), so the registry can't drift from the real routes.
  - Pro: no handler rewrite; stdlib intact; the DTO types stay the single shape source.
  - Con: the registry re-declares the route list (mitigated by the cross-check test); prose must be
    carried (see §5).
- **B — per-handler contract registration.** Extend each module's `Register` to also hand its
  contracts to a collector. Keeps route+contract co-located but touches every handler and enlarges the
  Register signature across modules.
- **C — annotations (swaggo).** `// @` comments per handler, parsed by a new tool. Verbose across 85
  routes, comments drift from code, new dependency. Rejected.
- **Recommendation: A**, with the mux cross-check test as the anti-drift guard.

## 5. Open decision for the owner — prose (249 fields)

Generation must not silently delete the 174 descriptions / 75 summaries. Options:
- **(i) Prose in Go**: carry description/summary as struct field doc-comments or a maintained
  `map[path|field]string` the generator reads. Go stays the single source. Cost: migrate ~249 strings
  once.
- **(ii) Overlay**: generate the machine-truth skeleton (paths, schemas, types, required/nullable,
  enums), then deep-merge a curated `openapi.overlay.yaml` holding only prose. Generator is more
  complex (merge semantics) but no mass migration.
- **Recommendation: (i)** — it is the only option that keeps a *single* source of truth (the audit's
  whole point); the migration is mechanical and one-time. **Owner to confirm**, since (ii) trades a
  one-time migration for permanent merge complexity.

## 6. How drift fails the gate

`make generate` runs the generator, writing `backend/openapi/openapi.yaml`. `check-generated` already
does `git diff --exit-code` over generated dirs; add `backend/openapi/openapi.yaml` to that set. A
handler/DTO change that isn't reflected in a regenerate → non-empty diff → `check-generated` red. The
§3.1 fail-closed assertion covers the one case a diff can't: a new custom-marshal type whose struct
tags *happen* to match its old schema — the build dies before emit. Orval still consumes the generated
yaml, so FE types regenerate from the same truth.

## 7. Multi-session slice boundaries (each self-contained, own proof)

A "ready slice" compiles, gates green, and leaves the tree shippable:

- **S0.1 — generator skeleton + one module (course), behind a parallel output.** Generator emits to
  `openapi.generated.yaml`; a test asserts it is semantically equal to the hand-written `openapi.yaml`
  for the course paths (using `openapi3` equality), proving fidelity *including* the `Block` override,
  before anything switches over. Proof: new test green; no production path changed yet.
- **S0.2 — cover all 85 routes** in the registry; the fidelity test widens to the whole document and
  must pass (this is where hand-vs-generated discrepancies surface and get reconciled — expect real
  findings, like the phase reports warn).
- **S0.3 — cutover**: `openapi.yaml` becomes generated, prose moved per §5, `make generate` +
  `check-generated` wired, Orval re-runs from it. Proof: ordered `make check` incl. `check-generated`
  and e2e green.
- **S0.4 — the fail-closed assertion test**: add a throwaway custom-marshal type in a test fixture,
  assert the generator fails; remove. Proof: red-on-unregistered demonstrated.

## 8. Risks / open questions

- `openapi3gen` fidelity for the existing 120 schemas (nullable via pointer, `oneOf`, enums,
  `additionalProperties`) may not match the hand-written spec 1:1; S0.2 is where that cost is paid and
  may surface many small reconciliations.
- Prose decision (§5) is a prerequisite for S0.3.
- If S0.2 reveals `openapi3gen` can't express a shape the hand spec uses, that schema needs a §3.1-style
  override too — the fail-closed registry generalizes to "any type the reflector can't do faithfully".

## 9. Verification plan

Each slice: ordered `make check` by impact; S0.3 (persisted contract + FE regen) needs full e2e.
Generator has unit tests (determinism: regenerate twice → byte-identical; fidelity: generated ≡ hand
spec; fail-closed: unregistered marshaler → build error). Independent review (Courses 2) per slice.

## 10. Plan-critique outcome (2026-10-27) — GATE RED, approach reconfirm needed

Four independent lenses (fresh subagents, blind to each other): **4 launched / 4 returned**. Tally:
**11 blocking, 7 non-blocking**. The blockers converge on one approach-level finding that supersedes
the detail; the plan is **not advanced to plan-defense** until the owner reconfirms direction.

**Decisive thread (logic ×2 + edge ×8 + complexity ×1):** the reflection generator does **not** achieve
fail-closed for the H2 class it exists to stop.
- Reflection is blind to far more than the one `Block` `MarshalJSON`: string-enum types (no `enum`
  emitted), the ~110-value error-code enum (`WriteError(w,r,status,"code")` bare literals — no Go type
  at all), `*T`→`nullable`+`required` (openapi3gen's default is the opposite), per-field array
  nullability (`nil` vs `make([]T,0)` — identical Go type), validation keywords
  (minLength/pattern/maximum/format — a class §5 didn't count), `additionalProperties` (false global but
  ClientLog open), Go-name↔spec-name divergence (`Section`→`LessonSection`), and one embedded type
  yielding two schemas (`Course` vs narrower `CourseDetail`). Each needs a **manual override that
  itself drifts silently** — the §3.1 assertion only fires on a *new unregistered* `json.Marshaler`,
  and the §4-A cross-check guards only the route *set*, not route→type. So "Go = single source of
  truth" largely evaporates into a dozen hand-maintained override classes.
- **The H2 class is already caught by vendored, proven code** (complexity lens, with evidence): 8
  modules run `openapi3filter.ValidateResponse` against `openapi.yaml`, and
  `course/.../document_handler_test.go:TestDocumentRecursiveSchemaRoundTrip` already validates the real
  `Block` `{"type":"list","ordered":true}` response (the H2/SA02 shape) against its schema. H2 survived
  from missing *coverage* at the time, not a missing mechanism — and the coverage is now built.
- Therefore the generator's **only** unique benefit over reuse is removing manual maintenance of the
  3694-line spec — a separate, larger goal than the audit §S0 drift-gate.

**Cheaper reuse-path (complexity lens recommendation):** (1) extend the existing `ValidateResponse`
harness to the uncovered `(route, status, shape)` gaps; (2) add a coverage-gate test — "every mux-
registered route has ≥1 `ValidateResponse` case" — reusing the `boundaries_test` mux-parse, pointed at
the validator not a generator. Closes the §S0 drift goal (handler/spec drift can't reach a green gate),
reusing proven code, with none of: 85-route registry, override registry, fidelity/determinism
machinery, 249-prose migration, overlay, cutover, Orval rewire.

**Non-blocking (fold in whichever path is chosen):** [slop] drop §5's "single source of truth" totem,
reframe on the real migration-vs-merge axis; [slop] collapse S0.4 into the §9 unit tests (→ 3 slices);
[slop] record the generate-vs-contract-tests rationale in §1; [logic] pin Go-gen before Orval in `make
generate` + add `openapi.yaml` to the `check-generated` diff set; [logic] one-DTO expressiveness spike
before treating fidelity as a readiness bar (resolve §7↔§8); [edge] registry `respType-per-status`
must allow a list → `oneOf` (login 200 = `AuthenticatedUser`|`TwoFactorRequired`); [edge] assert
generated paths ≡ current hand-spec paths at cutover.

**Open blocking: 11 — gate RED.** They do not individually fold into edits because they are dominated by
the approach reconfirm: if the owner's intent is only to close §S0 drift, the reuse-path is the
proportionate design and most generator blockers become moot; if the intent also includes retiring the
hand-maintained spec, the generator proceeds but the plan must first absorb the edge/logic blockers
(override the full class list — enums, error-codes, nullability, validation, naming, embeds — plus a
marshaler↔schema fidelity test and per-route `ValidateResponse` so the manual overrides can't drift).
Reconfirm with owner before plan-defense.

## 11. Revision v2 — owner chose FULL GENERATOR (2026-10-27), blockers absorbed

Owner reconfirmed generate-from-Go after seeing §10, **including** retiring the hand-maintained spec
(not only the drift-gate). That resolves the complexity blocker as an accepted tradeoff: the generator's
extra cost over the reuse-path buys elimination of 3694 hand-maintained lines, which the owner wants.
The remaining 10 logic/edge blockers are absorbed below. Supersedes §3.1, §4, §5, §7 where they conflict.

**11.1 Two-layer fail-closed guarantee (resolves logic-B1, logic-B2, and folds the reuse harness in).**
A generator alone is not fail-closed — reflection can silently emit a wrong schema and the diff stays
empty. So the gate is two layers that together close the H2 class:
- **Layer A — generation from Go.** Reflection for plain structs + explicit mechanisms (11.2) for every
  reflection-blind class, each with a build-time fail-closed assertion.
- **Layer B — per-route conformance (the binding guard).** Every mux-registered route must have ≥1
  `openapi3filter.ValidateResponse` case asserting the **real handler's bytes** validate against the
  **generated** spec; a coverage-gate test (reusing the `boundaries_test` mux-parse) fails if any route
  has none. This is what forces route→type correctness (logic-B2) and catches a stale override schema
  for `Block` or any type (logic-B1): if the generated schema drifts from what the handler emits,
  ValidateResponse reddens. The existing 8-module harness is extended to full coverage, not discarded.

**11.2 The reflection-blind override surface (resolves all 8 edge blockers).** Go remains the source;
each class names how its truth is *derived from Go* (not hand-kept) wherever possible, else is a declared
override the generator fail-closes on:
| Class | Mechanism (Go-derived where possible) | Fail-closed rule |
|---|---|---|
| custom `MarshalJSON` (Block) | override registry `map[reflect.Type]*openapi3.Schema` + Layer-B round-trip | any `json.Marshaler` in a DTO not in registry → build fails |
| string-enum `type X string` | registry maps the type → its `const` value set (single source: a test asserts the set equals the consts via a declared slice per type) | any named `string`-kind type in a DTO not registered → build fails |
| error-code enum (~110) | static-parse every `WriteError(w,r,status,"literal")` across `*/adapters/http` (like boundaries_test) → union is the enum | a literal not matching the emitted enum, or a non-constant code arg → build fails |
| `*T` nullability | rule: `*T` → `nullable:true` **and** stays `required`; non-pointer → required; `,omitempty` → optional | pinned openapi3gen config; asserted by Layer B |
| per-field array nil-vs-`make([]T,0)` | **not derivable from the type** → per-field annotation in the route registry (`nullableFields`) | a `[]T` field without an explicit nil/non-nil decision → build fails |
| validation keywords (minLength/pattern/maximum/format) | **not in Go** → curated overlay (11.3) | overlay keys must resolve to an emitted schema/field, else build fails |
| `additionalProperties` | default `false`; open-map whitelist (ClientLogRequest) in registry | a map-typed/open schema not whitelisted but emitted open → build fails |
| Go-name ↔ component-name (`Section`→`LessonSection`) | explicit name-map in the registry; named components + `$ref` (never inline); `openapi3gen` `ThrowErrorOnCycle`→ref for self-referential `Section.Columns` | a reflected named type with no component-name mapping → build fails |
| embedded type → two schemas (`Course`/narrower `CourseDetail`) | treat each endpoint response type as its own schema; `CourseDetail` is a distinct registry entry (narrower `status` enum, non-null `publishedAt`), not a reuse of `Course` | drops the "one schema per Go type" premise |

**11.3 Prose + validation = curated overlay; structure = generated (resolves §5, edge-5/edge-7 count).**
Go determines structure (paths, types, required, refs, enums-from-consts, error-codes-from-literals,
nullability-from-pointers); only what Go genuinely does not express goes to a curated
`openapi.overlay.yaml`: the ~249 prose fields (description/summary), validation keywords, per-field
array-nullability, `CourseDetail`'s narrowing, the additionalProperties whitelist. The generator
deep-merges overlay onto the generated skeleton and **fail-closes if any overlay key does not resolve to
an emitted path/schema/field** (so a renamed field can't leave orphaned prose). This is a cleaner source
split than §5(i)/(ii): structure can never drift (generated), curated metadata can never silently orphan
(merge assertion). `servers:` and `securitySchemes:` live in the overlay too.

**11.4 Determinism (keeps §3.2).** Sorted schema/path/required keys, stable `$ref` names from the
name-map, pinned YAML encoder; `make generate` twice → byte-identical; Go-gen runs **before** Orval in
the `generate` target; `backend/openapi/openapi.yaml` added to the `check-generated` diff set.

**11.5 Revised slices (folds slop S0.4 → unit tests; 3 build slices + spike).**
- **S0.0 — expressiveness spike (resolves logic-L4 §7↔§8):** one course DTO end-to-end through Layer A
  — prove `*priceDTO`/`*ratingDTO` nullability, a string-enum, named-component `$ref`, and the `Block`
  override reproduce the hand-spec exactly, or pin the override. Proof: a focused fidelity test green.
- **S0.1 — generator + full override surface (11.2) + Layer-B coverage-gate**, emitting to
  `openapi.generated.yaml` beside the hand spec; whole-document fidelity test (generated+overlay ≡ hand
  spec) — this is where reconciliations surface. Proof: fidelity + coverage-gate green; no cutover yet.
- **S0.2 — cutover:** `openapi.yaml` becomes generated+overlay; Orval re-runs from it; `make generate` +
  `check-generated` wired; assert generated paths ≡ prior hand-spec paths (edge non-blocking). Proof:
  ordered `make check` incl. `check-generated` + e2e green.
- Generator **unit tests** (not a slice): regenerate-twice byte-identical; each override type
  marshal→validate; unregistered `json.Marshaler`/string-enum/non-constant error-code → build fails;
  overlay orphan key → build fails. `respType-per-status` accepts a list → `oneOf` (login 200).
- §1 gains the generate-vs-contract-tests rationale (slop): contract tests prove conformance to a spec
  that still rots by hand; the generator additionally retires the 3694 hand-maintained lines — the
  owner's stated intent.

**Open blocking after v2: 0** (owner accepted the generator-vs-reuse tradeoff; the 10 logic/edge
blockers are absorbed in 11.1–11.5). Ready for plan-defense.

## 12. plan-defense (2026-10-27)

Two behaviors added to close defense gaps:
- **12-a (overlay integrity).** The generator parses `openapi.overlay.yaml` with a strict loader and
  **fail-closes on malformed YAML or an unknown overlay shape**, in addition to the orphan-key assertion
  (11.3). A missing overlay file is an explicit error, not an empty merge.
- **12-b (name-map injectivity).** The Go-name→component-name map (11.2) must be **injective**: two Go
  types mapping to the same component name → build fails. This prevents a silent schema collision that
  determinism alone would not catch.

"Input" to this tool is the Go source (types, routes, `WriteError` literals) + the overlay; the "user"
is the developer running `make generate`/`make check`; the failure mode is a **red build**, never a
silently-wrong spec.

| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | кривий інпут (unhandled type, non-constant error-code, unregistered marshaler/enum, `[]T` без nil-рішення) | §11.2 fail-closed rules + §11.5 unit tests | закрито |
| 2 | відмова залежності (openapi3gen не виражає схему; overlay малформ/відсутній; Orval на спеку) | §11.2 (override коли openapi3gen не може) + §12-a (overlay) + S0.0 spike + S0.2 (`make check`+e2e) | закрито |
| 3 | гонка й повторний виклик (`make generate` двічі / коли вже актуально) | §11.4 (regenerate-twice byte-identical → `check-generated` чистий); детермінований build-step, без мережі/конкурентності | закрито |
| 4 | порожні та граничні (0-route модуль, рекурсія `Section.Columns`, дубль-назви компонентів, 1-значний enum, route без ValidateResponse) | §11.1 Layer-B coverage-gate (route без кейсу → red) + §11.2 (cycle→ref, enum-from-consts) + §12-b (інжективність назв) | закрито |
| 5 | rollback | §12 (нижче) | закрито |

**Rollback (row 5).** Cutover (S0.2) лише замінює вміст `openapi.yaml` на generated+overlay — обидва у git.
Відкат = `git revert` коміту cutover: повертає ручний спек, Orval детерміновано регенерує FE з відкоченого
спеку. **Незворотного немає**: жодної data-міграції, жодного зовнішнього ефекту (pre-launch, test-mode, спек
— файл). Єдине «одностороннє» — видалення ручного супроводу спеку, але історія в git тримає попередній зміст.

**Дірок: 0 — гейт зелений.** План іде на рев'ю плану (Courses 2), далі — таски S0.0→S0.2.

## 13. Plan-review delta (Courses 2, 2b98c4a → 2026-10-27) — B1 absorbed, 4 fold-ins

Courses 2 approved the approach but found one BLOCKING-for-plan the §12 defense missed, plus 4
non-blocking fold-ins. Absorbed here; supersedes §11.1/§11.2/§11.3/§11.5 where in conflict.

**B1 — the error-code enum is single-layer, and §11.2's mechanism is undercomputable as written.**
- Layer A §11.2 parses `WriteError(w,r,status,"literal")`. But several modules wrap it: `course`,
  `identity`, `orders`, `adminstats`, `discussion` each have a local `writeError`, and `discussion`
  (`handler.go:755`) passes a **variable** `code := err.Error()` whose values are `errors.New("…")`
  consts in `discussion/domain/errors.go` (~26 codes) — the call-site parser never collects them, and
  the "non-constant code arg → build fails" rule would simply redden the build with no path. §2's
  "measured on main" missed this.
- Layer B coverage-gate is per-**route** (≥1 `ValidateResponse`), not per-(route,status,error-code);
  a rarely-reached code (`review_already_exists` on a race, `thread_full`, `report_duplicate`) is
  touched by no case, so Layer B never validates its enum membership.
- **Fix (plan):** (a) the error-code source parser also reads `errors.New("…")` / `Error = "…"` consts
  in `*/domain/errors.go` **and** recognizes each module's local `writeError` wrapper; the enum is the
  union of those plus direct `WriteError` literals. (b) **Record error-code enum fidelity explicitly as
  single-layer (Layer A parser), not two-layer** — its fail-closed rule is: any `WriteError`/`writeError`
  code argument that cannot be resolved to the parsed union → build fails; and the parser's coverage of
  wrappers + domain-errors is itself asserted (a test that every `domain/errors.go` code appears in the
  emitted enum). **S0.0 spike extended** to exercise the discussion code-source path before S0.1.

**Fold-ins:**
- **N1 (overlay is prose-only; validation keywords have a Go source).** Validation keywords
  (minLength/pattern/maximum) often DO have a Go truth (`domain.ValidatePassword`, regexes), so putting
  them in a hand overlay revives the very drift S0 fights. **Narrow the overlay to PROSE + truly-non-Go
  metadata** (descriptions/summaries, `servers`, `securitySchemes`, `CourseDetail` narrowing, per-field
  array nil-vs-make, additionalProperties whitelist). Validation-keyword fidelity is recorded as a
  **named limitation with mechanism** (derive from the Go validators in a later slice, or accept the
  overlay carries them and they are *not* claimed drift-proof) — drop the "can never drift" claim for
  that subset. Prose has no Go source, so it genuinely cannot drift.
- **N2 (restore per-module staging).** v2 §11.5 collapsed v1's staged S0.1(course)→S0.2(all 85) into
  one oversized S0.1. Restore it: **S0.1 = course end-to-end behind the parallel file (per-module
  fidelity); S0.1b…S0.1n widen module by module**; each lands green within a session (cf. the
  "Codex: impl and tests in separate runs" lesson). Cutover stays last.
- **N3 (non-JSON route exclusion).** `orders/handler.go:659` `w.Write(callback.Acknowledge())` is a
  non-JSON WayForPay callback response; the §4-A "respType per status" and the Layer-B "every route has
  a ValidateResponse" gate would fail on it. Add an explicit **non-JSON route exclusion list** (declared,
  so a new route isn't silently excluded).
- **N4 (determinism additions).** Also sort object `properties` (not just `required`). An encoder /
  `yaml.v3` / `kin-openapi` bump reflows YAML deterministically but produces a large diff — treat such a
  bump as a deliberate regen event (note, not a blocker).

**Open blocking after plan-review: 0** once B1(a)+(b) land in §11.1/§11.2 and the overlay is narrowed
(N1). Re-send the plan delta to Courses 2, then owner-approval, then S0.0.

## 14. S0.0 expressiveness spike — outcome (2026-10-27, owner-approved go)

Spike code (throwaway, `//go:build s0spike`): `backend/internal/course/adapters/http/s0_spike_test.go`
+ `s0_spike_errcode_test.go`. Run `go test -tags s0spike -run TestS0Spike -v ./internal/course/adapters/http/`.
**4/4 tests green.** Resolves the §7↔§8 open question: fidelity is achievable, and the override surface
is measured, not assumed. Evidence per claim below.

**14.1 Reproducible from Go (Layer A, no override) — confirmed green.**
- Reflection (openapi3gen v0.149.0) emits the 13 `courseDTO` properties with correct base types.
- `*T` → `nullable:true` is the library default (`publishedAt`/`price`/`rating`); non-pointer →
  non-nullable. Pins §11.2 pointer row **without custom config**.
- `time.Time` → `format:date-time`, `int64` → `format:int64` — library defaults.
- **`required` is NOT emitted by openapi3gen at all** — the generator owns it. Rule "required unless
  json tag has `,omitempty`" reproduces hand-spec `Course.required` (all 13) **exactly**. Confirms
  §11.2 and that required-derivation is our code.

**14.2 Override surface — measured census (matches §11.2, two refinements).** Each class below is NOT
derivable by reflection and needs its §11.2 mechanism: `format:uuid` (bare string), status `enum`
(plain `string` field, not `type X string` — enum lives in overlay/registry), validation keywords
(minLength/pattern/minimum — in domain validators, not DTO types → overlay per N1), `additionalProperties:false`
(library leaves unset), name-map (`Block`→`LessonBlock`, `Span`→`LessonSpan`), string-enum `BlockType`
(const-set registry). Two **new** findings refine the plan:
- **F1 — the §2 Block premise is imprecise.** The *published* `LessonBlock` schema is **permissive**
  (`ordered` is an optional bool allowed on any block); reflecting the `Block` **struct** reproduces it
  *modulo the `BlockType` enum*. The custom `MarshalJSON` (ordered-as-pointer) does **not** change the
  published component — H2/SA02 lived in a **stricter reader schema**, not this published schema. So the
  Block *override-registry* entry (§11.2 row 1) is **not required for the published schema**; only the
  **string-enum mechanism** for `BlockType` is. The `json.Marshaler` fail-closed assertion stays valuable
  as a guard, but its stated justification ("the one type reflection gets wrong") is wrong for the
  published surface. **Plan edit:** reclassify Block under string-enum, keep the marshaler assertion as a
  defensive guard not a correctness mechanism.
- **F2 — DTOs are unexported; §4-A central registry cannot reflect them cross-package.** `courseDTO`,
  `priceDTO`, … are package-private. A single registry file in a `openapigen` package **cannot**
  reference them. The generator must therefore run **per-module** (each adapter package exports its
  route→type bindings / contributes its reflected schemas) rather than from one central typed registry.
  This supersedes §4 approach-A's "one Go file lists all 85 routes": the route→type binding lives
  **inside each module's adapter package**, and a thin central collector aggregates + the mux cross-check
  stays. **Plan edit:** §4/§11 registry is per-module-contributed, not central-typed.

**14.3 B1 (error-code source) — proven green.** `go/parser` extracts all **21** `errors.New("…")` codes
from `discussion/domain/errors.go` (incl. `review_already_exists`, `report_already_open`,
`course_forbidden`). Confirms B1(a): the call-site literal scan yields **0** of these (discussion passes
`code := err.Error()`, a variable), so the enum **must** union domain/errors.go consts + direct
`WriteError` literals + per-module `writeError` wrappers, with a coverage test asserting every
domain/errors.go code ∈ emitted enum. Error-code fidelity stays **single-layer (Layer A parser)** per §13.

**14.4 N3 (non-JSON route) — proven green.** `orders/adapters/http/handler.go` has exactly **1** raw
`w.Write` (the WayForPay `Acknowledge()` callback) vs 13 `WriteJSON`. The callback route goes on a
**declared** non-JSON exclusion list so the Layer-B coverage-gate does not fail on it.

**S0.0 verdict:** fidelity achievable; override surface = §11.2 with F1/F2 refinements. **Next: Courses 2
review of this outcome (existing_session), then S0.1 (course end-to-end behind parallel file).** The
spike tests are throwaway (build-tagged, no build wired) — delete or evolve into the real generator's
fidelity test in S0.1.

## 15. F2 generator-architecture options (owner-requested +/− + pitfalls, 2026-10-27)

Owner chose: Courses 2 reviews §14 first; direction not fixed until their verdict. This is the decision
input. **Key de-escalation found while writing it:** the constraint is NOT "reflection can't see
unexported types." openapi3gen reflects a `reflect.Type`/`any` **value**; exportedness of the type
*identifier* is a compile-time source concern, not a runtime reflection one, and `courseDTO`'s **fields**
are all exported (`ID`, `Slug`, …). So any package handed a `courseDTO{}` value reflects it fine. The
real constraint is only that a central registry **cannot name `courseDTO` in its own source** — it must
**receive values** built inside the owning package. That difference is what separates the two options.

**Option 1 — per-module contribution (each adapter package hands values to a thin central collector).**
Each `*/adapters/http` package gets a small `openapi_contract.go` that builds its route→type bindings
(it can construct its own unexported DTOs) and registers them with a `platform/openapigen` collector it
is given. A top-level generator (cmd or test) runs every module's contribution, aggregates, applies the
override surface (§11.2), and emits. The mux cross-check runs **per module**.
- **+** DTOs stay unexported — encapsulation and the M8 isolation rule (just merged in #80) are
  untouched; no module can reach another's transport types. `platform/openapigen` is infra (like
  `httpserver`); modules may import it, it imports no module → no cycle, no ADR-013 breach.
- **+** Each module owns its contract next to its handlers; a route and its schema binding live together.
- **−** Touches all ~8 modules (one contract file each) and adds a per-handler/`Register` seam — the
  enlarged-signature cost the plan disliked about approach-B, in a milder form (a separate file, not a
  changed `Register`).
- **−** The "one Go file lists all 85 routes" simplicity of §4-A is gone; the route list is distributed.
- **Pitfalls:** (1) a new route added with no contract entry must **fail** the per-module mux cross-check
  — a global-only cross-check would let a module under-report silently (this is the §11.1 Layer-B
  coverage-gate's job; it must be scoped per module, not just per document). (2) The collector takes
  `any`/`reflect.Type`; a module passing the wrong value (pointer vs value, `*priceDTO` vs `priceDTO`)
  shifts nullability — needs a typed helper so the binding is checked at compile time where possible.
  (3) ordering: aggregation across modules must sort by (path, method) or the diff is unstable (§11.4).

**Option 3 — export the DTOs, keep the central §4-A registry.** Rename `courseDTO`→`CourseDTO` across
modules; one registry file imports every `adapters/http` package and lists the exported types.
- **+** §4-A as originally planned: a single central file, no per-module seam, the simplest collector.
- **−** Exporting transport DTOs widens each module's public API for no domain reason, and **directly
  invites the cross-module coupling M8 (#80) exists to forbid**: another module could now import
  `course/adapters/http.CourseDTO`. We'd have to add a boundaries_test rule banning exactly that — fighting
  our own just-merged isolation gate.
- **−** Large mechanical rename diff across all modules; exported names are now a maintenance contract.
- **Pitfalls:** (1) a central `openapigen` registry importing all 8 `adapters/http` packages is a package
  that depends on every module — fine only if it is a **leaf tool** (cmd/), never imported back by a
  module, or it cycles. (2) Export makes misuse compile cleanly, so the only guard is a test — strictly
  weaker than "unexported, cannot be named."

**Recommendation (pending Courses 2):** Option 1. It preserves the M8 isolation we just shipped, keeps
DTOs encapsulated, and the de-escalation above makes its collector thin (values in, not a type rewrite).
Its real cost is a per-module contract file + a per-module coverage-gate scope — both proportionate.
Option 3's central-file simplicity is bought by weakening module isolation, which is the wrong trade a
week after M8. **Not fixed until Courses 2's §14 verdict.**

## 16. Courses 2 verdict on S0.0 — CHANGES (2026-10-27), absorbed as S0.1 entry conditions

Owner confirmed FULL GENERATOR (2026-10-27, serves "stop hand-editing the spec"; reuse-path does not).
Courses 2 reran the spike (4/4, tests non-vacuous, §14 census exact) and returned **CHANGES**: F2→Option 1
**approved**; Q3 clean; but 1 blocking-as-stated + a systemic gap + a Q2 over-claim must land before S0.1.
These do not retract the spike measurement — they correct its *conclusions*. S0.0 is **not** merged as
"B1/override closed." Resolutions:

**16.1 Q2 (error-code resolve rule) — ACCEPTED (verified).** The spike proved only the `errors.New("…")`
form (discussion, 21 codes). The dominant form is a typed const `ErrX Error = "literal"` — confirmed in
adminstats, clientlog, identity, orders, playback, transcript, media, image. **clientlog** emits
`httpserver.WriteError(w, r, status, string(domain.ErrInvalidEvent), …)` (handler.go:55,68) — a
`string(domain.ErrX)` **cast argument**, neither a bare literal nor `err.Error()`, and clientlog has **no**
`writeError` wrapper (so §13's 5-wrapper list doesn't cover it). The enum **source** stays complete via the
`*/domain/errors.go` glob (it parses both `errors.New("…")` and `Error = "…"` const decls). But the Layer-A
**call-site fail-closed RESOLVE rule** must accept four arg shapes — bare literal, `err.Error()` variable
(discussion), `string(domain.ErrX)` cast (clientlog), and a direct `domain.ErrX` of kind `Error` — else S0.1
reddens legitimate coded calls. **S0.1 entry: the error-code parser handles const-decl + the three
non-literal call forms, with a test that every `*/domain/errors.go` code ∈ emitted enum AND every WriteError
call resolves.**

**16.2 Systemic — Layer B is responses-only; request-body schemas are single-layer — ACCEPTED.** §11.1's
"two-layer fail-closed" holds for responses (`openapi3filter.ValidateResponse`) but **request-body schemas
have no conformance backstop**: Layer A reflection+override is the only check (~10+ request bodies). This is
recorded, like the error-code enum, as **single-layer (Layer A)** honestly — not claimed drift-proof.
**S0.1 options (pick in S0.1):** (b) add `openapi3filter.ValidateRequest` coverage to Layer B for
request-carrying routes (makes it genuinely two-layer), or (c) record request schemas as Layer-A-only with
mechanism. Leaning (b) where a request fixture already exists; (c) otherwise with a named limitation.

**16.3 F1 (Block request-side) — PARTIALLY ACCEPTED; blocking-as-stated DISPUTED with evidence.** Courses 2:
dropping the Block override revives H2 on the request side, because `SaveLessonDocumentRequest` carries
`LessonDocument`(RichText/Block) under a strict reader contract and Layer B validates only responses.
**Verified against the hand spec:** `SaveLessonDocumentRequest` (openapi.yaml:1427/2208) → `LessonDocument`
(2130) → `LessonSection` → `LessonRichText` → `LessonBlock` (2167) — the **same permissive** `LessonBlock`
(`ordered: {type:boolean}`, optional). **There is no strict block schema anywhere in the hand spec, on
either the request or the response side.** So a generated permissive request schema is **faithful to the
current contract, not a regression**: it cannot "revive H2" relative to a strictness the openapi contract
never had. H2/SA02 was a **response-serialization** defect (bullet-list `ordered:false` dropped), fixed in
`Block.MarshalJSON` and covered by Layer-B ValidateResponse; the strict document invariant (ordered
required-on-list, forbidden elsewhere) lives in Go's `validateDocumentSection`, never in openapi.
**Therefore F1 stands for both request and response: the Block *override* is not required for fidelity to
the current spec; reclass Block under the `BlockType` string-enum; keep the `json.Marshaler` assertion as a
defensive guard.** The *valid kernel* of Courses 2's finding is 16.2 (requests are single-layer) — accepted
there. Making the generated schema **stricter than today** (a discriminated `oneOf` by block type) is a
**new capability**, not a fidelity requirement; if wanted, it is a scoped follow-up with its own override.
**Re-verdict requested from Courses 2 on whether F1 remains blocking given the permissive-request evidence.**

**16.4 Q1 (required rule on request DTOs) — ACCEPTED.** Proven only on `courseDTO` (response, all-13,
0 omitempty). **S0.1 fidelity test must cover ≥1 request DTO** (optional-pointer + null semantics), since
request-required divergence is single-layer per 16.2.

**16.5 F2 — Option 1 APPROVED by Courses 2.** Per-module contribution + thin `platform/openapigen`
collector keeps DTOs unexported, respects M8/ADR-013 (platform imports no module; aggregator is a leaf),
with the coverage-gate scoped **per module** (global-only would let a module under-report). This is the
S0.1 architecture.

**Status:** S0.0 measurement sound; conclusions corrected (16.1–16.4). F2 settled (Option 1). One open
discussion: F1 re-verdict (16.3). S0.1 entry conditions = 16.1 parser forms + 16.2 request-layer decision
+ 16.4 request-DTO fidelity coverage.

**16.6 Final verdict (Courses 2, 2026-10-27): S0.0 PASS, 0 blocking.** F1 withdrawn after line-by-line
hand-spec check (`SaveLessonDocumentRequest`:2208 and `LessonDocumentDraft`:2195 both `$ref` the one
`LessonDocument`:2130 → the single permissive `LessonBlock`:2168; no strict/oneOf block schema exists
request or response). Accepted: 16.1, 16.2 (systemic request single-layer survives independently of Block),
16.3 (Block → `BlockType` string-enum + `json.Marshaler` assertion as guard), 16.4, 16.5/F2 Option 1. Q3
clean. **Cleared for S0.1 under the 16.1/16.2/16.4/16.5/16.3 entry conditions + owner go.** Per-slice code
review returns to Courses 2. Future-only note: if S0.1 ever makes the spec stricter than the document's Go
validator (discriminated `oneOf` by block type), that strict schema becomes a new surface and 16.2's
request-conformance must cover it — not now.

**16.7 S0.1 staging clarification (2026-10-27, Codex STOP on brief↔plan conflict).** Codex (Run 1) stopped
pre-code: the brief scoped Run 1 to the course **catalog** (GET /courses, GET /courses/{slug}), but §16's
"S0.1 entry conditions" wording reads as *preconditions to start*. **Clarified:** 16.1 (error-code parser),
16.2 (request-layer), 16.4 (≥1 request DTO in fidelity) are **S0.1 completion-gate conditions, verified by
Courses 2's per-slice review across S0.1's staged runs (N2: course → S0.1b…S0.1n) before the S0.2 cutover —
not line-1 preconditions of the first run.** This changes no *what*: every condition stays mandatory before
S0.1 closes. Run 1 = catalog-only (response surface) is in-scope. Consistent with Courses 2's own framing
("checklist for S0.1") and N2's staged slices.
Also adopted (Codex ARGUE channel, correct): the brief's "no nullable slices in catalog" was imprecise —
§11.2 needs an **explicit** nil/non-nil decision per `[]T` field. Catalog has three
(`pageDTO.Items`→CoursePage.items, `courseDetailDTO.Program`, `programModuleDTO.Lessons`); all are
**non-nullable, required** in the hand spec (no `nullable:true`; handler builds with `make([]T,0)`), so each
is encoded as an **explicit non-null** decision — a `[]T` added later with no decision still fails closed.
