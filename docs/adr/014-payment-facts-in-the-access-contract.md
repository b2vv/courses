# ADR 014 — Payment facts belong in the shared access contract
Status: accepted, 2026-09-30. Decided by the user in session Courses 1. Supplements ADR 013,
whose revisit condition was a fourth module deciding access; `discussion` is now that fourth
module, and payment makes the shared decision depend on facts owned outside those four modules.

**Context.** ADR 013 made `access.CanView` the single pure rule used by `course`, `playback`,
`transcript`, and `discussion`, while each module continued to read and map its own facts. Before
payment, publication was enough for a stranger to view a course. Adding only an entitlement clause
would leave every published paid course open, because the existing `Published` clause would still
grant access. Preview is a fact about the selected lesson, while paid status is a fact about the
course and entitlement is a fact about the subject-course pair; all three affect the same access
decision.

**Options weighed.**

| Option | Rule and data flow | Cost |
|---|---|---|
| A. Payment decides access before calling the four modules | payment returns a verdict to each reader | runtime coupling to a fifth module and an extra read on the playback segment path |
| B. Each module extends its own access rule | four copies learn paid, entitlement, and preview independently | four places can drift, including the nginx `auth_request` path |
| **C. Extend the ADR 013 contract with neutral payment facts** | one pure rule; each module maps the same three facts | one join in each module's adapter when O03 starts supplying the facts |

**Decision: C.** `access.Course` gains the neutral boolean facts `Paid`, `Entitled`, and `Preview`.
The shared rule is:

`Administrator ∨ owner ∨ Entitled ∨ (Published ∧ (¬Paid ∨ Preview))`.

`access.PurchaseRequired` uses the same `Subject` and `Course` facts and reports the payment-specific
reason for denial:

`Published ∧ Paid ∧ ¬Preview ∧ ¬(Administrator ∨ owner ∨ Entitled)`.

It remains a pure standard-library predicate. It is not a domain method and does not add a fifth
`CanBeViewedBy`. In particular, `orders` will not decide course visibility. The invariant is
`PurchaseRequired(subject, course) ⟹ ¬CanView(subject, course)`.

**Behaviour before O03.** All three new booleans default to `false`. Until adapters and SQL supply
new facts, the rule reduces to ADR 013's current behaviour: administrators and owners always view;
otherwise a published course is viewable. O01 therefore changes the contract's capacity without
changing production access decisions.

**Guards.**
1. `contracts/access/access_test.go` runs the expanded truth table through both predicates and checks
   the implication above, including paid locked, paid entitled, paid preview, and archived entitled
   cases.
2. `platform/architecture/shared_rules_test.go` maps every row through all four domain fact types.
3. `platform/architecture/access_mapping_ast_test.go` continues to require exactly four
   `CanBeViewedBy` methods whose return is `access.CanView(…)`.
4. O03 must run the expanded table through each module's SQL so the new facts are proven on the
   production read paths rather than only in fixtures.

**Revisit when:** access needs I/O or time, preview ceases to be a lesson-level boolean, or another
module must make the same visibility decision. Any such change must preserve a single named rule and
must re-measure the per-segment playback read path before replacing the pure contract.
