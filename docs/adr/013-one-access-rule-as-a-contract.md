# ADR 013 — One access rule, held as a domain-safe contract; each module reads its own facts
Status: accepted, 2026-09-25. Decided by the user in session Courses 1 (option «B» of three
presented). Amends `boundaries_test.go` (domain may import `internal/contracts/…`) and the
reasoning in `docs/transcript/plan.md` §X2 that a shared function was impossible.

**Context.** «Who may view a course» exists in three copies with the same logic and different
actor and owner types: `course/domain/learning.go:80`, `playback/domain/authorization.go:33`,
`transcript/domain/access.go:39`. Only a behaviour test keeps them honest
(`platform/architecture/shared_rules_test.go`). Access is checked at four points — learning, the
transcript read, the playback session, and nginx `auth_request` on **every** playlist and segment
request (`playback/application/authorize.go:19-35`). Payment will be the last phase
(`work/PRODUCT-BRIEF.md` §24 item 21), and it adds not only a condition («the viewer has a
right») but **data**: whether this user holds a grant for this course.

**Options weighed.**

| Option | What payment then has to add | Cost now, and on the hot path |
|---|---|---|
| A. The rule lives in `course`, others call it through a contract port | one lookup, inside `course` | runtime coupling playback/transcript → course; **two** DB reads per segment (playback already reads the course, the provider reads it again with the grant) |
| **B. The rule is a pure function in `internal/contracts/access`; each module reads its own facts** | one clause in the rule + **one join** in each of three modules' SQL | no runtime coupling; **one** read per segment, with a join |
| C. Keep three copies, strengthen the test | three rules + three SQL changes | nothing now — which is what the user asked to avoid |

**Decision: B.** One rule, one place:

- `backend/internal/contracts/access` holds domain-safe types — a neutral subject
  (`UserID`, `Administrator`) and neutral course facts (`OwnerID`, `Published`) — and one pure
  function `CanView`. Standard library only; it imports no module.
- Each module keeps its own domain types and maps them into the contract. Its `CanBeViewedBy`
  becomes a one-line mapping onto `access.CanView`. The copies of the **logic** disappear, and the
  copies of the **mapping** are what the behaviour test now pins.
- Each module keeps reading its own facts in its own adapter. When payment arrives, the grant is
  one more fact: one field in the contract, one clause in `CanView`, and one join in each
  module's query. The mechanical part — three joins — is held by a gate, not by memory (below).

**What this deliberately does not do.** It adds no port, table or type for orders, money,
enrollment, grants, webhooks or refunds: nothing consumes them yet, and AGENTS.md forbids unused
interfaces. Those are documented seams (`docs/readiness/spec.md`).

**Why the guide's advice differs here.** The reference architecture the user shared keeps
business rules out of `contracts`. We depart from that consciously: this is a single stateless
predicate with no dependencies, and keeping it out of `contracts` costs a runtime call and a second
read on the segment path. Anything larger than a pure predicate does not belong in `contracts`.

**Guards that make B safe:**
1. `boundaries_test.go` allows a domain to import `courses/internal/contracts/…` and nothing else
   beyond the standard library; a new test forbids `contracts` from importing any module.
2. The behaviour test runs one truth table through `access.CanView` and through every module's
   mapping.
3. An integration test per module runs the same truth table through that module's **SQL** — the
   facts its adapter actually reads. When payment adds a join, this is the test that fails if one
   module is missed.

**Revisit when:** the rule stops being a pure predicate — for example, if it needs I/O or a clock —
or when a fourth module starts deciding access.
