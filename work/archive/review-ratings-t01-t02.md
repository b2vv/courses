# Review — ratings T01 + T02 (commit fb96633, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Against: `docs/ratings/spec.md` (AC1, AC2, AC4, AC5; G1,
G2, G5, G6, G7), `docs/ratings/plan.md` §4/§10/§11, brief `work/brief-codex-t01-t02.md`.

## Verdict

**Approve. 0 blocking, 0 substantive non-blocking.** Code and tests match the brief and the plan
line for line. The three focus points Courses 1 named — admin role never counts, order of refusals,
rounding — all check out. One forward note for T05 below.

Evidence run myself (production path for domain unit tests), not trusted from the commit message:

```
go test ./internal/contracts/rating/... ./internal/discussion/domain/... ./internal/orders/application/...
ok  courses/internal/contracts/rating      1.235s
ok  courses/internal/discussion/domain      0.958s
ok  courses/internal/orders/application     0.664s
```

## Focus point 1 — admin role access never counts as a right (G1, G2) — PASS

- `CanReview` (`review.go:21`) never references `actor.Administrator` and never routes through
  `access.CanView`/`PurchaseRequired` (which do exempt admins/owners — `access.go:27,31`). This is
  exactly the brief's requirement. Confirmed `Actor.Administrator` exists (`access.go:15`), so the
  omission is deliberate, not a missing field.
- Behaviourally proven in `review_test.go:TestCanReview`: `admin paid without grant → review_purchase_required`
  (Rt17), `admin free with progress → allowed`, `admin paid with grant → allowed`, `preview only watcher
  (paid, no grant, progress) → review_purchase_required` (Rt16). Role is never the deciding fact.
- Owner is refused even with grant+progress: `owner with grant and progress → review_owner_forbidden`
  (Rt15). ✔ G2.

## Focus point 2 — order of refusals (plan §4) — PASS

`CanReview` order matches §4 / brief exactly:
1. `UserID == ""` → `ErrForbidden`
2. `UserID == OwnerID` → `review_owner_forbidden` (G2)
3. `Status != Published && !Entitled` → `review_not_available` (draft; archived without grant)
4. `Paid && !Entitled` → `review_purchase_required` (G1, G2)
5. `!HasProgress` → `review_not_started` (R1)

Archived+grant and archived-free-without-grant and draft cases all land on the right code
(`review_test.go`). Five new error codes added in `errors.go` with the spec's exact strings.

## Focus point 3 — rounding — PASS

- `rating.Summary` (`rating.go`): `tenths = (starsSum*20 + count) / (2*count)` is correct integer
  half-up to one decimal. Verified by hand: 14/3 → 4.7, 18/4 → 4.5, 15/3 → 5.0, 87/20 (=4.35) → 4.4
  (the x.x5 case the brief asked for, present in `rating_test.go`). `visible = count >= 3`; below 3
  returns `(0, false)`. Invalid input (`count<0`, `starsSum<count`, `starsSum>5*count`) → `(0,false)`,
  tested. No overflow risk at realistic volumes (the 1e6-count case passes).
- `ReviewProgress.Percent` (`review.go:72`): `(Completed*100 + Total/2) / Total`, 0 when Total==0.
  I checked the odd-`Total` worry: the `Total/2` floor only lowers the rounding threshold below 0.5,
  but for integer completed the fraction `Completed*100/Total` can never be exactly .5 when `Total`
  is odd (even numerator vs odd product), so it never changes a result. Matches the brief table
  (1/8→13, 2/3→67, 2/4→50 for Rt12). Correct.

## T02 — no admin self-grant (AC2, G5, Rt24) — PASS

- `GrantCourse` (`service.go:135`): `userID == actor.UserID → ErrGrantSelfForbidden`, placed **after**
  the administrator check and **before** reason validation / any store call — exactly the brief. A
  non-admin self-grant still returns `grant_forbidden` (correct: self-rule is admin-specific).
- `ErrGrantSelfForbidden` added to `orders/domain/errors.go`; handler maps it to **403**
  (`handler.go:403`); OpenAPI enum gains `grant_self_forbidden`; Orval model regenerated. The brief's
  three test levels are present: unit (`service_test.go` — self-grant case has `wantCalls=0`, i.e.
  store untouched, asserted), HTTP contract (`handler_test.go` — 403 + `"code":"grant_self_forbidden"`
  + `assertContract`), integration (`repository_integration_test.go` — `before==after` grant count,
  own id source `subtestIDs("72")`, scoped to its own rows per the shared-id-slice warning).

### Checked for the hidden cross-subtest break (shared persisted state)

The old integration line granted the admin to **themselves** (`integrationAdminID → integrationCourseID`)
and expected success; that line is now replaced by the self-grant subtest, so that admin grant is no
longer created. I verified nothing downstream depended on it: the next subtest *"paid event settles
when active admin grant already exists"* creates its **own** grant for a freshly-inserted user, and
the grandfather check at the end counts grants on `integrationFreeID` (a different course). Clean
removal, no orphaned dependency. (This is the class of break the repo lessons warn about — reviewing
the diff alone would miss it, so I traced the subtest sequence.)

## Constitution (plan §10) — PASS

- `contracts/rating` imports nothing (pure std-lib-only). The arch test
  `contracts_test.go:TestContractsImportOnlyTheStandardLibrary` walks the whole `internal/contracts`
  tree via `filepath.WalkDir`, so it covers the new `rating` package automatically — no hardcoded
  list to miss it. ✔
- Scope respected: domain + the one orders guard only. No migration, no repository, no review HTTP —
  correct for T01/T02; the `review_*` codes are intentionally **not** in OpenAPI yet (their endpoints
  are T05). Only `grant_self_forbidden` needed OpenAPI now (existing endpoint). ✔

## Non-blocking — forward note for T05 (not a defect in this slice)

- `Summary` returns `average float64`. When T05 serializes `rating.average` to JSON, make sure a whole
  average emits one decimal as the UI contract implies (`★ 5.0`), since Go's `encoding/json` marshals
  `5.0` as `5`. Plan §5 already says "rounded to one decimal on the server" — just flagging that the
  float return type pushes that formatting responsibility onto the T05 serialization/DTO, so it needs
  an explicit test there (e.g. average 5.0 and 4.0). No action needed in T01/T02.

## Rows this slice is credited with (plan §12)

Rt10, Rt11, Rt15, Rt18, Rt24, and the unit parts of Rt06–Rt08 (threshold table) — all have passing
unit tables. Rt16/Rt17 unit coverage also present (their integration/contract levels come in T03/T05).
