# Reference: frontend patterns from Cassiopeia and Cargotrans

Extracted 2026-09-04 from two of this author's other codebases, as input for Phase 4
(admin) and later. Both are read-only references: they run stacks this repository
forbids (Vite, Vitest, ESLint — `AGENTS.md:24`), so nothing is copied. What transfers is
the **shape** of a decision, never the code.

| Project | Path | Architecture | Why it is worth reading |
|---|---|---|---|
| cassiopeia-admin-ui | `~/projects/cassiopeia/cassiopeia-admin-ui` | module-per-aggregate (`src/modules/*` with `views/controllers/lib/forms`), API sliced under `src/core/api/*` | mature admin with ~20 aggregates, per-entity subresources and CQRS writes |
| cargotrans-ui | `~/projects/cargo/cargotrans-ui` | Feature-Sliced Design (`app/entities/features/pages/widgets/shared`) | the same API-client shape (`shared/api/{core,middleware,providers,schema,wire}`) under a different layering |

Both arrived independently at the same API-boundary decomposition, which is the strongest
signal in here: it is not one team's taste.

---

## 1. Two-level result at the API boundary — already matches ours

Cassiopeia splits the boundary in two:

- `HttpWireResult<T>` (`core/api/_shared/http-wire-result.ts`) — transport plus schema
  validation. Failure carries `source: 'http' | 'validation'`.
- `ServiceResult<T>` (`core/api/_shared/service-result.ts`) — the domain-facing result,
  documented as *"mirrors `HttpWireResult` failure branches"*.

That is the seam this repository arrived at on 2026-09-04 in `adapters/api/transport.ts`:
`TransportOutcome` for the wire, each domain's own `Result`/`SessionResult` above it. The
independent convergence is worth recording, because it means the split is structural
rather than incidental.

**Where theirs is richer, and why it might matter here:**

| Their detail | Source | Relevance to courses |
|---|---|---|
| Failure carries the raw response `body`, so a domain service can attempt recovery on a specific status without the transport knowing domain shapes | `service-result.ts` `HttpServiceError.body` | Our `wireError` re-parses `errorSchema` per adapter. Equivalent outcome, more repetition. Revisit if a third adapter appears. |
| Failure carries the request `method`, with a long comment explaining that one guard needs it: the backend's "entity not found" body means "the id in the URL is gone" for a read and is a lie for a write | `service-result.ts` `HttpServiceError.method` | Not needed yet. Will matter in Phase 4, where a command can carry a stale lookup id. |
| `isOk` / `isErr` type guards | `ok-result.ts` | We narrow with `result.ok` inline. Fine at this size. |

Naming difference: they use `{ok, data}`, we use `{ok, value}`. Not worth churning.

## 2. Branded ids — a real gap here

`core/api/_shared/branded.ts` gives every aggregate a compile-time-distinct id:

- `Branded<T, B>` via a `unique symbol`, so `PersonId` and `OrganizationId` cannot be
  swapped even though both are strings at runtime.
- `brandedId('Person')` produces a Zod schema that brands while parsing.
- `newBrandedId` mints a client-side id for the "create = PUT to a minted id" pattern.
- `asBrandedId` is the only escape hatch, documented as trust-boundary-only.

In this repository every id is a bare `string`: `PublishedCourse.id`, `AuthenticatedUser.id`,
`Session.id`. With one aggregate per domain that costs nothing. Phase 4 introduces
`Course`, `Module` and `Lesson` together, all with string ids, all passed through the same
admin call sites — that is exactly the shape branding protects. **Decide before Phase 4,
not during it.**

Related: `guid.ts` documents why they use lenient `z.guid()` rather than `z.uuid()` —
their backend ids are .NET `Guid`s whose version nibble is not RFC 4122. Ours are
PostgreSQL `uuid`, so `z.uuid()` stays correct here; the note only explains why their
code looks different.

## 3. `EntityRef` — keeping slices from importing each other

`core/api/_shared/entity-ref.ts`: `{kind, id}`, deliberately minimal, *"use this instead
of importing the full domain type of another entity — keeps slices loosely coupled"*,
extended per use (`PersonRef = EntityRef<'person'> & {displayName: string}`).

Directly relevant to Phase 4: a `Lesson` needs to reference its `Module` and a `Module`
its `Course`. Without a rule, the admin slice ends up importing three full domain types
into one form.

## 4. Cache-staleness as a named policy, not a per-call number

`core/api/_shared/query-stale-defaults.ts` names three tiers — 15 min for volatile lists,
2 h for reference data, 24 h for session-like data — plus a default `gcTime`, with the
instruction to *"pick the smallest scope that fits the entity"*.

Here, staleness is set per query at the call site: `staleTime: 60_000` globally in
`router.tsx` and `staleTime: 0` on the session query. Two values is not yet a policy
problem, but Phase 4 adds course, module, lesson and publish-state queries at once.

## 5. Explicit invalidation helpers per resource

Cassiopeia keeps named invalidation next to the resource:
`core/api/persons/invalidate-after-person-delete.ts`,
`invalidate-person-detail-and-subresources.ts`.

This repository has no mutation beyond login/logout yet, and those replace the session key
directly. Phase 4 is the first place where one write invalidates several reads; naming
those functions beside the resource is cheaper than discovering the rule in three
components.

## 6. CQRS write helpers — for Phase 4 admin forms

`core/api/_shared/cqrs/` contains `buildCqrsFields` (omit undefined keys from a partial
write), `leaf-diff`, `decide-leaf-row` and `write-leaf-row`. The docblock on
`LeafRowDecision` is the useful part:

> The shared vocabulary for every form-shaped write in the app … Four hand-rolled copies
> of this decision existed before it was named; each one was missing a guard the others
> had.

That is the failure mode to expect in Phase 4: "should this row be written whole, as a
delta, or not at all" gets re-answered per form. Their answer is to name it once.

Whether courses needs a delta protocol at all depends on the admin API we design — a
full-resource PUT avoids the question entirely. Decide it in the Phase 4 spec.

## 7. Never throw inside `validateSearch`

`core/api/_shared/guid.ts` carries the sharpest rule in either codebase,
`optionalGuidSearchField`:

> safeParse-per-field, not `guid()` inside `z.object(...).parse()` of the whole search.
> An exception in `validateSearch` paints the **router's** error boundary — a white screen
> instead of the page, which is worse than any degradation the parameter was read for.
> And it is not only our own link that arrives here: a saved or truncated link, a
> hand-assembled URL, `?id=123` (TanStack hands you a number). So a broken value simply
> disappears, and the screen decides what a missing id means.

`app/router/catalog-search.ts:43` does the opposite: it throws `CatalogQueryError` on an
invalid parameter. That is a deliberate Phase 2 decision and it does not white-screen,
because the route declares `errorComponent: CatalogFailure` — but see the findings below
for what it does produce.

`app/router/login-search.ts` (added in I05) already follows the Cassiopeia rule: an
unusable `returnTo` disappears and the screen falls back.

## 8. Module anatomy — do not adopt

Cassiopeia's `modules/<aggregate>/{views,controllers,lib,forms,modals,i18n}` and
Cargotrans' FSD `entities/features/widgets` are both alternatives to this repository's
`domain/application/ports/adapters/presentation`. All three separate the same concerns.
Switching would be churn with no dependency-direction gain, and the machine-checked
boundaries here (`frontend/scripts/check-architecture.mjs`) are stricter than either.

What is worth borrowing is smaller: Cassiopeia keeps per-aggregate **i18n** next to the
module rather than in one file. Our `app/providers/i18n.tsx` already holds every string
for two locales in a single module and will not scale through Phase 4.
